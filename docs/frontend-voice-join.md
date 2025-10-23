# Frontend Voice Join Flow

This guide explains how a browser client should negotiate a voice connection with the GoChat SFU. The flow mirrors the [Pion `sfu-ws` example](https://github.com/pion/webrtc/tree/master/examples/sfu-ws) and the server code under `cmd/sfu/`.

## Message formats at a glance

For backward compatibility the SFU transmits every signaling payload in **two** shapes:

| Purpose       | Legacy envelope (`op`/`t`)                                                | Event message (`event`/`data`)             |
|---------------|----------------------------------------------------------------------------|--------------------------------------------|
| Offer         | `{ "op":7, "t":501, "d": { "sdp": "…", "type": "offer" } }`         | `{ "event":"offer", "data":"{…}" }`     |
| Answer        | `{ "op":7, "t":502, "d": { "sdp": "…", "type": "answer" } }`        | `{ "event":"answer", "data":"{…}" }`    |
| ICE candidate | `{ "op":7, "t":503, "d": { "candidate": "…", … } }`                   | `{ "event":"candidate", "data":"{…}" }` |

Both payloads contain identical SDP/candidate data. Consume whichever format fits your client and ignore/dedupe the other copy to avoid double-processing. 【F:cmd/sfu/sfu.go†L28-L72】

## Step-by-step sequence

1. **Open the signaling WebSocket**  
   Connect to `wss://<sfu-host>/signal` (or `ws://` locally) with the browser `WebSocket` API. 【F:cmd/sfu/app.go†L49-L70】

2. **Send the join envelope immediately**  
   The very first frame must be the RTC join envelope:

   ```json
   { "op": 7, "t": 500, "d": { "channel": <voice_channel_id>, "token": "<join_jwt>" } }
   ```

   The token is validated against the requested channel; invalid credentials close the socket. 【F:cmd/sfu/app.go†L74-L115】【F:cmd/sfu/proto.go†L9-L32】【F:cmd/sfu/app.go†L155-L175】

3. **Wait for the join acknowledgement**  
   Authorization success is reported via `{ "op":7, "t":500, "d": { "ok": true } }`. Do not start SDP negotiation until this ACK arrives—the server only attaches the peer to the room after flushing the acknowledgement. 【F:cmd/sfu/app.go†L150-L169】

4. **Handle the server offer and send your answer**  
   Immediately after the ACK the SFU adds the peer to the requested voice channel and emits an SDP offer (in both formats listed earlier). Apply the first copy you see with `pc.setRemoteDescription()` and discard the duplicate. 【F:cmd/sfu/app.go†L171-L198】【F:cmd/sfu/sfu.go†L126-L189】

   Generate an answer (`pc.createAnswer()` + `pc.setLocalDescription()`) and return it using either the envelope or the event message. Include the SDP `type` field (`"answer"`) so the SFU can validate the signaling transition. 【F:cmd/sfu/app.go†L200-L232】【F:cmd/sfu/proto.go†L25-L33】

5. **Trickle ICE candidates**  
   After `remoteDescription` is set, forward browser ICE candidates to the SFU in either format and call `pc.addIceCandidate` on the remote candidates you receive. Because each candidate arrives twice, apply one copy and drop the duplicate. Queue local candidates until the remote description is present to keep username fragments aligned. 【F:cmd/sfu/app.go†L118-L199】【F:cmd/sfu/sfu.go†L59-L83】

6. **Maintain the heartbeat**  
   Keep the WebSocket alive by sending heartbeat envelopes periodically:

   ```json
   { "op": 2, "d": { "nonce": "ping-<timestamp>", "ts": <timestamp> } }
   ```

   The SFU responds with a matching pong payload. 【F:cmd/sfu/app.go†L201-L231】

7. **Leave the voice channel**  
   Close the WebSocket or send `{ "op":7, "t":504 }` when leaving. The SFU tears down your `RTCPeerConnection`, removes forwarded tracks, and updates discovery metrics automatically. 【F:cmd/sfu/app.go†L118-L199】【F:cmd/sfu/sfu.go†L97-L124】

## Additional guidance

* **Single offer policy:** let the SFU trigger renegotiations. If you change local tracks, wait for the next server offer instead of creating one yourself.  
* **Candidate hygiene:** apply one copy of each candidate and discard duplicates from the alternate message stream.  
* **Error recovery:** if you receive an error envelope after joining, close the socket and request a fresh join token before retrying.

Following these steps keeps the browser and the SFU in the correct signaling roles, prevents ICE username-fragment mismatches, and matches the renegotiation behavior baked into the server implementation.
