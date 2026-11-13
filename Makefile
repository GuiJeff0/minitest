CLUSTER_NAME ?= minitest
KIND_CONFIG ?= k8s/cluster/kind/kind-config.yaml
TRAEFIK_VALUES ?= k8s/cluster/kind/traefik-values.yaml
IMAGE_NAME ?= minitest-api
IMAGE_TAG ?= dev
SOPS_AGE_KEY_FILE ?= ./keys.txt

.PHONY: help cluster-up traefik-install build load deploy-dev secrets-apply cluster-down

help:
	@echo "Alvos disponíveis:"
	@echo "  make cluster-up       - Cria o cluster kind ($(CLUSTER_NAME))"
	@echo "  make traefik-install  - Instala o Traefik via Helm no cluster kind"
	@echo "  make build            - Compila a imagem Docker $(IMAGE_NAME):$(IMAGE_TAG)"
	@echo "  make load             - Carrega a imagem $(IMAGE_NAME):$(IMAGE_TAG) no kind"
	@echo "  make deploy-dev       - Aplica os manifestos Kustomize do overlay dev"
	@echo "  make secrets-apply    - Descriptografa e aplica os Secrets via SOPS"
	@echo "  make cluster-down     - Remove o cluster kind ($(CLUSTER_NAME))"

cluster-up:
	kind create cluster --config $(KIND_CONFIG)
	kubectl cluster-info --context kind-$(CLUSTER_NAME)

traefik-install:
	helm repo add traefik https://traefik.github.io/charts || true
	helm repo update
	helm upgrade --install traefik traefik/traefik \
		--namespace traefik --create-namespace \
		-f $(TRAEFIK_VALUES)

build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) ./microservices/minitest-api

load:
	kind load docker-image $(IMAGE_NAME):$(IMAGE_TAG) --name $(CLUSTER_NAME)

deploy-dev:
	kubectl apply -k k8s/overlays/dev

secrets-apply:
	SOPS_AGE_KEY_FILE=$(SOPS_AGE_KEY_FILE) sops -d k8s/base/config/secret.enc.yaml | kubectl apply -f -
	SOPS_AGE_KEY_FILE=$(SOPS_AGE_KEY_FILE) sops -d k8s/base/databases/postgres/secret.enc.yaml | kubectl apply -f -
	SOPS_AGE_KEY_FILE=$(SOPS_AGE_KEY_FILE) sops -d k8s/base/apps/microservices/minitest-api/secret.enc.yaml | kubectl apply -f -

cluster-down:
	kind delete cluster --name $(CLUSTER_NAME)
