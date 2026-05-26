[<- Bots](README.md)

# Event Routing

The existing client topics remain in place on the user NATS bus. Bot delivery uses a dedicated bot NATS bus with partitioned event subjects:

```text
bot.event.p.{partition}.guild.{guildID}
bot.event.p.{partition}.guild.{guildID}.channel.{channelID}
bot.event.p.{partition}.user.{botUserID}.dm
```

`cmd/botrouter` owns partitions with Redis leases, consumes partitioned bot events, expands each guild event to all eligible installed bots, and publishes targeted delivery envelopes to bot WebSocket instances:

```text
bot.deliver.instance.{instanceID}
```

`cmd/botws` does not subscribe to guild or channel event subjects. Each bot WS process subscribes only to its instance delivery subject, and each identified bot session is registered in Redis with bot id, session id, instance id, shard id, shard count, and DM ownership.

## Routing Rules

- A guild event is consumed once by the router partition owner.
- The router lists current bots installed in that guild.
- Each installed bot is filtered independently by granted permissions, roles, and channel visibility.
- If a bot is sharded, the event is delivered only to `guild_id % shard_count`.
- If a bot is unsharded with `shard_count = 1`, all active bot instances receive the event.
- Direct messages route to shard `0`; unsharded bots receive them on all active instances.

This means a guild with ten bots fans one event out to all ten eligible bots, while each bot's own sharding layout is respected independently.

## Live Changes

Delivery decisions are live:

- Adding a bot to a guild makes it eligible for future routed events without reconnecting.
- Removing a bot from a guild stops future routed events without reconnecting.
- Role changes and channel visibility changes are evaluated by the router before delivery.

Delivery is live and at most once in v1. Bots should use runtime REST history routes to catch up after reconnecting.
