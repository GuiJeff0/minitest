# Projeto de Treinamento: Go API + Kubernetes + PostgreSQL (com SOPS/Age)

Este projeto de laboratório foca na integração entre uma API desenvolvida em Golang, um banco de dados PostgreSQL e ferramentas de infraestrutura executadas em um cluster Kubernetes.

Nesta versão, adicionamos a separação de variáveis de ambiente via **ConfigMap** e o gerenciamento seguro de credenciais utilizando **Kubernetes Secrets** criptografados com **Mozilla SOPS** e chaves **Age**.

O projeto também utiliza o **Traefik** como proxy reverso/load balancer e o **Portainer** para gerenciamento do cluster.

## 🛠️ Tecnologias Utilizadas

- **Golang (1.21)** e **PostgreSQL (15)**
- **Docker** (Multi-stage build)
- **Kubernetes** (Deployments, Services, ConfigMaps, Secrets, Ingress)
- **SOPS & Age**: Criptografia de Secrets direto no repositório.
- **Traefik**: Ingress Controller.
- **Portainer**: Interface gráfica (GUI).

---

## 📋 Pré-requisitos

Certifique-se de ter as seguintes ferramentas instaladas:
1. [Docker](https://docs.docker.com/get-docker/)
2. Cluster K8s local ([Minikube](https://minikube.sigs.k8s.io/docs/start/), [Kind](https://kind.sigs.k8s.io/) ou [k3d](https://k3d.io/))
3. [kubectl](https://kubernetes.io/docs/tasks/tools/)
4. [age](https://github.com/FiloSottile/age) (Para gerar as chaves de criptografia)
5. [sops](https://github.com/getsops/sops) (Para criptografar/descriptografar os manifestos)

> **Nota para Minikube:** Habilite o addon do Ingress antes de iniciar:
> ```bash
> minikube addons enable ingress
> ```

---

## 🚀 Como executar o projeto

### 1. Construir a imagem Docker
Construa a imagem da API Go localmente:
```bash
docker build -t api-go-k8s:latest .