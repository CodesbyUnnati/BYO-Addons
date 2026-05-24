SHELL := /bin/bash

IMG ?= byo-addons-operator:0.1.0
KUSTOMIZE ?= kustomize
HELM ?= helm
KUBECTL ?= kubectl
export GOCACHE ?= $(CURDIR)/.cache/go-build
export GOMODCACHE ?= $(CURDIR)/.cache/gomod

.PHONY: help
help: ## Show available targets.
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-24s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: fmt
fmt: ## Format Go sources.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

.PHONY: test
test: fmt vet ## Run unit tests.
	go test ./...

.PHONY: manifests
manifests: ## Regenerate CRDs when controller-gen is installed.
	@if command -v controller-gen >/dev/null 2>&1; then \
		controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases; \
	else \
		echo "controller-gen not installed; using checked-in manifests"; \
	fi

.PHONY: build
build: ## Build the manager binary.
	go build -o bin/manager ./cmd

.PHONY: docker-build
docker-build: ## Build operator image.
	docker build -t $(IMG) .

.PHONY: install
install: manifests ## Install CRDs into the current cluster.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: deploy
deploy: manifests ## Deploy operator into the current cluster.
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: ## Remove operator resources from current cluster.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete -f -

.PHONY: helm-lint
helm-lint: ## Validate the operator Helm chart.
	$(HELM) lint charts/byo-addons-operator

.PHONY: render
render: ## Render Kustomize and Helm locally.
	$(KUSTOMIZE) build config/default >/tmp/byo-addons-kustomize.yaml
	$(HELM) template byo-addons charts/byo-addons-operator >/tmp/byo-addons-helm.yaml
