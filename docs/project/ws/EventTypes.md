[<- Documentation](../README.md) - [WebSocket Events](README.md)

# Event Types

This page lists all event type values used in WebSocket payloads under the `t` field (e.g., `{ "op": 0, "t": 100, "d": {...} }`).

### Message Types (100–115)
| Type | Description               |
|------|---------------------------|
| 100  | Message Create            |
| 101  | Message Update            |
| 102  | Message Delete            |
| 103  | Guild Create              |
| 104  | Guild Update              |
| 105  | Guild Delete              |
| 106  | Channel Create            |
| 107  | Channel Update            |
| 108  | Channel Order Update      |
| 109  | Channel Delete            |
| 110  | Guild Role Create         |
| 111  | Guild Role Update         |
| 112  | Guild Role Delete         |
| 113  | Thread Create             |
| 114  | Thread Update             |
| 115  | Thread Delete             |

### Guild Member Types (200–204)
| Type | Description               |
|------|---------------------------|
| 200  | Guild Member Added        |
| 201  | Guild Member Update       |
| 202  | Guild Member Remove       |
| 203  | Guild Member Role Added   |
| 204  | Guild Member Role Removed |

### Channel Message Types (300)
| Type | Description           |
|------|-----------------------|
| 300  | Guild Channel Message |
| 301  | Channel Typing Event  |

### User Types (400–406)
| Type | Description               |
|------|---------------------------|
| 400  | User Update Read State    |
| 401  | User Update Settings      |
| 402  | Incoming Friend Request   |
| 403  | Friend Added              |
| 404  | Friend Removed            |
| 405  | User DM Message           |
| 406  | User Update               |

### RTC Events
- Most RTC (voice) events are documented in the voice section:
  - [SFU WebSocket Protocol](../voice/SFUProtocol.md)

- WS‑dispatched RTC control:

| Type | Description                    |
|------|--------------------------------|
| 513  | RTC Server Rebind (reconnect)  |

RTC Server Rebind (t=513)
- Purpose: Notify clients in a voice channel that the route (SFU instance) has changed.
- Action: Client should call API Join for that channel to get a fresh `sfu_url` and `sfu_token`.

RTC Binding Alive (t=509)
- Purpose: Keep the per‑channel SFU route (voice:route:{channel}) alive while peers are connected.
- Action: Client periodically sends `{ op:7, t:509, d:{ channel:int64 } }` on the WS connection; server refreshes the route TTL.
