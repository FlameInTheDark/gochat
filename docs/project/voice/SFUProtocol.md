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

| t    | Name                         | Direction         | Purpose |
|------|------------------------------|-------------------|---------|
| 500  | RTCJoin                      | Client → SFU      | Join a voice channel using short‑lived token |
| 501  | RTCOffer                     | Client/SFU        | SDP offer for negotiation |
| 502  | RTCAnswer                    | Client/SFU        | SDP answer for negotiation |
| 503  | RTCCandidate                 | Client/SFU        | Trickle ICE candidate |
| 504  | RTCLeave                     | Client → SFU      | Leave/close voice session |
| 505  | RTCMuteSelf                  | Client → SFU      | Toggle publishing of local microphone |
| 506  | RTCMuteUser                  | Client → SFU      | Local mute/unmute a specific remote user |
| 507  | RTCServerMuteUser            | Client → SFU      | Privileged: mute a user for everyone |
| 508  | RTCServerDeafenUser          | Client → SFU      | Privileged: deafen a user (receive no one) |
| 510  | RTCServerKickUser            | Client → SFU      | Privileged: kick a user from the room |
| 511  | RTCServerBlockUser           | Client → SFU      | Privileged: block/unblock user from the room |
| 512  | RTCMoved                     | SFU → Client      | Server notification to move to another channel |

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
