# Payloads

This page documents JSON objects used by the bot REST API and gateway.

Conventions:

- `int64` IDs are JSON numbers.
- Fields marked optional may be omitted or `null`, depending on the producer.
- Timestamps are RFC 3339 strings unless a field explicitly says Unix time.

## User

Public user profile returned to bots.

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | User ID |
| `name` | string | Display name |
| `discriminator` | string | Discriminator string |
| `bio` | string | Optional bio |
| `banner_color` | int | Optional banner color |
| `panel_color` | int | Optional panel color |
| `avatar` | [AvatarData](#avatardata) | Optional avatar metadata |
| `banner` | [BannerData](#bannerdata) | Optional banner metadata |
| `personal_note` | string | Optional note, when present in a user-facing payload |
| `is_bot` | boolean | Whether the user is a bot |

## UserBrief

Lightweight user object used in direct-message and friendship events.

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | User ID |
| `name` | string | Display name |
| `discriminator` | string | Discriminator string |
| `avatar` | int64 | Optional legacy avatar ID |
| `avatar_data` | [AvatarData](#avatardata) | Optional avatar metadata |

## AvatarData

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Avatar file ID |
| `url` | string | Public URL |
| `content_type` | string | Optional MIME type |
| `width` | int64 | Optional width in pixels |
| `height` | int64 | Optional height in pixels |
| `size` | int64 | File size in bytes |

## BannerData

| Field | Type | Description |
|-------|------|-------------|
| `exists` | boolean | Whether a banner exists |
| `id` | int64 | Optional banner file ID |
| `url` | string | Optional public URL |
| `content_type` | string | Optional MIME type |
| `width` | int64 | Optional width in pixels |
| `height` | int64 | Optional height in pixels |
| `size` | int64 | Optional file size in bytes |

## GuildResponse

Returned by `GET /guild`.

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Guild ID |
| `name` | string | Guild name |
| `owner` | int64 | Owner user ID |
| `public` | boolean | Whether the guild is public |
| `granted_permissions` | int64 | Permission grant for the bot install |

## Guild

Used by guild update events.

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Guild ID |
| `name` | string | Guild name |
| `icon` | [Icon](#icon) | Optional guild icon |
| `owner` | int64 | Owner user ID |
| `public` | boolean | Whether the guild is public |
| `permissions` | int64 | Permission bits in this context |
| `system_channel_id` | int64 | Optional system channel ID |

## Icon

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Icon file ID |
| `url` | string | Public URL |
| `filesize` | int64 | File size in bytes |
| `width` | int64 | Width in pixels |
| `height` | int64 | Height in pixels |

## Channel

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Channel ID |
| `type` | integer | [Channel type](#channeltype) |
| `guild_id` | int64 | Optional guild ID |
| `participant_id` | int64 | Optional DM participant ID |
| `creator_id` | int64 | Optional creator user ID |
| `member` | [ThreadMember](#threadmember) | Optional thread member data |
| `member_ids` | int64[] | Optional member IDs |
| `name` | string | Channel name |
| `parent_id` | int64 | Optional parent channel/category ID |
| `position` | integer | Sort position |
| `topic` | string | Optional topic |
| `permissions` | int64 | Optional permission bits |
| `private` | boolean | Whether the channel is private |
| `closed` | boolean | Whether the channel/thread is closed |
| `roles` | int64[] | Optional role IDs |
| `last_message_id` | int64 | Last message ID or `0` |
| `message_count` | int64 | Optional message count |
| `voice_region` | string | Optional voice region |
| `created_at` | string | Creation timestamp |

## ChannelType

| Value | Name |
|-------|------|
| `0` | Guild text |
| `1` | Guild voice |
| `2` | Guild category |
| `3` | Direct message |
| `4` | Group direct message |
| `5` | Thread |

## ThreadMember

| Field | Type | Description |
|-------|------|-------------|
| `user_id` | int64 | User ID |
| `join_timestamp` | string | Join timestamp |
| `flags` | integer | Thread member flags |

## ChannelOrder

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Channel ID |
| `position` | integer | New position |

## Role

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Role ID |
| `guild_id` | int64 | Guild ID |
| `name` | string | Role name |
| `color` | integer | Decimal RGB color |
| `permissions` | int64 | Permission bits |
| `position` | integer | Sort position |
| `hoist` | boolean | Whether the role is visually separated |

## Member

| Field | Type | Description |
|-------|------|-------------|
| `user` | [User](#user) | User profile |
| `username` | string | Optional guild nickname |
| `avatar` | int64 | Optional guild avatar ID |
| `join_at` | string | Guild join timestamp |
| `roles` | int64[] | Role IDs |

## Message

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Message ID |
| `channel_id` | int64 | Channel ID |
| `author` | [User](#user) | Author profile |
| `content` | string | Message content |
| `position` | int64 | Optional channel position |
| `nonce` | string | Optional client nonce |
| `attachments` | [Attachment](#attachment)[] | Attachments |
| `embeds` | [Embed](#embed)[] | Manual and generated embeds |
| `flags` | integer | Message flags |
| `type` | integer | Message type |
| `reference` | int64 | Optional referenced message ID |
| `reference_channel_id` | int64 | Optional referenced channel ID |
| `thread_id` | int64 | Optional thread ID |
| `thread` | [Channel](#channel) | Optional thread object |
| `reactions` | [MessageReaction](#messagereaction)[] | Reaction summaries |
| `updated_at` | string | Optional edit timestamp |

## Attachment

| Field | Type | Description |
|-------|------|-------------|
| `content_type` | string | Optional MIME type |
| `filename` | string | File name |
| `height` | int64 | Optional media height |
| `width` | int64 | Optional media width |
| `url` | string | Public URL |
| `preview_url` | string | Optional preview URL |
| `size` | int64 | File size in bytes |

## MessageSend

Request body for `POST /message/channel/{channel_id}`.

| Field | Type | Description |
|-------|------|-------------|
| `content` | string | Message text |
| `attachments` | int64[] | Existing attachment IDs |
| `embeds` | [Embed](#embed)[] | Manual embeds |
| `reference` | int64 | Optional message ID to reply to |

## MessageEdit

Request body for `PATCH /message/channel/{channel_id}/{message_id}`.

| Field | Type | Description |
|-------|------|-------------|
| `content` | string | Optional replacement content |
| `embeds` | [Embed](#embed)[] | Optional replacement manual embeds |
| `flags` | integer | Optional replacement flags |

## MessageReaction

| Field | Type | Description |
|-------|------|-------------|
| `count` | integer | Number of users with this reaction |
| `me` | boolean | Whether the bot user has this reaction |
| `emoji` | [MessageReactionEmoji](#messagereactionemoji) | Emoji |

## MessageReactionEmoji

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Custom emoji ID, omitted for Unicode emoji |
| `name` | string | Unicode emoji or custom emoji name |

## MessageReactionUsersPage

| Field | Type | Description |
|-------|------|-------------|
| `items` | [User](#user)[] | Users in this page |
| `next_after` | int64 | Cursor for the next page |

## GuildEmoji

`id` and `guild_id` are encoded as JSON strings in emoji event payloads.

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Emoji ID |
| `guild_id` | string | Guild ID |
| `name` | string | Emoji name |
| `animated` | boolean | Whether the emoji is animated |

## Embed

Manual embeds can be sent in `MessageSend` and `MessageEdit`.

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Optional title |
| `type` | string | Optional type: `rich`, `image`, `video`, `gifv`, `article`, `link` |
| `description` | string | Optional description |
| `url` | string | Optional canonical URL |
| `timestamp` | string | Optional timestamp |
| `color` | integer | Optional decimal RGB color, `0` to `16777215` |
| `footer` | [EmbedFooter](#embedfooter) | Optional footer |
| `image` | [EmbedMedia](#embedmedia) | Optional full-size image |
| `thumbnail` | [EmbedMedia](#embedmedia) | Optional thumbnail |
| `video` | [EmbedMedia](#embedmedia) | Optional video metadata |
| `provider` | [EmbedProvider](#embedprovider) | Optional provider metadata |
| `author` | [EmbedAuthor](#embedauthor) | Optional author metadata |
| `fields` | [EmbedField](#embedfield)[] | Up to 25 fields |

### Embed Limits

| Limit | Value |
|-------|-------|
| Embeds per message | `10` |
| Fields per embed | `25` |
| Total embed text across all embeds | `6000` characters |
| Title | `256` characters |
| Description | `4096` characters |
| Footer text | `2048` characters |
| Author name | `256` characters |
| Field name | `256` characters |
| Field value | `1024` characters |
| Color | `0` to `16777215` |

URLs must use `http`, `https`, or, for selected fields, `attachment`.

## EmbedFooter

| Field | Type | Description |
|-------|------|-------------|
| `text` | string | Required when footer is present |
| `icon_url` | string | Optional icon URL |
| `proxy_icon_url` | string | Optional proxied icon URL |

## EmbedMedia

| Field | Type | Description |
|-------|------|-------------|
| `url` | string | Media URL |
| `proxy_url` | string | Optional proxied URL |
| `height` | int64 | Optional non-negative height |
| `width` | int64 | Optional non-negative width |
| `content_type` | string | Optional MIME type |
| `placeholder` | string | Optional encoded placeholder |
| `placeholder_version` | integer | Optional non-negative placeholder version |
| `flags` | integer | Optional non-negative flags |

## EmbedProvider

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Provider name |
| `url` | string | Provider URL |

## EmbedAuthor

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Required when author is present |
| `url` | string | Optional author URL |
| `icon_url` | string | Optional icon URL |
| `proxy_icon_url` | string | Optional proxied icon URL |

## EmbedField

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Required field name |
| `value` | string | Required field value |
| `inline` | boolean | Optional inline display hint |

## ActiveStream

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Stream ID |
| `channel_id` | int64 | Channel ID |
| `source_type` | string | Stream source type |
| `audio_mode` | string | Audio mode |
| `started_at` | int64 | Unix timestamp |

## Ready

Gateway ready event payload.

| Field | Type | Description |
|-------|------|-------------|
| `bot` | [User](#user) | Bot user |
| `session_id` | string | Gateway session ID |
| `shard_id` | integer | Current shard ID |
| `shard_count` | integer | Total shard count |
| `guild_ids` | int64[] | Installed guilds assigned to this shard |
| `dm_channel_ids` | int64[] | DM channels assigned to this shard at identify time |
| `group_dm_channel_ids` | int64[] | Group DM channels assigned to this shard at identify time |
| `receives_dm_events` | boolean | Whether this shard receives DM event routing |
| `granted_permissions` | object | Map of guild ID to granted permission bits |

## HeartbeatInterval

Server hello payload after identify.

| Field | Type | Description |
|-------|------|-------------|
| `heartbeat_interval` | int64 | Milliseconds between heartbeats |
| `session_id` | string | Gateway session ID |

## HeartbeatAck

| Field | Type | Description |
|-------|------|-------------|
| `server_time` | int64 | Server Unix time in milliseconds |

## Presence

Initial identify presence.

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | `online`, `idle`, `dnd`, or `offline` |
| `custom_status_text` | string | Optional custom status, max 255 characters |

## PresenceUpdateRequest

Client payload for opcode `3`.

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | `online`, `idle`, `dnd`, or `offline` |
| `custom_status_text` | string | Optional custom status |

The shared struct also contains `platform`, `voice_channel_id`, `mute`, `deafen`, and `self_video`, but the bot gateway currently normalizes only `status` and `custom_status_text`.

## Dispatch Payloads

### MessageCreate

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `message` | [Message](#message) | Message created |

### MessageUpdate

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `message` | [Message](#message) | Updated message |

### MessageDelete

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `channel_id` | int64 | Channel ID |
| `message_id` | int64 | Deleted message ID |

### MessageReactionAdd

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `channel_id` | int64 | Channel ID |
| `message_id` | int64 | Message ID |
| `reaction` | [MessageReaction](#messagereaction) | Reaction summary |

### MessageReactionRemove

Same fields as [MessageReactionAdd](#messagereactionadd).

### GuildCreate

| Field | Type | Description |
|-------|------|-------------|
| `guild` | [Guild](#guild) | Guild |

### GuildUpdate

| Field | Type | Description |
|-------|------|-------------|
| `guild` | [Guild](#guild) | Updated guild |

### GuildDelete

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Deleted guild ID |

### ChannelCreate

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `channel` | [Channel](#channel) | Created channel |

### ChannelUpdate

Same fields as [ChannelCreate](#channelcreate), with the updated channel.

### ChannelOrderUpdate

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `channels` | [ChannelOrder](#channelorder)[] | Updated order entries |

### ChannelDelete

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `channel_type` | integer | Deleted channel type |
| `channel_id` | int64 | Deleted channel ID |

### GuildRoleCreate

| Field | Type | Description |
|-------|------|-------------|
| `role` | [Role](#role) | Created role |

### GuildRoleUpdate

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `role` | [Role](#role) | Updated role |

### GuildRoleDelete

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `role_id` | int64 | Deleted role ID |

### ThreadCreate

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `thread` | [Channel](#channel) | Created thread |

### ThreadUpdate

Same fields as [ThreadCreate](#threadcreate), with the updated thread.

### ThreadDelete

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `thread_id` | int64 | Deleted thread ID |

### GuildEmojiCreate

| Field | Type | Description |
|-------|------|-------------|
| `emoji` | [GuildEmoji](#guildemoji) | Created emoji |

### GuildEmojiUpdate

Same fields as [GuildEmojiCreate](#guildemojicreate), with the updated emoji.

### GuildEmojiDelete

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | string | Guild ID |
| `emoji_id` | string | Deleted emoji ID |

### GuildMemberAdd

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `user_id` | int64 | User ID |
| `member` | [Member](#member) | Added member |

### GuildMemberUpdate

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `member` | [Member](#member) | Updated member |

### GuildMemberRemove

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `user_id` | int64 | Removed user ID |

### GuildMemberAddRole

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `role_id` | int64 | Role ID |
| `user_id` | int64 | User ID |

### GuildMemberRemoveRole

Same fields as [GuildMemberAddRole](#guildmemberaddrole).

### GuildMemberJoinVoice

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `user_id` | int64 | User ID |
| `channel_id` | int64 | Voice channel ID |

### GuildMemberLeaveVoice

Same fields as [GuildMemberJoinVoice](#guildmemberjoinvoice).

### GuildMemberModeration

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `user_id` | int64 | Target user ID |
| `actor_id` | int64 | Moderator user ID |
| `action` | string | `kick`, `ban`, or `unban` |
| `reason` | string | Optional reason |

### VoiceRegionChanging

| Field | Type | Description |
|-------|------|-------------|
| `channel_id` | int64 | Voice channel ID |
| `region` | string | Target region |
| `delay_ms` | integer | Suggested reconnect delay |

### VoiceStateUpdate

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `user_id` | int64 | User ID |
| `channel_id` | int64 | Voice channel ID |
| `mute` | boolean | Mute state |
| `deafen` | boolean | Deafen state |
| `self_video` | boolean | Camera state |

### GuildMemberStartStream

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `channel_id` | int64 | Channel ID |
| `user_id` | int64 | User ID |
| `stream` | [ActiveStream](#activestream) | Stream metadata |

### GuildMemberStopStream

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `channel_id` | int64 | Channel ID |
| `user_id` | int64 | User ID |
| `stream_id` | int64 | Stream ID |
| `reason` | string | Optional reason |

### GuildStreamsRebind

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Guild ID |
| `channel_id` | int64 | Channel ID |
| `stream_ids` | int64[] | Stream IDs |
| `jitter_ms` | integer | Optional reconnect jitter |

### GuildChannelMessage

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `channel_id` | int64 | Channel ID |
| `message_id` | int64 | Message ID |

### ChannelUserTyping

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `channel_id` | int64 | Channel ID |
| `user_id` | int64 | Typing user ID |

### Mention

| Field | Type | Description |
|-------|------|-------------|
| `guild_id` | int64 | Optional guild ID |
| `channel_id` | int64 | Channel ID |
| `message_id` | int64 | Message ID |
| `author_id` | int64 | Message author ID |
| `type` | integer | Mention type |

### DMMessage

| Field | Type | Description |
|-------|------|-------------|
| `channel_id` | int64 | DM channel ID |
| `message_id` | int64 | Message ID |
| `from` | [UserBrief](#userbrief) | Sender |
| `message` | [Message](#message) | Optional full message |

### UpdateReadState

| Field | Type | Description |
|-------|------|-------------|
| `channel_id` | int64 | Channel ID |
| `message_id` | int64 | Last read message ID |

### UpdateUserSettings

| Field | Type | Description |
|-------|------|-------------|
| `settings` | object | User settings object |

### IncomingFriendRequest

| Field | Type | Description |
|-------|------|-------------|
| `from` | [UserBrief](#userbrief) | Requesting user |

### FriendAdded

| Field | Type | Description |
|-------|------|-------------|
| `friend` | [UserBrief](#userbrief) | Friend user |

### FriendRemoved

Same fields as [FriendAdded](#friendadded).

### UpdateUser

| Field | Type | Description |
|-------|------|-------------|
| `user` | [User](#user) | Updated public user profile |

### UserAuthRevoked

| Field | Type | Description |
|-------|------|-------------|
| `session_version` | int64 | New session version |

## DM Call Payloads

### DMCallSummary

| Field | Type | Description |
|-------|------|-------------|
| `call_id` | int64 | Call ID |
| `channel_id` | int64 | DM channel ID |
| `caller_id` | int64 | Caller user ID |
| `recipient_id` | int64 | Recipient user ID |
| `region` | string | Optional voice region |
| `participants` | object | Optional map of user ID to state value |
| `started_at` | int64 | Unix timestamp |
| `solo_since` | int64 | Optional Unix timestamp |
| `dismissed` | boolean | Whether the call is dismissed |

### DMCallStarted

| Field | Type | Description |
|-------|------|-------------|
| `call` | [DMCallSummary](#dmcallsummary) | Call state |

### DMCallJoined

| Field | Type | Description |
|-------|------|-------------|
| `call` | [DMCallSummary](#dmcallsummary) | Call state |
| `user_id` | int64 | Joining user ID |

### DMCallDeclined

Same fields as [DMCallJoined](#dmcalljoined).

### DMCallLeft

Same fields as [DMCallJoined](#dmcalljoined).

### DMCallEnded

| Field | Type | Description |
|-------|------|-------------|
| `call` | [DMCallSummary](#dmcallsummary) | Call state |
| `reason` | string | Optional end reason |

### DMCallStreamStarted

| Field | Type | Description |
|-------|------|-------------|
| `call` | [DMCallSummary](#dmcallsummary) | Call state |
| `user_id` | int64 | Streaming user ID |
| `stream` | [ActiveStream](#activestream) | Stream metadata |

### DMCallStreamStopped

| Field | Type | Description |
|-------|------|-------------|
| `call` | [DMCallSummary](#dmcallsummary) | Call state |
| `user_id` | int64 | User ID |
| `stream_id` | int64 | Stream ID |
| `reason` | string | Optional reason |

## RTC Payloads

These are shared protocol payloads and are not currently handled as bot gateway client commands.

### VoiceMove

| Field | Type | Description |
|-------|------|-------------|
| `channel` | int64 | Target channel ID |
| `sfu_url` | string | SFU URL |
| `sfu_token` | string | SFU token |

### VoiceRebind

| Field | Type | Description |
|-------|------|-------------|
| `channel` | int64 | Channel ID |
| `jitter_ms` | integer | Optional reconnect jitter |
