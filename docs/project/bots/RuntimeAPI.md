[<- Bots](README.md)

# Runtime REST API

Bot runtime REST uses the dedicated `cmd/botapi` service.

Base path:

```text
/bot/api/v1
```

Authentication accepts only:

```http
Authorization: Bot <token>
```

`Bearer` is not accepted for bot runtime routes. Missing, malformed, revoked, invalid, disabled, or non-bot tokens are rejected before runtime work is performed.

Runtime handlers are organized by responsibility:

| Entity | Package | Responsibility |
|--------|---------|----------------|
| `user` | `cmd/botapi/endpoints/user` | Current bot account context |
| `guild` | `cmd/botapi/endpoints/guild` | Installed guilds and visible guild channels |
| `message` | `cmd/botapi/endpoints/message` | Messages, typing, reactions, read acknowledgements |

## Route Coverage

| Method | Route | Purpose |
|--------|-------|---------|
| `GET` | `/bot/api/v1/user/me` | Current bot account and bot config |
| `GET` | `/bot/api/v1/guild` | Guilds where the bot is installed |
| `GET` | `/bot/api/v1/guild/:guild_id/channels` | Guild channels visible to the bot |
| `POST` | `/bot/api/v1/message/channel/:channel_id` | Send a message |
| `GET` | `/bot/api/v1/message/channel/:channel_id` | Read message history |
| `PATCH` | `/bot/api/v1/message/channel/:channel_id/:message_id` | Edit a message |
| `DELETE` | `/bot/api/v1/message/channel/:channel_id/:message_id` | Delete a message |
| `POST` | `/bot/api/v1/message/channel/:channel_id/:message_id/ack` | Advance bot read state |
| `POST` | `/bot/api/v1/message/channel/:channel_id/typing` | Publish typing indicator |
| `PUT` | `/bot/api/v1/message/channel/:channel_id/:message_id/reactions/:reaction_name` | Add reaction |
| `DELETE` | `/bot/api/v1/message/channel/:channel_id/:message_id/reactions/:reaction_name` | Remove reaction |
| `GET` | `/bot/api/v1/message/channel/:channel_id/:message_id/reactions/:reaction_name` | List users for one reaction |

The route set avoids special path characters and does not use `@` in any route path.

## Direct Messages

Bots can receive one-to-one direct messages sent to the bot user through the bot gateway. The delivered payload includes `channel_id`, `message_id`, sender data, and, when available, the full `message` DTO. Bots reply using the same send route:

```text
POST /bot/api/v1/message/channel/:channel_id
```

The bot does not need to subscribe to the DM channel to receive DM events.
