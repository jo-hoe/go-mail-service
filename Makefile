include help.mk

# get content of .env as environment variables
include .env
export

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
IMAGE_NAME := go-mail-service
IMAGE_VERSION := 1.3.1

.DEFAULT_GOAL := start

.PHONY: start-cluster
start-cluster:
	@k3d cluster create --config ${ROOT_DIR}k3d/mailcluster.yaml
	@helm install gomailservice --set service.port=$(API_PORT) \
	 							--set defaultSenderMailAddress=$(DEFAULT_FROM_ADDRESS) \
								--set defaultSenderName=$(DEFAULT_FROM_NAME) \
								--set sendgrid.apiKey=$(SENDGRID_API_KEY) \
								--set service.enabled=false \
								--set image.repository=registry.localhost:5000/${IMAGE_NAME} \
								${ROOT_DIR}charts/go-mail-service
	@kubectl apply -f ${ROOT_DIR}k3d/service.yaml

.PHONY: push-to-registry
push-to-registry: ## build and push docker image to registry
	@docker build -f ${ROOT_DIR}Dockerfile . -t ${IMAGE_NAME}
	@docker tag ${IMAGE_NAME} localhost:5000/${IMAGE_NAME}:${IMAGE_VERSION}
	@docker push localhost:5000/${IMAGE_NAME}:${IMAGE_VERSION}

.PHONY: start-k3d
start-k3d: start-cluster push-to-registry ## starts k3d and deploys local image with loadbalancer

.PHONY: stop-k3d
stop-k3d: ## stop K3d
	@k3d cluster delete --config ${ROOT_DIR}k3d/mailcluster.yaml

.PHONY: restart-k3d
restart-k3d: stop-k3d start-k3d # restart the cluster

.PHONY: start
start: ## rebuild and start via docker
	@docker compose up --build

.PHONY: generate-helm-docs
generate-helm-docs: ## generates helm docu in /helm folder 
	@docker run --rm --volume "$(ROOT_DIR)charts/go-mail-service:/helm-docs" jnorwood/helm-docs:latest