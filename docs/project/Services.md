# Services Overview

This project is composed of several services located under the `cmd/` directory. Each service is a separate application with a focused responsibility. Below is a brief overview to help you navigate and understand their roles and primary dependencies.

## API (`cmd/api`)
- Purpose: Public HTTP API gateway for the platform (guilds, channels, messages, search, voice control).
- Key features:
  - REST endpoints for core resources and actions.
  - Issues short‑lived SFU tokens for voice join/move flows.
  - Manages voice region overrides and selects SFU instances via discovery.
  - Publishes/consumes events via NATS.
- Dependencies: Scylla/Cassandra, PostgreSQL, Redis/KeyDB (cache), NATS, OpenSearch (via Indexer), etcd (discovery).

## Auth (`cmd/auth`)
- Purpose: Authentication and account lifecycle.
- Key features:
  - Login, registration, token refresh (access/refresh), password reset flows.
  - Email delivery via pluggable providers (SMTP, SendPulse, Resend, or log-only).
- Dependencies: PostgreSQL, Redis/KeyDB (cache).

## WebSocket Gateway (`cmd/ws`)
- Purpose: Persistent WebSocket gateway for client real‑time updates.
- Key features:
  - Bridges NATS topics to user connections (subscribe/publish per user/guild/channel).
  - Presence heartbeats and aggregation, session tracking, and metrics (`/metrics`).
  - Validates client tokens and enforces access on subscriptions.
- Dependencies: NATS, Scylla/Cassandra, PostgreSQL, Redis/KeyDB (presence/cache).

## SFU (`cmd/sfu`)
- Purpose: Voice Selective Forwarding Unit with WebRTC media relay and WS signaling.
- Key features:
  - WebSocket signaling endpoint at `/sfu/signal`.
  - Validates short‑lived SFU tokens and enforces voice permissions (speak/video/connect).
  - Admin controls: kick, block/unblock, and move notifications.
  - Reports load (peer count) via periodic heartbeats.
- Discovery & heartbeat:
  - SFU sends `POST /api/v1/webhook/sfu/heartbeat` to the Webhook service with header `X-Webhook-Token: <JWT>`.
  - Webhook validates the token (HS256, claims: `{ typ:"sfu", id:"<service_id>" }`) and writes/refreshes the instance in discovery (etcd).
  - API reads instances from etcd when serving JoinVoice. No fallback to origin; returns 503 when no instance exists.
- Dependencies: Webhook (for discovery), etcd (backing store for discovery), optional STUN servers.
 - Config: `webhook_url`, pre-generated `webhook_token` (HS256 JWT), and `service_id` (must match token `id`).

## Webhook (`cmd/webhook`)
- Purpose: Secure integration surface for internal events (currently: SFU discovery heartbeat, attachment finalize).
- Endpoints:
  - `POST /api/v1/webhook/sfu/heartbeat` — body: `{ id, region, url, load }`, header: `X-Webhook-Token: <JWT>`.
  - `POST /api/v1/webhook/attachments/finalize` — updates attachment metadata after upload completes.
- Auth: HS256 JWT in `X-Webhook-Token` with claims `{ typ, id }`; no expiration is required.
- Config: `jwt_secret`, `etcd_endpoints`, `etcd_prefix`, and optional Cassandra cluster for attachments.
- Writes SFU instances into etcd for API discovery; SFU does not talk to etcd directly when webhook is used.
 - Token generation: use `cmd/tools` → `gen-token --type sfu --secret <jwt_secret> [--id <service_id>]` and set the result as SFU `webhook_token`.

## Attachments (`cmd/attachments`)
- Purpose: File upload service for message attachments, avatars, and icons.
- Key features:
  - Upload endpoints with size/type validation and metadata persistence.
  - S3‑compatible storage integration and public URL computation.
  - Emits events (e.g., avatar/icon updates) via NATS.
- Dependencies: Scylla/Cassandra, PostgreSQL, S3‑compatible storage, NATS.

## Indexer (`cmd/indexer`)
- Purpose: Asynchronous search indexing worker.
- Key features:
  - Subscribes to NATS topics for message index, update, and delete events.
  - Writes to OpenSearch for full‑text search.
- Dependencies: NATS, OpenSearch.
