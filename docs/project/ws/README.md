[<- Documentation](../README.md)

# WebSocket Gateway

The WebSocket gateway (`cmd/ws`) is the real-time event delivery layer. Clients maintain a single persistent WebSocket connection through which they receive all server-push events (messages, guild updates, presence, voice notifications) and send control commands (heartbeats, subscriptions, presence updates).

## Endpoint

```
GET /subscribe
```

Traefik routes `/ws` → `ws:3100` (with `StripPrefix`), so the external URL is `wss://example.com/ws/subscribe`.

> [!IMPORTANT]
> **Two separate WebSocket connections exist in gochat:**
>
> | Connection | Service | Endpoint | Purpose |
> |------------|---------|----------|---------|
> | **Gateway WS** | `cmd/ws` (port 3100) | `/subscribe` | Chat events, presence, subscriptions, heartbeat |
> | **Voice SFU WS** | `cmd/sfu` (port 3300) | `/signal` | WebRTC signaling, media negotiation, speaking indicators |
>
> These are **independent connections** — a client has one Gateway WS open at all times, and opens a separate SFU WS only when joining a voice channel. This documentation covers the **Gateway WS**. For the SFU WS protocol, see [Voice Protocol](../voice/) docs.

## Features

- **zlib-stream compression** — append `?compress=zlib-stream` to receive binary frames compressed with zlib (best-speed level).
- **Connection Hub** — one shared NATS subscription per topic across all local connections, with in-memory fan-out.
- **Writer pump** — async outbound channel (256 buffer) with non-blocking drop for slow consumers.
- **Prometheus metrics** — `gochat_ws_active_clients` gauge; scraped at `/metrics`.

## Documentation

- [Connection Lifecycle](ConnectionLifecycle.md) — Authentication, heartbeat, and disconnect flows
- [Event Message Structure](EventMessageStructure.md) — Envelope format, OP codes, and client-sent payloads
- [Event Types](EventTypes.md) — Complete list of `t` (event type) values and their payloads
- [Subscription Model](SubscriptionModel.md) — How guild, channel, and presence subscriptions work