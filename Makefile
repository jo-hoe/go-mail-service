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
	golangci-lint run --config ${ROOT_DIR}.golangci.yml

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
	@helm upgrade --install gomailservice \
								--set logLevel=$(LOG_LEVEL) \
								--set sender.address=$(DEFAULT_FROM_ADDRESS) \
								--set sender.name=$(DEFAULT_FROM_NAME) \
								--set http.port=$(HTTP_PORT) \
								--set smtp.port=$(SMTP_PORT) \
								--set smtp.domain=$(SMTP_DOMAIN) \
								--set smtp.auth.required=$(SMTP_AUTH_REQUIRED) \
								--set smtp.auth.username=$(SMTP_AUTH_USERNAME) \
								--set smtp.auth.password=$(SMTP_AUTH_PASSWORD) \
								--set provider.noop.enabled=$(IS_NOOP_ENABLED) \
								--set provider.sendgrid.enabled=$(IS_SENDGRID_ENABLED) \
								--set provider.sendgrid.apiKey=$(SENDGRID_API_KEY) \
								--set provider.mailjet.enabled=$(IS_MAILJET_ENABLED) \
								--set provider.mailjet.apiKeyPublic=$(MAILJET_API_KEY_PUBLIC) \
								--set provider.mailjet.apiKeyPrivate=$(MAILJET_API_KEY_PRIVATE) \
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
start-docker: ## build and start via plain docker, mounting local/config.yaml
	@docker build -f ${ROOT_DIR}Dockerfile . -t ${IMAGE_NAME}
	@docker run --rm -p $(HTTP_PORT):$(HTTP_PORT) -p $(SMTP_PORT):$(SMTP_PORT) \
		-v ${ROOT_DIR}local/config.yaml:/config/config.yaml:ro \
		${IMAGE_NAME}

.PHONY: generate-helm-docs
generate-helm-docs: ## generates helm docu in /helm folder 
	@docker run --rm --volume "$(ROOT_DIR)charts/go-mail-service:/helm-docs" jnorwood/helm-docs:latest
