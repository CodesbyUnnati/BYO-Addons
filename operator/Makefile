SHELL := /bin/bash

.PHONY: build
build: ## Build the operator binary.
	go build ./...

.PHONY: test
test: ## Run unit tests.
	go test ./...

.PHONY: generate
generate: ## Generate deepcopy code.
	controller-gen object paths=./api/...

.PHONY: manifests
manifests: ## Regenerate CRD definitions.
	controller-gen crd paths=./api/... output:crd:artifacts:config=config/crd/bases
