# GoChat File Upload Flow

This document outlines the file upload architecture and flow for the GoChat project. File handling is split across the main `api` service and a dedicated `attachments` microservice to offload heavy payload processing and streaming.

## High-Level Architecture

The upload system is designed around a two-step process:
1. **Metadata Allocation (Main API):** The client requests to create an attachment, avatar, icon, or guild emoji. The API returns an ID and placeholder metadata.
2. **Binary Upload (Attachments Service):** The client performs a `POST` request with the actual file body to the `attachments` microservice, which processes the file, uploads it to S3-compatible storage, and finalizes the record.

## Detailed Flows

### 1. Message Attachments (`/api/v1/message/channel/{channel_id}/attachment`)
**Endpoint:** `POST /api/v1/upload/attachments/{channel_id}/{attachment_id}` (Attachments Service)

- **Initiation:** The main API validates upload limits and channel permissions, generates an attachment ID, and creates a pending record in the database.
- **Upload:** The client posts `multipart/form-data` or a binary stream to the attachments service.
- **Processing:**
  - Peeks at the payload to determine `Content-Type`.
  - Uploads the original file directly to S3 via an S3 presigned URL.
  - **Images and Videos:** If the file is an image or video, it runs `ffmpeg` to generate a `350x350` WebP preview frame by using `ffmpegExtractWebP`. It also probes the original dimensions by using `ffprobe`.
- **Finalization:** Validates that the total size does not exceed the user's limit after upload. It then marks the attachment as `done` in the database, saves dimensions, and returns public URLs.

### 2. User Avatars (`/api/v1/user/me/avatar`)
**Endpoint:** `POST /api/v1/upload/avatars/{user_id}/{avatar_id}` (Attachments Service)

- **Initiation:** The API creates a pending avatar record in Cassandra.
- **Upload:** The client posts the binary body.
- **Processing:**
  - Ensures the `Content-Type` is an image.
  - Streams the image directly through `ffmpeg` by using `ffmpegToWebPStreamLimited` to convert it to WebP scaled to a maximum of `128x128` pixels, enforcing a stricter size limit of `250 KB`.
  - Uploads only the optimized WebP to S3. The original file is not stored.
- **Finalization:** Saves dimensions, marks the avatar as `done`, sets it as the user's active avatar, and triggers a user update event via NATS.

### 3. Guild Icons (`/api/v1/guild/{guild_id}/icon`)
**Endpoint:** `POST /api/v1/upload/icons/{guild_id}/{icon_id}` (Attachments Service)

- **Initiation:** Validates that the caller is the guild owner and creates a pending icon.
- **Upload:** The client streams the binary body.
- **Processing:** Similar to avatars, the service streams the image through `ffmpegToWebPStreamLimited` to produce a `128x128` WebP image capped at `250 KB`.
- **Finalization:** Uploads the optimized image to S3, finalizes the row in Cassandra, assigns the icon to the guild in PostgreSQL, and broadcasts a guild update event.

### 4. Guild Emoji (`/api/v1/guild/{guild_id}/emojis`)
**Endpoint:** `POST /api/v1/upload/emojis/{guild_id}/{emoji_id}` (Attachments Service)

- **Initiation:** The main API validates the emoji name, declared file size, content type, guild-local uniqueness, upload permission, and the per-guild active placeholder cap before creating a pending row in Citus.
- **Upload:** The client posts the image as the raw request body.
- **Processing:**
  - Validates that the actual payload is an image and that the source dimensions do not exceed `128x128`.
  - Detects animation for GIF, APNG, and animated WebP inputs.
  - Converts the upload to WebP. Animated inputs are encoded with animated WebP output.
  - Produces three variants: `master`, `96`, and `44`.
  - Stores them with deterministic ID-only keys:
    - `emojis/{emoji_id}/master.webp`
    - `emojis/{emoji_id}/96.webp`
    - `emojis/{emoji_id}/44.webp`
- **Finalization:**
  - Enforces the final per-guild quota after animation detection: `50` static and `50` animated ready emoji.
  - Marks the emoji as `done` in Citus and updates the `emoji_lookup` row used for compose-time validation.
  - Invalidates KeyDB caches for `emoji:id:{emojiId}` and `emoji:guild:{guildId}`.
  - Broadcasts `Guild Emoji Create` after the upload is finalized.
- **Failure handling:** If processing, storage upload, or finalization fails, the service deletes any uploaded objects and removes the placeholder rows.

### Public Guild Emoji Delivery
**Endpoint:** `GET /emoji/{emoji_id}.webp?size={n}`

This route is public and intentionally hot-path friendly:
- No auth check
- No Citus lookup
- No KeyDB lookup
- Immediate redirect to the deterministic S3 or CDN object URL

Variant selection rules:
- no `size`: `master.webp`
- `size=44`: `44.webp`
- `size=96`: `96.webp`
- any other positive size: choose the closest of `44`, `96`, and `master` where `master` is treated as `128`
- invalid or non-positive `size`: `master.webp`

Because the route only redirects to deterministic object keys, pending or deleted emoji naturally return `404` from object storage if the file does not exist.

## Storage Backend

- **Backend:** S3-compatible object storage such as AWS S3 or MinIO.
- **Client:** Uses the `github.com/FlameInTheDark/gochat/internal/s3` wrapper.
- **Delivery:** Files are served externally using the base URL defined by `S3ExternalURL`. Public URLs are assembled from that base URL and the object key.
- **Key format:**
  - Attachments: `media/{channel_id}/{attachment_id}/original` and `media/{channel_id}/{attachment_id}/preview.webp`
  - Avatars: `avatars/{user_id}/{avatar_id}.webp`
  - Icons: `icons/{guild_id}/{icon_id}.webp`
  - Guild emoji: `emojis/{emoji_id}/master.webp`, `emojis/{emoji_id}/96.webp`, and `emojis/{emoji_id}/44.webp`

## Post-Processing Utilities

The attachments microservice relies heavily on local binaries for fast processing:
- `ffmpeg`: Used for scaling, preview extraction, static WebP encoding, and animated WebP encoding.
- `ffprobe`: Used for extracting intrinsic dimensions of videos and large images.
