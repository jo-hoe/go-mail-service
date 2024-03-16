include help.mk

.DEFAULT_GOAL := start

.PHONY: start
start: ## rebuild and start via docker
	docker compose up --build