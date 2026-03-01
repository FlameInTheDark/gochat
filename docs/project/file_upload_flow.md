# GoChat File Upload Flow

This document outlines the file upload architecture and flow for the GoChat project. The file handling is split across the main `api` service and a dedicated `attachments` microservice to offload heavy payload processing and streaming.

## High-Level Architecture

The upload system is designed around a two-step process:
1. **Metadata Allocation (Main API)**: The client requests to create an attachment/avatar/icon. The API returns an ID and placeholder metadata.
2. **Binary Upload (Attachments Service)**: The client performs a `POST` request with the actual file binary to the `attachments` microservice, which processes the file, uploads it to S3-compatible storage, and finalizes the record.

## Detailed Flows

### 1. Message Attachments (`/api/v1/message/channel/{channel_id}/attachment`)
**Endpoint**: `POST /upload/attachments/{channel_id}/{attachment_id}` (Attachments Service)

* **Initiation**: The main API validates upload limits and channel permissions, generates an attachment ID, and creates a pending record in the database.
* **Upload**: The client posts the `multipart/form-data` or binary stream to the attachments service.
* **Processing**:
  * Peeks at the payload to determine `Content-Type`.
  * Uploads the original file directly to S3 via an S3 Presigned URL.
  * **Images & Videos**: If the file is an image or video, it runs `ffmpeg` to generate a 350x350 WebP preview frame (using `ffmpegExtractWebP`) and uploads this preview to S3. It also probes the original dimensions using `ffprobe`.
* **Finalization**: Validates that the total size doesn't exceed the user's limit *after* upload. It then marks the attachment as `done` in the database, saves dimensions, and returns the public URLs.

### 2. User Avatars (`/api/v1/user/me/avatar`)
**Endpoint**: `POST /upload/avatars/{user_id}/{avatar_id}` (Attachments Service)

* **Initiation**: The API creates a pending avatar record in Cassandra.
* **Upload**: Client posts the binary.
* **Processing**:
  * Ensures the `Content-Type` is an image.
  * Streams the image directly through `ffmpeg` (`ffmpegToWebPStreamLimited`) to convert it to a WebP format scaled to a maximum of 128x128 pixels, enforcing a stricter file size limit (max 250KB limit passed directly to ffmpeg `-fs`).
  * Uploads ONLY the optimized WebP to S3 (no original stored).
* **Finalization**: Saves dimensions (probed from image decode or `ffprobe`), marks the avatar as `done`, sets it as the user's active avatar, and triggers a user update event via the NATS message queue.

### 3. Guild Icons (`/api/v1/guild/{guild_id}/icon`)
**Endpoint**: `POST /upload/icons/{guild_id}/{icon_id}` (Attachments Service)

* **Initiation**: Validates if the user is the guild owner and creates a pending icon.
* **Upload**: Client streams the binary.
* **Processing**: Similar to avatars, streams through `ffmpegToWebPStreamLimited` to produce a 128x128 WebP image, capped at 250KB.
* **Finalization**: Uploads to S3, finalizes the row in Cassandra, assigns the icon to the Guild in PostgreSQL, and broadcasts a NATS message `UpdateGuild` to all connected clients natively.

## Storage Backend

* **Backend**: S3-compatible object storage (e.g., AWS S3, MinIO).
* **Client**: Uses the `github.com/FlameInTheDark/gochat/internal/s3` wrapper.
* **Delivery**: Files are served externally using the base URL defined in the configuration by `S3ExternalURL`. Previews and originals are assembled using this base and their respective S3 Object Keys.
* **Keys format**:
  * Attachments: `media/{channel_id}/{attachment_id}/{filename}` and `media/{channel_id}/{attachment_id}/preview.webp`
  * Avatars: `avatars/{user_id}/{avatar_id}.webp`
  * Icons: `icons/{guild_id}/{icon_id}.webp`

## Post-Processing Utilities

The attachment microservice relies heavily on local binaries for fast processing:
* `ffmpeg`: Used for scaling `force_original_aspect_ratio=decrease`, extracting video frames to a `image2pipe` stream, and forcing `.webp` encoding.
* `ffprobe`: Used for extracting intrinsic dimensions (`width`, `height`) of videos and large images.
