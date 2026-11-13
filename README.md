# Projeto de Treinamento: Go API + Kubernetes + PostgreSQL (kind, Kustomize, SOPS/Age)

Este projeto de laboratório foca na integração entre um microserviço desenvolvido em Golang (`microservices/minitest-api/`), um banco de dados PostgreSQL e manifestos Kubernetes estruturados com **Kustomize** (`k8s/`), executados localmente em um cluster **kind** (*Kubernetes in Docker*) com isolamento de **Namespaces** (`apps`, `databases`, `portainer`) e gerenciamento seguro de credenciais com **Mozilla SOPS** e chaves **Age**.

O projeto também utiliza o **Traefik** (`IngressRoute`) como proxy reverso/load balancer e o **Portainer** para gerenciamento do cluster.

## 🛠️ Tecnologias Utilizadas

- **Golang (1.22)** e **PostgreSQL (15)**
- **Docker** (Multi-stage build com imagem `distroless` não-root)
- **kind** (*Kubernetes in Docker*) + **Helm** (Traefik Ingress Controller)
- **Kubernetes + Kustomize** (`base/` + `overlays/` por ambiente: `dev`, `staging`, `prod`)
- **SOPS & Age**: Criptografia de Secrets direto no repositório (`*.enc.yaml`).
- **Traefik**: Ingress Controller (`traefik.io/v1alpha1` `IngressRoute`).
- **Portainer**: Interface gráfica (GUI) para administração do cluster.

---

## 🌳 Estrutura do Repositório

```text
minitest/
├── microservices/
│   └── minitest-api/
│       ├── cmd/api/main.go
│       ├── internal/{config,database,domain,repository,service,handler}/
│       ├── migrations/001_init.sql
│       ├── Dockerfile
│       ├── go.mod
│       └── go.sum
├── k8s/
│   ├── cluster/
│   │   └── kind/
│   │       ├── kind-config.yaml
│   │       └── traefik-values.yaml
│   ├── base/
│   │   ├── namespaces.yaml
│   │   ├── config/
│   │   ├── databases/postgres/
│   │   ├── portainer/
│   │   └── apps/microservices/minitest-api/
│   └── overlays/
│       ├── dev/
│       ├── staging/
│       └── prod/
├── Makefile
├── .sops.yaml
├── .gitignore
└── README.md
```

---

## 📋 Pré-requisitos

Certifique-se de ter as seguintes ferramentas instaladas:
1. [Go (1.22+)](https://go.dev/dl/)
2. [Docker](https://docs.docker.com/get-docker/)
3. [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
4. [kubectl](https://kubernetes.io/docs/tasks/tools/) (v1.21+ já inclui Kustomize)
5. [Helm](https://helm.sh/docs/intro/install/)
6. [age](https://github.com/FiloSottile/age) (Para gerar as chaves de criptografia)
7. [sops](https://github.com/getsops/sops) (Para criptografar/descriptografar os manifestos)

---

## 🔐 Configuração de Secrets (SOPS + Age)

1. Gere um par de chaves Age (a chave privada `keys.txt` é ignorada pelo `.gitignore`):
```bash
age-keygen -o keys.txt
```

2. Copie a chave pública (`age1...`) gerada no output e atualize o arquivo `.sops.yaml`.

3. Crie os valores reais em arquivos temporários `*.dec.yaml` (também ignorados pelo `.gitignore`) e criptografe cada Secret:
```bash
sops --encrypt k8s/base/config/secret.dec.yaml > k8s/base/config/secret.enc.yaml
sops --encrypt k8s/base/databases/postgres/secret.dec.yaml > k8s/base/databases/postgres/secret.enc.yaml
sops --encrypt k8s/base/apps/microservices/minitest-api/secret.dec.yaml > k8s/base/apps/microservices/minitest-api/secret.enc.yaml
```

4. Para descriptografar ao aplicar no cluster (informando sua chave privada via `SOPS_AGE_KEY_FILE`):
```bash
export SOPS_AGE_KEY_FILE=./keys.txt
sops -d k8s/base/config/secret.enc.yaml | kubectl apply -f -
sops -d k8s/base/databases/postgres/secret.enc.yaml | kubectl apply -f -
sops -d k8s/base/apps/microservices/minitest-api/secret.enc.yaml | kubectl apply -f -
```

---

## 🚀 Como executar o projeto no kind

> Caso tenha o `k3s` ativo usando as portas `80`/`443` no host, pare-o (`sudo systemctl stop k3s`) ou altere `hostPort` em `k8s/cluster/kind/kind-config.yaml` para `8080`/`8443` antes de criar o cluster.

### 1. Subir o cluster kind e instalar o Traefik via Helm
```bash
make cluster-up
make traefik-install
```
Ou diretamente:
```bash
kind create cluster --config k8s/cluster/kind/kind-config.yaml
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install traefik traefik/traefik \
  --namespace traefik --create-namespace \
  -f k8s/cluster/kind/traefik-values.yaml
```

### 2. Compilar, testar e carregar a imagem no kind
```bash
cd microservices/minitest-api
go test ./...
go build ./...
cd ../..

make build
make load
```

### 3. Aplicar manifestos Kustomize e Secrets
```bash
make deploy-dev
make secrets-apply
```
Ou diretamente:
```bash
kubectl apply -k k8s/overlays/dev
export SOPS_AGE_KEY_FILE=./keys.txt
sops -d k8s/base/config/secret.enc.yaml | kubectl apply -f -
sops -d k8s/base/databases/postgres/secret.enc.yaml | kubectl apply -f -
sops -d k8s/base/apps/microservices/minitest-api/secret.enc.yaml | kubectl apply -f -
```

### 4. Validar endpoints e acessar o Portainer
Testar a API via Traefik (overlay `dev` utiliza o host `dev.minitest.local`):
```bash
curl -H "Host: dev.minitest.local" http://localhost/healthz
curl -H "Host: dev.minitest.local" http://localhost/readyz
```

Acessar o Portainer via `port-forward`:
```bash
kubectl port-forward -n portainer svc/portainer-service 9000:9000
```

### 5. Endpoints da API
- `GET /` — Informações básicas do serviço e ambiente
- `GET /healthz` — Liveness probe
- `GET /readyz` — Readiness probe (verifica conectividade com PostgreSQL)
- `GET /api/v1/items` — Lista itens cadastrados
- `POST /api/v1/items` — Cria um novo item (`{"name": "exemplo"}`)