[<- Documentation](../README.md) - [Voice](README.md)

# SFU Event Payloads

This document enumerates the payload shapes for each RTC event (`op=7`). Field names and types match the JSON wire format.

Legend
- Dir: C→S (Client to SFU), S→C (SFU to Client), Both for bidirectional

| t    | Name                | Dir   | Payload schema |
|------|---------------------|-------|----------------|
| 500  | RTCJoin             | C→S   | `{ channel:int64, token:string }` — token is short‑lived (`typ=sfu`, `aud=sfu`) and may include `moved:bool` (grants connect + audio/video publish for the session) |
| 500  | RTCJoin (ack)       | S→C   | `{ ok:boolean }` |
| 501  | RTCOffer            | Both  | `{ sdp:string }` |
| 502  | RTCAnswer           | Both  | `{ sdp:string }` |
| 503  | RTCCandidate        | Both  | `{ candidate:string, sdpMid?:string, sdpMLineIndex?:int }` |
| 504  | RTCLeave            | C→S   | `{}` (empty `d`) |
| 505  | RTCMuteSelf         | C→S   | `{ muted:boolean }` |
| 506  | RTCMuteUser         | C→S   | `{ user:int64, muted:boolean }` (local‑only) |
| 507  | RTCServerMuteUser   | C→S   | `{ user:int64, muted:boolean }` (privileged) |
| 508  | RTCServerDeafenUser | C→S   | `{ user:int64, deafened:boolean }` (privileged) |
| 510  | RTCServerKickUser   | C→S   | `{ user:int64 }` (privileged). Server replies by notifying the user and closing their WS. |
| 511  | RTCServerBlockUser  | C→S   | `{ user:int64, block:boolean }` (privileged). Blocks/unblocks joining this channel. |
| 512  | RTCMoved            | S→C   | `{ channel:int64 }` — client should reconnect to the indicated channel |
| 514  | RTCSpeaking         | S→C   | `{ user_id:int64, speaking:int }` (1=active, 0=inactive) |

Heartbeat (separate op=2)
- Client → SFU: `{ op:2, d:{ nonce?:any, ts?:int } }`
- SFU → Client: `{ op:2, d:{ pong:true, server_ts:int, nonce?:any, ts?:any } }`

Notes
- Privileged actions require `PermVoiceMuteMembers`, `PermVoiceDeafenMembers`, or `PermVoiceMoveMembers`, or `PermAdministrator` (override).
- When a user is force‑moved via the API, the token for the target channel includes `moved=true`, which allows bypassing a room‑level block.

## Client Interaction Examples (int64 IDs)

All IDs (users, channels) are 64‑bit integers in this system. The JSON below uses int64 values in the `user` and `channel` fields.

Self mute microphone (local publish toggle)
```json
{ "op": 7, "t": 505, "d": { "muted": true } }
```

Locally mute another user (stop receiving their media only for this client)
```json
{ "op": 7, "t": 506, "d": { "user": 2230469276416868352, "muted": true } }
```

Server‑wide mute a user (privileged)
```json
{ "op": 7, "t": 507, "d": { "user": 2230469276416868352, "muted": true } }
```

Server‑wide deafen a user (privileged)
```json
{ "op": 7, "t": 508, "d": { "user": 2230469276416868352, "deafened": true } }
```

Keep route alive for the current channel over WS (not SFU signaling)
```json
{ "op": 7, "t": 509, "d": { "channel": 2230469276416868352 } }
```

Rebind notice from WS (route changed) — client should call API Join for the channel
```json
{ "op": 7, "t": 513, "d": { "channel": 2230469276416868352 } }
```

Local per‑user volume (browser example)
```js
// ontrack handler: map stream to userId and keep gain node per user
pc.ontrack = (ev) => {
  const stream = ev.streams[0];
  // stream.id is "user-<user_id>" (e.g., "user-2230469276416868352")
  const userId = Number(stream.id.slice(5));
  const ctx = new AudioContext();
  const src = ctx.createMediaStreamSource(stream);
  const gain = ctx.createGain();
  gain.gain.value = 1.0; // default volume
  src.connect(gain).connect(ctx.destination);
  userAudio[userId] = { stream, gain };
};

// Change volume (0.0 .. 1.0) for a specific user id (int64)
function setUserVolume(userId, vol) {
  if (userAudio[userId]) userAudio[userId].gain.gain.value = vol;
}

// Locally mute a user via SFU (saves bandwidth)
ws.send(JSON.stringify({ op: 7, t: 506, d: { user: 2230469276416868352, muted: true } }));
```
Speaking indicator (broadcast to peers)
```json
{ "op": 7, "t": 514, "d": { "user_id": 2230469276416868352, "speaking": 1 } }
```

Simple (non‑envelope) events accepted by the SFU

- Request renegotiation (server offer):
```json
{ "event": "negotiate" }
```

- Speaking on/off:
```json
{ "event": "speaking", "data": "1" }
{ "event": "speaking", "data": "0" }
```

The speaking simple event also accepts a JSON string payload:
```json
{ "event": "speaking", "data": "{\"speaking\":1}" }
```
