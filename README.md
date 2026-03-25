# GoChat

<p align="center">
  <strong>Distributed real-time chat and voice backend written in Go.</strong><br />
  REST API, WebSocket delivery, file uploads, search indexing, generated embeds, and SFU-based voice.
</p>

<p align="center">
  <a href="docs/project/README.md"><img alt="Docs" src="https://img.shields.io/badge/Docs-Project%20Guide-0f172a?style=for-the-badge" /></a>
  <a href="docs/api/swagger.json"><img alt="API" src="https://img.shields.io/badge/API-Swagger-f59e0b?style=for-the-badge" /></a>
  <a href="https://github.com/FlameInTheDark/gochat-react"><img alt="UI" src="https://img.shields.io/badge/UI-gochat--react-2563eb?style=for-the-badge" /></a>
<a href="https://github.com/FlameInTheDark/gochat-electron"><img alt="UI" src="https://img.shields.io/badge/DESKTOP-gochat--electron-2563eb?style=for-the-badge" /></a>
  <a href="https://github.com/FlameInTheDark/gochat-deployment"><img alt="Deployment" src="https://img.shields.io/badge/Deployment-gochat--deployment-059669?style=for-the-badge" /></a>
</p>

<p align="center">
  <a href="CHANGELOG.md">Changelog</a>
  |
  <a href="clients/api/goclient/">Go client</a>
  |
  <a href="clients/api/jsclient/">TypeScript client</a>
  |
  <a href="LICENSE">License</a>
</p>

## Overview

GoChat is a service-oriented messaging platform built around focused Go services. This repository contains the backend services, local development stack, generated API clients, and project documentation for messaging, presence, uploads, search, webhooks, and voice.

## Quick Links

- [Project documentation](docs/project/README.md)
- [Services overview](docs/project/Services.md)
- [Observability docs](docs/project/observability/README.md)
- [WebSocket docs](docs/project/ws/README.md)
- [Voice docs](docs/project/voice/README.md)
- [Presence system](docs/project/Presence.md)
- [Tools CLI](docs/project/Tools.md)
- [OpenAPI schema](docs/api/swagger.json)
- [Go API client](clients/api/goclient/)
- [TypeScript API client](clients/api/jsclient/)

## What Ships Here

| Area | What it covers |
| --- | --- |
| Core services | API, auth, WebSocket gateway, attachments, webhook, indexer, embedder, and SFU |
| Data layer | Citus/PostgreSQL for relational state and ScyllaDB for message-heavy workloads |
| Messaging | NATS-backed event flow for real-time delivery and background workers |
| Media | S3-compatible object storage for uploads, avatars, icons, and attachment assets |
| Search | OpenSearch indexing pipeline driven by the `indexer` service |
| Discovery | etcd-backed voice service discovery through the `webhook` service |
| Clients | Generated Go and TypeScript API clients under `clients/api/` |
| Operations | Docker Compose stack, OpenObserve assets, OTEL collector wiring, and helper CLI tooling |

## Architecture

```mermaid
flowchart LR
    Client[Client Applications]
    Traefik[Traefik Edge]

    subgraph App[Application Services]
        API[API]
        Auth[Auth]
        WS[WebSocket Gateway]
        Attachments[Attachments]
        Webhook[Webhook]
        Indexer[Indexer]
        Embedder[Embedder]
    end

    subgraph Voice[External Voice Plane]
        SFU[SFU]
    end

    subgraph Data[State and Infra]
        PG[Citus / PostgreSQL]
        Scylla[ScyllaDB]
        KeyDB[KeyDB]
        NATS[NATS]
        Search[OpenSearch]
        S3[S3-compatible Storage]
        Etcd[etcd]
    end

    Client --> Traefik
    Traefik --> API
    Traefik --> Auth
    Traefik --> WS
    Traefik --> Attachments
    Traefik --> Webhook

    API --> PG
    API --> Scylla
    API --> KeyDB
    API --> NATS
    API --> Search
    API --> Etcd

    Auth --> PG
    Auth --> KeyDB

    WS --> NATS
    WS --> Scylla
    WS --> PG
    WS --> KeyDB

    Attachments --> Scylla
    Attachments --> PG
    Attachments --> NATS
    Attachments --> S3

    Webhook --> Scylla
    Webhook --> Etcd

    Indexer --> NATS
    Indexer --> Search

    Embedder --> NATS
    Embedder --> Scylla

    Client --> SFU
    SFU --> Webhook
```

## Services

| Service | Path | Responsibility |
| --- | --- | --- |
| API | `cmd/api` | Main REST surface for users, guilds, channels, messages, invites, search, uploads, and voice control |
| Auth | `cmd/auth` | Registration, login, refresh tokens, email flows, and password reset |
| WebSocket Gateway | `cmd/ws` | Real-time subscriptions, event delivery, presence updates, and session handling |
| Attachments | `cmd/attachments` | Upload pipeline for attachments, avatars, icons, and related metadata |
| Webhook | `cmd/webhook` | Internal webhook surface for trusted service callbacks such as SFU heartbeats and attachment finalization |
| SFU | `cmd/sfu` | External WebRTC media relay and signaling service for voice channels |
| Indexer | `cmd/indexer` | Consumes message events and writes search documents to OpenSearch |
| Embedder | `cmd/embedder` | Builds generated message embeds from remote metadata and republishes updates |
| Tools | `cmd/tools` | Operational helpers such as webhook token generation |

## Feature Surface

- Account lifecycle with token-based authentication
- Guilds, channels, roles, invites, bans, and custom emoji
- Direct messages, message history, mentions, attachments, and embed generation
- Presence updates and real-time event fanout over WebSocket
- Search indexing and query flow through OpenSearch
- Voice channel join flow with region-aware SFU discovery
- Generated Go and TypeScript API clients from the OpenAPI schema

## Stack

- Go `1.25.1`
- Fiber, Fiber WebSocket, and Pion WebRTC
- Citus/PostgreSQL for relational data
- ScyllaDB for message timelines and attachment-heavy data
- NATS for async messaging between services
- KeyDB for cache and presence/session state
- OpenSearch for full-text search
- S3-compatible storage for media assets
- etcd for service discovery
- Traefik, OpenObserve, OpenTelemetry Collector, and OpenSearch Dashboards for local operations

## Repository Layout

```text
cmd/             runnable services and operational tools
internal/        shared packages for transport, storage, search, mail, presence, and server wiring
db/              PostgreSQL and ScyllaDB migrations
docs/            project docs and generated OpenAPI output
clients/api/     generated Go and TypeScript API clients
compose.yaml     reference local development stack
Makefile         bootstrap, migration, client generation, and rebuild commands
```

## Getting Started

### Prerequisites

- Go `1.25.1` or newer
- Docker and Docker Compose
- GNU Make
- `migrate` CLI when you want to create migration files locally (`make tools` installs it)

### Fast Path

```bash
make setup
```

`make setup` installs local tooling, starts the reference stack, initializes ScyllaDB, and applies both PostgreSQL and ScyllaDB migrations.

### Manual Bootstrap

Start the local infrastructure:

```bash
docker compose up -d
docker compose exec scylla bash ./init-scylladb.sh
docker compose -p gochat up --scale citus-worker=3 -d
```

Apply migrations locally with the existing Make targets:

```bash
make migrate
```

Build the versioned migration image locally when you want to test the same packaging used in CI:

```bash
make build_migration_image
make migrate_image PG_ADDRESS="postgres://postgres@host.docker.internal/gochat" CASSANDRA_ADDRESS="cassandra://host.docker.internal/gochat?x-multi-statement=true"
```

When you run the migration image from Docker, the connection strings need container-reachable hosts such as Compose service DNS names or `host.docker.internal`, not host-local `127.0.0.1`.

For deployments, GitHub Actions now publishes `ghcr.io/<owner>/gochat-migrations:<tag>` for releases and `ghcr.io/<owner>/gochat-migrations:dev` from the `dev` branch. The image contains the exact migration files for that version and defaults to applying both migration sets with `up`.

This image applies versioned schema migrations. Database bootstrap that is outside the migration files, such as creating the ScyllaDB keyspace or enabling Citus, still needs to be completed before the container runs.

You can scope it to a single database or change the command:

```bash
docker run --rm \
  -e PG_ADDRESS="postgres://postgres@postgres/gochat?sslmode=disable" \
  -e CASSANDRA_ADDRESS="cassandra://scylla/gochat?x-multi-statement=true" \
  ghcr.io/<owner>/gochat-migrations:v1.2.3

docker run --rm \
  -e MIGRATION_SCOPE=postgres \
  -e PG_ADDRESS="postgres://postgres@postgres/gochat?sslmode=disable" \
  ghcr.io/<owner>/gochat-migrations:v1.2.3 down 1
```

The container accepts:

- `PG_ADDRESS` for PostgreSQL migrations
- `CASSANDRA_ADDRESS` for Cassandra or ScyllaDB migrations
- `MIGRATION_SCOPE=all|postgres|pg|cassandra|scylla`
- `MIGRATION_COMMAND` as a default command when you prefer env-driven invocation

Review the example configuration files before running services locally:

- `api_config.example.yaml`
- `auth_config.example.yaml`
- `attachments_config.example.yaml`
- `ws_config.example.yaml`
- `sfu_config.example.yaml` for standalone SFU deployments outside Compose
- `webhook_config.example.yaml`
- `indexer_config.example.yaml`
- `embedder_config.example.yaml`

## Running Services

Run individual services directly with Go:

```bash
go run ./cmd/api
go run ./cmd/auth
go run ./cmd/ws
go run ./cmd/attachments
go run ./cmd/webhook
go run ./cmd/indexer
go run ./cmd/embedder
```

Run the SFU separately when you need voice media in a non-Compose environment:

```bash
go run ./cmd/sfu
```

Useful Make targets:

- `make up` to start the Compose stack and initialize ScyllaDB
- `make down` to stop the stack
- `make migrate` to apply both database migration sets
- `make build_migration_image` to build the versioned migration container locally
- `make migrate_image` to run both migration sets through the container locally
- `make swag` to rebuild `docs/api/swagger.json`
- `make client` to regenerate Go and TypeScript clients
- `make rebuild_all` to rebuild the application containers

## Documentation and Clients

| Resource                  | Link                                                                     |
|---------------------------|--------------------------------------------------------------------------|
| Project docs              | [docs/project/README.md](docs/project/README.md)                         |
| Service documentation     | [docs/project/Services.md](docs/project/Services.md)                     |
| Observability             | [docs/project/observability/README.md](docs/project/observability/README.md) |
| WebSocket protocol        | [docs/project/ws/README.md](docs/project/ws/README.md)                   |
| Voice and SFU docs        | [docs/project/voice/README.md](docs/project/voice/README.md)             |
| Presence model            | [docs/project/Presence.md](docs/project/Presence.md)                     |
| OpenAPI schema            | [docs/api/swagger.json](docs/api/swagger.json)                           |
| Go API client             | [clients/api/goclient/](clients/api/goclient/)                           |
| TypeScript API client     | [clients/api/jsclient/](clients/api/jsclient/)                           |
| Frontend repository       | [gochat-react](https://github.com/FlameInTheDark/gochat-react)           |
| Desktop client repository | [gochat-electron](https://github.com/FlameInTheDark/gochat-electron)     |
| Deployment repository     | [gochat-deployment](https://github.com/FlameInTheDark/gochat-deployment) |

## Local Observability

The supported local observability workflow is OpenObserve plus the OpenTelemetry Collector:

- Start fresh with `docker compose down --remove-orphans` and then `docker compose up -d`.
- Bootstrap dashboards and alerts with `go run ./cmd/tools observability bootstrap --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123`.
- Run the smoke check with `go run ./cmd/tools observability smoke --url http://localhost:5080 --org default --user root@example.com --password Complexpass#123`.
- OpenObserve is available on `http://localhost:5080`.
- OTEL collector health is available on `http://localhost:13133/`.
- Traefik dashboard remains available on `http://localhost:8080`.
- OpenSearch Dashboards remains available on `http://localhost:5601`.

PostgreSQL health in the local stack is now reported through native service-side probe telemetry rather than a Prometheus exporter bridge.

## License

MIT. See [LICENSE](LICENSE).
