[<- Documentation](../README.md) - [WebSocket Events](README.md)

# Subscription Model

The WebSocket gateway uses a shared NATS subscription model. Rather than each connection creating its own NATS subscriptions, a centralized **Hub** manages one NATS subscription per unique topic and fans messages out to all local connections in-memory.

---

## Architecture

```
┌─────────────────────────────────────────────┐
│  WS Server Process                          │
│                                             │
│  ┌─────┐  ┌─────┐  ┌─────┐                 │
│  │Conn1│  │Conn2│  │Conn3│  ... (clients)   │
│  └──┬──┘  └──┬──┘  └──┬──┘                 │
│     │        │        │                     │
│  ┌──▼────────▼────────▼──┐                  │
│  │      Subscriber       │  per-connection  │
│  │  (key → topic map)    │  subscription    │
│  └──────────┬────────────┘  tracking        │
│             │                               │
│  ┌──────────▼────────────┐                  │
│  │         Hub           │  shared NATS     │
│  │  (topic → [conns])    │  subscription    │
│  └──────────┬────────────┘  management      │
│             │                               │
│  ┌──────────▼────────────┐                  │
│  │     NATS Client       │                  │
│  └───────────────────────┘                  │
└─────────────────────────────────────────────┘
```

---

## Topic Types

| Topic Pattern | Subscribed When | Events Delivered |
|---------------|----------------|-----------------|
| `user.{userId}` | Hello (automatic) | Read state, settings, friend events, DMs, user updates, VoiceMove |
| `guild.{guildId}` | Hello (automatic for all guilds) + OP 5 | Guild/channel/role/member/voice events |
| `channel.{channelId}` | OP 5 Channel Subscription | Messages, typing indicators, channel-specific events |
| `presence.user.{userId}` | OP 6 Presence Subscription | Presence status changes for watched users |

---

## Subscription Operations

### Automatic (on Hello)

After authentication, the server subscribes the connection to:
1. `user.{userId}` — personal events.
2. `guild.{guildId}` for every guild the user is a member of.

These are registered using the guild ID as both the key and topic identifier.

### Manual (OP 5 — Channel Subscription)

Client subscribes to specific channels for real-time typing and message events:

```json
{ "op": 5, "d": { "channel": 123, "guilds": [456] } }
```

**Channel permission check order:**
1. Is it a guild channel? → Check `PermServerViewChannels`.
2. Is it a DM? → Check user is a participant.
3. Is it a Group DM? → Check user is a participant.
4. None match → subscription rejected (silent, logged server-side).

For guild IDs in the `guilds` array, the server verifies the user is a member before subscribing.

### Manual (OP 6 — Presence Subscription)

Client manages a set of user IDs to watch for presence changes:

| Operation | Behavior |
|-----------|----------|
| `clear: true` | Unsubscribe from all presence topics |
| `set: [ids]` | Replace entire watch list (unsubscribe all, then subscribe to new set) |
| `add: [ids]` | Add users to watch list (skip if already watching) |
| `remove: [ids]` | Stop watching specific users |

On every `add` or `set` operation, the server immediately sends a **presence snapshot** (current cached status) for each newly-watched user, so the client doesn't have to wait for the next status change.

---

## Hub Internals

### Registration

```
Hub.Register(conn, topic):
  1. Lock hub
  2. If topic exists → add conn to topic's conn set → done
  3. If topic is new:
     a. Create topicEntry with conn
     b. Create shared NATS subscription
     c. NATS callback: for each conn in set → conn.Send(msg.Data)
```

### Unregistration

```
Hub.Unregister(conn, topic):
  1. Remove conn from topic's conn set
  2. If set is now empty:
     a. Delete topic from hub
     b. Unsubscribe shared NATS subscription
```

### Connection Close

```
Hub.UnregisterAll(conn):
  1. Scan all topics for this conn
  2. Call Unregister for each
```

### Key Properties

- **Non-blocking delivery:** `conn.Send()` uses a buffered channel with `select/default` — if the buffer is full, the message is dropped. The heartbeat/ping timeout will eventually evict dead connections.
- **Deduplication:** The Subscriber layer prevents duplicate subscriptions. If `Subscribe("channel", "channel.123")` is called again with the same key and topic, it's a no-op.
- **Thread safety:** The Hub uses `sync.RWMutex`; each topic entry has its own `sync.RWMutex` for the connection set.
