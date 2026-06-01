# REST API

The bot REST API is JSON over HTTP.

Default base URL:

```text
https://gochat.anticode.dev/bot/api/v1
```

All endpoints require:

```http
Authorization: Bot <gcb_token>
```

## Endpoint Summary

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/user/me` | Current bot account and bot config |
| `GET` | `/guild` | Guilds where the bot is installed |
| `GET` | `/guild/{guild_id}/channels` | Channels visible to the bot in one guild |
| `GET` | `/message/channel/{channel_id}` | Read message history |
| `POST` | `/message/channel/{channel_id}` | Send a message |
| `PATCH` | `/message/channel/{channel_id}/{message_id}` | Edit a message |
| `DELETE` | `/message/channel/{channel_id}/{message_id}` | Delete a message |
| `POST` | `/message/channel/{channel_id}/{message_id}/ack` | Mark a channel read |
| `POST` | `/message/channel/{channel_id}/typing` | Publish a typing indicator |
| `PUT` | `/message/channel/{channel_id}/{message_id}/reactions/{reaction_name}` | Add the bot's reaction |
| `DELETE` | `/message/channel/{channel_id}/{message_id}/reactions/{reaction_name}` | Remove the bot's reaction |
| `GET` | `/message/channel/{channel_id}/{message_id}/reactions/{reaction_name}` | List users for one reaction |

Path IDs are signed 64-bit integers encoded as decimal strings.

## Errors

Common responses:

| Status | Meaning |
|--------|---------|
| `400` | Invalid path parameter, query parameter, or JSON body |
| `401` | Missing or invalid bot authentication |
| `403` | Bot is authenticated but cannot access the target guild/channel/message |
| `404` | Target message was not found |
| `500` | Server could not complete the request |

Error bodies are plain strings today. Clients should not depend on a structured error object.

## Pagination

Message history uses cursor pagination with `from` and `direction`.

Reaction users use cursor pagination with `after`.

Default and maximum limits:

| Endpoint | Default | Maximum |
|----------|---------|---------|
| List messages | `50` | `100` |
| List reaction users | `50` | `100` |

Message history rejects `limit <= 0` and `limit > 100`. Reaction users clamp invalid or too-large limits to the supported range.

## `GET /user/me`

Returns the authenticated bot account and bot configuration.

### Response

```json
{
  "bot_user_id": 1001,
  "owner_user_id": 42,
  "user": {
    "id": 1001,
    "name": "ExampleBot",
    "discriminator": "0001",
    "bio": "I help with support",
    "banner_color": 3447003,
    "panel_color": 16777215,
    "is_bot": true
  },
  "description": "Support automation bot",
  "public": true,
  "default_permissions": 1024
}
```

### Status Codes

| Status | Meaning |
|--------|---------|
| `200` | Success |
| `401` | Missing or invalid bot token |

## `GET /guild`

Lists guilds where the bot is installed. Results are sorted by guild ID.

### Response

```json
[
  {
    "id": 2230469276416868352,
    "name": "GoChat Community",
    "owner": 42,
    "public": true,
    "granted_permissions": 3072
  }
]
```

### Status Codes

| Status | Meaning |
|--------|---------|
| `200` | Success |
| `401` | Missing or invalid bot token |
| `500` | Bot guild list could not be loaded |

## `GET /guild/{guild_id}/channels`

Lists guild channels visible to the bot. The bot must be installed in the guild and have `PermServerViewChannels` for each returned channel.

Returned channel types:

| Value | Meaning |
|-------|---------|
| `0` | Guild text channel |
| `1` | Guild voice channel |
| `2` | Guild category |
| `5` | Thread |

### Response

```json
[
  {
    "id": 2230469276416868353,
    "type": 0,
    "guild_id": 2230469276416868352,
    "name": "general",
    "parent_id": null,
    "creator_id": null,
    "position": 1,
    "topic": "Welcome",
    "permissions": 3072,
    "private": false,
    "closed": false,
    "last_message_id": 2230469276416868354,
    "voice_region": null,
    "created_at": "2026-06-01T09:00:00Z"
  }
]
```

### Status Codes

| Status | Meaning |
|--------|---------|
| `200` | Success |
| `400` | Invalid `guild_id` |
| `401` | Missing or invalid bot token |
| `403` | Bot is not installed in the guild |
| `500` | Guild channel lookup failed |

## `GET /message/channel/{channel_id}`

Lists messages visible to the bot.

Required permission:

```text
PermServerViewChannels + PermTextReadMessageHistory
```

### Query Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `from` | int64 | channel last message ID | Message ID cursor |
| `limit` | int | `50` | Page size, from `1` to `100` |
| `direction` | string | `before` | One of `before`, `after`, `around` |

When `from` is omitted and the channel has no last message, the response is an empty array.

### Example

```http
GET /bot/api/v1/message/channel/2230469276416868353?from=2230469276416868354&direction=before&limit=50
Authorization: Bot gcb_xxx
```

### Response

```json
[
  {
    "id": 2230469276416868354,
    "channel_id": 2230469276416868353,
    "author": {
      "id": 1001,
      "name": "ExampleBot",
      "discriminator": "0001",
      "is_bot": true
    },
    "content": "hello",
    "position": 12,
    "embeds": [],
    "flags": 0,
    "type": 0,
    "reactions": []
  }
]
```

### Status Codes

| Status | Meaning |
|--------|---------|
| `200` | Success |
| `400` | Invalid channel ID, limit, or direction |
| `401` | Missing or invalid bot token |
| `403` | Bot cannot access channel or read history |
| `500` | Message lookup failed |

## `POST /message/channel/{channel_id}`

Sends a message as the bot.

Required permission:

```text
PermServerViewChannels + PermTextSendMessage
```

### Request Body

At least one of `content`, `attachments`, or `embeds` is required.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `content` | string | No | Message text |
| `attachments` | int64[] | No | Existing attachment IDs |
| `embeds` | Embed[] | No | Manual embeds |
| `reference` | int64 | No | Message ID to reply to in the same channel |

```json
{
  "content": "Build finished.",
  "embeds": [
    {
      "type": "rich",
      "title": "Deploy",
      "description": "Production deploy completed.",
      "color": 65280
    }
  ],
  "reference": 2230469276416868354
}
```

### Response

Returns the created [Message](payloads.md#message).

### Side Effects

- Reserves a message position.
- Stores the message.
- Updates channel last message.
- Updates guild channel last message for guild channels.
- Advances the bot read state to the new message.
- Publishes a `MESSAGE_CREATE` event.

### Status Codes

| Status | Meaning |
|--------|---------|
| `200` | Message created |
| `400` | Invalid body, empty message, or invalid embed |
| `401` | Missing or invalid bot token |
| `403` | Bot cannot access channel or send messages |
| `500` | Message creation failed |

## `PATCH /message/channel/{channel_id}/{message_id}`

Edits a message.

Required permission to access the channel:

```text
PermServerViewChannels + PermTextReadMessageHistory
```

Bots can edit their own messages. In guild channels, bots can edit other users' editable messages only when they also have `PermTextManageMessages`.

### Request Body

All fields are optional. Fields that are omitted keep their current value.

| Field | Type | Description |
|-------|------|-------------|
| `content` | string | New message content |
| `embeds` | Embed[] | Replacement manual embeds |
| `flags` | int | Replacement normalized message flags |

```json
{
  "content": "Updated content",
  "flags": 0
}
```

### Response

Returns the updated [Message](payloads.md#message).

### Side Effects

Publishes a `MESSAGE_UPDATE` event.

### Status Codes

| Status | Meaning |
|--------|---------|
| `200` | Message updated |
| `400` | Invalid body, invalid embed, or non-editable message type |
| `401` | Missing or invalid bot token |
| `403` | Bot cannot access or edit this message |
| `404` | Message not found |
| `500` | Message update failed |

## `DELETE /message/channel/{channel_id}/{message_id}`

Deletes a message.

Required permission to access the channel:

```text
PermServerViewChannels + PermTextReadMessageHistory
```

Bots can delete their own messages. In guild channels, bots can delete other users' messages only when they also have `PermTextManageMessages`.

### Response

No body.

### Side Effects

- Deletes the message.
- Repairs channel last message if needed.
- Publishes a `MESSAGE_DELETE` event.

### Status Codes

| Status | Meaning |
|--------|---------|
| `204` | Message deleted |
| `400` | Invalid path parameter |
| `401` | Missing or invalid bot token |
| `403` | Bot cannot access or delete this message |
| `404` | Message not found |
| `500` | Message deletion failed |

## `POST /message/channel/{channel_id}/{message_id}/ack`

Marks the channel read for the bot user.

Required permission:

```text
PermServerViewChannels + PermTextReadMessageHistory
```

### Response

No body.

### Side Effects

- Updates the bot's read state for the channel.
- Publishes a read-state update on the user event bus for the bot user. Bot gateway DM routing currently does not deliver this event to bot sessions.

### Status Codes

| Status | Meaning |
|--------|---------|
| `204` | Read state updated |
| `400` | Invalid path parameter |
| `401` | Missing or invalid bot token |
| `403` | Bot cannot access channel or read history |
| `500` | Read-state update failed |

## `POST /message/channel/{channel_id}/typing`

Publishes a typing indicator for the bot user.

Required permission:

```text
PermServerViewChannels + PermTextSendMessage
```

### Response

No body.

### Side Effects

Publishes a `CHANNEL_USER_TYPING` event to eligible clients and bots.

### Status Codes

| Status | Meaning |
|--------|---------|
| `204` | Typing indicator published |
| `400` | Invalid channel ID |
| `401` | Missing or invalid bot token |
| `403` | Bot cannot access channel or send messages |
| `500` | Permission check or publish failed |

## Reactions

`reaction_name` is path-unescaped by the server.

Supported forms:

| Form | Example | Meaning |
|------|---------|---------|
| Unicode emoji | `%E2%9D%A4%EF%B8%8F` | Built-in/system emoji |
| Custom emoji | `party%3A2230469276416868352` | Emoji name plus custom emoji ID |

The server stores reaction buckets as `u:<emoji>` for Unicode emoji and `c:<emoji_id>` for custom emoji.

## `PUT /message/channel/{channel_id}/{message_id}/reactions/{reaction_name}`

Adds or updates the bot's reaction to a message.

Required permission:

```text
PermServerViewChannels + PermTextReadMessageHistory + PermTextAddReactions
```

### Response

```json
{
  "count": 1,
  "me": true,
  "emoji": {
    "id": 2230469276416868352,
    "name": "party"
  }
}
```

### Side Effects

Publishes a `MESSAGE_REACTION_ADD` event.

### Status Codes

| Status | Meaning |
|--------|---------|
| `200` | Reaction added |
| `400` | Invalid path parameter or reaction name |
| `401` | Missing or invalid bot token |
| `403` | Bot cannot access channel, read history, or add reactions |
| `404` | Message not found |
| `500` | Reaction update failed |

## `DELETE /message/channel/{channel_id}/{message_id}/reactions/{reaction_name}`

Removes the bot's reaction from a message. If the bot did not have that reaction, the endpoint still returns `204`.

Required permission:

```text
PermServerViewChannels + PermTextReadMessageHistory
```

### Response

No body.

### Side Effects

Publishes a `MESSAGE_REACTION_REMOVE` event only when an existing bot reaction was removed.

### Status Codes

| Status | Meaning |
|--------|---------|
| `204` | Reaction absent or removed |
| `400` | Invalid path parameter or reaction name |
| `401` | Missing or invalid bot token |
| `403` | Bot cannot access channel or read history |
| `500` | Reaction removal failed |

## `GET /message/channel/{channel_id}/{message_id}/reactions/{reaction_name}`

Lists users who reacted with one reaction.

Required permission:

```text
PermServerViewChannels + PermTextReadMessageHistory
```

### Query Parameters

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `after` | int64 | omitted | Reaction ID cursor. Must be greater than zero when present |
| `limit` | int | `50` | Page size. Values above `100` are clamped |

### Response

```json
{
  "items": [
    {
      "id": 42,
      "name": "alice",
      "discriminator": "0001",
      "is_bot": false
    }
  ],
  "next_after": 9001
}
```

When there is no next page, `next_after` is omitted.

### Status Codes

| Status | Meaning |
|--------|---------|
| `200` | Success |
| `400` | Invalid path parameter, reaction name, or `after` cursor |
| `401` | Missing or invalid bot token |
| `403` | Bot cannot access channel or read history |
| `404` | Message not found |
| `500` | Reaction user lookup failed |

## Permissions By Operation

| Operation | Required permissions |
|-----------|----------------------|
| List guild channels | `PermServerViewChannels` per returned channel |
| List messages | `PermServerViewChannels`, `PermTextReadMessageHistory` |
| Send message | `PermServerViewChannels`, `PermTextSendMessage` |
| Edit own message | `PermServerViewChannels`, `PermTextReadMessageHistory` |
| Edit another user's message | Own-message requirements plus `PermTextManageMessages` in guild channels |
| Delete own message | `PermServerViewChannels`, `PermTextReadMessageHistory` |
| Delete another user's message | Own-message requirements plus `PermTextManageMessages` in guild channels |
| Ack read state | `PermServerViewChannels`, `PermTextReadMessageHistory` |
| Typing | `PermServerViewChannels`, `PermTextSendMessage` |
| Add reaction | `PermServerViewChannels`, `PermTextReadMessageHistory`, `PermTextAddReactions` |
| Remove own reaction | `PermServerViewChannels`, `PermTextReadMessageHistory` |
| List reaction users | `PermServerViewChannels`, `PermTextReadMessageHistory` |

The effective permission set is still capped by the bot guild install grant.
