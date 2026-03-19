[<- Documentation](../README.md) - [Channels](README.md)

# Threads

This document describes how threads work in gochat.

## Overview

Threads are channels of type `5` (`ChannelTypeThread`) that belong to a guild text channel and are anchored to exactly one source message.

Key properties:

- A thread can only be created from a message in a guild text channel.
- A source message can have only one thread.
- Threads cannot be created inside other threads.
- Threads inherit permissions from their parent channel.
- Messages inside threads receive the same `position` field as messages in any other channel.
- Threads are discoverable through the parent-channel thread listing endpoint and direct channel lookup by thread ID.
- Thread message events are delivered only to clients subscribed to the thread channel itself.
- Thread lifecycle events are delivered to guild subscribers both as generic channel lifecycle events and as dedicated thread lifecycle events.

## Channel Model

Thread channels use the normal `dto.Channel` payload with a few thread-specific fields:

```json
{
  "id": 2230469276416868352,
  "type": 5,
  "guild_id": 2230469276400000000,
  "creator_id": 2230469276300000000,
  "member": {
    "user_id": 2230469276300000000,
    "join_timestamp": "2026-03-11T10:00:00Z",
    "flags": 0
  },
  "member_ids": [
    2230469276300000000,
    2230469276300000001
  ],
  "name": "release discussion",
  "parent_id": 2230469276200000000,
  "position": 0,
  "topic": "Follow-up work for the release",
  "permissions": 274877906944,
  "private": false,
  "closed": false,
  "last_message_id": 2230469276500000000,
  "message_count": 2,
  "created_at": "2026-03-11T10:00:00Z"
}
```

Important fields:

| Field | Meaning |
|-------|---------|
| `type` | Always `5` for threads |
| `parent_id` | The guild text channel the thread belongs to |
| `creator_id` | User who created the thread |
| `member` | Current user's thread membership when the channel is returned over HTTP |
| `member_ids` | User IDs of members who have joined the thread, ordered by join time |
| `closed` | If `true`, the thread is read-only |
| `message_count` | Approximate thread message count returned as the stored Postgres base plus any pending KeyDB delta |

## Message Model

Four message fields are important for threads:

| Field | Meaning |
|-------|---------|
| `reference` | The source message referenced by a reply, thread-created message, or thread-initial message |
| `reference_channel_id` | The channel where the referenced source message lives |
| `thread_id` | The thread attached to a source message, or linked by a thread-created message |
| `thread` | Nested thread channel metadata, attached when the message points to a live thread |

Thread messages also carry the normal `position` field:

- it is channel-local to the thread
- it is assigned when the message is created
- it never changes after create
- it may have gaps if a reserved cache block is only partially used
- the copied thread-initial message gets the first thread position, the starter message gets the next one

When a thread is created:

1. The original source message in the parent channel keeps `type = 0`, but is updated so:
   - `thread_id` points to the new thread
   - `thread` contains the nested thread channel metadata
   - clients receive a normal `Message Update` event for that source message with both fields attached
2. The first message inside the thread is a copy of the source message, stored in the thread as a thread-initial message.
   - it preserves the original author, content, attachments, embeds, and flags
   - it points back to the source message with `reference` and `reference_channel_id`
   - it is informational and cannot be edited
3. The second message inside the thread is the new user-authored starter message supplied in the create-thread request.
4. An informative thread-created message is posted in the parent channel with:
   - `reference = source message id`
   - `reference_channel_id = parent channel id`
   - `thread_id = new thread id`
   - `type = 3` (`Thread Created`)
   - `content = current thread name`
   - the message is informational and cannot be edited
5. Message payloads for:
   - the original type-0 source message in the parent channel
   - the type-3 thread-created message in the parent channel
   include both `thread_id` and nested `thread` metadata so clients can render/open the thread without another lookup.
   - nested `thread` metadata includes `member_ids`
   - nested `thread` metadata includes `message_count`
   - nested `thread` metadata does not include current-user `member` state because message payloads are shared

## Creating Threads

### Endpoint

```http
POST /message/channel/{channel_id}/{message_id}/thread
```

- `channel_id`: parent guild text channel
- `message_id`: source message inside that channel

### Request Body

```json
{
  "name": "release discussion",
  "content": "Let's track the rollout here",
  "nonce": "draft-1",
  "attachments": [],
  "mentions": [],
  "embeds": []
}
```

Rules:

- The thread starter payload cannot be empty.
- The request must include message content, attachments, or embeds.
- `name` is optional.
- `nonce` is optional and applies only to the creator's starter message inside the thread.
- `enforce_nonce` is not supported on thread creation. The create-thread route accepts `nonce` only as a client correlation token for the starter event.
- If `nonce` is set, only the author's own thread `Message Create` event receives it. Other thread subscribers receive the same starter message without `nonce`.
- If `name` is empty, the server uses the source message content.
- If the source message content is empty, the server falls back to the starter message content.
- Thread names are capped at 256 characters.

### Starter Attachments

Attachments for the first thread message are uploaded against the parent channel first, then copied into the new thread during creation.

### Concurrency

Thread creation is atomic with respect to the source message:

- Only one request can attach a thread to a given source message.
- If two users try to create a thread from the same message at the same time, one succeeds and the other receives `409 Conflict`.

## Discovery

`GET /guild/{guild_id}/channel` returns only regular guild channels and does not include threads.

Threads can be discovered in two ways:

- `GET /guild/{guild_id}/channel/{channel_id}/threads`
  Returns threads whose `parent_id` matches the given channel.
- `GET /guild/{guild_id}/channel/{thread_id}`
  Fetches a specific thread directly when the client already knows the thread ID.
  Returns `404 Not Found` if the thread does not exist in that guild or was deleted.

## Message Positions

Message positions are not thread-specific; they exist in any channel type, including parent guild channels and threads.

Implementation details:

- each channel has a durable Postgres `message_position` high watermark
- the API allocates message positions from KeyDB-backed blocks to avoid a database write for every message
- if cache state is lost, allocation resumes from the durable channel watermark, so positions stay monotonic
- unused reserved values may be skipped after restart or cache loss; clients must treat `position` as ordered navigation metadata, not as a contiguous row number
- historical messages can be backfilled with `gctools messages backfill-positions`

## Lifecycle Events

Guild subscribers receive two parallel lifecycle streams for threads:

- generic channel lifecycle events:
  - `Channel Create` (`t=106`)
  - `Channel Update` (`t=107`)
  - `Channel Delete` (`t=109`)
- dedicated thread lifecycle events:
  - `Thread Create` (`t=113`)
  - `Thread Update` (`t=114`)
  - `Thread Delete` (`t=115`)

Both streams describe the same thread lifecycle changes. The dedicated thread events exist so clients can handle thread state separately without filtering regular guild channel events by `channel.type = 5`.

## Membership

Threads keep explicit membership for notification and unread behavior.

Rules:

- The thread creator is automatically joined when the thread is created.
- Sending a message in a thread automatically joins the sender if they were not already joined.
- Leaving a thread stops future thread activity notifications for that user, but does not remove ordinary access to read the thread if the parent channel is still visible.
- Thread channel payloads include `member_ids` so clients can render who has joined the thread without a separate member-list request.
- The same `member_ids` list is also attached in shared nested `message.thread` metadata for source messages and thread-created messages.

### Membership Endpoints

```http
PUT /guild/{guild_id}/channel/{thread_id}/thread-member/me
DELETE /guild/{guild_id}/channel/{thread_id}/thread-member/me
```

Both endpoints require the user to be able to view the thread through the parent guild channel.

## Permissions

Threads inherit the permission model of the parent guild text channel.

### Create

Creating a thread requires:

- guild membership
- `PermTextCreateThreads`, or an equivalent higher privilege such as administrator/server owner

### Read

Reading a thread requires the same channel visibility and read-history access as the parent channel.

### Send

Sending inside a thread requires:

- inherited access to the parent channel
- `PermTextSendMessageInThreads`
- `closed = false`

### Manage

The following users can manage a thread:

- the thread creator
- users with `PermTextManageThreads`
- users with higher privileges implied by the permission system, such as administrator/server owner

Management actions include:

- rename thread
- update topic
- close thread
- reopen thread
- delete thread

## Closing Threads

Threads are closed by patching the channel:

```http
PATCH /guild/{guild_id}/channel/{thread_id}
```

Example:

```json
{
  "closed": true
}
```

Behavior of closed threads:

- old messages remain readable
- sending new messages is rejected
- typing events are rejected
- new attachment uploads for thread messages are rejected
- message edits and deletes inside the closed thread are rejected

## Updating Name and Topic

Thread metadata is updated through the normal guild channel patch route:

```http
PATCH /guild/{guild_id}/channel/{thread_id}
```

Example:

```json
{
  "name": "release qa",
  "topic": "QA findings and rollout blockers"
}
```

Notes:

- Threads do not accept `private` updates.
- Thread permissions cannot diverge from the parent channel.
- When the thread name changes, the parent thread-created message content is updated to the new thread name.
- That rename also emits a normal `Message Update` event for the parent-channel type-3 thread-created message.

## Deleting Threads

Threads are deleted through the normal guild channel delete route:

```http
DELETE /guild/{guild_id}/channel/{thread_id}
```

Deleting a thread removes the channel and its stored thread messages.

## Notifications And Unread

gochat now separates thread message delivery from thread activity notifications:

- Live thread message events still go only to clients subscribed to `channel.{threadId}`.
- Thread activity notifications are sent only to joined thread members.
- `guilds_last_messages` in user settings continues to exclude threads.
- Joined thread unread state is exposed separately through `threads_last_messages`.
- Joined thread membership layout is exposed through `joined_threads`, shaped as `guild_id -> parent_channel_id -> [thread_id, ...]`, with each thread ID list sorted ascending.
- Direct user mentions inside a thread still notify the explicitly mentioned user.
- Thread `@role`, `@everyone`, and `@here` notifications are scoped to users who are currently joined to that thread.

## Message Count

Thread channel payloads expose `message_count` as an approximate stored counter.

Behavior:

- A new thread persists an initial count of `2` in Postgres because it contains the thread-initial copy plus the creator's starter message.
- New thread messages do not update Postgres immediately. They append a delta in KeyDB instead.
- Reads return the stored Postgres base plus any pending KeyDB delta, so active threads show fresh counts without a SQL write per message.
- A background flusher periodically merges cached KeyDB deltas back into Postgres.
- Pending KeyDB deltas survive normal API process restarts and are flushed by a later API instance, so counters are eventually persisted instead of being lost on every restart.
- Message deletes are not applied immediately to the counter, so `message_count` should be treated as approximate.
- Older threads created before this field existed may start at `0` until new activity or a future backfill updates the stored base count.

## WebSocket Behavior

Thread lifecycle uses the normal guild channel events:

- create thread -> `Channel Create` on `guild.{guildId}` with `channel.type = 5`
- update thread -> `Channel Update` on `guild.{guildId}` with `channel.type = 5`
- delete thread -> `Channel Delete` on `guild.{guildId}` with `channel_type = 5`

Thread creation also emits message events:

- source message update on `channel.{parentChannelId}` because `thread_id` is attached to the source message
- parent thread-created message create on `channel.{parentChannelId}` with `type = 3`
- thread-initial message create on `channel.{threadId}`
- starter thread message create on `channel.{threadId}`
  The starter event includes `nonce` only for the author when the create-thread request provided one.

Important subscription rule:

- subscribing to the parent channel does not subscribe the client to thread message traffic
- to receive new messages in a thread, the client must subscribe to `channel.{threadId}`
- to receive live events from both the parent channel and the thread, the client should send both IDs in the WebSocket OP 5 `channels` list

Important notification rule:

- regular guild-channel activity events are guild-scoped
- thread activity events are user-scoped and delivered only to joined thread members
- joining a thread affects notifications and unread summaries, not raw thread message transport
