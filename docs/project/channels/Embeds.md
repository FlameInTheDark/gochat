[<- Channels](README.md)
# Message Embeds

GoChat supports Discord-like message embeds. A message can contain multiple embeds, and each embed follows the same high-level shape as a Discord embed object.

## Overview

Embeds in GoChat are split into two sources:

- `embeds`: manual embeds provided directly by the client in `SendMessageRequest` or `MessageUpdateRequest`.
- `auto_embeds`: generated embeds produced by the `embedder` service after URL unfurling.

API responses and WebSocket events do not expose these two arrays separately. Clients always receive a single merged `embeds` array.

## Message Lifecycle

### Manual embeds

- `POST /api/v1/message/channel/{channel_id}` accepts `embeds` as a full array of manual embeds.
- `PATCH /api/v1/message/channel/{channel_id}/{message_id}` accepts `embeds` as a full replacement for the manual embed array.
- Sending `"embeds": []` removes all manual embeds from the message.

### Generated embeds

When message content contains one or more URLs:

1. The API stores the message immediately.
2. The API publishes an `embed.make` NATS event.
3. The `embedder` service fetches metadata from the URL.
4. The service builds embeds from YouTube, discovered oEmbed, Open Graph, or Twitter Card metadata.
   - For YouTube videos, `youtube_embed_base_url` controls whether generated embeds use the standard `https://www.youtube.com/embed/VIDEO_ID` form or a privacy-friendly base such as `https://www.youtube-nocookie.com/embed/VIDEO_ID`.
5. The generated embeds are saved into `auto_embeds`.
6. The service publishes a normal `MessageUpdate` event so connected clients receive the new embeds.

Generated embeds are cleared when:

- the message content changes,
- the user enables embed suppression,
- the worker re-generates embeds and no longer finds valid results.

## Flags

`MessageUpdateRequest.flags` is a bitmask.

Currently defined flags:

- `4` (`1 << 2`): suppress generated embeds for this message.

When the suppress flag is set:

- existing generated embeds are removed,
- future URL unfurling is skipped,
- manual embeds remain untouched.

When the flag is removed and the message content still contains URLs, the API re-enqueues embed generation.

## Validation Rules

GoChat validates embeds on the backend before storing them.

### Limits

- Maximum embeds per message: `10`
- Maximum fields per embed: `25`
- Maximum combined text across all embeds in a message: `6000` characters
- Title length: `256`
- Description length: `4096`
- Footer text length: `2048`
- Author name length: `256`
- Field name length: `256`
- Field value length: `1024`
- Color range: `0..16777215`

### Provider-specific generated colors

Generated embeds may set a provider color automatically:

- YouTube embeds use red: `#FF0000`
- Twitter/X embeds use blue: `#1DA1F2`

### Allowed embed types

- `rich`
- `image`
- `video`
- `gifv`
- `article`
- `link`

### Allowed URL schemes

Embed URLs accept:

- `http`
- `https`
- `attachment` for fields where Discord-style attachment URLs make sense

## Supported Structure

Each embed may contain:

- `title`
- `type`
- `description`
- `url`
- `timestamp`
- `color`
- `footer`
- `image`
- `thumbnail`
- `video`
- `provider`
- `author`
- `fields`

Nested objects follow Discord naming closely:

- `footer`: `text`, `icon_url`, `proxy_icon_url`
- `author`: `name`, `url`, `icon_url`, `proxy_icon_url`
- `provider`: `name`, `url`
- `image` / `thumbnail` / `video`: `url`, `proxy_url`, `width`, `height`, `content_type`, `placeholder`, `placeholder_version`, `flags`
- `fields[]`: `name`, `value`, `inline`

## Response Semantics

The message DTO returned by REST and WebSocket uses this shape:

```json
{
  "id": "2230469276416868352",
  "channel_id": "2230469276416868352",
  "content": "https://www.youtube.com/watch?v=OgfdyH4iaps",
  "embeds": [
    {
      "type": "video",
      "url": "https://www.youtube.com/watch?v=OgfdyH4iaps",
      "title": "Why is Microsoft updating their text editors!? | TheStandup",
      "provider": {
        "name": "YouTube",
        "url": "https://www.youtube.com"
      },
      "thumbnail": {
        "url": "https://i.ytimg.com/vi/OgfdyH4iaps/maxresdefault.jpg"
      },
      "video": {
        "url": "https://www.youtube.com/embed/OgfdyH4iaps"
      }
    }
  ],
  "flags": 0,
  "type": 0
}
```

`embeds` is always the merged view:

- manual embeds first,
- generated embeds after them,
- capped to the global message embed limit.

## Example: Manual embed on send

```json
{
  "content": "Release notes",
  "embeds": [
    {
      "type": "rich",
      "title": "GoChat 1.0",
      "description": "Embed support is live.",
      "color": 65280,
      "fields": [
        {
          "name": "Status",
          "value": "Stable",
          "inline": true
        }
      ]
    }
  ]
}
```

## Example: Remove generated embeds and prevent regeneration

```json
{
  "embeds": [],
  "flags": 4
}
```

This request removes manual embeds, clears generated embeds, and prevents future URL unfurling for the message until the flag is removed.
