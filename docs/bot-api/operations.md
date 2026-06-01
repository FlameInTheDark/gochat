# Operations

This page is for operators running the bot runtime services.

## Services

| Service | Binary/package | Purpose |
|---------|----------------|---------|
| Bot REST API | `cmd/botapi` | Authenticated bot HTTP API |
| Bot Gateway | `cmd/botws` | Bot WebSocket gateway |
| Bot Router | `cmd/botrouter` | Routes internal bot events to connected bot gateway sessions |

## Configuration Files

| Service | Example config |
|---------|----------------|
| Bot REST API | `botapi_config.example.yaml` |
| Bot Gateway | `botws_config.example.yaml` |
| Bot Router | `botrouter_config.example.yaml` |

Local configs without `.example` exist for development.

## Runtime Dependencies

| Dependency | Used by | Purpose |
|------------|---------|---------|
| PostgreSQL | All bot services | Bot records, tokens, guild installs, users, channels, permissions |
| ScyllaDB/Cassandra | Bot REST API | Messages, read states, reactions |
| KeyDB/Redis | Bot REST API, bot gateway, router | Idempotency, gateway sessions, shard leases, router leases, presence |
| NATS | Bot REST API, bot gateway, router | Realtime event bus |
| Bot NATS | Bot REST API, bot gateway, router | Dedicated bot event routing bus |

## Network Endpoints

| Service | Default local address | Public path |
|---------|-----------------------|-------------|
| Bot REST API | `:3102` by config | `/bot/api/v1` |
| Bot Gateway | `:3101` by config | `/bot/ws` |

The public deployment can terminate TLS and route both paths under:

```text
https://gochat.anticode.dev
```

## Event Routing Pipeline

1. A domain handler publishes a normal internal event.
2. Bot-aware publishers also publish to the bot NATS bus.
3. Bot NATS subject encodes partition and target scope.
4. `cmd/botrouter` owns event partitions through Redis leases.
5. Router loads installed bots for the guild or bot user for a DM.
6. Router evaluates current permissions and shard targeting.
7. Router publishes a delivery envelope to `bot.deliver.instance.{instanceID}`.
8. `cmd/botws` forwards the event only to matching session IDs.

Bot event subjects:

```text
bot.event.p.{partition}.guild.{guildID}
bot.event.p.{partition}.guild.{guildID}.channel.{channelID}
bot.event.p.{partition}.user.{botUserID}.dm
```

Bot gateway delivery subject:

```text
bot.deliver.instance.{instanceID}
```

## Gateway Session Registry

Identified sessions are stored in Redis with:

- `session_id`
- `bot_user_id`
- `instance_id`
- `shard_id`
- `shard_count`
- whether the session receives DM events

The gateway touches the session on heartbeat. When the connection closes, the gateway unregisters the session and clears the bot session presence.

## Shard Leases

When `shard_count > 1`, the gateway uses a Redis lease key:

```text
botgw:lease:{botUserID}:{shardCount}:{shardID}
```

Only one active connection may hold a lease for a sharded slot. The heartbeat path extends the lease.

## Router Partitions

Bot router consumes partitioned event subjects. Partition ownership is protected by Redis leases so multiple router instances can share work without double-routing the same partition.

Default partition count is `1024` unless configured otherwise.

## Presence

Bot presence is stored per active gateway session using platform `bot`. The gateway:

- Sets initial presence during identify.
- Updates presence on opcode `3`.
- Touches presence TTL on heartbeat.
- Clears session presence on disconnect.

Presence TTL is derived from heartbeat timeout:

```text
ttl_seconds = max(1, heartbeat_timeout_ms * 2 / 1000)
```

## Delivery Guarantees

Bot gateway delivery is:

- Live.
- At most once.
- Targeted to current gateway sessions.
- Not replayed after reconnect.

For recovery, bots should persist their own last processed message IDs and call:

```text
GET /bot/api/v1/message/channel/{channel_id}?from={last_id}&direction=after&limit=100
```

## Health Checks

The bot gateway exposes:

```text
GET /healthz
```

The bot REST API uses the shared server stack and dependency startup probes.

## Troubleshooting

### `401` On REST Or Gateway

Check:

- Header is exactly `Authorization: Bot <token>`.
- Token starts with `gcb_`.
- Bot is not disabled.
- Token record still exists.
- Bot user still exists and is marked as a bot.

### Gateway Closes After Connect

Check:

- The client sends identify opcode `1` immediately.
- `shard_count > 0`.
- `0 <= shard_id < shard_count`.
- All active connections for the bot use the same `shard_count`.
- No duplicate sharded connection is already active.
- Presence status is one of `online`, `idle`, `dnd`, `offline`.

### Bot Does Not Receive Guild Events

Check:

- Bot is installed in the guild.
- Bot gateway session is identified and heartbeating.
- Router service is running and owns bot event partitions.
- Bot has the needed permission grant and channel-level permission.
- Event guild ID targets the bot's shard: `guild_id % shard_count == shard_id`.

### Bot Does Not Receive DM Events

Check:

- Event type is `USER_DM_MESSAGE`.
- Bot is connected on shard `0`.
- For unsharded bots, `shard_count = 1`.
- Bot gateway session has not expired from the registry.

### Message REST Calls Return `403`

Check:

- Bot can view the channel.
- Message history routes require read history.
- Send and typing routes require send message.
- Add reaction requires add reactions.
- Editing/deleting someone else's guild message requires manage messages.
- The install grant is not masking out a role permission.

## Operational Safety

- Run more than one bot router instance for partition failover.
- Keep gateway and router clocks reasonably synchronized for heartbeat and lease behavior.
- Treat Bot NATS as production-critical for bot event delivery.
- Monitor gateway disconnect rates, heartbeat failures, router partition churn, and REST 401/403/5xx rates.
- Encourage bot developers to perform REST catch-up after reconnects.
