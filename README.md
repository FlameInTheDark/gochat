# GoChat

GoChat is a real-time chat platform written in Go, designed around a set of focused services communicating over HTTP, WebSockets, and message queues.

[Documentation](docs/project/README.md)

## Architecture overview

GoChat follows a service-oriented architecture with a clear separation of concerns:

- API service (cmd/api): Public REST API for user and chat management; orchestrates business logic.
- Auth service (cmd/auth): Authentication/authorization, token issuing/validation.
- WS Gateway (cmd/ws): WebSocket gateway for real-time messaging; authenticates clients and bridges to messaging backend.
- Indexer (cmd/indexer): Consumes message events and builds the search index.
- Message Queue (internal/mq): NATS provides reliable message delivery between services, enabling loose coupling and asynchronous communication for real-time features and event-driven architecture.
- Database (PostgreSQL + ScyllaDB): PostgreSQL is the primary relational store (users, guilds/channels, memberships, auth); ScyllaDB handles high-throughput message timelines and attachment metadata. PostgreSQL schema and migrations under db/pgmigrations and ScyllaDB migrations are under db/migrations.
- Search (OpenSearch): Full-text search for messages, driven by the Indexer.
- Object Storage (S3-compatible): File and attachment storage via internal/s3.
- Caching (internal/cache): In-memory helpers and adapters; used for hot paths where applicable.
- Mailer (internal/mailer): Pluggable email delivery (SMTP, SendPulse) with templates for registration and password resets.
- Edge/Proxy: Traefik fronts services in the Docker Compose deployment.

## Core services

1. API (cmd/api)
   - Responsibilities: general API endpoints for client business logic.
   - Endpoints: see cmd/api/endpoints for route handlers and internal/dto for payloads.
   - Persistence: PostgreSQL via internal/database/pgdb + internal/database/pgentities; ScyllaDB via internal/database/db + internal/database/entities (messages, attachments). Migrations for PostgreSQL in db/pgmigrations.
   - Outbound integrations: publishes events to MQ (internal/indexmq, internal/mq).

2. Auth (cmd/auth)
   - Responsibilities: login, registration, password reset flows.
   - Security: issues JWT tokens.
   - Configuration: auth_config.yaml.

3. WS Gateway (cmd/ws)
   - Responsibilities: upgrades HTTP to WebSocket, authenticates connections (cmd/ws/auth) using JWT token provided by auth service, routes messages to/from MQ, and pushes real-time updates to clients.
   - Subscribers/Handlers: see cmd/ws/subscriber and cmd/ws/handler for message processing.
   - Configuration: ws_config.yaml.

4. Indexer (cmd/indexer)
   - Responsibilities: consumes message-related events from MQ and updates the search backend (OpenSearch).
   - Configuration: indexer_config.yaml.

## Supporting infrastructure

- Message Queue (internal/mq)
  - Backend: internal/mq/nats provides the NATS implementation used by all services in the reference deployment.
  - Message types: internal/mq/mqmsg defines event payloads exchanged between services.

- Database (PostgreSQL + ScyllaDB)
  - PostgreSQL:
    - Code: internal/database/pgdb and internal/database/pgentities.
    - Migrations: db/pgmigrations and db/migrations.
  - ScyllaDB:
    - Code: internal/database/db (CQL connector) and internal/database/entities.
    - Usage: high-throughput message and attachment storage; queried by API/WS; uses gocql driver.

- Search (OpenSearch)
  - Code: internal/solr currently contains the search client/index management used by the Indexer; the deployment uses OpenSearch.
  - Usage: primarily by the Indexer to build/query message indexes.

- Object Storage (S3-compatible)
  - Code: internal/s3 provides an adapter for S3/MinIO.
  - Usage: media uploads and file attachments associated with chats and messages.

- Mailer
  - Code: internal/mailer with pluggable providers under internal/mailer/providers.
  - Providers: SMTP (internal/mailer/providers/smtp) and SendPulse (internal/mailer/providers/sendpulse).
  - Templates: password_reset.tmpl, email_notify.tmpl; rendering in internal/mailer/template.go.

- Caching
  - Code: internal/cache with in-memory helpers (e.g., internal/cache/vkc).

- Server bootstrap
  - Code: internal/server provides common server wiring (HTTP, middleware), internal/shutter for graceful shutdown.

## Data and control flows

1. Registration and email confirmation
   - API receives registration, persists user in PostgreSQL, and sends a confirmation email via Mailer.
   - User clicks the confirmation link; token is validated by Auth/API and the account is activated.

2. Authentication
   - Client submits credentials to Auth, which returns a token used for REST and WS authentication.
   - WS Gateway validates tokens on connection and for subsequent privileged actions.

3. Messaging
   - Client sends a message via API; API publishes a message event to MQ (NATS).
   - API persists message metadata in ScyllaDB when needed; Indexer consumes events from MQ and updates OpenSearch for search.
   - WS broadcasts the message to relevant recipients in real time.

4. Attachments
   - Client requests an upload persistent URL for the S3 service from the API; API stores metadata in ScyllaDB and after upload to S3 receives an event to verify the file upload.

5. Search
   - API exposes search endpoints; queries go to OpenSearch, results are combined with relational data as needed. (in development)

## ScyllaDB in GoChat

ScyllaDB is used alongside PostgreSQL to handle high-throughput, append-heavy data:
- What is stored in ScyllaDB:
  - Messages and message timelines (see internal/database/entities/message, uses CQL and id buckets via internal/idgen.GetBucket).
  - Attachments ephemeral metadata (see internal/database/entities/attachment). Temporary upload rows use TTL and are finalized when upload completes.
- Access patterns:
  - API writes/reads messages and attachments; WS reads to serve real-time history windows.
  - Entities are implemented with gocql; the connector lives in internal/database/db (NewCQLCon).
- Configuration:
  - api_config.yaml and ws_config.yaml contain:
    - cluster: ["scylla"] — list of Scylla nodes/hosts.
    - cluster_keyspace: "gochat" — keyspace used by the app.
  - compose.yaml includes the scylla service exposed on 9042 for CQL.
- Complementary PostgreSQL usage:
  - Relational data such as users, guilds, channels, memberships, roles, and auth-related records live in PostgreSQL and are modeled under internal/database/pgentities. SQL access is via internal/database/pgdb.

## Configuration layout

Top-level config files illustrate configuration shapes for each service:
- api_config.yaml: API service settings (DB, MQ, Mailer, S3, search, etc.).
- auth_config.yaml: Auth service settings.
- ws_config.yaml: WebSocket gateway settings.
- indexer_config.yaml: Indexer settings (NATS + OpenSearch).

Reference deployment manifest:
- compose.yaml: Reference Docker Compose stack for local/dev.
- Helm Chart: Planned in feature after most of the core services are implemented.
