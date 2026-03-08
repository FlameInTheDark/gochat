[<- Documentation](../../README.md) - [Project Documentation](README.md)

# Presence System

This document outlines the architecture and workflows for user presence, including general online status and voice channel presence (with speech detection).

## 1. General Presence Statuses

Users can have one of the following basic presence statuses, defined in `internal/presence/model.go`:
- `online`
- `idle`
- `dnd` (Do Not Disturb)
- `offline`

In addition, a custom text status (`custom_status_text`) can be set.

### 1.1 Storage and Aggregation
Presence is tracked per-session to support multiple connected devices simultaneously (e.g. mobile phone and desktop application):
- **Session Presence:** Stored in Redis as a hash (`presence:sessions:{userID}`), where each field is a `sessionID`. Each session has a TTL that is constantly refreshed by the WebSocket Gateway.
- **Aggregation:** When queried via `store.Aggregate()`, the system aggregates all active sessions for a user. It chooses the "priority" active status among them:
  1. `dnd` has the highest priority. If any session has `dnd`, the aggregative status is `dnd`.
  2. `online` is next.
  3. `idle` follows.
  4. If no active sessions remain, the status is `offline`.
- **Overrides:** Users can manually override their status globally (e.g., sticking to `dnd` or going "Invisible"). Overrides take absolute precedence and are stored in Redis under `presence:override:{userID}`.

### 1.2 Updating Presence
Clients send a `PresenceUpdateRequest` over their WebSocket connection as a regular websocket event payload. When the WS gateway receives it, it calls `store.UpsertSession` to update the session presence in Redis.

### 1.3 Broadcasting Presence
Other users are informed of presence changes via the Notification System (NATS). The WS gateway broadcasts `OP 3` Dispatch message to subscribers who have requested presence updates via OP 6 (`PresenceSubscription`).

The `OP 3` payload looks like:
```json
{
  "user_id": 2226021950625415200,
  "status": "online",
  "custom_status_text": "Coding...",
  "since": 1700000000,
  "client_status": {
    "web": "online",
    "desktop": "idle"
  }
}
```

---

## 2. Voice Channel Presence

When a user joins a voice channel, their presence is augmented with a `voice_channel_id` which designates the channel they are in.

### 2.1 Voice Channel Join Flow
1. **API Call:** The client requests to join a voice channel via the REST API (`/guild/{id}/voice/{channel_id}/join`). This returns a short-lived SFU JWT token.
2. **SFU Connection:** The client connects to the SFU server via WebSocket (`/signal`) and performs the `RTCJoin` handshake.
3. **Webhook Callback:** Upon successful WebRTC connection, the SFU sends an internal HTTP webhook (`/api/v1/webhook/sfu/voice/join`) to the GoChat API.
4. **Presence & Guild Events:** The API processes this webhook, marking the user's presence as being in `voice_channel_id`. It publishes a `GuildMemberJoinVoice` (Type 205) event over NATS to all guild members.
5. **Presence Update:** The WS gateway updates the session's voice channel in Redis via `SetSessionVoiceChannel` and triggers a new `OP 3` Presence Update broadcast, which will now include `"voice_channel_id": 2230469276416868352`.

### 2.2 Speech Detection (Speaking Indicator)

Inside a voice channel, users frequently start and stop talking. This high-frequency "speaking" state is handled entirely by the SFU signaling plane and is **not** persisted to Redis or broadcasted over the main Gateway WS, as that would overwhelm the general event bus.

Speech detection follows the **SFU WebSocket Protocol**:

1. **Client Event:** When a user's client detects audio input (e.g., via volume thresholding within WebRTC, or browser mic activity), it sends a simple JSON event directly to the SFU's `/signal` WebSocket:
   ```json
   { "event": "speaking", "data": "1" } 
   ```
   *(data is "1" when speaking starts, and "0" when it stops)*

2. **Broadcasting:** The SFU receives this, updates its internal state, and immediately broadcasts it to all other peers connected to the same voice channel using the standard `op/t/d` envelope (Type 514 - `RTCSpeaking`):
   ```json
   {
     "op": 7,
     "t": 514,
     "d": {
       "user_id": 123456,
       "speaking": 1
     }
   }
   ```

3. **Client UI:** Receiving clients intercept this OP 7 message, map the `user_id` to their voice channel component, and highlight the speaker's avatar (e.g., adding a green ring) to indicate active speech.
