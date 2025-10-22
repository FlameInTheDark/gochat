# Frontend Voice Join Flow

This document describes how a browser client should negotiate a voice connection with the GoChat SFU. The sequence mirrors the [`sfu-ws` example from Pion](https://github.com/pion/webrtc/tree/master/examples/sfu-ws) and matches the current implementation under `cmd/sfu/`.

## 1. Open the signaling WebSocket

Connect to `wss://<sfu-host>/signal` (or `ws://` when testing locally). The connection upgrades to WebSocket via Fiber, so standard browser `WebSocket` APIs work out of the box. 【F:cmd/sfu/app.go†L49-L70】

## 2. Send the join envelope immediately

The first message must be an RTC join envelope using the legacy opcode shape:

```json
{
  "op": 7,
  "t": 500,
  "d": {
    "channel": <voice_channel_id>,
    "token": "<one_time_join_jwt>"
  }
}
```

* `op` corresponds to `OPCodeRTC`.
* `t` is `EventTypeRTCJoin`.
* `channel` is the voice-channel ID the client wants to join.
* `token` is the signed JWT delivered by the API for that user/channel.

The server validates the token and rejects the socket if the envelope is malformed or the credentials do not match. 【F:cmd/sfu/app.go†L74-L115】【F:cmd/sfu/proto.go†L9-L32】【F:cmd/sfu/app.go†L155-L175】

## 3. Handle the join acknowledgement

If authorization succeeds, the SFU replies with:

```json
{
  "op": 7,
  "t": 500,
  "d": { "ok": true }
}
```

Wait for this acknowledgement before attempting to negotiate media. 【F:cmd/sfu/app.go†L137-L150】

## 4. Negotiate SDP (server offers, client answers)

The SFU always acts as the offerer. Right after the join it sends an SDP offer twice:

1. As a legacy envelope (`op:7`, `t:501`, payload `{ "sdp": "..." }`).
2. As a convenience event message `{ "event": "offer", "data": "<same payload json>" }`.

Consume whichever format fits your client – both reference the same SDP. Apply the offer via `pc.setRemoteDescription()` and generate a local answer.

Return the answer to the SFU using either of the supported shapes:

* Legacy envelope
  ```json
  { "op": 7, "t": 502, "d": { "sdp": "<answer>" } }
  ```
* Event message
  ```json
  { "event": "answer", "data": "{\"sdp\":\"<answer>\"}" }
  ```

The SFU applies either form and will reject duplicate or out-of-order answers. Make sure the `RTCPeerConnection` stays open until `setRemoteDescription` resolves successfully on the server. 【F:cmd/sfu/app.go†L118-L199】【F:cmd/sfu/sfu.go†L43-L83】【F:cmd/sfu/sfu.go†L126-L189】

## 5. Trickle ICE candidates

Both sides trickle ICE candidates using the same dual-message pattern as the offer:

* Envelope: `{ "op": 7, "t": 503, "d": { "candidate": "...", "sdpMid": "0", "sdpMLineIndex": 0 } }`
* Event: `{ "event": "candidate", "data": "{...}" }`

Call `pc.addIceCandidate` for each received candidate, and send your local candidates only after the remote description is in place to avoid username-fragment mismatches. The SFU forwards all candidates to connected peers. 【F:cmd/sfu/app.go†L118-L199】【F:cmd/sfu/sfu.go†L59-L83】

## 6. Keep the connection alive

Send heartbeat envelopes periodically while connected:

```json
{ "op": 2, "d": { "nonce": "ping-<timestamp>", "ts": <timestamp> } }
```

The SFU automatically responds with a matching `pong` payload. Continue heartbeats at your usual gateway cadence (e.g., every 10 seconds) to keep intermediate proxies from timing out the socket. 【F:cmd/sfu/app.go†L201-L231】

## 7. Leaving the channel

To leave, either close the WebSocket or send a legacy leave envelope (`op:7`, `t:504`). The server tears down the `RTCPeerConnection`, removes your forwarded tracks, and updates discovery load metrics automatically. 【F:cmd/sfu/app.go†L118-L199】【F:cmd/sfu/sfu.go†L97-L124】

## 8. Additional guidance

* **One offer at a time:** let the SFU drive renegotiations. If you need to add local tracks, wait for a server offer after updating your media state; do not initiate a client offer.
* **Candidate ordering:** queue locally generated ICE candidates until `pc.remoteDescription` is non-null.
* **Error handling:** if you receive an error envelope after joining, close the socket and fetch a fresh join token before retrying.

Following this flow keeps the browser client and the SFU in the correct signaling roles, prevents ICE username-fragment mismatches, and matches the renegotiation behavior baked into the server implementation.
