include help.mk

# get content of .env as environment variables
include .env
export

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
LOCAL_REGISTRY := localhost:5000
LOCAL_REGISTRY_HELM := registry.${LOCAL_REGISTRY}
IMAGE_NAME := go-mail-service
IMAGE_VERSION := latest

.DEFAULT_GOAL := start

.PHONY: update
update: ## update dependencies
	@git pull origin main
	@go mod tidy

.PHONY: lint
lint: ## run linters
	golangci-lint run -E dupl -E gocyclo -E gosec -E misspell

.PHONY: test
test: ## run tests
	@go test ./...

.PHONY: start-cluster
start-cluster:
	@k3d cluster create --config ${ROOT_DIR}k3d/mailcluster.yaml

.PHONY: push-to-registry
push-to-registry: ## build and push docker image to registry
	@docker build -f ${ROOT_DIR}Dockerfile . -t ${IMAGE_NAME}:${IMAGE_VERSION}
	@docker tag ${IMAGE_NAME}:${IMAGE_VERSION} ${LOCAL_REGISTRY}/${IMAGE_NAME}:${IMAGE_VERSION}
	@docker push ${LOCAL_REGISTRY}/${IMAGE_NAME}:${IMAGE_VERSION}

.PHONY: start-k3d
start-k3d: start-cluster push-to-registry deploy-chart ## starts k3d and deploys local image with loadbalancer

.PHONY: deploy-chart
deploy-chart: ## deploy helm chart after image is pushed
	@helm upgrade --install gomailservice --set service.port=$(API_PORT) \
	 							--set defaultSenderMailAddress=$(DEFAULT_FROM_ADDRESS) \
								--set defaultSenderName=$(DEFAULT_FROM_NAME) \
								--set sendgridService.enabled=$(IS_SENDGRID_ENABLED) \
								--set sendgridService.apiKey=$(SENDGRID_API_KEY) \
								--set mailjetService.enabled=$(IS_MAILJET_ENABLED) \
								--set mailjetService.apiKeyPublic=$(MAILJET_API_KEY_PUBLIC) \
								--set mailjetService.apiKeyPrivate=$(MAILJET_API_KEY_PRIVATE) \
								--set service.enabled=false \
								--set image.repository=${LOCAL_REGISTRY_HELM}/${IMAGE_NAME} \
								--set image.tag=${IMAGE_VERSION} \
								${ROOT_DIR}charts/go-mail-service
	@kubectl apply -f ${ROOT_DIR}k3d/service.yaml

.PHONY: stop-k3d
stop-k3d: ## stop K3d
	@k3d cluster delete --config ${ROOT_DIR}k3d/mailcluster.yaml

.PHONY: restart-k3d
restart-k3d: stop-k3d start-k3d # restart the cluster

.PHONY: start-docker
start-docker: ## build and start via plain docker with .env file
	@docker build -f ${ROOT_DIR}Dockerfile . -t ${IMAGE_NAME}
	@docker run --rm -p $(API_PORT):$(API_PORT) --env-file .env ${IMAGE_NAME}

.PHONY: generate-helm-docs
generate-helm-docs: ## generates helm docu in /helm folder 
	@docker run --rm --volume "$(ROOT_DIR)charts/go-mail-service:/helm-docs" jnorwood/helm-docs:latest
