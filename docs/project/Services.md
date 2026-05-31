# Services Overview

This project is composed of several services located under the `cmd/` directory. Each service is a separate application with a focused responsibility. Below is a brief overview to help you navigate and understand their roles and primary dependencies.

## API (`cmd/api`)
- Purpose: Public HTTP API gateway for the platform (guilds, channels, messages, search, voice and stream control).
- Key features:
  - REST endpoints for core resources and actions.
  - User-authenticated bot management and guild-admin bot installation/removal routes.
  - Issues short‑lived SFU tokens for voice join/move flows.
  - Issues separate short-lived stream tokens for screen/app sharing publishers and viewers.
  - Manages voice region overrides and selects SFU instances via discovery.
  - Selects stream instances in the same effective region as the voice channel.
  - Publishes client events via user NATS and bot fan-out events via bot NATS.
- Dependencies: Scylla/Cassandra, YugabyteDB YSQL, Redis/KeyDB (cache), user NATS, bot NATS, OpenSearch (via Indexer), etcd (discovery).

## Auth (`cmd/auth`)
- Purpose: Authentication and account lifecycle.
- Key features:
  - Login, registration, token refresh (access/refresh), password reset, password change, and TOTP-based two-factor authentication flows.
  - Login challenges backed by Redis/KeyDB for second-step verification, recovery codes, and email recovery fallback.
  - Session-versioned JWT issuance so password and 2FA mutations revoke older access, refresh, and WebSocket sessions.
  - Email delivery via pluggable providers (SMTP, SendPulse, Resend, or log-only).
- Dependencies: YugabyteDB YSQL, Redis/KeyDB (cache and MFA state), NATS (session revocation fan-out).

## WebSocket Gateway (`cmd/ws`)
- Purpose: Persistent WebSocket gateway for client real‑time updates.
- Key features:
  - Bridges NATS topics to user connections (subscribe/publish per user/guild/channel).
  - Presence heartbeats and aggregation, session tracking, and OTEL-based telemetry shipped to OpenObserve.
  - Validates client tokens and enforces access on subscriptions.
- Dependencies: NATS, Scylla/Cassandra, YugabyteDB YSQL, Redis/KeyDB (presence/cache).

## Bot API (`cmd/botapi`)
- Purpose: Dedicated REST API for bot runtime traffic.
- Key features:
  - Authenticates only bot runtime tokens with `Authorization: Bot <token>`.
  - Exposes bot account context, installed guilds, visible channels, message send/edit/delete/read, typing, reactions, and read-state routes.
  - Uses bot-specific endpoint packages under `cmd/botapi/endpoints` and does not mount user API handlers.
  - Enforces normal guild/channel permissions capped by the guild install grant.
- Dependencies: Scylla/Cassandra, YugabyteDB YSQL, Redis/KeyDB, user NATS, bot NATS.

## Bot Event Router (`cmd/botrouter`)
- Purpose: Partitioned fan-out service for bot gateway events.
- Key features:
  - Owns `bot.event.p.*` partitions through Redis leases.
  - Expands each guild event to all currently installed eligible bots.
  - Applies grant caps, role permissions, and channel visibility before delivery.
  - Publishes compact delivery envelopes to `bot.deliver.instance.*` subjects.
- Dependencies: bot NATS, YugabyteDB YSQL, Redis/KeyDB.

## Bot WebSocket Gateway (`cmd/botws`)
- Purpose: Dedicated WebSocket gateway for bot event delivery.
- Deployment config: `botws_config.yaml` mounted as `/dist/config.yaml`.
- Key features:
  - Authenticates upgrades with `Authorization: Bot <token>`.
  - Registers active bot sessions in Redis for router targeting.
  - Uses guild-based sharding so each guild event stream is delivered to one bot shard, or to all active instances when unsharded.
  - Subscribes only to its instance delivery subject on bot NATS.
- Dependencies: bot NATS, YugabyteDB YSQL, Redis/KeyDB.

## SFU (`cmd/sfu`)
- Purpose: Voice Selective Forwarding Unit with WebRTC media relay and WS signaling.
- Deployment: external to local Compose. The SFU is expected to run as a standalone service and self-ship telemetry to OpenObserve.
- Key features:
  - WebSocket signaling endpoint at `/signal` (or the ingress-prefixed equivalent).
  - Validates short‑lived SFU tokens and enforces voice permissions (speak/video/connect).
  - Admin controls: kick, block/unblock, and move notifications.
  - Reports load (peer count) via periodic heartbeats.
- Discovery & heartbeat:
  - SFU sends `POST /api/v1/webhook/sfu/heartbeat` to the Webhook service with header `X-Webhook-Token: <JWT>`.
  - Webhook validates the token (HS256, claims: `{ typ:"sfu", id:"<service_id>" }`) and writes/refreshes the instance in discovery (etcd).
  - API reads instances from etcd when serving JoinVoice. No fallback to origin; returns 503 when no instance exists.
- Dependencies: Webhook (for discovery), etcd (backing store for discovery), optional STUN servers.
- Config: `webhook_url`, pre-generated `webhook_token` (HS256 JWT), `service_id` (must match token `id`), and optional `dtls_certificate_file` / `dtls_private_key_file` for reusable WebRTC DTLS certificates.
- DTLS cert generation: `go run ./cmd/tools certificates dtls generate --cert-out ./certs/sfu.crt --key-out ./certs/sfu.key`
- Observability: see `docs/project/observability/ExternalSFU.md` for the direct OTLP and direct log-shipping contract.

## Webhook (`cmd/webhook`)
- Purpose: Secure integration surface for internal events (SFU discovery heartbeat, stream discovery/lifecycle, attachment finalize).
- Endpoints:
  - `POST /api/v1/webhook/sfu/heartbeat` — body: `{ id, region, url, load }`, header: `X-Webhook-Token: <JWT>`.
  - `POST /api/v1/webhook/stream/heartbeat` — body: `{ id, region, url, load }`, header: `X-Webhook-Token: <JWT>`.
  - `POST /api/v1/webhook/stream/start` — marks a stream active and publishes stream presence/events.
  - `POST /api/v1/webhook/stream/stop` — clears active stream state and publishes stream stop presence/events.
  - `POST /api/v1/webhook/stream/alive` — refreshes active stream route/meta/presence TTLs.
  - `POST /api/v1/webhook/attachments/finalize` — updates attachment metadata after upload completes.
- Auth: HS256 JWT in `X-Webhook-Token` with claims `{ typ, id }`; no expiration is required.
- Config: `jwt_secret`, `etcd_endpoints`, `etcd_prefix`, `stream_etcd_prefix`, and optional Cassandra cluster for attachments.
- Writes SFU and stream instances into etcd for API discovery; media services do not talk to etcd directly when webhook is used.
- Token generation: use `go run ./cmd/tools tokens webhook generate --type sfu --secret <jwt_secret> [--id <service_id>]` or the same command with `--type stream`, then set the result as the media service `webhook_token`.

## Stream (`cmd/stream`)
- Purpose: Screen/app sharing media service for voice channels.
- Deployment: external to local Compose, like `cmd/sfu`. Run it as a standalone service for local WebRTC testing.
- Key features:
  - WebSocket signaling endpoint at `/signal`; GoChat clients use `/signal?v=2`.
  - Validates one-minute stream JWTs signed with shared `auth_secret` and bound to this stream service's `service_id`.
  - Enforces publisher/viewer roles: publishers may send video and optional audio; viewers are receive-only.
  - Forwards high-resolution screen/app RTP without transcoding or recording.
  - Supports DAVE over the same v2 signaling opcode surface used by voice.
  - Reports load and active stream lifecycle through Webhook callbacks.
- Discovery & heartbeat:
  - Stream service sends `POST /api/v1/webhook/stream/heartbeat` to the Webhook service with header `X-Webhook-Token: <JWT>`.
  - Webhook validates the token (HS256, claims: `{ typ:"stream", id:"<service_id>" }`) and writes/refreshes the instance in stream discovery (etcd).
  - API reads instances from `stream_etcd_prefix` and only selects instances in the effective voice region for the channel. No cross-region fallback is allowed for streams.
- Active stream callbacks:
  - `POST /api/v1/webhook/stream/start` marks the stream active after publisher media is established.
  - `POST /api/v1/webhook/stream/stop` clears stream state after publisher stop/disconnect.
  - `POST /api/v1/webhook/stream/alive` refreshes stream route/meta/presence TTLs while media is active.
- Dependencies: Webhook (for discovery and lifecycle), etcd (discovery store), Redis/KeyDB through Webhook/API for active stream state, optional STUN servers.
- Config: `auth_secret`, `region`, `public_base_url`, `webhook_url`, `webhook_token`, `service_id`, optional DTLS certificate files, optional UDP port range, DAVE settings, and stream bitrate ceilings.
- Token generation: use `go run ./cmd/tools tokens webhook generate --type stream --secret <jwt_secret> [--id <service_id>]` and set the result as stream `webhook_token`.
- Full pipeline: see `docs/project/voice/Streaming.md`.

## Attachments (`cmd/attachments`)
- Purpose: File upload service for message attachments, avatars, and icons.
- Key features:
  - Upload endpoints with size/type validation and metadata persistence.
  - S3‑compatible storage integration and public URL computation.
  - Emits events (e.g., avatar/icon updates) via NATS.
- Dependencies: Scylla/Cassandra, YugabyteDB YSQL, S3-compatible storage, NATS.

## Indexer (`cmd/indexer`)
- Purpose: Asynchronous search indexing worker.
- Key features:
  - Subscribes to NATS topics for message index, update, and delete events.
  - Writes to OpenSearch for full‑text search.
- Dependencies: NATS, OpenSearch.


## Embedder (`cmd/embedder`)
- Purpose: Asynchronous URL unfurling worker for message embeds.
- Key features:
  - Subscribes to `embed.make` events from the API.
  - Builds generated embeds from YouTube, oEmbed, Open Graph, and Twitter Card metadata.
  - Stores generated embeds separately from manual embeds and emits a normal `MessageUpdate` event after regeneration.
  - Blocks private or loopback fetch targets by default to reduce SSRF risk.
- Dependencies: NATS, Scylla/Cassandra, outbound HTTP(S).
