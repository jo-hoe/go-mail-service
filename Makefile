include help.mk

.DEFAULT_GOAL := start

.PHONY: start-k3d
start-k3d: ## start helm chart in K3d
	@k3d create -publish 80:80

.PHONY: stop-k3d
stop-k3d: ## stop K3d
	@k3d create -publish 80:80

.PHONY: start
start: ## rebuild and start via docker
	docker compose up --build