SHELL := /bin/bash

help: ## This help message
	@echo "Usage: make [target]"
	@echo "Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY:run
run: ## Run
	@exec go run *.go ./data

.PHONY:build
build: ## Build application
	@exec go build -o SimpleSearch *.go

