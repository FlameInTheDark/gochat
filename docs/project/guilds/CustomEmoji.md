[<- Documentation](../README.md) - [Guilds](README.md)

# Custom Guild Emoji

Custom guild emoji let a guild upload image-based expressions that are referenced as `<:name:id>` in messages. The asset itself is fetched only by `emoji_id`, while permission to insert that emoji into a message is checked against membership in the source guild.

## Design Goals

- Global asset URL by `emoji_id`
- Guild-local unique names
- DB-free public media fetch
- Guild-scoped upload and moderation permissions
- Cached metadata reads for compose-time validation and user settings

## Storage Model

Emoji metadata lives in PostgreSQL/Citus.

- `guild_emojis` is distributed by `guild_id` and colocated with guild relational data.
- `emoji_lookup` is distributed by `id` for direct lookup by `emoji_id` during message compose.

This data intentionally stays in Citus instead of ScyllaDB because the feature depends on relational guarantees: guild-local unique names, per-guild static and animated quotas, guild membership checks, and guild-scoped delete and rename operations.

The attachments service stores three deterministic S3 objects per emoji:

- `emojis/{emoji_id}/master.webp`
- `emojis/{emoji_id}/96.webp`
- `emojis/{emoji_id}/44.webp`

`guild_id` is intentionally not part of the object key. `emoji_id` is globally unique and the public fetch path must stay cheap.

On guild delete, the API loads all emoji IDs from Citus, removes their metadata rows, and deletes these deterministic object keys from storage.

## Name Rules and Limits

- Allowed characters: `A-Z`, `a-z`, `0-9`, `-`
- Validation regex: `^[A-Za-z0-9-]+$`
- Names are unique per guild after lowercase normalization
- Max declared upload size: `256 KB`
- Max source dimensions: `128x128`
- Max ready static emojis per guild: `50`
- Max ready animated emojis per guild: `50`
- Max total active rows per guild (ready plus unexpired pending): `100`

Accepted uploads are converted to WebP. Animated GIF, APNG, and animated WebP inputs stay animated after conversion.

## Permissions

- Upload requires guild owner, `Administrator`, or `Create Expressions`
- Rename and delete require guild owner, `Administrator`, or `Manage Expressions`
- `Manage Expressions` does not allow upload
- Neither permission is granted by the default guild permission set

## API Flow

### 1. Create placeholder

`POST /api/v1/guild/{guild_id}/emojis`

Request:

```json
{
  "name": "party-cat",
  "file_size": 182331,
  "content_type": "image/gif"
}
```

Response:

```json
{
  "id": "2230469276416868352",
  "guild_id": "2226022078304223200",
  "name": "party-cat"
}
```

This reserves the emoji ID, validates the name, checks guild-local uniqueness, and creates a pending row with an upload expiration time.

### 2. Upload binary

`POST /api/v1/upload/emojis/{guild_id}/{emoji_id}`

- Send the image as the raw request body
- The attachments service validates content type, declared size, actual dimensions, and placeholder expiry
- The binary is converted to WebP and three variants are written
- Final quota enforcement happens after animation detection, because static and animated limits are separate

A successful upload returns `201 Created`. Re-uploading an already finalized emoji returns `204 No Content`.

### 3. List ready emojis

`GET /api/v1/guild/{guild_id}/emojis`

Only guild members can list emojis. The response contains ready emojis only:

```json
[
  {
    "id": "2230469276416868352",
    "guild_id": "2226022078304223200",
    "name": "party-cat",
    "animated": true
  }
]
```

### 4. Rename or delete

- `PATCH /api/v1/guild/{guild_id}/emojis/{emoji_id}`
- `DELETE /api/v1/guild/{guild_id}/emojis/{emoji_id}`

Rename request:

```json
{
  "name": "party-cat-fast"
}
```

Delete removes both Citus rows and all three object variants.

## Public Asset Route

`GET /emoji/{emoji_id}.webp?size={n}`

This route is intentionally hot-path friendly:

- no auth
- no Citus lookup
- no KeyDB lookup
- direct redirect to the deterministic S3 or CDN object key

Size selection:

- no `size`: use `master.webp`
- exact `44` or `96`: use that variant
- other positive values: choose the closest of `44`, `96`, and `master`
- `master` is treated as `128` for distance comparisons
- ties prefer the larger variant

Examples:

- `/emoji/2230469276416868352.webp`
- `/emoji/2230469276416868352.webp?size=44`
- `/emoji/2230469276416868352.webp?size=80` redirects to `96.webp`

If the object does not exist because the emoji is still pending or has been deleted, object storage returns `404`.

## User Settings

`GET /api/v1/user/me/settings` now includes a top-level `guild_emojis` map:

```json
{
  "guild_emojis": {
    "2226022078304223200": [
      {
        "name": "party-cat",
        "id": "2230469276416868352"
      }
    ],
    "2226022078304223201": []
  }
}
```

Notes:

- The Go type is `map[int64][]EmojiRef`
- JSON object keys are strings on the wire
- Only ready emojis from guilds the user belongs to are included
- No separate `emoji_storage_url` setting is needed because `/emoji/{emoji_id}.webp` is the public fetch entry point

## Message Compose Rules

The backend only rewrites canonical custom emoji tags:

- accepted form: `<:name:id>`
- bare `:name:` is left as plain text

On message create and update:

1. Parse each `<:name:id>` tag
2. Resolve `emoji_id` through KeyDB or `emoji_lookup`
3. Verify the emoji is ready
4. Verify the sender is still a member of the emoji's source guild

Sanitization result:

- accessible emoji: rewrite to canonical `<:stored-name:id>`
- missing, pending, deleted, or inaccessible emoji: downgrade to `:original-name:`

This lets clients optimistically insert emoji tags while the API remains the authority for access control.

## Cache Behavior

KeyDB is used only for metadata reads that are hot but still permissioned:

- `emoji:id:{emojiId}` for compose-time lookup
- `emoji:guild:{guildId}` for guild emoji lists and user settings

Default TTLs:

- lookup cache: `3600s`
- negative lookup cache: `60s`
- guild list cache: `600s`

Cache entries are invalidated on finalize, rename, delete, and guild delete.

## Realtime Events

Guild subscribers receive:

- `Guild Emoji Create`
- `Guild Emoji Update`
- `Guild Emoji Delete`

See [WebSocket Event Types](../ws/EventTypes.md) for event numbers and payloads.
