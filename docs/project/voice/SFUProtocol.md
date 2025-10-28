[<- Documentation](../README.md) - [Voice](README.md)

# SFU WebSocket Protocol

This document describes the signaling protocol between clients and the SFU. It mirrors the GoChat WS message envelope to minimize payload size and keep a consistent client surface.

Envelope

- `op` (int): Operation code (e.g., OPCodeRTC)
- `t` (int): Event/subtype under the operation (for RTC)
- `d` (object): Event payload

Operation Codes

- `op = 7` — RTC signaling and voice control (OPCodeRTC)
- `op = 2` — Heartbeat/ping (compatible ping/pong)

RTC Events (op = 7)

| t    | Name                         | Direction         | Purpose                                                         |
|------|------------------------------|-------------------|-----------------------------------------------------------------|
| 500  | RTCJoin                      | Client → SFU      | Join a voice channel using short‑lived token                    |
| 501  | RTCOffer                     | Client/SFU        | SDP offer for negotiation                                       |
| 502  | RTCAnswer                    | Client/SFU        | SDP answer for negotiation                                      |
| 503  | RTCCandidate                 | Client/SFU        | Trickle ICE candidate                                           |
| 504  | RTCLeave                     | Client → SFU      | Leave/close voice session                                       |
| 505  | RTCMuteSelf                  | Client → SFU      | Toggle publishing of local microphone                           |
| 506  | RTCMuteUser                  | Client → SFU      | Local mute/unmute a specific remote user                        |
| 507  | RTCServerMuteUser            | Client → SFU      | Privileged: mute a user for everyone                            |
| 508  | RTCServerDeafenUser          | Client → SFU      | Privileged: deafen a user (receive no one)                      |
| 510  | RTCServerKickUser            | Client → SFU      | Privileged: kick a user from the room                           |
| 511  | RTCServerBlockUser           | Client → SFU      | Privileged: block/unblock user from the room                    |
| 512  | RTCMoved                     | SFU → Client      | Server notification to move to another channel                  |
| 514  | RTCSpeaking                  | SFU → Client      | Speaking indicator broadcast `{ user_id:int64, speaking:0\|1 }` |

Ping/Pong (op = 2)

- Client → SFU: `{op:2, d:{nonce?:any, ts?:int}}`
- SFU → Client: `{op:2, d:{pong:true, server_ts:int, nonce?:any, ts?:int}}`

Notes

- The SFU requires a short-lived token (`typ=sfu`, `aud=sfu`) with fields:
  - `channel_id`: voice channel to join
  - `perms`: permission bitmask for the user in this channel (from API)
- Enforced permissions:
  - `PermVoiceConnect` — required to join
  - `PermVoiceSpeak` — required to publish audio
  - `PermVoiceVideo` — required to publish video
  - `PermVoiceMuteMembers` — required for `t=507`
  - `PermVoiceDeafenMembers` — required for `t=508`
  - `PermVoiceMoveMembers` — required to kick/move members (handled by higher-level API/workflow)
  - `PermAdministrator` — overrides all the above to positive
- Token field `moved=true` lets a blocked user join (forced move) and grants audio/video publish permissions for the session.

- The SFU typically sends the first offer immediately after a user joins the room. Clients respond with an `RTCAnswer`, but may still initiate their own offer when local conditions change (codec/device switch).

## Convenience WS Events (non‑envelope)

SFU also accepts a small set of simple JSON control messages outside the `op/t/d` envelope for ease of use:

- Request server‑initiated offer (renegotiation):
  - Client → SFU: `{ "event":"negotiate" }`
  - SFU reacts by sending `{ op:7, t:501, d:{ sdp:"<OFFER>" } }`.

- Speaking indicator (client → server; server re‑broadcasts):
  - Client → SFU: `{ "event":"speaking", "data":"1" }` or `{ "event":"speaking", "data":"0" }`
  - Also accepted: `{ "event":"speaking", "data":"{\"speaking\":1}" }`
  - SFU → other peers: `{ op:7, t:514, d:{ user_id:<int64>, speaking:1|0 } }`

These convenience events are optional; you can implement everything with the full `op/t/d` protocol if preferred.

## Media IDs (stream/track)

- For every inbound remote track, the SFU forwards media using a stream id tagged with the sender’s user id: `stream.id = "u:<user_id>"`.
- The outbound local track id may be normalized to ensure uniqueness per user: `track.id = "<user_id>-<original>"`.
- Frontend can map `ontrack` events to users by reading `e.streams[0].id` and parsing the number after `u:`.

## Media Limits & Enforcement

Config keys in `sfu_config.yaml`:
- `max_audio_bitrate_kbps` (int, default 0): When > 0, the SFU will cap audio bitrate by injecting SDP constraints (adds/updates `b=TIAS`, `b=AS`, and Opus `fmtp maxaveragebitrate`).
- `enforce_audio_bitrate` (bool, default false): If true, the SFU monitors inbound audio RTP and disconnects peers exceeding the cap for sustained windows (two consecutive seconds).
- `audio_bitrate_margin_percent` (int 0..100, default 15): Tolerance over the cap to account for headers, jitter, and short spikes.

Notes:
- Enforcement uses network bytes (RTP + headers). If you see false positives, increase the margin or cap.

## Offer/Answer Flow

1. **Join & Ack** — Client connects to `/signal`, sends `RTCJoin`, and receives `{op:7, t:500, d:{ok:true}}` once the token and channel permissions are validated.
2. **Server Offer** — The room synchronizer (`signalPeers`) immediately creates an SDP offer for every peer with a stable signaling state and pushes it via `{op:7, t:501, d:{sdp:"<OFFER>"}}`.
3. **Client Answer** — Clients set the remote description, create an answer, and reply with `{op:7, t:502, d:{sdp:"<ANSWER>"}}`. When the SFU applies the answer it reschedules `signalPeers`, ensuring any pending tracks are attached.
4. **Media Fan-out** — For each inbound publisher track the SFU clones RTP packets into a shared `TrackLocalStaticRTP` and reuses it across subscribers. Anytime membership or mute state changes, `signalPeers` refreshes RTPSenders and triggers a new offer.
5. **Trickle ICE** — Both sides continue to exchange `{op:7, t:503}` candidates until the transports are connected. A PLI broadcast runs after every resync so late joiners request keyframes from active speakers.

Notes:
- Binding keepalive (t=509) is handled by the WS service. See WS Event Types for t=509.
- Route rebind notifications are dispatched by the WS service (t=513). See WS Event Types.

See also: [SFU Event Payloads](SFUEventPayloads.md) and [SFU Permissions](SFUPermissions.md).
