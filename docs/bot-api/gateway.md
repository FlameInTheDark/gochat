# Gateway

The bot gateway is a WebSocket API for realtime bot events.

Default endpoint:

```text
wss://gochat.anticode.dev/bot/ws
```

Local service default:

```text
ws://127.0.0.1:3101/bot/ws
```

The gateway uses the same bot authorization header as REST:

```http
Authorization: Bot <gcb_token>
```

## Connection Lifecycle

1. Open a WebSocket connection to `/bot/ws` with the bot authorization header.
2. Send an identify frame with opcode `1`.
3. Receive a server hello frame with the heartbeat interval and session ID.
4. Receive a `GATEWAY_READY` dispatch with shard assignment state.
5. Send heartbeat frames with opcode `2` at the configured interval.
6. Receive heartbeat acknowledgements with opcode `8`.
7. Process dispatch events with opcode `0`.
8. Reconnect and use REST history if the connection drops.

## Gateway Envelope

All frames are JSON envelopes:

```json
{
  "op": 0,
  "t": 100,
  "d": {}
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `op` | integer | Yes | Gateway opcode |
| `t` | integer | Dispatch only | Event type. Omitted for non-dispatch control frames |
| `d` | object | Yes | Payload for the opcode or event |

The server also accepts `data` as an alias for `d` when reading client frames. New clients should send `d`.

## Opcodes

| Value | Name | Direction | Current bot gateway behavior |
|-------|------|-----------|------------------------------|
| `0` | `DISPATCH` | Server to client | Realtime event payload |
| `1` | `HELLO` | Client to server, then server to client | Client identify; server heartbeat configuration |
| `2` | `HEARTBEAT` | Client to server | Keeps the session and shard lease alive |
| `3` | `PRESENCE_UPDATE` | Client to server | Updates the bot user's visible presence |
| `4` | `GUILD_UPDATE_SUBSCRIPTION` | Reserved | Not handled by bot gateway today |
| `5` | `CHANNEL_SUBSCRIPTION` | Reserved | Not handled by bot gateway today |
| `6` | `PRESENCE_SUBSCRIPTION` | Reserved | Not handled by bot gateway today |
| `7` | `RTC` | Reserved/shared protocol | Not handled by bot gateway today |
| `8` | `HEARTBEAT_ACK` | Server to client | Acknowledges a heartbeat |

The bot gateway server currently handles client opcodes `1`, `2`, and `3`. Unknown client opcodes are ignored.

## Identify

The first client frame must be `HELLO` with shard identity.

```json
{
  "op": 1,
  "d": {
    "shard_id": 0,
    "shard_count": 1,
    "session_id": "optional-existing-session-id",
    "presence": {
      "status": "online",
      "custom_status_text": "Processing queue"
    }
  }
}
```

### Identify Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `shard_id` | integer | Yes | Zero-based shard number |
| `shard_count` | integer | Yes | Total shard count. Must be greater than zero |
| `session_id` | string | No | Optional client-supplied session ID to continue using |
| `presence` | object | No | Initial presence. Defaults to `online` |

Validation:

- `shard_count` must be greater than `0`.
- `shard_id` must be `>= 0` and `< shard_count`.
- All active sessions for one bot must use the same `shard_count`.
- When `shard_count > 1`, only one active connection is allowed for each `{bot_user_id, shard_count, shard_id}`.
- Sending identify again on the same connection closes the connection with a policy violation.

## Server Hello

After successful identify, the server sends opcode `1` back to the client:

```json
{
  "op": 1,
  "d": {
    "heartbeat_interval": 45000,
    "session_id": "9cfd9d62-9b31-4f56-8eb7-b7b9c5a82b15"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `heartbeat_interval` | integer | Heartbeat interval in milliseconds |
| `session_id` | string | Gateway session ID assigned or accepted by the server |

## Ready Dispatch

Immediately after server hello, the gateway sends event type `1`.

```json
{
  "op": 0,
  "t": 1,
  "d": {
    "bot": {
      "id": 1001,
      "name": "ExampleBot",
      "is_bot": true
    },
    "session_id": "9cfd9d62-9b31-4f56-8eb7-b7b9c5a82b15",
    "shard_id": 0,
    "shard_count": 1,
    "guild_ids": [2230469276416868352],
    "dm_channel_ids": [2230469276416868353],
    "group_dm_channel_ids": [],
    "receives_dm_events": true,
    "granted_permissions": {
      "2230469276416868352": 3072
    }
  }
}
```

Ready payload fields are documented in [Payloads](payloads.md#ready).

## Heartbeats

Send a heartbeat every `heartbeat_interval` milliseconds after receiving server hello:

```json
{
  "op": 2,
  "d": {
    "client_time": 1780300000000
  }
}
```

The server responds:

```json
{
  "op": 8,
  "d": {
    "server_time": 1780300000050
  }
}
```

Heartbeat effects:

- Extends the WebSocket read deadline.
- Extends a shard lease when sharding is enabled.
- Touches the registered bot gateway session.
- Touches bot presence TTL if presence is active.

If the server does not receive heartbeat traffic in time, the read deadline expires and the connection closes.

## Presence Update

Bots can update their own visible presence over the gateway:

```json
{
  "op": 3,
  "d": {
    "status": "idle",
    "custom_status_text": "Processing queue"
  }
}
```

Allowed statuses:

| Value | Meaning |
|-------|---------|
| `online` | Bot is online |
| `idle` | Bot is idle |
| `dnd` | Bot is in do-not-disturb mode |
| `offline` | Bot appears offline |

Rules:

- Identify is required before presence update.
- Empty status defaults to `online`.
- Status is lowercased and trimmed.
- `custom_status_text` is trimmed and limited to 255 Unicode characters.
- Invalid status or too-long custom text closes the connection with unsupported data.
- The gateway does not send a direct acknowledgement for presence updates.

## Sharding

Guild dispatch assignment:

```text
guild_id % shard_count == shard_id
```

DM dispatch assignment:

```text
shard_id == 0
```

Ready payloads include only guild IDs and DM/group DM channel IDs assigned to the connecting shard.

When `shard_count = 1`, multiple active unsharded sessions can be connected. Each eligible session can receive routed events.

When `shard_count > 1`, duplicate `{bot_user_id, shard_count, shard_id}` connections are rejected.

## Event Delivery

The bot gateway process subscribes to an internal instance delivery subject. It sends only events targeted to the current session ID.

Delivery properties:

- Live delivery only.
- At most once.
- No sequence numbers.
- No resume replay.
- No explicit subscribe operation is needed from bot clients.
- Missed message state should be recovered through `GET /message/channel/{channel_id}`.

## Close Conditions

The gateway can close or drop the connection when:

| Condition | Result |
|-----------|--------|
| Missing or invalid auth during upgrade | HTTP `401` before WebSocket upgrade |
| Missing bot principal after upgrade | WebSocket policy violation |
| Invalid identify JSON | Unsupported data close |
| Invalid shard info | Policy violation close |
| Duplicate identify | Policy violation close |
| Shard count mismatch | Policy violation close |
| Duplicate sharded connection | Policy violation close |
| Invalid presence update | Unsupported data close |
| Heartbeats stop | Read timeout and disconnect |
| Internal subscription or identify failure | Internal server error close |

## Client Responsibilities

- Identify immediately after connecting.
- Heartbeat on schedule.
- Reconnect with backoff after disconnects.
- Fetch REST history after reconnecting if the bot needs gap recovery.
- Keep handler code fast; slow processing should be moved to worker goroutines or queues.
