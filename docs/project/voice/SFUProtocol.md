[<- Documentation](../README.md) - [Voice](README.md)

# SFU WebSocket Protocol

This document describes the signaling contract exposed by the SFU `/signal` WebSocket.

## Endpoint And Version Selection

- Connect to `wss://<sfu>/signal`.
- If `v` is absent, the SFU uses legacy signaling `v=1`.
- `?v=1` explicitly selects the legacy flow.
- `?v=2` selects the v2 voice gateway flow.
- Any other explicit `v` value is rejected with HTTP `400` before the WebSocket upgrade.
- Reconnects caused by move, rebind, or transient failure should reuse the same signaling version that was already in use.

## `v=1` Legacy Signaling

`v=1` keeps the old RTC envelope:

```json
{
  "op": 7,
  "t": 500,
  "d": {}
}
```

### Legacy Opcodes

| `op` | Name |
|------|------|
| `1` | `Hello` |
| `2` | `Heartbeat` |
| `7` | `RTC` |

### Legacy RTC Events

| `t` | Name | Direction | Purpose |
|-----|------|-----------|---------|
| `500` | `RTCJoin` | Client -> SFU | Join with `{ channel, token }` |
| `501` | `RTCOffer` | SFU -> Client | Server-driven offer |
| `502` | `RTCAnswer` | Client -> SFU | Client answer |
| `503` | `RTCCandidate` | Both | Trickle ICE candidate |
| `504` | `RTCLeave` | Client -> SFU | Leave the voice session |
| `505-514` | Control events | Mixed | Mute, deafen, kick, block, move, rebind, speaking |

### `v=1` Notes

- `v=1` still accepts the older simple JSON compatibility messages such as `{"event":"answer"}` and `{"event":"speaking"}`.
- Initial bootstrap is server-offer based.
- Later renegotiation also stays on `501/502/503`.

## `v=2` Voice Gateway

`v=2` does not use the `op=7,t=*` RTC envelope for the public bootstrap path. It uses dedicated voice gateway opcodes instead.

### Text JSON Opcodes

| `op` | Name | Direction | Purpose |
|------|------|-----------|---------|
| `0` | `Identify` | Client -> SFU | Authenticate and describe DAVE capabilities |
| `1` | `Select Protocol` | Client -> SFU | Send local SDP offer or answer |
| `2` | `Ready` | SFU -> Client | ICE servers, codecs, publish permissions, DAVE settings |
| `3` | `Heartbeat` | Client -> SFU | Keep the session alive |
| `4` | `Session Description` | SFU -> Client | Return SDP answer or later server offer |
| `5` | `Speaking` | Both | Speaking state updates |
| `6` | `Heartbeat ACK` | SFU -> Client | Reply to `Heartbeat` |
| `7` | `Resume` | Client -> SFU | Rebind a recent `v=2` session |
| `8` | `Hello` | SFU -> Client | Session id and heartbeat interval |
| `9` | `Resumed` | SFU -> Client | Resume success |
| `11` | `Clients Connect` | SFU -> Client | User ids that joined the media session |
| `13` | `Client Disconnect` | SFU -> Client | User id that left the media session |
| `21` | `DAVE Prepare Transition` | SFU -> Client | Announce downgrade or protocol switch |
| `22` | `DAVE Execute Transition` | SFU -> Client | Transition execution point |
| `23` | `DAVE Transition Ready` | Client -> SFU | Client is ready for the transition |
| `24` | `DAVE Prepare Epoch` | SFU -> Client | Announce MLS group creation or recreation |
| `31` | `DAVE MLS Invalid Commit Welcome` | Client -> SFU | Ask the gateway to recreate state after invalid MLS material |

### Binary Opcodes

| Opcode | Name | Direction |
|--------|------|-----------|
| `25` | `DAVE MLS External Sender Package` | SFU -> Client |
| `26` | `DAVE MLS Key Package` | Client -> SFU |
| `27` | `DAVE MLS Proposals` | SFU -> Client |
| `28` | `DAVE MLS Commit Welcome` | Client -> SFU |
| `29` | `DAVE MLS Announce Commit Transition` | SFU -> Client |
| `30` | `DAVE MLS Welcome` | SFU -> Client |

## `v=2` Bootstrap Order

The connection order matches the GoChat v2 voice gateway flow:

1. Client connects to `/signal?v=2`.
2. Client sends `Identify (0)`.
3. SFU replies with `Hello (8)`.
4. SFU replies with `Ready (2)`.
5. Client creates a local offer and sends `Select Protocol (1)`.
6. SFU applies the offer and replies with `Session Description (4)` carrying the answer.
7. After the session is established, the SFU sends `Clients Connect (11)` and `Client Disconnect (13)` membership notifications as needed.
8. DAVE transitions use `21-31` plus binary `25-30`.

### `Hello`

Server -> client:

```json
{
  "op": 8,
  "d": {
    "v": 2,
    "heartbeat_interval": 15000,
    "session_id": "c37c8c76-6f88-4b2b-99ea-f4a6f44f1b7b"
  }
}
```

### `Identify`

Client -> SFU:

```json
{
  "op": 0,
  "d": {
    "channel_id": 789,
    "token": "<sfu jwt>",
    "max_dave_protocol_version": 1,
    "supports_encoded_transforms": true,
    "video": true,
    "streams": [
      { "type": "video", "rid": "100", "quality": 100 }
    ]
  }
}
```

- `channel_id` and `token` are required.
- `max_dave_protocol_version` and `supports_encoded_transforms` control whether the session is DAVE-capable.
- `identity_key` is optional and reserved for later verification UX.

### `Ready`

Server -> client:

```json
{
  "op": 2,
  "d": {
    "ice_servers": [{ "urls": ["stun:stun.example.net:3478"] }],
    "supported_codecs": [
      { "name": "opus", "type": "audio", "payload_type": 111 },
      { "name": "VP8", "type": "video", "payload_type": 96 },
      { "name": "VP9", "type": "video", "payload_type": 98 }
    ],
    "can_publish_audio": true,
    "can_publish_video": true,
    "max_audio_bitrate_kbps": 64,
    "experiments": [],
    "dave_enabled": true,
    "dave_required": false,
    "allow_av1_under_dave": false
  }
}
```

### `Select Protocol`

Client -> SFU:

```json
{
  "op": 1,
  "d": {
    "protocol": "webrtc",
    "type": "offer",
    "sdp": "v=0\r\n...",
    "rtc_connection_id": "280c8df4-eced-4244-9ea3-9d64b4bfc653"
  }
}
```

The initial `Select Protocol` must carry an offer. Later renegotiation on `v=2` also reuses `op=1`:

- client answer to a server offer: `type = "answer"`
- client-initiated renegotiation: `type = "offer"`

### `Session Description`

Server -> client:

```json
{
  "op": 4,
  "d": {
    "type": "answer",
    "sdp": "v=0\r\n...",
    "rtc_connection_id": "280c8df4-eced-4244-9ea3-9d64b4bfc653",
    "media_session_id": "2462af133bf785af961d4419f9d28850",
    "audio_codec": "opus",
    "video_codec": "VP8",
    "dave_protocol_version": 0,
    "dave_epoch": 0
  }
}
```

The same opcode is also used later when the SFU sends a renegotiation offer:

- `type = "offer"` for a server-initiated renegotiation
- `type = "answer"` for a bootstrap or negotiated reply

## Heartbeat

Client ping:

```json
{
  "op": 3,
  "d": {
    "t": 1717171717,
    "seq_ack": 0
  }
}
```

Server ack:

```json
{
  "op": 6,
  "d": {
    "t": 1717171717
  }
}
```

The SFU closes the socket if no heartbeat arrives for `heartbeat_interval + 10s`.

## DAVE / E2EE Behavior On `v=2`

- If every connected participant is DAVE-capable and DAVE is enabled, the channel uses protocol version `1`.
- If any connected participant is not DAVE-capable, the channel stays on or downgrades to protocol version `0`.
- GoChat clients should attach encoded transforms from the start and begin in passthrough mode.
- `Clients Connect (11)` and `Client Disconnect (13)` provide the membership view used by the DAVE state machine.

### Upgrade / Group Recreation

1. SFU sends `DAVE Prepare Epoch (24)`.
2. SFU sends binary `25`.
3. Pending members send binary `26`.
4. SFU broadcasts binary `27`.
5. A committing member sends binary `28`.
6. SFU broadcasts binary `29` and targeted binary `30`.
7. Clients send `23`.
8. SFU sends `22`.

### Downgrade

1. SFU sends `21` with `protocol_version = 0`.
2. DAVE participants switch receivers to passthrough mode and send `23`.
3. SFU sends `22`.

## Close Codes

`v=2` uses close codes instead of `RTCError` envelopes.

| Code | Meaning |
|------|---------|
| `4001` | Unknown opcode |
| `4002` | Invalid payload |
| `4003` | Unauthorized or blocked |
| `4009` | Heartbeat timeout |
| `4016` | Resume session expired |
| `4017` | DAVE is required but unsupported |
| `4020` | Message arrived in the wrong phase |
| `4021` | Unsupported protocol or transport mode |

## Notes

- Public `v=2` no longer exposes the old custom `530-539` RTC envelope flow.
- Public `v=2` also no longer uses public `501/502/503` for renegotiation; it stays on `1/4`.
- Privileged moderation events still use the existing legacy RTC control extension paths internally until dedicated gateway opcodes are introduced.
- `v=1` remains the only default when the query parameter is absent. Invalid `v` values do not silently downgrade.
