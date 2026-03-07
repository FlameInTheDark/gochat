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

**Payload (t=101, Message Update):**
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
    "content": "Hello, edited!",
    "updated_at": "2026-01-15T10:30:00Z"
  }
}
```

**Payload (t=102, Message Delete):**
```json
{
  "guild_id": 2226022078304223200,
  "channel_id": 2226022078341972000,
  "message_id": 2228801793842741200
}
```

---

## Guild Events (103–105)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 103 | Guild Create | `guild.{guildId}` | New guild created |
| 104 | Guild Update | `guild.{guildId}` | Guild properties changed |
| 105 | Guild Delete | `guild.{guildId}` | Guild deleted |

**Payload (t=103, Guild Create):**
```json
{
  "guild": {
    "id": 2226022078304223200,
    "name": "My Server",
    "icon": null,
    "owner": 2226021950625415200,
    "public": false,
    "permissions": 7927905
  }
}
```

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

**Payload (t=105, Guild Delete):**
```json
{
  "guild_id": 2226022078304223200
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

**Payload (t=106, Channel Create):**
```json
{
  "guild_id": 2226022078304223200,
  "channel": {
    "id": 2226022078341972000,
    "type": 0,
    "guild_id": 2226022078304223200,
    "name": "general",
    "position": 0,
    "topic": null,
    "private": false,
    "last_message_id": 0,
    "created_at": "2026-01-15T10:00:00Z"
  }
}
```

**Payload (t=107, Channel Update):**
```json
{
  "guild_id": 2226022078304223200,
  "channel": {
    "id": 2226022078341972000,
    "type": 0,
    "guild_id": 2226022078304223200,
    "name": "general-chat",
    "position": 0,
    "topic": "General discussion",
    "private": false,
    "last_message_id": 2228801793842741200,
    "created_at": "2026-01-15T10:00:00Z"
  }
}
```

**Payload (t=108, Channel Order Update):**
```json
{
  "guild_id": 2226022078304223200,
  "channels": [
    { "id": 2226022078341972000, "position": 0 },
    { "id": 2226022078341972001, "position": 1 },
    { "id": 2226022078341972002, "position": 2 }
  ]
}
```

**Payload (t=109, Channel Delete):**
```json
{
  "guild_id": 2226022078304223200,
  "channel_type": 0,
  "channel_id": 2226022078341972000
}
```

---

## Guild Role Events (110–112)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 110 | Guild Role Create | `guild.{guildId}` | Role created |
| 111 | Guild Role Update | `guild.{guildId}` | Role permissions/properties changed |
| 112 | Guild Role Delete | `guild.{guildId}` | Role removed |

**Payload (t=110, Guild Role Create):**
```json
{
  "role": {
    "id": 2230469276416868352,
    "guild_id": 2226022078304223200,
    "name": "Moderator",
    "color": 3447003,
    "permissions": 274877910022
  }
}
```

**Payload (t=111, Guild Role Update):**
```json
{
  "role": {
    "id": 2230469276416868352,
    "guild_id": 2226022078304223200,
    "name": "Super Moderator",
    "color": 15158332,
    "permissions": 274877910022
  }
}
```

**Payload (t=112, Guild Role Delete):**
```json
{
  "guild_id": 2226022078304223200,
  "role_id": 2230469276416868352
}
```

---

## Thread Events (113–115)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 113 | Thread Create | `guild.{guildId}` | Thread created |
| 114 | Thread Update | `guild.{guildId}` | Thread properties changed |
| 115 | Thread Delete | `guild.{guildId}` | Thread deleted |

**Payload (t=113, Thread Create):**
```json
{
  "guild_id": 2226022078304223200,
  "channel": {
    "id": 2226022078341973000,
    "type": 2,
    "guild_id": 2226022078304223200,
    "name": "thread-discussion",
    "parent_id": 2226022078341972000,
    "position": 0,
    "topic": null,
    "private": false,
    "last_message_id": 0,
    "created_at": "2026-01-15T10:30:00Z"
  }
}
```

**Payload (t=114, Thread Update):**
```json
{
  "guild_id": 2226022078304223200,
  "channel": {
    "id": 2226022078341973000,
    "type": 2,
    "guild_id": 2226022078304223200,
    "name": "updated-thread-name",
    "parent_id": 2226022078341972000,
    "position": 0,
    "topic": "Updated topic",
    "private": false,
    "last_message_id": 2228801793842741300,
    "created_at": "2026-01-15T10:30:00Z"
  }
}
```

**Payload (t=115, Thread Delete):**
```json
{
  "guild_id": 2226022078304223200,
  "channel_type": 2,
  "channel_id": 2226022078341973000
}
```

---

## Guild Member Events (200–209)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 200 | Guild Member Added | `guild.{guildId}` | User joined guild |
| 201 | Guild Member Update | `guild.{guildId}` | Member nickname/properties changed |
| 202 | Guild Member Remove | `guild.{guildId}` | Member left/kicked from guild |
| 203 | Guild Member Role Added | `guild.{guildId}` | Role assigned to member |
| 204 | Guild Member Role Removed | `guild.{guildId}` | Role removed from member |
| 205 | Guild Member Join Voice | `guild.{guildId}` | Member joined a voice channel |
| 206 | Guild Member Leave Voice | `guild.{guildId}` | Member left a voice channel |
| 209 | Voice State Update | `guild.{guildId}` | User's mute/deafen status changed in voice channel |

**Payload (t=200, Guild Member Added):**
```json
{
  "guild_id": 2226022078304223200,
  "user_id": 2226021950625415200,
  "member": {
    "user": {
      "id": 2226021950625415200,
      "name": "NewUser",
      "discriminator": "newuser",
      "avatar": null
    },
    "username": "FancyNickname",
    "avatar": null,
    "join_at": "2026-01-15T10:00:00Z",
    "roles": []
  }
}
```

**Payload (t=201, Guild Member Update):**
```json
{
  "guild_id": 2226022078304223200,
  "member": {
    "user": {
      "id": 2226021950625415200,
      "name": "NewUser",
      "discriminator": "newuser",
      "avatar": null
    },
    "username": "UpdatedNickname",
    "avatar": null,
    "join_at": "2026-01-15T10:00:00Z",
    "roles": [2230469276416868352]
  }
}
```

**Payload (t=202, Guild Member Remove):**
```json
{
  "guild_id": 2226022078304223200,
  "user_id": 2226021950625415200
}
```

**Payload (t=203, Guild Member Role Added):**
```json
{
  "guild_id": 2226022078304223200,
  "role_id": 2230469276416868352,
  "user_id": 2226021950625415200
}
```

**Payload (t=204, Guild Member Role Removed):**
```json
{
  "guild_id": 2226022078304223200,
  "role_id": 2230469276416868352,
  "user_id": 2226021950625415200
}
```

**Payload (t=205, Guild Member Join Voice):**
```json
{
  "guild_id": 2226022078304223200,
  "user_id": 2226021950625415200,
  "channel_id": 2230469276416868352
}
```

**Payload (t=206, Guild Member Leave Voice):**
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

## Voice State Events (209)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 209 | Voice State Update | `guild.{guildId}` | User's mute/deafen status changed in voice channel |

**Payload (t=209, Voice State Update):**
```json
{
  "guild_id": 2226022078304223200,
  "user_id": 2226021950625415200,
  "channel_id": 2230469276416868352,
  "mute": true,
  "deafen": false
}
```

**Notes:**
- This event is sent when a user mutes/unmutes themselves or is server-muted/deafened
- The `mute` field indicates if the user is muted (cannot speak)
- The `deafen` field indicates if the user is deafened (cannot hear others)
- This event is broadcast to all guild members, not just those in the voice channel

---

## Channel Message Events (300–302)

| Type | Name | NATS Topic | Description |
|------|------|------------|-------------|
| 300 | Guild Channel Message | `channel.{channelId}` | Activity notification for a guild channel |
| 301 | Channel Typing Event | `channel.{channelId}` | User started typing |
| 302 | Mention | `user.{userId}` | User was mentioned |

**Payload (t=300, Guild Channel Message):**
```json
{
  "guild_id": 2226022078304223200,
  "channel_id": 2226022078341972000,
  "message_id": 2228801793842741200
}
```

**Payload (t=301, Channel Typing):**
```json
{
  "channel_id": 2226022078341972000,
  "user_id": 2226021950625415200
}
```

**Payload (t=302, Mention):**
```json
{
  "guild_id": 2226022078304223200,
  "channel_id": 2226022078341972000,
  "message_id": 2228801793842741200,
  "author_id": 2226021950625415201,
  "type": 0
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

**Payload (t=401, User Update Settings):**
```json
{
  "settings": {
    "theme": "dark",
    "locale": "en-US",
    "notifications": {
      "enable": true,
      "sound": true
    }
  }
}
```

**Payload (t=402, Incoming Friend Request):**
```json
{
  "from": {
    "id": 2226021950625415200,
    "name": "FriendRequester",
    "discriminator": "frienduser",
    "avatar": null
  }
}
```

**Payload (t=403, Friend Added):**
```json
{
  "friend": {
    "id": 2226021950625415200,
    "name": "NewFriend",
    "discriminator": "newfriend",
    "avatar": null
  }
}
```

**Payload (t=404, Friend Removed):**
```json
{
  "friend": {
    "id": 2226021950625415200,
    "name": "OldFriend",
    "discriminator": "oldfriend",
    "avatar": null
  }
}
```

**Payload (t=405, User DM Message):**
```json
{
  "channel_id": 2226022078341974000,
  "message_id": 2228801793842741300,
  "from": {
    "id": 2226021950625415200,
    "name": "SenderName",
    "discriminator": "senderuser",
    "avatar": null
  }
}
```

**Payload (t=406, User Update):**
```json
{
  "user": {
    "id": 2226021950625415200,
    "name": "UpdatedName",
    "discriminator": "updatedname",
    "avatar": null
  }
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
  "mute": false,
  "deafen": false,
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
| `mute` | bool | Whether the user is muted in voice (only present if in voice channel) |
| `deafen` | bool | Whether the user is deafened in voice (only present if in voice channel) |
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
