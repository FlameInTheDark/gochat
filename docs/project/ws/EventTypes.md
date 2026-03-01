[<- Documentation](../README.md) - [WebSocket Events](README.md)

# Event Types

When the server sends a **Dispatch** message (`op: 0`), the `t` field identifies the event type. This page lists all event type values, their payloads, and which NATS topic delivers them.

> [!NOTE]
> All events on this page (100–406) are delivered over the **Gateway WebSocket** (`/subscribe`). Voice/WebRTC signaling events (500–515) are exchanged over the separate **SFU WebSocket** (`/signal`) — see [SFU Protocol](../voice/SFUProtocol.md). Only a few voice-related control events (509, 512, 513) pass through the Gateway WS as noted in the [RTC Events](#rtc-events-500515-gateway-ws-only) section.

---

## Message Events (100–102)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 100 | Message Create | `channel.{channelId}` | New message posted |
| 101 | Message Update | `channel.{channelId}` | Message edited |
| 102 | Message Delete | `channel.{channelId}` | Message removed |

**Payload (t=100, Message Create):**
```json
{
  "guild_id": 2226022078304223200,
  "message": {
    "id": 2228801793842741200,
    "channel_id": 2226022078341972000,
    "author": {
      "id": 2226021950625415200,
      "name": "FlameInTheDark",
      "discriminator": "flameinthedark"
    },
    "content": "Hello"
  }
}
```

---

## Guild Events (103–105)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 103 | Guild Create | `guild.{guildId}` | New guild created |
| 104 | Guild Update | `guild.{guildId}` | Guild properties changed |
| 105 | Guild Delete | `guild.{guildId}` | Guild deleted |

**Payload (t=104, Guild Update):**
```json
{
  "guild": {
    "id": 2226022078304223200,
    "name": "My Server",
    "icon": "..."
  }
}
```

---

## Channel Events (106–109)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 106 | Channel Create | `guild.{guildId}` | Channel created in guild |
| 107 | Channel Update | `guild.{guildId}` | Channel properties changed |
| 108 | Channel Order Update | `guild.{guildId}` | Channel ordering/position changed |
| 109 | Channel Delete | `guild.{guildId}` | Channel deleted |

---

## Guild Role Events (110–112)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 110 | Guild Role Create | `guild.{guildId}` | Role created |
| 111 | Guild Role Update | `guild.{guildId}` | Role permissions/properties changed |
| 112 | Guild Role Delete | `guild.{guildId}` | Role removed |

---

## Thread Events (113–115)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 113 | Thread Create | `guild.{guildId}` | Thread created |
| 114 | Thread Update | `guild.{guildId}` | Thread properties changed |
| 115 | Thread Delete | `guild.{guildId}` | Thread deleted |

---

## Guild Member Events (200–206)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 200 | Guild Member Added | `guild.{guildId}` | User joined guild |
| 201 | Guild Member Update | `guild.{guildId}` | Member nickname/properties changed |
| 202 | Guild Member Remove | `guild.{guildId}` | Member left/kicked from guild |
| 203 | Guild Member Role Added | `guild.{guildId}` | Role assigned to member |
| 204 | Guild Member Role Removed | `guild.{guildId}` | Role removed from member |
| 205 | Guild Member Join Voice | `guild.{guildId}` | Member joined a voice channel |
| 206 | Guild Member Leave Voice | `guild.{guildId}` | Member left a voice channel |

**Payload (t=205, Guild Member Join Voice):**
```json
{
  "guild_id": 2226022078304223200,
  "user_id": 2226021950625415200,
  "channel_id": 2230469276416868352
}
```

---

## Voice Region Events (208)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 208 | Guild Voice Region Changing | `guild.{guildId}` | SFU region migration starting |

**Payload:**
```json
{
  "channel_id": 2230469276416868352,
  "region": "eu-west",
  "delay_ms": 3000
}
```

**Client action:** Display countdown or preparation notice. After `delay_ms` milliseconds, a `VoiceRebind` (t=513) event will follow for clients in that channel. Clients should prepare to call `JoinVoice` API again after the rebind fires.

---

## Channel Message Events (300–302)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 300 | Guild Channel Message | `channel.{channelId}` | Activity notification for a guild channel |
| 301 | Channel Typing Event | `channel.{channelId}` | User started typing |
| 302 | Mention | `user.{userId}` | User was mentioned |

**Payload (t=301, Channel Typing):**
```json
{
  "channel_id": 2226022078341972000,
  "user_id": 2226021950625415200
}
```

---

## User Events (400–406)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 400 | User Update Read State | `user.{userId}` | Read state marker advanced |
| 401 | User Update Settings | `user.{userId}` | User settings changed |
| 402 | Incoming Friend Request | `user.{userId}` | Friend request received |
| 403 | Friend Added | `user.{userId}` | Friend accepted |
| 404 | Friend Removed | `user.{userId}` | Friend removed |
| 405 | User DM Message | `user.{userId}` | New DM message |
| 406 | User Update | `user.{userId}` | User profile changed |

**Payload (t=400, Read State Update):**
```json
{
  "channel_id": 2226022078341972000,
  "message_id": 2228801793842741200
}
```

---

## Presence Events (OP 3 Dispatch)

Presence updates are dispatched with `op: 3` (not `op: 0`). They are delivered via NATS topic `presence.user.{userId}` to clients that have subscribed via [OP 6](EventMessageStructure.md#op-6--presence-subscription).

**Payload:**
```json
{
  "user_id": 2226021950625415200,
  "status": "online",
  "custom_status_text": "Coding...",
  "since": 1700000000,
  "voice_channel_id": 2230469276416868352,
  "client_status": {
    "web": "online"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `user_id` | int64 | User whose presence changed |
| `status` | string | `"online"`, `"idle"`, `"dnd"`, or `"offline"` |
| `custom_status_text` | string | Free-text status message |
| `since` | int64 | Unix timestamp when status was set |
| `voice_channel_id` | int64 | Voice channel user is in (omitted if none) |
| `client_status` | map | Per-platform status (optional) |

---

## RTC Events (500–515, Gateway WS Only)

> [!IMPORTANT]
> The full RTC signaling protocol (Join, Offer, Answer, Candidate, Speaking, Mute, Deafen, Kick, Block — events 500–515) is handled over the **separate SFU WebSocket** connection (`/signal` on port 3300). See [SFU Protocol](../voice/SFUProtocol.md) for that protocol.
>
> Only the following **3 voice control events** pass through the **Gateway WS** (`/subscribe`):

| Type | Name | NATS Topic | Direction | Description |
|------|------|------------|-----------|-------------|
| 509 | RTC Binding Alive | — | Client → Gateway | Keep `voice:route:{channelId}` TTL alive in Redis |
| 512 | RTC Moved | `user.{userId}` | Gateway → Client | Admin moved user to another channel; includes new SFU URL |
| 513 | RTC Server Rebind | `guild.{guildId}` | Gateway → Client | SFU route changed (region migration); reconnect required |

**Payload (t=513, VoiceRebind):**
```json
{
  "channel": 2230469276416868352,
  "jitter_ms": 3000
}
```

**Client action:** Wait `rand(0, jitter_ms)` milliseconds, then call API `JoinVoice` for that channel to get a fresh `sfu_url` and `sfu_token`. During a region migration, the next `JoinVoice` returns a 5-minute JWT (instead of standard 2-minute) while `voice:rebind:{channelId}` remains in Redis.

**Payload (t=512, VoiceMove):**
```json
{
  "channel": 999,
  "sfu_url": "wss://sfu.gochat.io/signal",
  "sfu_token": "eyJhbGci..."
}
```

**Client action:** Disconnect from current SFU, connect to the new `sfu_url`, and send Join with the new token.

For the full SFU WebSocket protocol (events 500–515), see [SFU Protocol](../voice/SFUProtocol.md).
