# GoChat Documentation

GoChat is a realtime chat system with guilds, channels, direct messages, voice, presence, bot runtime APIs, and WebSocket event delivery.

This documentation is organized for two audiences:

- **Bot developers** who want to build integrations against the public bot REST API and bot gateway.
- **GoChat maintainers** who need service, routing, storage, and operations details.

## Start Here

If you are building a bot:

1. Read the [Bot API overview](bot-api/index.md).
2. Create or obtain a bot token from the GoChat bot management flow.
3. Follow the [Quick Start](bot-api/quickstart.md) with the Go client.
4. Use the [REST API reference](bot-api/rest.md) and [Gateway reference](bot-api/gateway.md) for production integrations.

If you are maintaining GoChat itself:

1. Read the [project overview](project/README.md).
2. Review [Services](project/Services.md), [Architecture](project/Architecture.md), and [Database](project/Database.md).
3. For bot runtime internals, see [Project Internals > Bots](project/bots/README.md).

## Bot API At A Glance

| Surface | Base | Purpose |
|---------|------|---------|
| Bot REST API | `https://gochat.anticode.dev/bot/api/v1` | Read bot context, list installed guilds/channels, send/edit/delete/read messages, typing, reactions |
| Bot Gateway | `wss://gochat.anticode.dev/bot/ws` | Identify a bot shard, maintain heartbeats, receive realtime events, update bot presence |
| Go client | `clients/bot/goclient` | Typed Go session, REST helpers, gateway handlers, examples |

Bot runtime authentication always uses:

```http
Authorization: Bot <gcb_token>
```

Bearer user tokens are not accepted by the bot runtime.
