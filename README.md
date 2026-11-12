# Projeto de Treinamento: Go API + Kubernetes + PostgreSQL (com Kustomize, SOPS/Age)

Este projeto de laboratório foca na integração entre um microserviço desenvolvido em Golang (`microservices/minitest-api/`), um banco de dados PostgreSQL e manifestos Kubernetes estruturados com **Kustomize** (`k8s/`), utilizando isolamento de **Namespaces** (`apps`, `databases`, `portainer`) e gerenciamento seguro de credenciais com **Mozilla SOPS** e chaves **Age**.

O projeto também utiliza o **Traefik** (`IngressRoute`) como proxy reverso/load balancer e o **Portainer** para gerenciamento do cluster.

## 🛠️ Tecnologias Utilizadas

- **Golang (1.22)** e **PostgreSQL (15)**
- **Docker** (Multi-stage build com imagem `distroless` não-root)
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
├── .sops.yaml
├── .gitignore
└── README.md
```

---

## 📋 Pré-requisitos

Certifique-se de ter as seguintes ferramentas instaladas:
1. [Go (1.22+)](https://go.dev/dl/)
2. [Docker](https://docs.docker.com/get-docker/)
3. Cluster K8s local ([Minikube](https://minikube.sigs.k8s.io/docs/start/), [Kind](https://kind.sigs.k8s.io/) ou [k3d](https://k3d.io/)) com Traefik habilitado
4. [kubectl](https://kubernetes.io/docs/tasks/tools/) (v1.21+ já inclui Kustomize)
5. [age](https://github.com/FiloSottile/age) (Para gerar as chaves de criptografia)
6. [sops](https://github.com/getsops/sops) (Para criptografar/descriptografar os manifestos)

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

4. Para descriptografar antes do deploy (informando sua chave privada via `SOPS_AGE_KEY_FILE`):
```bash
export SOPS_AGE_KEY_FILE=./keys.txt
sops --decrypt k8s/base/databases/postgres/secret.enc.yaml
```

---

## 🚀 Como executar o projeto

### 1. Compilar e testar o microserviço Go localmente
```bash
cd microservices/minitest-api
go test ./...
go build ./...
```

### 2. Construir a imagem Docker
Construa a imagem da API Go a partir do novo diretório do microserviço:
```bash
docker build -t minitest-api:dev ./microservices/minitest-api
```

### 3. Visualizar e aplicar os manifestos via Kustomize
Renderize os manifestos para validar o overlay desejado (`dev`, `staging` ou `prod`):
```bash
kubectl kustomize k8s/overlays/dev
```

Aplique os recursos no cluster Kubernetes:
```bash
kubectl apply -k k8s/overlays/dev
```

### 4. Endpoints da API
- `GET /` — Informações básicas do serviço e ambiente
- `GET /healthz` — Liveness probe
- `GET /readyz` — Readiness probe (verifica conectividade com PostgreSQL)
- `GET /api/v1/items` — Lista itens cadastrados
- `POST /api/v1/items` — Cria um novo item (`{"name": "exemplo"}`)
cd microservices/minitest-api
go test ./...
go build ./...
```

### 2. Construir a imagem Docker
Construa a imagem da API Go a partir do novo diretório do microserviço:
```bash
docker build -t minitest-api:dev ./microservices/minitest-api
```

### 3. Visualizar e aplicar os manifestos via Kustomize
Renderize os manifestos para validar o overlay desejado (`dev`, `staging` ou `prod`):
```bash
kubectl kustomize k8s/overlays/dev
```

Aplique os recursos no cluster Kubernetes:
```bash
kubectl apply -k k8s/overlays/dev
```

### 4. Endpoints da API
- `GET /` — Informações básicas do serviço e ambiente
- `GET /healthz` — Liveness probe
- `GET /readyz` — Readiness probe (verifica conectividade com PostgreSQL)
- `GET /api/v1/items` — Lista itens cadastrados
- `POST /api/v1/items` — Cria um novo item (`{"name": "exemplo"}`)