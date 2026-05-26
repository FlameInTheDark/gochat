[<- Bots](README.md)

# Gateway

The bot WebSocket gateway is a separate service from the regular client WebSocket gateway.

Deployment config:

```text
botws_config.yaml
botws_config.example.yaml
```

Endpoint:

```text
GET /bot/ws
```

Authentication uses the bot runtime header:

```http
Authorization: Bot <token>
```

The gateway uses the common envelope structure:

```json
{
  "op": 0,
  "t": "EventType",
  "d": {}
}
```

## Identify

Identify payload:

```json
{
  "op": 1,
  "d": {
    "shard_id": 0,
    "shard_count": 1,
    "session_id": "optional-client-session-id",
    "presence": {
      "status": "online",
      "custom_status_text": "optional text"
    }
  }
}
```

After identify, the gateway returns a heartbeat interval and a ready dispatch containing:

- bot profile
- session id
- shard id
- shard count
- installed guild ids assigned to that shard
- direct/group DM channel ids visible at identify time
- whether this shard owns direct-message notifications
- granted permission map per guild

## Presence Update

Bots can update their own visible presence programmatically over the gateway:

```json
{
  "op": 3,
  "d": {
    "status": "idle",
    "custom_status_text": "Processing queue"
  }
}
```

Allowed statuses are `online`, `idle`, `dnd`, and `offline`. Presence is stored per active bot session and then aggregated for users who can see the bot.

## Sharding

Bots can connect multiple gateway instances by shard. Guild assignment is deterministic:

```text
guild_id % shard_count == shard_id
```

When `shard_count > 1`, only one active connection is allowed for each `{bot_id, shard_count, shard_id}`. The gateway enforces this with Redis leases. With `shard_count = 1`, multiple active instances are allowed and all receive eligible events.
