[<- Documentation](../README.md) - [Voice](README.md)

# Direct-Message Calls

Direct-message calls are private 1:1 voice sessions for direct DM channels. They reuse the existing voice SFU and stream service, but their API, events, and permissions are scoped only to the two users in the DM pair. Group DMs are out of scope.

## Bootstrap And State

Active DM calls are returned in the normal settings bootstrap:

```http
GET /api/v1/user/me/settings
```

The response includes `dm_calls`, an array of active calls visible to the current user. There is no separate active-calls route; clients should hydrate from settings and then keep state fresh from private user WebSocket events.

Each call summary contains:

| Field | Meaning |
| --- | --- |
| `call_id` | Stable active call ID. |
| `channel_id` | Direct DM channel ID. |
| `caller_id` | User who started the call. |
| `recipient_id` | Other DM participant. |
| `region` | Voice region selected for the call route. |
| `participants` | Map of joined user ID to join timestamp. |
| `started_at` | Unix timestamp when the call started. |
| `solo_since` | Unix timestamp when only one participant remained, if applicable. |
| `dismissed` | Whether this call was locally dismissed for the current user. |

DM call state is not exposed through public presence. Only the caller and recipient can receive the call summary or call events.

## REST API

All endpoints require a normal user access token and only work for direct DM channels where the current user is one of the two participants.

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/user/me/channels/{channel_id}/call` | Start or resume a DM call and join the caller. |
| `POST` | `/user/me/channels/{channel_id}/call/join` | Join an active DM call. |
| `POST` | `/user/me/channels/{channel_id}/call/decline` | Dismiss the incoming call locally. |
| `DELETE` | `/user/me/channels/{channel_id}/call` | Leave the current call participation. |
| `GET` | `/user/me/channels/{channel_id}/call/streams` | List active streams in the DM call. |
| `POST` | `/user/me/channels/{channel_id}/call/streams` | Start or resume the caller's DM call stream. |
| `POST` | `/user/me/channels/{channel_id}/call/streams/{stream_id}/join` | Join an active DM call stream as viewer. |
| `DELETE` | `/user/me/channels/{channel_id}/call/streams/{stream_id}` | Stop the caller's own DM call stream. |

`POST /call` and `POST /call/join` return the call summary, selected `sfu_url`, `sfu_token`, and `region`.

## Lifecycle

- Starting a call creates or resumes the active call for that direct DM channel and joins the caller.
- Joining adds the current user to `participants` and returns the same SFU route selected for the call.
- Declining only hides the incoming call for that user. It does not end the call for the other participant.
- Leaving removes the current user from `participants`.
- If no participants remain, the call ends immediately.
- If one participant remains, `solo_since` is set and the call is cleaned up after 3 minutes.

The 3-minute solo cleanup applies only to DM calls. Guild voice channels keep their existing behavior and route lifetime model.

## Voice Region Selection

DM calls use the caller's `voice.preferred_region` user setting to select the initial route. When the caller setting is missing, empty, or `"auto"`, the API preserves the existing automatic region behavior. The recipient always joins the already selected route.

For guild voice channels, an explicit channel region still wins. If the channel is automatic, the first joining user's preferred region can select the route; `"auto"` preserves the existing default behavior.

## SFU And Stream Permissions

DM call media tokens grant only the media permissions needed for a private call:

- `PermVoiceConnect`
- `PermVoiceSpeak`
- `PermVoiceVideo`

DM calls do not grant guild moderation permissions such as mute members, deafen members, move members, kick, block, or administrator. The SFU and stream service should treat DM call tokens as media-scoped tokens for the two call participants.

## Screen Sharing

DM call streams use the same stream service as guild voice-channel streaming, but the scope is the DM call instead of a guild voice channel. Only the caller and recipient can start, list, join, or stop streams for that call.

DM stream tokens use the normal stream token shape with the DM call channel as `channel_id` and no guild permission surface. Publisher tokens can publish screen/app media; viewer tokens are receive-only.

## WebSocket Events

DM call events are sent over the Gateway WebSocket on the private `user.{userId}` topic for the two DM participants only:

| Type | Name | Purpose |
| --- | --- | --- |
| `408` | `UserDMCallStarted` | A DM call was started or resumed. |
| `409` | `UserDMCallJoined` | A participant joined the call. |
| `410` | `UserDMCallDeclined` | A participant dismissed the incoming call. |
| `411` | `UserDMCallLeft` | A participant left the call. |
| `412` | `UserDMCallEnded` | The call ended. |
| `413` | `UserDMCallStreamStarted` | A participant started or resumed screen sharing. |
| `414` | `UserDMCallStreamStopped` | A participant stopped screen sharing. |

See [WebSocket Event Types](../ws/EventTypes.md) for payload examples.
