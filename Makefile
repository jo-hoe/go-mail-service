include help.mk

# get content of .env as environment variables
include .env
export

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.DEFAULT_GOAL := start

.PHONY: start-k3d
start-k3d: ## start helm chart in K3d
	@k3d cluster create --config ${ROOT_DIR}k3d/mailcluster.yaml
	@helm install gomailservice --set service.port=$(API_PORT) \
	 							--set defaultSenderMailAddress=$(DEFAULT_FROM_ADDRESS) \
								--set defaultSenderName=$(DEFAULT_FROM_NAME) \
								--set sendgrid.apiKey=$(SENDGRID_API_KEY) \
								--set service.enabled=false \
								${ROOT_DIR}helm/
	@kubectl apply -f ${ROOT_DIR}k3d/service.yaml

.PHONY: stop-k3d
stop-k3d: ## stop K3d
	@k3d cluster delete --config ${ROOT_DIR}k3d/mailcluster.yaml

.PHONY: start
start: ## rebuild and start via docker
	@docker compose up --build

.PHONY: generate-helm-docs
generate-helm-docs: ## generates helm docu in /helm folder 
	@docker run --rm --volume "$(ROOT_DIR)/helm:/helm-docs" jnorwood/helm-docs:latest