[<- Bots](README.md)

# Operations

Bot traffic is split from user WebSocket traffic so large bots and bot event spikes can scale independently.

## NATS Buses

Bot events use the dedicated endpoint configured with:

```text
bot_nats_conn_string
BOT_NATS_CONN_STRING
```

Normal user WebSocket traffic uses:

```text
nats_conn_string
NATS_CONN_STRING
```

`cmd/botws` connects to both buses: bot NATS for instance delivery and core NATS for presence publishes.

## Service Flow

- `cmd/api` publishes message and guild events to both the user NATS bus and the bot NATS bus.
- `cmd/botapi` publishes bot-authored message events to both buses so regular clients and other permitted bots can receive them.
- `cmd/ws` remains on the user NATS bus.
- `cmd/botrouter` consumes partitioned bot events from the bot NATS bus.
- `cmd/botws` consumes only instance delivery envelopes from the bot NATS bus.
- Docker Compose includes a separate `bot-nats` service so bot event load can be scaled independently from user WebSocket event load.

## Makefile

Backend rebuild helpers include bot services:

```text
make docker-rebuild-botapi
make docker-rebuild-botws
make docker-rebuild-botrouter
```

Use these when iterating on bot runtime services without rebuilding unrelated services.
