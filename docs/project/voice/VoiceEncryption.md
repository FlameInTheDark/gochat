[<- Documentation](../README.md) - [Voice](README.md)

# Voice End-to-End Encryption (DAVE)

This document describes the current GoChat voice E2EE direction for `/signal?v=2`.

GoChat now uses a DAVE control plane on the v2 voice gateway:

- JSON voice gateway opcodes for bootstrap, membership, and transitions
- binary DAVE opcodes `25-30` for MLS-related payload delivery
- protocol version `0` for transport-only media
- protocol version `1` for DAVE-enabled media sessions

## Current Server Model

### What Is Implemented

- `v=2` handshake uses `Identify (0) -> Hello (8) -> Ready (2) -> Select Protocol (1) -> Session Description (4)`
- DAVE capability tracking is per channel
- membership changes are surfaced through `Clients Connect (11)` and `Client Disconnect (13)`
- the SFU can:
  - recreate a DAVE group with `24/25/26/27/28/29/30/23/22`
  - downgrade to transport-only with `21/23/22`
  - resume a recent `v=2` session and return the active `dave_protocol_version`
- binary DAVE envelopes are encoded in pure Go under `internal/voice/dave/wire`

### What Is Still Deliberately Lightweight

- the current MLS validation layer is structural, not a full RFC 9420 cryptographic verifier
- the gateway validates transition phase, required payload presence, and wire shape, then relays opaque MLS payload bytes
- persistent identity upload, verification UI, and out-of-band trust UX remain phase 2

GoChat clients may still expose a session verification code in the first rollout, but that code should be documented honestly:

- if it is derived from ephemeral session identity only, it proves that participants are in the same encrypted DAVE session and epoch
- it does not yet provide the stronger "same long-term identity as last time" guarantee that stable identity keys and explicit trust verification would provide

That means the wire and transition behavior are in place for GoChat clients, while full MLS cryptographic inspection can be tightened incrementally behind the same server interfaces.

## Protocol Version Policy

- If every connected participant is DAVE-capable and DAVE is enabled, the channel can use protocol version `1`.
- If any connected participant is not DAVE-capable, the channel stays on or transitions back to protocol version `0`.
- DAVE-capable clients must attach encoded transforms from the start and begin in passthrough mode.

## DAVE Transition Flows

### Upgrade / Group Creation

1. SFU sends `DAVE Prepare Epoch (24)` with `protocol_version = 1`.
2. SFU sends binary `External Sender Package (25)`.
3. Pending members send binary `Key Package (26)`.
4. SFU broadcasts binary `Proposals (27)`.
5. A committing member sends binary `Commit Welcome (28)`.
6. SFU broadcasts binary `Announce Commit Transition (29)` and targeted binary `Welcome (30)`.
7. Clients send `Transition Ready (23)`.
8. SFU sends `Execute Transition (22)`.

### Downgrade To Transport-Only

1. A non-DAVE participant joins.
2. SFU sends `Prepare Transition (21)` with `protocol_version = 0`.
3. DAVE participants switch receive transforms to passthrough and send `23`.
4. SFU sends `22`.

### Invalid Commit / Welcome Recovery

1. A client sends `Invalid Commit Welcome (31)`.
2. The gateway recreates the DAVE group state.
3. Clients receive a fresh `24` and binary `25`.
4. New key packages and a new commit flow are produced.

## Binary DAVE Messages

The server exposes these binary opcodes:

| Opcode | Meaning |
|--------|---------|
| `25` | External sender package |
| `26` | Key package |
| `27` | Proposal batch |
| `28` | Commit with optional welcome |
| `29` | Commit transition announcement |
| `30` | Welcome for a pending member |

GoChat encodes these gateway envelopes using MLS-style variable-length vectors so the wire is deterministic and testable. Golden-byte tests live under `internal/voice/dave/wire`.

## Config

```yaml
dave_enabled: true
dave_required_default: false
dave_transition_timeout_ms: 2000
dave_old_ratchet_window_ms: 10000
dave_allow_av1: false
```

### Notes

- `dave_required_default: true` rejects non-DAVE `v=2` clients with close code `4017`
- `dave_allow_av1` is `false` by default for the first rollout
- the old-ratchet window mirrors the DAVE guidance for in-flight media during transitions

## Media Path Notes

- SRTP transport encryption between client and SFU is still retained
- DAVE sits above transport encryption at the encoded-frame layer
- the current server includes protocol-frame helpers under `internal/voice/dave/frame`
- the existing RTP fan-out remains the primary forwarding path

The current rollout focuses on connection establishment, voice-gateway parity, and DAVE transition state. Per-receiver protocol-frame dropping during mixed capability transitions is intentionally isolated behind the new frame package so that it can be wired more deeply into fan-out without changing the public wire again.

## Client Expectations

GoChat DAVE-capable clients should:

1. Use `/signal?v=2`
2. Send `max_dave_protocol_version`
3. Send `supports_encoded_transforms`
4. Attach encoded transforms immediately
5. Treat `dave_protocol_version = 0` as passthrough mode
6. Send binary `26` and `28` when the gateway requests them
7. Send `23` only when the receive side is ready to switch

## Scope Boundaries

This rollout does not yet promise:

- third-party service interoperability
- full MLS cryptographic validation inside the gateway
- persistent identity verification UI
- AV1 DAVE support by default

Those are follow-on improvements, not wire-breaking changes.
