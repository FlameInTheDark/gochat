# Frontend Voice Join Flow

This guide explains how a browser client should negotiate a voice connection with the GoChat SFU. The flow mirrors the [Pion `sfu-ws` example](https://github.com/pion/webrtc/tree/master/examples/sfu-ws) and the server code under `cmd/sfu/`.

## Signaling envelopes

All WebRTC signaling travels in a single envelope format that uses the familiar `op` and `t` fields:

| Purpose       | Envelope payload (`op`/`t`)                                               |
|---------------|----------------------------------------------------------------------------|
| Offer         | `{ "op":7, "t":501, "d": { "sdp": "…", "type": "offer" } }`         |
| Answer        | `{ "op":7, "t":502, "d": { "sdp": "…", "type": "answer" } }`        |
| ICE candidate | `{ "op":7, "t":503, "d": { "candidate": "…", … } }`                   |

Legacy `event`/`data` frames are no longer produced, so clients only need to send and receive the envelopes above.

## Step-by-step sequence

1. **Open the signaling WebSocket**
   Connect to `wss://<sfu-host>/signal` (or `ws://` locally) with the browser `WebSocket` API.

2. **Send the join envelope immediately**
   The very first frame must be the RTC join envelope:

   ```json
   { "op": 7, "t": 500, "d": { "channel": <voice_channel_id>, "token": "<join_jwt>" } }
   ```

   The token is validated against the requested channel; invalid credentials close the socket.

3. **Wait for the join acknowledgement**
   Authorization success is reported via `{ "op":7, "t":500, "d": { "ok": true } }`. Do not start SDP negotiation until this ACK arrives.

4. **Handle the server offer and send your answer**
   Immediately after the ACK the SFU adds the peer to the requested voice channel and emits an SDP offer. Apply the offer with `pc.setRemoteDescription()` as soon as it arrives.

   Generate an answer (`pc.createAnswer()` + `pc.setLocalDescription()`) and return it in the same envelope format. Include the SDP `type` field (`"answer"`) so the SFU can validate the signaling transition.

5. **Trickle ICE candidates**
   After `remoteDescription` is set, forward browser ICE candidates to the SFU and call `pc.addIceCandidate` on the remote candidates you receive. Each candidate arrives exactly once, so you can process envelopes sequentially without deduplication. Queue local candidates until the remote description is present to keep username fragments aligned.

6. **Maintain the heartbeat**
   Keep the WebSocket alive by sending heartbeat envelopes periodically:

   ```json
   { "op": 2, "d": { "nonce": "ping-<timestamp>", "ts": <timestamp> } }
   ```

   The SFU responds with a matching pong payload.

7. **Leave the voice channel**
   Close the WebSocket or send `{ "op":7, "t":504 }` when leaving. The SFU tears down your `RTCPeerConnection`, removes forwarded tracks, and updates discovery metrics automatically.

## Additional guidance

* **Single offer policy:** let the SFU trigger renegotiations. If you change local tracks, wait for the next server offer instead of creating one yourself.
* **Candidate hygiene:** forward each candidate as soon as it appears; the SFU delivers one envelope per candidate.
* **Error recovery:** if you receive an error envelope after joining, close the socket and request a fresh join token before retrying.

Following these steps keeps the browser and the SFU in the correct signaling roles, prevents ICE username-fragment mismatches, and matches the renegotiation behavior baked into the server implementation.
