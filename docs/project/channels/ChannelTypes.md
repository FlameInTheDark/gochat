[<- Documentation](../README.md) - [Channels](README.md)

# Channel Types

This document describes the different channel types available in gochat, their properties, and structure.

## Channel Type Values

| Type | Constant | Name | Description |
|------|----------|------|-------------|
| 0 | `ChannelTypeGuild` | Text Channel | Standard text channel within a guild |
| 1 | `ChannelTypeGuildVoice` | Voice Channel | Voice channel for audio/video communication |
| 2 | `ChannelTypeGuildCategory` | Category | Container for organizing channels |
| 3 | `ChannelTypeDM` | DM Channel | Direct message between two users |
| 4 | `ChannelTypeGroupDM` | Group DM | Direct message group with multiple users |
| 5 | `ChannelTypeThread` | Thread | Conversation thread attached to a message |

---

## Channel Structure

```json
{
  "id": 2226022078341972000,
  "type": 0,
  "guild_id": 2226022078304223200,
  "name": "general",
  "position": 0,
  "topic": "General discussion channel",
  "private": false,
  "roles": [],
  "last_message_id": 2228801793842741200,
  "created_at": "2026-01-15T10:00:00Z"
}
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Unique channel identifier (snowflake) |
| `type` | int | Channel type (0-5, see table above) |
| `guild_id` | int64? | Guild ID for guild channels (null for DMs) |
| `participant_id` | int64? | For DM channels: the other participant's user ID |
| `name` | string | Channel name (lowercase, no spaces) |
| `parent_id` | int64? | Parent category ID for nested channels |
| `position` | int | Sorting position within the guild |
| `topic` | string? | Channel topic/description |
| `permissions` | int64? | Channel-specific permission overrides |
| `private` | bool | If true, only visible to specific roles |
| `roles` | int64[] | Role IDs with access to private channels |
| `last_message_id` | int64 | ID of the most recent message |
| `voice_region` | string? | Voice region for voice channels (e.g., "us-east", "eu-west") |
| `created_at` | string (ISO8601) | When the channel was created |

---

## Type 0: Text Channel (`ChannelTypeGuild`)

Standard text channel for sending messages within a guild.

**Features:**
- Send/receive text messages
- Upload file attachments
- Message threading
- Typing indicators
- Read receipts

**Example:**
```json
{
  "id": 2226022078341972000,
  "type": 0,
  "guild_id": 2226022078304223200,
  "name": "general",
  "position": 0,
  "topic": "General chat for everyone",
  "private": false,
  "roles": [],
  "last_message_id": 2228801793842741200,
  "created_at": "2026-01-15T10:00:00Z"
}
```

---

## Type 1: Voice Channel (`ChannelTypeGuildVoice`)

Voice channels for real-time audio communication using WebRTC.

**Features:**
- Voice communication (WebRTC)
- Video streaming (WebRTC)
- Screen sharing
- Speaking indicators
- Server-side mute/deafen
- Per-user volume control

**Additional Fields:**
- `voice_region`: Selects the SFU (Selective Forwarding Unit) region for low latency

**Example:**
```json
{
  "id": 2226022078341972001,
  "type": 1,
  "guild_id": 2226022078304223200,
  "name": "general-voice",
  "position": 1,
  "topic": null,
  "voice_region": "us-east",
  "private": false,
  "roles": [],
  "last_message_id": 0,
  "created_at": "2026-01-15T10:00:00Z"
}
```

**Voice Regions:**
- `us-east` - US East Coast
- `us-west` - US West Coast
- `eu-west` - Western Europe
- `eu-central` - Central Europe
- `ap-northeast` - Asia Pacific Northeast
- `ap-southeast` - Asia Pacific Southeast

---

## Type 2: Category (`ChannelTypeGuildCategory`)

Organizational container to group channels together.

**Features:**
- Groups channels visually in the UI
- Can contain any channel type except other categories
- Channels reference the category via `parent_id`
- Cannot send messages directly

**Example:**
```json
{
  "id": 2226022078341971000,
  "type": 2,
  "guild_id": 2226022078304223200,
  "name": "Text Channels",
  "position": 0,
  "topic": null,
  "private": false,
  "roles": [],
  "last_message_id": 0,
  "created_at": "2026-01-15T10:00:00Z"
}
```

**Channel with Parent Category:**
```json
{
  "id": 2226022078341972000,
  "type": 0,
  "guild_id": 2226022078304223200,
  "name": "announcements",
  "parent_id": 2226022078341971000,
  "position": 0,
  "topic": "Official announcements",
  "private": false,
  "roles": [],
  "last_message_id": 2228801793842741200,
  "created_at": "2026-01-15T10:00:00Z"
}
```

---

## Type 3: DM Channel (`ChannelTypeDM`)

One-to-one direct message channel between two users.

**Features:**
- Private conversation between exactly 2 users
- Not associated with any guild
- Cannot be created manually (auto-created on first message)
- Supports all text channel features

**Example:**
```json
{
  "id": 2226022078341974000,
  "type": 3,
  "participant_id": 2226021950625415201,
  "name": "dm-2226021950625415200-2226021950625415201",
  "position": 0,
  "topic": null,
  "private": true,
  "last_message_id": 2228801793842741300,
  "created_at": "2026-01-15T10:00:00Z"
}
```

> [!NOTE]
> `participant_id` indicates the other user in the DM conversation.

---

## Type 4: Group DM (`ChannelTypeGroupDM`)

Direct message channel with multiple participants (group chat).

**Features:**
- Private conversation with 2+ users
- Not associated with any guild
- Supports all text channel features
- Can be created via API

**Example:**
```json
{
  "id": 2226022078341975000,
  "type": 4,
  "name": "friends-group",
  "position": 0,
  "topic": "Group chat for the weekend trip",
  "private": true,
  "last_message_id": 2228801793842741400,
  "created_at": "2026-01-15T10:00:00Z"
}
```

---

## Type 5: Thread (`ChannelTypeThread`)

A conversation thread attached to a specific message.

**Features:**
- Created from an existing message
- Organized discussion on a specific topic
- `parent_id` references the source channel
- Archived automatically after inactivity
- Supports all text channel features

**Example:**
```json
{
  "id": 2226022078341973000,
  "type": 5,
  "guild_id": 2226022078304223200,
  "name": "thread-feature-ideas",
  "parent_id": 2226022078341972000,
  "position": 0,
  "topic": null,
  "private": false,
  "last_message_id": 2228801793842741500,
  "created_at": "2026-01-15T10:30:00Z"
}
```

> [!NOTE]
> Thread channels use `parent_id` to reference the original channel, not a category.

---

## Private Channels

Channels can be marked as `private: true` to restrict access to specific roles.

**Example Private Channel:**
```json
{
  "id": 2226022078341972002,
  "type": 0,
  "guild_id": 2226022078304223200,
  "name": "moderator-only",
  "position": 2,
  "topic": "Staff discussion channel",
  "private": true,
  "roles": [2230469276416868352, 2230469276416868353],
  "last_message_id": 2228801793842741600,
  "created_at": "2026-01-15T10:00:00Z"
}
```

Private channels are only visible to users who have at least one of the specified roles.

---

## WebSocket Events

Channel-related WebSocket events:

| Event Type | Name | Description |
|------------|------|-------------|
| 106 | Channel Create | New channel created |
| 107 | Channel Update | Channel properties changed |
| 108 | Channel Order Update | Channel positions reordered |
| 109 | Channel Delete | Channel deleted |

See [EventTypes.md](../ws/EventTypes.md) for full payload details.

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/guild/{guild_id}/channel` | Create a guild channel |
| POST | `/guild/{guild_id}/category` | Create a category |
| PATCH | `/guild/{guild_id}/channel/{channel_id}` | Update channel |
| DELETE | `/guild/{guild_id}/channel/{channel_id}` | Delete channel |
| POST | `/guild/{guild_id}/channels/order` | Reorder channels |
| POST | `/guild/{guild_id}/voice` | Join voice channel |
