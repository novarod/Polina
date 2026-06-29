<p align="center">
  <h1 align="center">Polina · Go + Hexagonal Architecture</h1>
</p>

<p align="center">
  Mission-orchestration backend for Unreal Engine 5, a "Figma for Quests". Business rules stay in plain Go and the framework stays at the edges.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white" alt="Go 1.25" />
  <img src="https://img.shields.io/badge/Echo-v4-00ADD8?logo=go&logoColor=white" alt="Echo v4" />
  <img src="https://img.shields.io/badge/PostgreSQL-17-4169E1?logo=postgresql&logoColor=white" alt="PostgreSQL 17" />
  <img src="https://img.shields.io/badge/license-MIT-green.svg" alt="License MIT" />
</p>

<p align="center">
  <strong>English</strong> · <a href="./README.pt-BR.md">Português</a>
</p>

## Description

Polina is a backend for authoring and serving game-quest logic to Unreal Engine 5 without
recompiling the game binary. It is built with Hexagonal Architecture (ports and adapters): the domain
layer doesn't depend on Echo, pgx or HTTP. The business rules are plain Go, and the web framework,
the PostgreSQL adapter and the JWT/bcrypt code are all infrastructure plugged into interfaces (ports).

For now the code implements the **auth**, **user**, **member**, **organization**, **workspace** and
**mission** (core: quest graph + DAG validation) domains. Mission versioning/publish and the UE5
engine endpoint are the next cycle. The infrastructure (CI, Docker, migrations, linting, commit hooks)
is already set up, so adding a new domain doesn't mean redoing the foundation.

## Architecture

The code is split into a framework-agnostic core and an infrastructure shell.

```
apps/api/
├── cmd/server/             # composition root (reads env, starts the server)
├── internal/
│   ├── domain/             # entities & rules, no framework imports
│   │   ├── member/         # Role value object (VIEWER < DESIGNER < ADMIN)
│   │   ├── mission/        # name/desc validation + structural graph (DAG) validation
│   │   ├── organization/   # name/slug validation
│   │   ├── shared/         # pagination
│   │   └── workspace/      # name/description validation
│   ├── application/        # use cases (one struct per use case)
│   │   ├── auth/           # register, login
│   │   ├── authz/          # reusable org-scoped authorization
│   │   ├── mission/        # create, list, get, update, update-graph, delete
│   │   ├── organization/   # create, list, get, update, delete
│   │   ├── token/          # JWT session claims (shared issuer/verifier type)
│   │   └── workspace/      # create, list, get, update, delete (tenant-scoped)
│   ├── ports/              # repository & transaction interfaces (the ports)
│   └── adapters/           # the outside world
│       ├── http/           # Echo handlers, middleware (auth, rate limit)
│       └── postgres/       # pgx repositories, Store + transaction manager
├── pkg/                    # apierr, dag (quest graph validator), hash (SHA-256)
└── db/migrations/          # golang-migrate SQL
```

Some decisions behind the structure:

- The domain layer never imports Echo or pgx. The PR template checks for it, so it isn't left to
  good intentions. That keeps the business logic easy to test on its own and the framework replaceable.
- Each use case is a single struct behind an interface, wired in `internal/server`, which keeps the
  handlers thin.
- Authorization is read fresh from the database. `authz.RequireOrgRole` looks up the caller's
  membership per request instead of trusting the token, so a revoked or downgraded role takes effect
  immediately. The login route has a tighter rate limit than the rest.
- Multi-write use cases run inside a transaction through a `Querier` abstraction (satisfied by both
  the pool and a `pgx.Tx`) and a `WithinTx` manager. For example, creating an organization and its
  first ADMIN member is atomic.
- Organization is the multi-tenant boundary. Every tenant-owned row is soft-deleted via `deleted_at`.

## Stack

- **Runtime:** Go 1.25
- **Framework:** Echo v4 (HTTP)
- **Database:** PostgreSQL 17 via pgx v5 (raw SQL, no ORM), golang-migrate
- **Auth & security:** JWT (HS256), bcrypt, per-IP/route rate limiting, CORS allowlist
- **Validation:** go-playground/validator plus domain validators
- **Tests:** standard `testing` plus testify (unit and integration behind a build tag)
- **Tooling:** golangci-lint, gofmt, lefthook (Conventional Commits and pre-commit checks), Dependabot, GitHub Actions, Docker

## Prerequisites

- Go 1.25
- PostgreSQL 17, or Docker if you'd rather not install it locally
- [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI for migrations
  (`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`)

## Project setup

```bash
cd apps/api
cp .env.example .env   # then fill in the values
go mod download
```

Apply the migrations (with a Postgres running and `DB_URL` set, or via Docker below):

```bash
make migrate
```

## Compile and run the project

```bash
cd apps/api

# development
make run            # go run ./cmd/server

# production build
make build          # outputs ./bin/server
```

The API starts on the port defined by `PORT` (default `8080`).

### With Docker

The compose file brings up the API together with a PostgreSQL container and runs the migrations on
startup (the API runs with hot-reload via `air`):

```bash
docker compose up --build
```

- API at `http://localhost:8080`
- PostgreSQL at `localhost:5432`

## Environment variables

| Variable                       | Description                                                |
| ------------------------------ | ---------------------------------------------------------- |
| `DATABASE_URL`                 | PostgreSQL connection string (**required**)                |
| `JWT_SECRET`                   | JWT signing secret (**required**)                          |
| `JWT_EXPIRY_HOURS`             | Token lifetime in hours (default `24`)                     |
| `BCRYPT_ROUNDS`                | bcrypt cost factor (default `12`)                          |
| `PORT`                         | HTTP port (default `8080`)                                 |
| `FRONTEND_URL`                 | Allowed CORS origin (default `http://localhost:3000`)      |
| `THROTTLE_LIMIT`               | Default requests-per-minute limit (default `30`)           |
| `ENGINE_THROTTLE_LIMIT`        | Rate limit for the UE5 engine routes (reserved, not read yet) |
| `ENGINE_LAST_USED_THROTTLE_MS` | Throttle for engine "last used" updates (reserved, not read yet) |

The two `ENGINE_*` variables are reserved for the upcoming engine (UE5 plugin) module and are not
read by the code yet.

## Run tests

```bash
cd apps/api

# unit tests (race detector)
make test

# integration tests (requires a running PostgreSQL)
make test-integration
```

Unit tests cover the domain validators, use cases (with in-memory fakes) and the HTTP handlers. The
integration suite exercises the repositories and transactional use cases against a real database
(create org plus admin member, cascade delete, slug uniqueness).

## API documentation

In non-production environments, interactive Swagger UI is served at:

```
http://localhost:8080/swagger/index.html
```

The OpenAPI spec lives in `apps/api/docs/` (generated by `make generate` with swaggo/swag and checked
for drift in CI). A health endpoint is always available:

```
GET /health   ->   200 {"status":"ok"}
```

## Code quality

```bash
cd apps/api
gofmt -l .            # formatting (must be empty)
golangci-lint run     # linters
```

Commits follow the [Conventional Commits](https://www.conventionalcommits.org/) spec, enforced by a
[lefthook](https://github.com/evilmartians/lefthook) `commit-msg` hook. The `pre-commit` hook runs
`gofmt`, `go vet` and `golangci-lint` on staged Go files. Install the hooks once with:

```bash
lefthook install
```

CI runs format, vet, lint, build, the full test suite (unit and integration), a Docker build and a
migration up/down check on every push and pull request.

## Deployment

Build the production image with the multi-stage `apps/api/Dockerfile` (it compiles a static binary,
runs as a non-root user and ships a healthcheck). On deploy, apply migrations before starting the
app:

```bash
migrate -path db/migrations -database "$DATABASE_URL" up
```

Provide the environment variables through your orchestrator; they're never baked into the image.

## License

[MIT](./LICENSE).
