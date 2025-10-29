[<- Documentation](../README.md) - [WebSocket Events](README.md)

# Event Message Structure

Here's the structure of the event message:

```json
{
  "op": 0, // Opcode of the message. See the list of op codes below.
  "d": { // Data of the event. There are many different types. This one is the New Message Event.
    "guild_id": 2226022078304223200,
    "message": {
      "id": 2228801793842741200,
      "channel_id": 2226022078341972000,
      "author": {
        "id": 2226021950625415200,
        "name": "FlameInTheDark",
        "discriminator": "flameinthedark"
      },
      "content": "Hello"
    }
  },
  "t": 100 // Event message type. See the list of event types.
}
```

### OP Code Types
| Type | Description                           |
|------|---------------------------------------|
| 0    | Dispatch                              |
| 1    | Hello (auth and hearth beat interval) |
| 2    | Hearth Beat                           |
| 3    | Presence Update                       |
| 4    | Guild Update Subscription             |
| 5    | Channel Subscription                  |

### Presence Update (op=3) Payload
- Request body:
```json
{ "status": "online|idle|dnd|offline",
  "platform": "web|mobile|desktop", // Has no effect at the moment, just for info
  "custom_status_text": "string",
  "voice_channel_id": 2230469276416868352 // optional; set to number to indicate voice presence, omit or 0 to clear. If was set before, should be 0 to clear
}
```

Notes:
- When `voice_channel_id` is provided and positive, it is stored for the current session and included in aggregated presence updates so others can see which voice channel the user is in. Should be set to 0 to clear if was presented before.
- Voice presence can also be maintained by periodically sending `op=7, t=509` (RTC BindingAlive), which refreshes the per‑channel route and sets session voice channel automatically.
