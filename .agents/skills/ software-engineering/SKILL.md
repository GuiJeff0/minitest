---
name: software-engineering
description: Aplica padrões de engenharia de software (arquitetura limpa, qualidade de código, testes, segurança, Git/Conventional Commits, observabilidade e infraestrutura declarativa) ao planejar, implementar, refatorar ou revisar código e manifestos no projeto.
---

# Padrão de Engenharia de Software

Siga estas diretrizes em qualquer alteração de código, infraestrutura, documentação ou revisão de Pull Request neste repositório.

---

## 1. Princípios de Design e Arquitetura

- **Separação de Responsabilidades (SoC):** Isole camadas de entrada/transporte (HTTP, CLI, gRPC), regras de negócio (`service`/`domain`), persistência (`repository`/`database`) e configuração (`config`).
- **Injeção de Dependências Explícita:** Receba dependências (banco, clients externos, loggers) via construtores/interfaces em vez de usar variáveis globais ou singletons implícitos.
- **Simplicidade antes de Abstração (KISS & YAGNI):** Evite criar abstrações genéricas ou camadas extras sem necessidade real; prefira código legível, coeso e direto.
- **Falhe Rápido e com Contexto (Fail Fast):** Valide pré-condições e entradas na borda do sistema. Nunca engula erros silenciosamente — propague ou trate com contexto suficiente para diagnóstico.
- **Mudanças Incrementais:** Mantenha alterações focadas no escopo da tarefa/issue. Preserve comentários e documentações existentes que não tenham relação com a mudança.

---

## 2. Padrões de Código e Microserviços (Go / Backend)

- **Estrutura de Pacotes:** Siga o layout padrão (`cmd/<app>/main.go` como entrypoint fino + `internal/` para encapsular lógica privada do módulo).
- **Propagação de Contexto e Timeouts:** Aceite `context.Context` como primeiro parâmetro em operações de I/O (HTTP, SQL, filas) e defina timeouts explícitos em servidores e clientes externos.
- **Ciclo de Vida e Graceful Shutdown:** Trate sinais do sistema (`SIGINT`, `SIGTERM`), interrompa novas requisições e drene conexões ativas antes de encerrar o processo.
- **Contratos de Saúde para Orquestradores:** Exponha endpoints dedicados:
  - `/healthz` (*Liveness*): indica se o processo está de pé.
  - `/readyz` (*Readiness*): valida se dependências críticas (ex: PostgreSQL) estão acessíveis.
- **Logs Estruturados:** Registre eventos relevantes com contexto útil (método, rota, latência, erro), nunca registrando senhas, tokens ou dados sensíveis.

---

## 3. Infraestrutura, Containers e Kubernetes

- **Infraestrutura Declarativa:** Versione todos os recursos em `k8s/` seguindo Kustomize (`base/` + `overlays/dev|staging|prod/`) com convenção de um arquivo por `Kind` (`deployment.yaml`, `service.yaml`, `configmap.yaml`, etc.).
- **Isolamento por Namespaces:** Separe aplicações (`apps`), bancos de dados (`databases`) e ferramentas operacionais (`portainer`/infra). Use FQDNs internos (`<service>.<namespace>.svc.cluster.local`) para comunicação entre namespaces.
- **Segurança em Containers:**
  - Use *multi-stage builds* e imagens base enxutas (`alpine`/`distroless`).
  - Execute containers com usuário não-root (`USER appuser` / `runAsNonRoot: true`).
  - Defina `resources.requests` e `resources.limits` (CPU e memória) e configure `livenessProbe` / `readinessProbe`.
  - Evite tag `:latest` em ambientes `staging` e `prod`.

---

## 4. Segurança e Gestão de Segredos

- **Zero Segredos em Texto Claro no Git:** Nunca versione senhas, chaves privadas, tokens ou arquivos `.env` reais no repositório.
- **Criptografia Versionada:** Utilize **SOPS + Age** (`*.enc.yaml`) para secrets declarativos no Kubernetes e mantenha chaves privadas (`*.agekey`, `keys.txt`) e arquivos `*.dec.yaml` explicitamente bloqueados no `.gitignore`.
- **Menor Privilégio (Least Privilege):** Conceda apenas as permissões mínimas necessárias em RBAC, acessos de banco e tokens de CI/CD.
- **Validação de Entrada:** Sanitize e valide payloads JSON, parâmetros de rota e variáveis de ambiente antes de utilizá-los.

---

## 5. Testes e Verificação

- **Pirâmide de Testes:**
  - **Unitários:** Cubra regras de negócio e validações usando *table-driven tests* e mocks/fakes nas interfaces de repositório.
  - **Integração:** Valide queries SQL, migrations e contratos HTTP nos limites do serviço.
- **Verificação Obrigatória antes de Concluir:**
  - Execute build (`go build ./...`), testes (`go test ./...`) e validação dos manifestos (`kubectl kustomize k8s/overlays/dev`) sempre que aplicável.
  - Nunca afirme que algo passou sem executar o comando de verificação. Se algum ambiente/ferramenta não estiver disponível, reporte explicitamente o que não foi possível testar.

---

## 6. Fluxo Git, Commits e Pull Requests

- **Branches:** Nunca trabalhe ou faça push direto na `main`. Crie branches descritivas a partir da branch principal (ex: `feat/estrutura-pastas`, `fix/db-timeout`).
- **Conventional Commits:** Padronize mensagens de commit:
  - `feat:` nova funcionalidade
  - `fix:` correção de bug
  - `refactor:` refatoração sem mudança de comportamento
  - `docs:` documentação
  - `chore:` manutenção, build, dependências ou configs
  - `test:` adição ou ajuste de testes
- **Pull Requests:** Mantenha PRs pequenos, vinculados a uma issue, com resumo objetivo da mudança, impacto e evidências de verificação.

---

## 7. Checklist Final (Definition of Done)

Antes de considerar qualquer tarefa concluída, verifique:

- [ ] A mudança respeita a estrutura de pastas e separação de camadas do projeto.
- [ ] Não há segredos, credenciais ou chaves privadas expostas no diff.
- [ ] Tratamento de erros, timeouts e logs não expõem dados sensíveis.
- [ ] Build, linter e testes relevantes passaram após a última edição.
- [ ] Documentação (`README.md` / migrations / manifestos) foi atualizada se o contrato ou setup mudou.
