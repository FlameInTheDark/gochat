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