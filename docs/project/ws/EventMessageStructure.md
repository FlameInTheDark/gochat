[<- Documentation](../README.md) - [WebSocket Events](README.md)

# Event Message Structure

> [!NOTE]
> This page documents the **Gateway WebSocket** (`/subscribe` on port 3100) used for chat, presence, and subscriptions. The **SFU WebSocket** (`/signal` on port 3300) is a separate connection used exclusively for voice/WebRTC signaling — see [SFU Protocol](../voice/SFUProtocol.md) for its message format.

All messages exchanged over the Gateway WebSocket share a single JSON envelope:

```json
{
  "op": 0,
  "d":  { ... },
  "t":  100
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `op` | int | ✅ | Operation code — determines how the message is routed |
| `d` | object / json | ✅ | Payload data (structure varies by op + t) |
| `t` | int | ❌ | Event type — only meaningful when `op = 0` (Dispatch) or `op = 7` (RTC). Omitted for control ops |

> **Note:** The server accepts both `"d"` and `"data"` as the payload key (for client convenience).

---

## OP Codes (Client → Server)

| OP | Name | Description | Payload |
|----|------|-------------|---------|
| 1 | **Hello** | First message after WS connect; authenticates the session | See [Hello](#op-1--hello) |
| 2 | **Heartbeat** | Keep-alive sent on the server-defined interval | See [Heartbeat](#op-2--heartbeat) |
| 3 | **Presence Update** | Update own presence status | See [Presence Update](#op-3--presence-update) |
| 5 | **Channel Subscription** | Subscribe to channel and/or guild events | See [Channel Subscription](#op-5--channel-subscription) |
| 6 | **Presence Subscription** | Manage which users' presence you track | See [Presence Subscription](#op-6--presence-subscription) |
| 7 | **RTC** | WebRTC/voice signaling (send to SFU or keep-alive) | See [RTC](#op-7--rtc) |

## OP Codes (Server → Client)

| OP | Name | Description |
|----|------|-------------|
| 0 | **Dispatch** | Server-push event (message created, guild updated, etc.). Always includes `t` field |
| 1 | **Hello Reply** | Response to Hello — contains heartbeat interval and session ID |
| 3 | **Presence Update** | Dispatched presence snapshot for a subscribed user |
| 7 | **RTC** | Voice/WebRTC events (offer, candidate, speaking, mute, kick, etc.) |

---

## Client → Server Payloads

### OP 1 — Hello

Sent immediately after WebSocket upgrade. If not received within **5 seconds**, the server closes the connection.

```json
{
  "op": 1,
  "d": {
    "token": "eyJhbGci...",
    "heartbeat_session_id": "optional-previous-session-id"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `token` | string | ✅ | JWT access token (from `/auth/login` or `/auth/refresh`) |
| `heartbeat_session_id` | string | ❌ | Session ID from a previous connection to resume presence session |

**Server response (OP 1):**
```json
{
  "op": 1,
  "d": {
    "heartbeat_interval": 30000,
    "session_id": "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `heartbeat_interval` | int64 | Heartbeat interval in **milliseconds** |
| `session_id` | string | UUID-style session ID; pass as `heartbeat_session_id` on reconnect |

On success the server also:
1. Fetches user profile and guild list in parallel.
2. Subscribes the connection to the personal `user.{userId}` NATS topic.
3. Subscribes to all user's guilds (`guild.{guildId}`).
4. Starts the heartbeat timeout timer.

---

### OP 2 — Heartbeat

Must be sent periodically per the `heartbeat_interval` from Hello. Missing heartbeats (with 10s grace) cause disconnection.

```json
{
  "op": 2,
  "d": {
    "e": 42
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `e` | int64 | Last event ID acknowledged by the client (monotonically increasing) |

The server resets the heartbeat timer only if `e >= lastEventId`. On each heartbeat, the server also refreshes the session's presence TTL in Redis (throttled to at most once per 10 seconds).

---

### OP 3 — Presence Update

Update own presence status. Presence is **not** auto-set on Hello — the client must explicitly send this op.

```json
{
  "op": 3,
  "d": {
    "status": "online",
    "platform": "web",
    "custom_status_text": "Coding...",
    "voice_channel_id": 2230469276416868352
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `status` | string | ✅ | `"online"`, `"idle"`, `"dnd"`, or `"offline"` (invisible mode) |
| `platform` | string | ❌ | `"web"`, `"mobile"`, `"desktop"` — informational only |
| `custom_status_text` | string | ❌ | Free-text status message |
| `voice_channel_id` | int64 | ❌ | Set to a channel ID to indicate voice presence; set to `0` to clear |

**Behavior:**
- `"offline"` sets a global override — the user appears offline to all watchers, even though the session is active.
- Any other valid status clears the offline override and upserts the session.
- The aggregated presence is published to NATS (`presence.user.{userId}`) for all presence subscribers.

---

### OP 5 — Channel Subscription

Subscribe to channel-specific events (typing, messages) and/or additional guild topics.

```json
{
  "op": 5,
  "d": {
    "channel": 2226022078341972000,
    "guilds": [2226022078304223200]
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `channel` | int64 | ❌ | Channel ID to subscribe to (guild channel, DM, or group DM) |
| `guilds` | int64[] | ❌ | Additional guild IDs to subscribe to |

**Permission checks for channel subscriptions:**
1. **Guild channel** — must have `PermServerViewChannels` on that channel.
2. **DM channel** — must be a participant.
3. **Group DM** — must be a participant.

If none match, the subscription is silently rejected (logged server-side).

---

### OP 6 — Presence Subscription

Manage which users' presence updates you receive in real-time.

```json
{
  "op": 6,
  "d": {
    "add": [123456789, 987654321],
    "remove": [111111111],
    "set": [222222222, 333333333],
    "clear": false
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `add` | int64[] | User IDs to start watching |
| `remove` | int64[] | User IDs to stop watching |
| `set` | int64[] | Replace the entire watch list with these IDs |
| `clear` | bool | If `true`, unsubscribe from all presence currently watched |

**Processing order:** `clear` → `set` → `add` → `remove`.

When a user ID is added (via `add` or `set`), the server immediately sends a **presence snapshot** for that user so the client has the current status without waiting for a change.

---

### OP 7 — RTC (Gateway WS — limited)

Used for voice-related **control** over the Gateway WS connection. Only `t=509` (RTCBindingAlive) is sent **client → Gateway**. The full RTC signaling (join, offer, answer, candidate, speaking, mute, etc.) happens on the **separate SFU WebSocket** (`/signal` on port 3300) — see [SFU Protocol](../voice/SFUProtocol.md).

**RTCBindingAlive (t=509) — Keep voice route alive:**
```json
{
  "op": 7,
  "t": 509,
  "d": {
    "channel": 2230469276416868352
  }
}
```

This refreshes `voice:route:{channelId}` TTL (60s) in Redis and updates the session's voice channel presence. Clients should send this periodically while in a voice channel.
