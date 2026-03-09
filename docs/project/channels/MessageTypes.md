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
| `attachments` | array | List of file attachments |
| `type` | int | Message type (0=Chat, 1=Reply, 2=Join) |
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
  "attachments": [],
  "type": 0,
  "updated_at": null
}
```

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
> Reply messages include a reference to the original message. The `reference` field contains the ID of the message being replied to.

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

## Edited Messages

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
| GET | `/message/channel/{channel_id}` | Get messages from channel |
| PATCH | `/message/channel/{channel_id}/{message_id}` | Edit a message |
| DELETE | `/message/channel/{channel_id}/{message_id}` | Delete a message |
| POST | `/message/channel/{channel_id}/typing` | Send typing indicator |
| POST | `/message/channel/{channel_id}/attachment` | Upload file attachment |
| POST | `/message/channel/{channel_id}/{message_id}/ack` | Acknowledge/read message |
