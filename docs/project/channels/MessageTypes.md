[<- Documentation](../README.md) - [Channels](README.md)

# Message Types

This document describes the different message types and their payloads in the gochat system.

## Message Types

Messages can have different types based on their purpose:

| Type | Name | Description |
|------|------|-------------|
| 0 | Chat | Regular text message sent by a user |
| 1 | Reply | Message that references/replies to another message |
| 2 | Join | System message indicating a user joined the guild |
| 3 | Thread Created | Informative message posted in the parent channel when a thread is created |
| 4 | Thread Initial | Informative copy of the source message posted as the first message inside a new thread |

## Message Structure

### Standard Message Payload

```json
{
  "id": 2228801793842741200,
  "channel_id": 2226022078341972000,
  "author": {
    "id": 2226021950625415200,
    "name": "FlameInTheDark",
    "discriminator": "flameinthedark",
    "avatar": null
  },
  "content": "Hello world!",
  "position": 512,
  "nonce": "draft-1",
  "attachments": [],
  "type": 0,
  "updated_at": null
}
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Unique message identifier (snowflake) |
| `channel_id` | int64 | ID of the channel where the message was sent |
| `author` | [User](#user-structure) | The user who sent the message |
| `content` | string | The message text content |
| `position` | int64? | Monotonic channel-local message position used for navigation. It is assigned on create, never changes later, may have gaps, and does not decrease when messages are deleted. Older messages may omit it until backfilled. |
| `nonce` | string or integer? | Ephemeral client nonce echoed in the immediate send response and in the author's own `Message Create` WebSocket event. It is not stored with historical messages and is stripped from channel events delivered to other users. |
| `attachments` | array | List of file attachments |
| `type` | int | Message type (0=Chat, 1=Reply, 2=Join, 3=Thread Created, 4=Thread Initial) |
| `reference` | int64? | Referenced source message ID for replies and thread-created messages |
| `reference_channel_id` | int64? | Channel ID of the referenced source message |
| `thread_id` | int64? | Thread attached to this message or linked by a thread-created message. A normal type-0 source message gets this after a thread is created from it. |
| `thread` | [Channel](ChannelTypes.md#channel-structure)? | Nested thread metadata when the message points to a live thread. A normal type-0 source message gets this after a thread is created from it. This nested thread object is for thread navigation metadata: it includes `member_ids` and approximate `message_count`, but does not include current-user thread membership. |
| `updated_at` | string (ISO8601) | Timestamp when the message was last edited (null if never edited) |

## User Structure

```json
{
  "id": 2226021950625415200,
  "name": "FlameInTheDark",
  "discriminator": "flameinthedark",
  "avatar": {
    "url": "https://cdn.gochat.io/avatars/...",
    "content_type": "image/png",
    "width": 512,
    "height": 512,
    "size": 102400
  }
}
```

## Attachment Structure

```json
{
  "content_type": "image/png",
  "filename": "screenshot.png",
  "height": 600,
  "width": 800,
  "url": "https://cdn.gochat.io/attachments/...",
  "preview_url": "https://cdn.gochat.io/previews/...",
  "size": 1048576
}
```

### Attachment Fields

| Field | Type | Description |
|-------|------|-------------|
| `content_type` | string | MIME type of the file (e.g., `image/png`, `video/mp4`) |
| `filename` | string | Original file name |
| `height` | int64 | Image/video height in pixels (optional) |
| `width` | int64 | Image/video width in pixels (optional) |
| `url` | string | Direct URL to download the full file |
| `preview_url` | string | URL to a preview/thumbnail version (optional) |
| `size` | int64 | File size in bytes |

## Message Type Examples

### Type 0: Chat Message

Standard user message:

```json
{
  "id": 2228801793842741200,
  "channel_id": 2226022078341972000,
  "author": {
    "id": 2226021950625415200,
    "name": "FlameInTheDark",
    "discriminator": "flameinthedark",
    "avatar": null
  },
  "content": "Hey everyone! How's it going?",
  "position": 512,
  "nonce": "draft-1",
  "attachments": [],
  "type": 0,
  "updated_at": null
}
```

> [!NOTE]
> `nonce` is optional. Clients can send it on `POST /message/channel/{channel_id}` to correlate optimistic sends. `enforce_nonce = true` makes the send idempotent for a short window within that channel.

> [!NOTE]
> `position` exists on messages in any channel type. It is channel-local, monotonic, and immutable after the message is created. Positions are allocated lazily from cache-backed blocks, so clients must not assume they are contiguous.

### Type 0: Chat with Attachments

Message with file attachments:

```json
{
  "id": 2228801793842741300,
  "channel_id": 2226022078341972000,
  "author": {
    "id": 2226021950625415200,
    "name": "FlameInTheDark",
    "discriminator": "flameinthedark",
    "avatar": null
  },
  "content": "Check out this image!",
  "attachments": [
    {
      "content_type": "image/jpeg",
      "filename": "vacation.jpg",
      "height": 1080,
      "width": 1920,
      "url": "https://cdn.gochat.io/attachments/2228801793842741300/vacation.jpg",
      "preview_url": "https://cdn.gochat.io/previews/2228801793842741300/vacation_preview.webp",
      "size": 2097152
    }
  ],
  "type": 0,
  "updated_at": null
}
```

### Type 1: Reply Message

> [!NOTE]
> Reply messages include a reference to the original message. Replies are limited to the same channel, so `reference_channel_id` always matches the message's own `channel_id`.

```json
{
  "id": 2228801793842741400,
  "channel_id": 2226022078341972000,
  "author": {
    "id": 2226021950625415201,
    "name": "AnotherUser",
    "discriminator": "anotheruser",
    "avatar": null
  },
  "content": "Thanks for sharing!",
  "attachments": [],
  "type": 1,
  "reference": 2228801793842741300,
  "reference_channel_id": 2226022078341972000,
  "updated_at": null
}
```

### Type 2: Join System Message

System message sent when a user joins a guild:

```json
{
  "id": 2228801793842741500,
  "channel_id": 2226022078341972000,
  "author": {
    "id": 2226021950625415202,
    "name": "NewMember",
    "discriminator": "newmember",
    "avatar": null
  },
  "content": "NewMember joined the server",
  "attachments": [],
  "type": 2,
  "updated_at": null
}
```

### Type 0: Chat Message With Attached Thread

When a thread is created from a regular parent-channel message, that original message stays `type = 0` and gains thread linkage:

```json
{
  "id": 2228801793842741300,
  "channel_id": 2226022078341972000,
  "author": {
    "id": 2226021950625415200,
    "name": "FlameInTheDark",
    "discriminator": "flameinthedark",
    "avatar": null
  },
  "content": "Can we track this separately?",
  "attachments": [],
  "type": 0,
  "thread_id": 2226022078341973000,
  "thread": {
    "id": 2226022078341973000,
    "type": 5,
    "guild_id": 2226022078304223200,
    "creator_id": 2226021950625415200,
    "name": "release discussion",
    "parent_id": 2226022078341972000,
    "position": 0,
    "closed": false,
    "last_message_id": 2228801793842741700,
    "created_at": "2026-03-11T10:00:00Z"
  },
  "updated_at": null
}
```

### Type 3: Thread Created Message

Informative message posted in the parent channel after a thread is created:

```json
{
  "id": 2228801793842741600,
  "channel_id": 2226022078341972000,
  "author": {
    "id": 2226021950625415200,
    "name": "FlameInTheDark",
    "discriminator": "flameinthedark",
    "avatar": null
  },
  "content": "release discussion",
  "attachments": [],
  "type": 3,
  "reference_channel_id": 2226022078341972000,
  "reference": 2228801793842741300,
  "thread_id": 2226022078341973000,
  "thread": {
    "id": 2226022078341973000,
    "type": 5,
    "guild_id": 2226022078304223200,
    "creator_id": 2226021950625415200,
    "name": "release discussion",
    "parent_id": 2226022078341972000,
    "position": 0,
    "closed": false,
    "last_message_id": 2228801793842741700,
    "created_at": "2026-03-11T10:00:00Z"
  },
  "updated_at": null
}
```

> [!NOTE]
> Thread-created messages are informational and are not editable.
> Their `content` mirrors the current thread name.
> If the thread is renamed later, this message content is updated too.

### Type 4: Thread Initial Message

Informative copy of the source message stored as the first message inside the thread:

```json
{
  "id": 2228801793842741700,
  "channel_id": 2226022078341973000,
  "author": {
    "id": 2226021950625415200,
    "name": "FlameInTheDark",
    "discriminator": "flameinthedark",
    "avatar": null
  },
  "content": "Original message content",
  "attachments": [],
  "type": 4,
  "reference_channel_id": 2226022078341972000,
  "reference": 2228801793842741300,
  "updated_at": null
}
```

> [!NOTE]
> Thread-initial messages are informational and are not editable.
> They preserve the source message author/content/attachments/embeds while pointing back to the original message in the parent channel.

## Edited Messages

> [!NOTE]
> Only user-editable message types can be edited. Informational/system messages such as `Join`, `Thread Created`, and `Thread Initial` are not editable.

When a message is edited, the `updated_at` field contains the timestamp of the edit:

```json
{
  "id": 2228801793842741200,
  "channel_id": 2226022078341972000,
  "author": {
    "id": 2226021950625415200,
    "name": "FlameInTheDark",
    "discriminator": "flameinthedark",
    "avatar": null
  },
  "content": "This is the edited content",
  "attachments": [],
  "type": 0,
  "updated_at": "2026-01-15T14:30:00Z"
}
```

## WebSocket Events

Message-related WebSocket events use the following event types:

| Event Type | Name | Description |
|------------|------|-------------|
| 100 | Message Create | New message posted |
| 101 | Message Update | Message was edited |
| 102 | Message Delete | Message was deleted |

See [EventTypes.md](../ws/EventTypes.md) for full payload details.

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/message/channel/{channel_id}` | Send a new message |
| POST | `/message/channel/{channel_id}/{message_id}/thread` | Create a thread from a source message and send the first thread message |
| GET | `/message/channel/{channel_id}` | Get messages from channel |
| PATCH | `/message/channel/{channel_id}/{message_id}` | Edit a message |
| DELETE | `/message/channel/{channel_id}/{message_id}` | Delete a message |
| POST | `/message/channel/{channel_id}/typing` | Send typing indicator |
| POST | `/message/channel/{channel_id}/attachment` | Upload file attachment |
| POST | `/message/channel/{channel_id}/{message_id}/ack` | Acknowledge/read message |

### Send Message Request

`POST /message/channel/{channel_id}` accepts the normal message payload plus two nonce-related fields:

```json
{
  "content": "Hello world!",
  "nonce": "draft-1",
  "enforce_nonce": true,
  "reference": 2230469276416868351,
  "attachments": [],
  "mentions": [],
  "embeds": []
}
```

Rules:

- `nonce` is optional.
- `nonce` may be either a string or an integer.
- `nonce` is capped at 25 characters.
- `enforce_nonce` is optional.
- `enforce_nonce = true` requires `nonce`.
- When `enforce_nonce = true`, gochat treats `nonce` as a short-lived idempotency key for that author in that channel.
- If the same author re-sends the same `nonce` in the same channel during the active window, the server returns the existing message instead of creating a new one.
- `reference` is optional.
- When `reference` is set, the new message is stored as type `1` (`Reply`).
- Replies are limited to the same channel. The referenced message must already exist in the exact `channel_id` being posted to.
- `nonce` is not stored as durable message history metadata. It is only echoed in the immediate send response and in the author's own live WebSocket event.

### Reply Flow

Reply creation uses the normal message-send endpoint:

```http
POST /message/channel/{channel_id}
```

Example request:

```json
{
  "content": "Thanks for sharing!",
  "reference": 2228801793842741300
}
```

Flow:

1. The client sends a normal message payload with `reference = source_message_id`.
2. The server validates that the referenced message already exists in the exact same `channel_id`.
3. If the target message is missing or belongs to another channel, the request is rejected with `400 Bad Request`.
4. If validation succeeds, the server stores the new message as `type = 1` (`Reply`).
5. The stored reply keeps:
   - `reference = source_message_id`
   - `reference_channel_id = current channel_id`
6. The HTTP response returns the new reply message with those fields attached.
7. Clients subscribed to `channel.{channel_id}` receive a normal `Message Create` event (`t = 100`) whose `message.type = 1`.

Reply events do not use a separate websocket event type. They are ordinary message-create events with reply metadata on the `message` object.

### Get Messages Response Order

`GET /message/channel/{channel_id}` returns messages in different orders depending on `direction`.

Defaults:

- if `direction` is omitted, the server uses `before`
- if `direction = before` and `from` is omitted, the server starts from the channel's current `last_message_id`

Order rules:

- `direction=before`
  Returns messages from newest to oldest.
  The `from` message is included when it exists.
- `direction=after`
  Returns messages from oldest to newest.
  The `from` message is included when it exists.
- `direction=around`
  Returns the `from` message first, then older messages in descending order, then newer messages in ascending order.
  The `from` message is included exactly once.

Examples with `from = 100`:

- `before` -> `[100, 99, 98, 97]`
- `after` -> `[100, 101, 102, 103]`
- `around` -> `[100, 99, 98, 101, 102]`

See [Threads](Threads.md) for thread-specific message behavior.
