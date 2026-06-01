# Bot API Overview

The GoChat Bot API lets bot users connect to GoChat as first-class bot accounts. A bot can use HTTP for explicit actions and a WebSocket gateway for realtime events.

The current public bot runtime has three parts:

| Part | Protocol | What it does |
|------|----------|--------------|
| REST API | HTTPS JSON | Bot identity, installed guilds, visible channels, messages, read state, typing, reactions |
| Gateway | WebSocket JSON | Shard identify, ready payload, heartbeat, realtime event dispatch, bot presence updates |
| Go client | Go module | Convenience library for REST and gateway use |

Default production endpoints:

```text
REST:    https://gochat.anticode.dev/bot/api/v1
Gateway: wss://gochat.anticode.dev/bot/ws
```

Local services can use different ports. The Go client can override either the whole service URL or REST/gateway URLs separately.

## Concepts

### Bot User

A bot is represented by a normal GoChat user record with bot metadata and token records attached to it. Bot tokens are generated with the `gcb_` prefix and must be sent as a bot authorization header.

### Guild Installation

Bots are installed into guilds. Runtime guild access is limited to installed guilds. Channel access is then calculated from normal guild roles/channel overrides and capped by the permission grant stored on the installation.

### Gateway Session

A bot gateway connection is an identified session. The client sends shard information, the server returns heartbeat configuration, and then sends a `GATEWAY_READY` dispatch containing the bot account and the guilds/channels assigned to that shard.

### Sharding

A bot may open multiple gateway connections as shards. Guild dispatch is assigned by:

```text
guild_id % shard_count == shard_id
```

Direct-message notifications are owned by shard `0`.

### Delivery Model

Bot gateway delivery is live and at most once. Bots should use REST history endpoints to recover missed message state after reconnecting.

## Documentation Structure

- [Quick Start](quickstart.md): build and run a minimal bot.
- [Authentication](authentication.md): token format, headers, failure modes.
- [REST API](rest.md): every current bot REST endpoint.
- [Gateway](gateway.md): WebSocket lifecycle, opcodes, sharding, presence.
- [Events](events.md): event envelope, event type table, payload examples.
- [Payloads](payloads.md): shared JSON object schemas and limits.
- [Go Client](go-client.md): using `clients/bot/goclient`.
- [Operations](operations.md): deployment, routing, recovery, troubleshooting.
