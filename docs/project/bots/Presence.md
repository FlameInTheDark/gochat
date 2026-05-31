[<- Bots](README.md)

# Presence

Bot presence is user-visible presence for bot users. A bot appears online to users who can see it when at least one bot gateway session for that bot has an online-like status.

Presence is driven by `cmd/botws`:

- identify creates a bot presence session, defaulting to `online`
- heartbeat refreshes the session TTL
- disconnect removes that session
- presence update changes this session's status and custom text

Statuses:

| Status | Meaning |
|--------|---------|
| `online` | Active bot instance |
| `idle` | Bot is connected but idle |
| `dnd` | Bot is connected and busy |
| `offline` | Bot intentionally appears offline for that runtime session |

Multiple bot instances aggregate like normal multi-device user presence. If any active bot instance is `dnd`, the aggregate is `dnd`; otherwise any `online` instance makes the bot online; otherwise any `idle` instance makes it idle. If all active sessions are offline or expired, the bot appears offline.

Presence is stored in the shared presence Redis data model and published to the normal presence subject:

```text
presence.user.{botUserID}
```

Only presence updates use the normal NATS bus. Bot event delivery remains on `bot-nats`.
