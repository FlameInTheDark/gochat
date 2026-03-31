# Contributing to gochat

Thank you for your interest in contributing. This document covers how to set up the project, the development workflow, and the standards every contribution must meet.

---

## Table of Contents

1. [Project overview](#1-project-overview)
2. [Prerequisites](#2-prerequisites)
3. [Local development setup](#3-local-development-setup)
4. [Repository layout](#4-repository-layout)
5. [Development workflow](#5-development-workflow)
6. [Code standards](#6-code-standards)
7. [Testing](#7-testing)
8. [Database migrations](#8-database-migrations)
9. [Adding a new endpoint](#9-adding-a-new-endpoint)
10. [Adding a new service](#10-adding-a-new-service)
11. [Pull request checklist](#11-pull-request-checklist)
12. [Issue reporting](#12-issue-reporting)

---

## 1. Project overview

gochat is a Discord-like real-time chat platform built as a Go monorepo. It is composed of several independent services:

| Service | Directory | Purpose |
|---------|-----------|---------|
| API | `cmd/api` | Main REST API (guilds, channels, messages, voice) |
| Auth | `cmd/auth` | Authentication, TOTP, password recovery |
| WebSocket | `cmd/ws` | Real-time event fan-out over WebSocket (NATS-backed) |
| SFU | `cmd/sfu` | WebRTC Selective Forwarding Unit for voice/video |
| Webhook | `cmd/webhook` | Internal event receiver (SFU heartbeats, presence) |
| Attachments | `cmd/attachments` | File upload / S3 proxy |
| Embedder | `cmd/embedder` | URL embed generation |
| Indexer | `cmd/indexer` | Full-text search indexing |

Shared library code lives in `internal/`. Services communicate via NATS. Persistent data is split between PostgreSQL (relational) and ScyllaDB/Cassandra (messages, reactions). Redis/KeyDB is the cache layer.

---

## 2. Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.22+ | `go version` |
| Docker & Docker Compose | latest | All infrastructure runs in containers |
| `migrate` CLI | v4.19.1 | `make tools` installs it |
| `swag` CLI | latest | `make tools` installs it — for Swagger generation |
| `golangci-lint` | v1.57+ | See [install guide](https://golangci-lint.run/usage/install/) |

---

## 3. Local development setup

### 3.1 Start infrastructure

```bash
# Clone the repo
git clone https://github.com/FlameInTheDark/gochat.git
cd gochat

# Start all containers (Postgres, ScyllaDB, Redis, NATS, etcd, Traefik)
make up

# Run all migrations (Postgres + Cassandra)
make migrate
```

### 3.2 Configure a service

Each service reads its configuration from a `*_config.yaml` file at the **repository root** or from environment variables. Example files are committed for every service:

```bash
cp api_config.example.yaml api_config.yaml
# edit api_config.yaml — for non-local environments, set auth_secret to a random 32+ char string
```

All config example files follow the pattern `<service>_config.example.yaml`:
`api`, `auth`, `ws`, `sfu`, `webhook`, `attachments`, `embedder`, `indexer`, `telemetry_gateway`.

**Auth secret guidance:**
- `auth_secret` — required; use a random 32+ character value outside local development

### 3.3 Run a service

```bash
make run          # API server on :3100
make run_ws       # WebSocket server
make run_embedder # Embed generator
```

Or directly:
```bash
go run ./cmd/api
go run ./cmd/ws
```

### 3.4 Regenerate Swagger docs

Run after any handler annotation change:
```bash
make swag
```

Generated output goes to `docs/api/swagger.json`. Commit this file along with your handler changes.

---

## 4. Repository layout

```
gochat/
├── cmd/                        # Service entry points
│   ├── api/                    # REST API
│   │   ├── config/             # Config struct + loader
│   │   ├── endpoints/          # Route handlers, grouped by domain
│   │   │   ├── guild/          # Guild, channel, role, voice, emoji
│   │   │   ├── message/        # Messages, threads, attachments
│   │   │   └── user/           # User profile, friends, settings
│   │   ├── app.go              # Dependency wiring + server setup
│   │   └── main.go
│   └── ...                     # auth, ws, sfu, webhook, attachments, embedder, indexer
│
├── internal/                   # Shared library code
│   ├── cache/                  # Cache interface + Redis implementation
│   │   └── kvs/                # go-redis backed implementation
│   ├── database/
│   │   ├── model/              # Plain Go structs matching DB schema
│   │   ├── pgentities/         # PostgreSQL repositories (one package per table group)
│   │   └── entities/           # ScyllaDB repositories
│   ├── dto/                    # API response types (never used in DB layer)
│   ├── mq/                     # NATS publisher/subscriber helpers
│   │   └── mqmsg/              # Message payload types
│   ├── observability/          # OpenTelemetry, slog integration
│   ├── permissions/            # Permission bitmask helpers
│   └── helper/                 # JWT, HTTP helpers
│
├── migration/
│   ├── postgres/               # PostgreSQL migration files (.sql)
│   └── cassandra/              # ScyllaDB migration files (.cql)
│
├── docs/
│   ├── api/                    # Generated Swagger JSON (swagger.json)
│   ├── CODESTYLE.md            # Code style rules (read before contributing)
│   └── project/                # Architecture and feature documentation
│       ├── README.md
│       ├── Services.md
│       ├── Database.md
│       ├── AuthSecurity.md
│       ├── Presence.md
│       ├── channels/           # Channel types, message types, threads, embeds
│       ├── guilds/             # Roles, permissions, moderation, custom emoji
│       ├── voice/              # SFU protocol, architecture, encryption
│       └── observability/      # Signals, runbooks, dashboards, local dev
│
├── clients/                    # Generated API clients (do not edit by hand)
│   └── api/
│       ├── goclient/           # Go client (openapi-generator)
│       └── jsclient/           # TypeScript/Axios client (openapi-generator)
│
├── monitoring/                 # Grafana dashboards, Prometheus config
├── init/                       # Container init scripts (e.g. ScyllaDB init)
├── vendor/                     # Vendored Go dependencies (go mod vendor)
│
│   # Per-service config and Dockerfiles — all at the repository root:
├── api_config.example.yaml     # Example configs (committed)
├── api_config.yaml             # Local overrides (git-ignored)
├── api.Dockerfile              # One Dockerfile per service
├── ...                         # (auth, ws, sfu, webhook, attachments, embedder, indexer)
├── compose.yaml                # Docker Compose for local development
├── migration.Dockerfile        # Runs golang-migrate for DB migrations
├── otel-collector-config.yaml  # OpenTelemetry collector config
├── doc.go                      # Swagger entry point (swag init -g doc.go)
└── Makefile
```

### Key conventions

- `internal/database/model/` — pure data structs, no business logic, no methods.
- `internal/dto/` — API response types, separate from model types. Never return model types directly from handlers.
- `internal/database/pgentities/<name>/entity.go` — always exposes an interface (`Channel`, `Guild`, etc.) and an unexported `Entity` concrete type.

---

## 5. Development workflow

### 5.1 Branching

```
main        ← stable, tagged releases
dev         ← integration branch, all PRs target this
feature/*   ← new features (branch from dev)
fix/*       ← bug fixes (branch from dev)
refactor/*  ← refactoring (branch from dev)
```

Branch names use kebab-case: `feature/guild-thread-pagination`.

### 5.2 Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description>

[optional body]
[optional footer]
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`, `build`.

Scope is the service or package: `api`, `ws`, `sfu`, `cache`, `embedgen`, `auth`.

Examples:
```
feat(api): add thread pagination endpoint
fix(cache): use SetNX to prevent race on voice join
refactor(message): decompose SendMessage handler into helper functions
test(guild): add table-driven tests for role position updates
```

Breaking changes: append `!` after the type/scope and include a `BREAKING CHANGE:` footer.

### 5.3 Keeping your branch up to date

```bash
git fetch origin
git rebase origin/dev
```

Do not merge `dev` into your feature branch — rebase only.

---

## 6. Code standards

Read **[`docs/CODESTYLE.md`](docs/CODESTYLE.md)** in full before writing code. The highlights:

- All HTTP errors use `fiber.NewError(status, ErrConstant)`. Never pass `err.Error()` to a fiber error.
- Internal errors always wrap with `fmt.Errorf("verb noun: %w", err)`.
- Goroutines inside handlers use `observability.BackgroundFromContext(c.UserContext())`.
- Handler structs are named `handler` (unexported) with an exported `New(...)` constructor.
- Required secrets must fail startup if empty, and weak/local placeholder values should emit warnings.
- All initialisms are ALL_CAPS: `URL`, `ID`, `HTTP`, `NATS`, `JWT`.

### Linting

```bash
golangci-lint run ./...
```

The repository `.golangci.yml` enforces: `errcheck`, `govet`, `staticcheck`, `revive`, `contextcheck`.

Zero linter warnings are required for merge. Do not use `//nolint` directives without a comment explaining why.

### Formatting

```bash
gofmt -w .
goimports -w .
```

CI will reject unformatted code. Configure your editor to run these on save.

---

## 7. Testing

### 7.1 Running tests

```bash
# All unit tests
go test ./... -count=1

# With race detector (required before opening a PR)
go test ./... -count=1 -race

# Single package
go test ./cmd/api/endpoints/guild/... -v -run TestRolePositions

# Integration tests (requires running infrastructure)
go test ./... -count=1 -tags integration
```

### 7.2 What to test

| Layer | Test type | Notes |
|-------|-----------|-------|
| Handler | Unit (fake dependencies) | Table-driven, cover error paths |
| Repository | Integration (`-tags integration`) | Real DB in Docker |
| Business logic helpers | Unit | Pure functions, no fakes needed |
| Cache helpers | Unit (fake cache) | Use `testutil.Noop` base |

### 7.3 Writing test fakes

Use the `testutil.Noop` base type for cache fakes. Override only what your test exercises:

```go
import "github.com/FlameInTheDark/gochat/internal/cache/testutil"

type fakeCache struct {
    testutil.Noop
    stored map[string][]byte
}

func (f *fakeCache) GetJSON(ctx context.Context, key string, v interface{}) error {
    data, ok := f.stored[key]
    if !ok {
        return redis.Nil
    }
    return json.Unmarshal(data, v)
}
```

### 7.4 Test coverage expectations

- New features: all happy-path and primary error-path branches covered.
- Bug fixes: a regression test that would have caught the bug.
- Refactors: coverage must not decrease.

### 7.5 Benchmarks

Add benchmarks for functions that process large data sets or are on hot paths:

```go
func BenchmarkSomething(b *testing.B) {
    for i := 0; i < b.N; i++ {
        doSomething()
    }
}
```

Run with: `go test ./... -bench=. -benchmem`

---

## 8. Database migrations

### 8.1 Creating a migration

```bash
# PostgreSQL
make add_migration_postgres name=add_user_display_name

# ScyllaDB / Cassandra
make add_migration_cassandra name=add_reaction_index
```

This creates sequentially numbered files in `migration/postgres/` or `migration/cassandra/`.

### 8.2 Rules for migrations

- **Every migration must be reversible.** Provide both `up` and `down` files.
- **One logical change per migration.** Do not combine unrelated schema changes.
- **Never modify a merged migration.** If you need to correct a merged migration, create a new one.
- **Test both directions** before opening a PR: `make migrate` then `make migrate_down`, then `make migrate` again.
- **Avoid locking operations** on large tables in production migrations. Prefer `NOT VALID` constraints and background validation.

### 8.3 Applying migrations locally

```bash
make migrate          # apply all pending (both PG and Scylla)
make migrate_pg       # PostgreSQL only
make migrate_scylla   # ScyllaDB only

make migrate_pg_rollback    # roll back one PG migration
make migrate_scylla_rollback # roll back one Scylla migration
```

---

## 9. Adding a new endpoint

### 9.1 File structure

Create a new package under the appropriate service's `endpoints/` directory, or add to an existing package if the endpoint belongs to an existing domain:

```
cmd/api/endpoints/guild/
  handlers.go    ← add your handler method here
  scheme.go      ← add request/response types and Err* constants
```

### 9.2 Step-by-step

1. **Define request/response types** in `scheme.go` with JSON tags and an `ozzo-validation` `Validate()` method.
2. **Define error constants** as `const Err* = "..."` in `scheme.go` (or `errors.go` if the file grows large).
3. **Write the handler method** on `handler` in `handlers.go` following the parse → validate permissions → execute pattern.
4. **Add Swagger annotations** above the handler method.
5. **Register the route** in the entity's `Register(router fiber.Router)` method.
6. **Write tests** in `handlers_test.go` or a new `*_test.go` file. Cover the happy path and key error paths.
7. **Regenerate Swagger**: `make swag`.

### 9.3 Permission checks

Every endpoint that accesses guild or channel data must check permissions before returning any data:

```go
perms, err := h.pm.GetEffectivePermissions(c.UserContext(), guildID, userID)
if err != nil {
    return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetPermission)
}
if !perms.Has(permissions.ReadMessages) {
    return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
}
```

---

## 10. Adding a new service

1. Create `cmd/<service>/` with `main.go`, `app.go`, and `config/config.go`.
2. Follow the same startup pattern as existing services: `observability.Init` → `shutter.NewShutter` → `NewApp` → `Start`.
3. Add `run_<service>:` and `rebuild_<service>:` targets to `Makefile`.
4. Add the service to `compose.yaml` with proper health checks.
5. Add a `<service>.Dockerfile` at the repository root following the existing multi-stage build pattern.
6. Document the service's responsibility and external dependencies in its `README.md` (or in a `docs/` file if the service is complex).

---

## 11. Pull request checklist

Before marking your PR ready for review, verify:

**Code quality**
- [ ] `golangci-lint run ./...` passes with zero warnings
- [ ] `gofmt` / `goimports` applied
- [ ] No `err.Error()` passed directly to `fiber.NewError`
- [ ] All new goroutines in handlers use `observability.BackgroundFromContext`
- [ ] No required config fields have insecure defaults

**Tests**
- [ ] `go test ./... -count=1 -race` passes
- [ ] New functionality has unit tests
- [ ] Bug fixes include a regression test

**Database**
- [ ] Migrations have both `up` and `down` files
- [ ] Migrations tested in both directions locally

**Documentation**
- [ ] `make swag` run if any handler annotations changed
- [ ] Swagger JSON committed (`docs/api/swagger.json`)
- [ ] Relevant `docs/project/` doc updated if behaviour or architecture changed

**PR description**
- [ ] Links the issue being addressed (`Closes #123`)
- [ ] Describes the "why", not just the "what"
- [ ] Calls out any breaking changes or required deployment steps

---

## 12. Issue reporting

### Bug reports

Include:
- Service name and version (git SHA is fine)
- Steps to reproduce
- Expected vs. actual behavior
- Relevant log output (redact any tokens or credentials)
- Configuration snippet if relevant (redact secrets)

### Feature requests

Include:
- The problem you are trying to solve
- Your proposed solution (optional)
- Alternatives considered

### Security vulnerabilities

Do **not** open a public issue for security vulnerabilities. Email the maintainers directly or use GitHub's private security advisory feature.
