include help.mk

# get content of .env as environment variables
include .env
export

# get root dir
ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

.DEFAULT_GOAL := start

.PHONY: start-k3d
start-k3d: ## start helm chart in K3d
	@k3d cluster create mailcluster --port '8080:80@loadbalancer'
	@helm install gomailservice --set service.port=$(API_PORT) \
	 							--set defaultSenderMailAddress=$(DEFAULT_FROM_ADDRESS) \
								--set defaultSenderName=$(DEFAULT_FROM_NAME) \
								--set sendgrid.apiKey=$(SENDGRID_API_KEY) \
								${ROOT_DIR}helm\

.PHONY: stop-k3d
stop-k3d: ## stop K3d
	@k3d cluster delete mailcluster

.PHONY: start
start: ## rebuild and start via docker
	@docker compose up --build