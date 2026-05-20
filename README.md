<div align="center">

<img src="docs/assets/banner.svg" alt="GoChat" width="100%" />

<br/>

[![Docs](https://img.shields.io/badge/Docs-Project%20Guide-0f172a?style=for-the-badge)](docs/project/README.md)
[![API](https://img.shields.io/badge/API-Swagger-f59e0b?style=for-the-badge)](docs/api/swagger.json)
[![UI](https://img.shields.io/badge/UI-gochat--react-818cf8?style=for-the-badge)](https://github.com/FlameInTheDark/gochat-react)
[![Desktop](https://img.shields.io/badge/Desktop-gochat--electron-818cf8?style=for-the-badge)](https://github.com/FlameInTheDark/gochat-electron)
[![Deployment](https://img.shields.io/badge/Deployment-gochat--deployment-059669?style=for-the-badge)](https://github.com/FlameInTheDark/gochat-deployment)

<br/>

**Distributed real-time chat and voice backend written in Go.**

REST API · WebSocket delivery · File uploads · Full-text search · Link-preview embeds · WebRTC voice and streaming

<br/>

[Changelog](CHANGELOG.md) &nbsp;·&nbsp; [Go client](clients/api/goclient/) &nbsp;·&nbsp; [TypeScript client](clients/api/jsclient/) &nbsp;·&nbsp; [License](LICENSE)

</div>

---

## ![](docs/assets/icons/layout-dashboard.svg) Architecture

GoChat is service-oriented — each binary has a single focused responsibility. Services communicate through NATS for async events and share YugabyteDB YSQL, ScyllaDB, and KeyDB for state.

→ [Full diagram, data-store reference, and voice flow walkthrough](docs/project/Architecture.md)

---

## ![](docs/assets/icons/server.svg) Services

| Service | Path | Responsibility |
| :-- | :-- | :-- |
| **API** | `cmd/api` | Public REST surface — guilds, channels, messages, search, uploads, voice control |
| **Auth** | `cmd/auth` | Registration, login, token refresh, email flows, password reset |
| **WebSocket Gateway** | `cmd/ws` | Real-time event delivery, presence updates, session management |
| **Attachments** | `cmd/attachments` | Upload pipeline for files, avatars, and icons; S3 storage and metadata |
| **Webhook** | `cmd/webhook` | Internal callbacks — SFU/stream heartbeats, stream lifecycle, and attachment finalization |
| **SFU** | `cmd/sfu` | WebRTC media relay and WebSocket signaling for voice channels |
| **Stream** | `cmd/stream` | WebRTC media relay and WebSocket signaling for voice-channel screen/app sharing |
| **Indexer** | `cmd/indexer` | Consumes NATS message events, writes search documents to OpenSearch |
| **Embedder** | `cmd/embedder` | Builds link-preview embeds from remote metadata |
| **Telemetry Gateway** | `cmd/telemetrygateway` | OTEL proxy — collects signals from all services, forwards to observability backend |
| **Tools** | `cmd/tools` | Operational helpers: observability bootstrap, webhook token generation |

---

## ![](docs/assets/icons/zap.svg) Features

- Account lifecycle with JWT-based authentication and email flows
- Guilds, channels, roles, permissions, invites, bans, and custom emoji
- Direct messages, threads, message history, mentions, and reactions
- File attachments, avatars, and icons via S3-compatible storage
- Link-preview embed generation from remote metadata
- Presence updates and real-time event fanout over WebSocket
- Full-text search indexing and query through OpenSearch
- Voice channels with region-aware SFU discovery, WebRTC relay, and separate screen/app streaming
- Structured observability — distributed traces, metrics, and logs via OTEL

---

## ![](docs/assets/icons/layers.svg) Stack

| Area | Technology |
| :-- | :-- |
| Language | Go `1.25.8` |
| HTTP / WebSocket | Fiber v2, Fiber WebSocket |
| Voice / WebRTC | Pion WebRTC |
| Relational DB | YugabyteDB YSQL |
| Wide-column DB | ScyllaDB |
| Cache / sessions | Redis / KeyDB |
| Message bus | NATS |
| Search | OpenSearch |
| Object storage | S3-compatible |
| Service discovery | etcd |
| Reverse proxy | Traefik |
| Observability | OpenTelemetry + OpenObserve |

---

## ![](docs/assets/icons/rocket.svg) Getting Started

### Prerequisites

- Go `1.25.8` or newer
- Docker and Docker Compose
- GNU Make
- `migrate` CLI for creating migration files locally (`make tools` installs it)

### Quick setup

```bash
make setup
```

Installs local tooling, starts the full Compose stack, initializes ScyllaDB, and applies all migrations.

### Manual bootstrap

```bash
# Start infrastructure
docker compose up -d
docker compose exec scylla bash ./init-scylladb.sh

# Apply all migrations
make migrate
```

To test the migration image used in CI:

```bash
make build_migration_image
make migrate_image \
  YUGABYTE_ADDRESS="postgres://yugabyte:yugabyte@host.docker.internal:5433/gochat?sslmode=disable" \
  CASSANDRA_ADDRESS="cassandra://host.docker.internal/gochat?x-multi-statement=true"
```

CI publishes `ghcr.io/<owner>/gochat-migrations:<tag>` for releases and `ghcr.io/<owner>/gochat-migrations:dev` from the `dev` branch. Database bootstrap steps outside the migration files (ScyllaDB keyspace creation and YugabyteDB database creation) must be completed before running the container.

The container accepts `YUGABYTE_ADDRESS`, `CITUS_ADDRESS`, `PG_ADDRESS` for backward compatibility, `CASSANDRA_ADDRESS`, and `MIGRATION_SCOPE=all|yugabyte|yb|ysql|citus|postgres|pg|cassandra|scylla`.

Local Compose creates the `gochat` YugabyteDB database with `YUGABYTE_COLOCATION=false` by default. This keeps core relational metadata sharded for high-load deployments; enable colocation only for small isolated test databases or tiny reference datasets.

### Service configuration

Copy and edit the example config for each service:

```
api_config.example.yaml          ws_config.example.yaml
auth_config.example.yaml         sfu_config.example.yaml
stream_config.example.yaml       telemetry_gateway_config.example.yaml
attachments_config.example.yaml  indexer_config.example.yaml
webhook_config.example.yaml      embedder_config.example.yaml
```

---

## ![](docs/assets/icons/terminal.svg) Running Services

```bash
go run ./cmd/api
go run ./cmd/auth
go run ./cmd/ws
go run ./cmd/attachments
go run ./cmd/webhook
go run ./cmd/indexer
go run ./cmd/embedder
go run ./cmd/telemetrygateway
go run ./cmd/sfu          # voice media; runs separately from Compose
go run ./cmd/stream       # screen/app streaming media; runs separately from Compose
```

Useful Make targets:

| Target | What it does |
| :-- | :-- |
| `make up` | Start the Compose stack and initialize ScyllaDB |
| `make down` | Stop the stack |
| `make migrate` | Apply all database migrations |
| `make swag` | Rebuild `docs/api/swagger.json` |
| `make client` | Regenerate Go and TypeScript API clients |
| `make rebuild_all` | Rebuild API, Auth, Indexer, Embedder, and WS containers |
| `make build_migration_image` | Build the versioned migration container locally |

---

## ![](docs/assets/icons/activity.svg) Observability

The local stack uses OpenObserve and the OpenTelemetry Collector. Bootstrap dashboards with the Tools CLI:

```bash
go run ./cmd/tools observability bootstrap \
  --url http://localhost:5080 --org default \
  --user root@example.com --password Complexpass#123

go run ./cmd/tools observability smoke \
  --url http://localhost:5080 --org default \
  --user root@example.com --password Complexpass#123
```

| Endpoint | URL |
| :-- | :-- |
| OpenObserve | http://localhost:5080 |
| OTEL Collector health | http://localhost:13133/ |
| Traefik dashboard | http://localhost:8080 |
| OpenSearch Dashboards | http://localhost:5601 |

---

## ![](docs/assets/icons/book-open.svg) Documentation

| | |
| :-- | :-- |
| [Architecture](docs/project/Architecture.md) | Full service diagram, data-store reference, voice flow |
| [Services overview](docs/project/Services.md) | Per-service responsibilities and config reference |
| [Channels & messages](docs/project/channels/README.md) | Channel types, message types, threads, embeds |
| [Guilds, roles & permissions](docs/project/guilds/README.md) | Roles, permissions bitmask, moderation, custom emoji |
| [Presence system](docs/project/Presence.md) | Presence state model and delivery |
| [WebSocket protocol](docs/project/ws/README.md) | Event types, subscription model, connection lifecycle |
| [Voice & SFU](docs/project/voice/README.md) | WebRTC signaling, SFU protocol, permissions |
| [Voice-channel streaming](docs/project/voice/Streaming.md) | Screen/app streaming service, lifecycle, presence, and region migration |
| [Direct-message calls](docs/project/voice/DMCalls.md) | Private 1:1 voice calls, settings bootstrap, pair-scoped events, and DM call streaming |
| [Observability](docs/project/observability/README.md) | OTEL signals, dashboards, runbooks, external SFU |
| [Auth security](docs/project/AuthSecurity.md) | Token design, expiry, refresh flow |
| [Database schema](docs/project/Database.md) | YugabyteDB YSQL and ScyllaDB schema diagrams |
| [Tools CLI](docs/project/Tools.md) | Operational helper commands |
| [OpenAPI schema](docs/api/swagger.json) | Machine-readable API spec |
| [Go API client](clients/api/goclient/) | Generated Go client |
| [TypeScript API client](clients/api/jsclient/) | Generated TypeScript client |
| [Frontend repo](https://github.com/FlameInTheDark/gochat-react) | React web client |
| [Desktop client](https://github.com/FlameInTheDark/gochat-electron) | Electron desktop app |
| [Deployment repo](https://github.com/FlameInTheDark/gochat-deployment) | Production deployment manifests |

---

## Repository Layout

```
cmd/             runnable services and operational tools
internal/        shared packages (transport, storage, search, mail, presence, server wiring)
migration/       YugabyteDB YSQL, legacy PostgreSQL/Citus, and ScyllaDB migrations
docs/            project documentation and generated OpenAPI schema
clients/api/     generated Go and TypeScript API clients
compose.yaml     local development stack
Makefile         bootstrap, migration, client generation, and rebuild targets
```

---

<div align="center">

MIT License · See [LICENSE](LICENSE)

</div>
