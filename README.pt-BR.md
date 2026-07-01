<p align="center">
  <h1 align="center">Polina · Go + Arquitetura Hexagonal</h1>
</p>

<p align="center">
  Backend de orquestração de missões para Unreal Engine 5, um "Figma para Quests". As regras de negócio ficam em Go puro e o framework fica nas bordas.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white" alt="Go 1.25" />
  <img src="https://img.shields.io/badge/Echo-v4-00ADD8?logo=go&logoColor=white" alt="Echo v4" />
  <img src="https://img.shields.io/badge/PostgreSQL-17-4169E1?logo=postgresql&logoColor=white" alt="PostgreSQL 17" />
  <img src="https://img.shields.io/badge/license-MIT-green.svg" alt="License MIT" />
</p>

<p align="center">
  <a href="./README.md">English</a> · <strong>Português</strong>
</p>

## Descrição

Polina é um backend para criar e servir a lógica de quests de jogos à Unreal Engine 5 sem recompilar
o binário do jogo. É construído com Arquitetura Hexagonal (ports e adapters): a camada de domínio não
depende de Echo, pgx ou HTTP. As regras de negócio são Go puro, e o framework web, o adapter de
PostgreSQL e o código de JWT/bcrypt são infraestrutura plugada em interfaces (ports).

Por enquanto o código implementa os domínios **auth**, **user**, **member**, **organization**,
**workspace**, **mission** (core: grafo de quest + validação DAG, mais versionamento/publish: um grafo
validado é compilado num contrato de runtime imutável e hasheado) e **engine** (API keys por
organização + os endpoints `x-api-key` que o plugin UE5 consulta para o hash ativo e o contrato) —
fechando o loop over-the-air. A infraestrutura (CI, Docker, migrations, lint, hooks de commit) já está
pronta, então adicionar um novo domínio não significa refazer a fundação.

## Arquitetura

O código é dividido entre um núcleo agnóstico de framework e uma casca de infraestrutura.

```
apps/api/
├── cmd/server/             # composition root (lê env, sobe o servidor)
├── internal/
│   ├── domain/             # entidades & regras, sem imports de framework
│   │   ├── member/         # value object Role (VIEWER < DESIGNER < ADMIN)
│   │   ├── mission/        # validação de name/desc, validação do grafo (DAG), compile do contrato
│   │   ├── organization/   # validação de name/slug
│   │   ├── shared/         # paginação
│   │   └── workspace/      # validação de name/description
│   ├── application/        # use cases (um struct por use case)
│   │   ├── apikey/         # API keys da organização: create (raw uma vez), list, revoke
│   │   ├── auth/           # register, login
│   │   ├── authz/          # autorização escopada por org, reutilizável
│   │   ├── engine/         # leitura do hash ativo + contrato para o plugin UE5
│   │   ├── mission/        # create, list, get, update, update-graph, delete, publish, versions
│   │   ├── organization/   # create, list, get, update, delete
│   │   ├── token/          # claims do JWT (tipo compartilhado emissor/verificador)
│   │   └── workspace/      # create, list, get, update, delete (escopado por tenant)
│   ├── ports/              # interfaces de repositório & transação (os ports)
│   └── adapters/           # o mundo externo
│       ├── http/           # handlers Echo, middleware (auth JWT, auth x-api-key, rate limit)
│       └── postgres/       # repositórios pgx, Store + gerenciador de transação
├── pkg/                    # apierr, dag (validador de grafo de quest), hash (SHA-256)
└── db/migrations/          # SQL do golang-migrate
```

Algumas decisões por trás da estrutura:

- A camada de domínio nunca importa Echo ou pgx. O template de PR verifica isso, então não fica na
  base da boa intenção. Isso mantém a regra de negócio fácil de testar isolada e o framework
  substituível.
- Cada use case é um único struct atrás de uma interface, montado em `internal/server`, o que mantém
  os handlers finos.
- A autorização é lida fresca do banco. `authz.RequireOrgRole` consulta a membership do solicitante a
  cada requisição em vez de confiar no token, então um papel revogado ou rebaixado vale na hora. A
  rota de login tem rate limit mais apertado que o resto.
- Use cases com múltiplas escritas rodam dentro de uma transação via uma abstração `Querier`
  (satisfeita tanto pelo pool quanto por um `pgx.Tx`) e um gerenciador `WithinTx`. Por exemplo, criar
  uma organização e seu primeiro membro ADMIN é atômico.
- Organization é a fronteira multi-tenant. Toda linha pertencente a um tenant usa soft delete via
  `deleted_at`.

## Stack

- **Runtime:** Go 1.25
- **Framework:** Echo v4 (HTTP)
- **Banco:** PostgreSQL 17 via pgx v5 (SQL cru, sem ORM), golang-migrate
- **Auth & segurança:** JWT (HS256), bcrypt, rate limiting por IP/rota, allowlist de CORS
- **Validação:** go-playground/validator mais validadores de domínio
- **Testes:** `testing` padrão mais testify (unit e integração atrás de build tag)
- **Tooling:** golangci-lint, gofmt, lefthook (Conventional Commits e checks de pre-commit), Dependabot, GitHub Actions, Docker

## Pré-requisitos

- Go 1.25
- PostgreSQL 17, ou Docker se preferir não instalar localmente
- CLI do [`golang-migrate`](https://github.com/golang-migrate/migrate) para as migrations
  (`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`)

## Setup do projeto

```bash
cd apps/api
cp .env.example .env   # depois preencha os valores
go mod download
```

Aplique as migrations (com um Postgres rodando e `DB_URL` setado, ou via Docker abaixo):

```bash
make migrate
```

## Compilar e rodar o projeto

```bash
cd apps/api

# desenvolvimento
make run            # go run ./cmd/server

# build de produção
make build          # gera ./bin/server
```

A API sobe na porta definida por `PORT` (default `8080`).

### Com Docker

O compose sobe a API junto de um container PostgreSQL e roda as migrations no startup (a API roda com
hot-reload via `air`):

```bash
docker compose up --build
```

- API em `http://localhost:8080`
- PostgreSQL em `localhost:5432`

## Variáveis de ambiente

| Variável                       | Descrição                                                  |
| ------------------------------ | ---------------------------------------------------------- |
| `DATABASE_URL`                 | String de conexão do PostgreSQL (**obrigatória**)          |
| `JWT_SECRET`                   | Segredo de assinatura do JWT (**obrigatório**)             |
| `JWT_EXPIRY_HOURS`             | Validade do token em horas (default `24`)                  |
| `BCRYPT_ROUNDS`                | Fator de custo do bcrypt (default `12`)                    |
| `PORT`                         | Porta HTTP (default `8080`)                                |
| `FRONTEND_URL`                 | Origem permitida no CORS (default `http://localhost:3000`) |
| `THROTTLE_LIMIT`               | Limite padrão de requisições por minuto (default `30`)     |
| `ENGINE_THROTTLE_LIMIT`        | Rate limit por API key das rotas de engine UE5 (default `600`) |
| `ENGINE_LAST_USED_THROTTLE_MS` | Intervalo mínimo entre escritas de `last_used_at` por key (default `60000`) |

## Rodar os testes

```bash
cd apps/api

# testes unitários (com race detector)
make test

# testes de integração (exigem um PostgreSQL rodando)
make test-integration
```

Os testes unitários cobrem os validadores de domínio, os use cases (com fakes em memória) e os
handlers HTTP. A suíte de integração exercita os repositórios e os use cases transacionais contra um
banco real (criar org mais membro admin, delete em cascata, unicidade de slug).

## Documentação da API

Em ambientes que não são de produção, a UI interativa do Swagger é servida em:

```
http://localhost:8080/swagger/index.html
```

O spec OpenAPI fica em `apps/api/docs/` (gerado por `make generate` com swaggo/swag e checado contra
drift na CI). Um endpoint de health está sempre disponível:

```
GET /health   ->   200 {"status":"ok"}
```

## Qualidade de código

```bash
cd apps/api
gofmt -l .            # formatação (deve ser vazio)
golangci-lint run     # linters
```

Os commits seguem o padrão [Conventional Commits](https://www.conventionalcommits.org/), imposto por
um hook `commit-msg` do [lefthook](https://github.com/evilmartians/lefthook). O hook `pre-commit` roda
`gofmt`, `go vet` e `golangci-lint` nos arquivos Go staged. Instale os hooks uma vez com:

```bash
lefthook install
```

A CI roda formatação, vet, lint, build, a suíte de testes completa (unit e integração), um build de
Docker e uma checagem de migration up/down a cada push e pull request.

## Deploy

Faça o build da imagem de produção com o `apps/api/Dockerfile` multi-stage (compila um binário
estático, roda como usuário não-root e traz um healthcheck). No deploy, aplique as migrations antes de
iniciar a aplicação:

```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```

Forneça as variáveis de ambiente pelo seu orquestrador; elas nunca são embutidas na imagem.

## Licença

[MIT](./LICENSE).
