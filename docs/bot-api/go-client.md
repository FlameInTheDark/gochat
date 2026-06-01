# Go Client

The Go bot client lives in:

```text
clients/bot/goclient
```

Module path:

```text
github.com/FlameInTheDark/gochat/clients/bot/goclient
```

The client provides:

- Session construction and endpoint configuration.
- Bot REST helpers for every current bot REST endpoint.
- WebSocket gateway identify, heartbeat, reconnect, and presence update support.
- Typed event handlers.
- Example projects.

## Defaults

```go
const (
	DefaultEndpoint        = "https://gochat.anticode.dev"
	DefaultAPIEndpoint     = "https://gochat.anticode.dev"
	DefaultGatewayEndpoint = "wss://gochat.anticode.dev/bot/ws"
)
```

## Create A Session

```go
session, err := goclient.New("gcb_your_token")
```

The token may be raw or already prefixed:

```go
goclient.New("gcb_your_token")
goclient.New("Bot gcb_your_token")
```

## Endpoint Configuration

Use one service base URL:

```go
session, err := goclient.New(
	"gcb_your_token",
	goclient.WithEndpoint("https://gochat.example.com"),
)
```

Use separate REST and gateway URLs:

```go
session, err := goclient.New(
	"gcb_your_token",
	goclient.WithAPIEndpoint("http://127.0.0.1:3102"),
	goclient.WithGatewayEndpoint("ws://127.0.0.1:3101/bot/ws"),
)
```

Set shard identity:

```go
session, err := goclient.New(
	"gcb_your_token",
	goclient.WithShard(0, 4),
)
```

## REST Methods

| Method | REST endpoint |
|--------|---------------|
| `UserMe(ctx)` | `GET /user/me` |
| `Me(ctx)` | Alias for `UserMe` |
| `Guilds(ctx)` | `GET /guild` |
| `GuildChannels(ctx, guildID)` | `GET /guild/{guild_id}/channels` |
| `ChannelMessages(ctx, channelID, limit, beforeID, afterID, aroundID)` | `GET /message/channel/{channel_id}` |
| `ChannelMessageSend(ctx, channelID, content)` | `POST /message/channel/{channel_id}` |
| `ChannelMessageSendComplex(ctx, channelID, data)` | `POST /message/channel/{channel_id}` |
| `ChannelMessageEdit(ctx, channelID, messageID, content)` | `PATCH /message/channel/{channel_id}/{message_id}` |
| `ChannelMessageEditComplex(ctx, channelID, messageID, data)` | `PATCH /message/channel/{channel_id}/{message_id}` |
| `ChannelMessageDelete(ctx, channelID, messageID)` | `DELETE /message/channel/{channel_id}/{message_id}` |
| `ChannelMessageAck(ctx, channelID, messageID)` | `POST /message/channel/{channel_id}/{message_id}/ack` |
| `ChannelTyping(ctx, channelID)` | `POST /message/channel/{channel_id}/typing` |
| `MessageReactionAdd(ctx, channelID, messageID, reactionName)` | `PUT /message/channel/{channel_id}/{message_id}/reactions/{reaction_name}` |
| `MessageReactionRemove(ctx, channelID, messageID, reactionName)` | `DELETE /message/channel/{channel_id}/{message_id}/reactions/{reaction_name}` |
| `MessageReactionUsers(ctx, channelID, messageID, reactionName, after, limit)` | `GET /message/channel/{channel_id}/{message_id}/reactions/{reaction_name}` |

## Raw REST Requests

Use `Request` when the server adds an endpoint before the client grows a helper:

```go
var out json.RawMessage
err := session.Request(ctx, http.MethodGet, "/bot/api/v1/user/me", nil, nil, &out)
```

Request options can set headers:

```go
err := session.ChannelMessageSend(
	ctx,
	channelID,
	"hello",
	goclient.WithIdempotencyKey("deploy-2026-06-01"),
)
```

## Gateway

Open a gateway connection:

```go
if err := session.Open(); err != nil {
	return err
}
defer session.Close()
```

Open with context:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
err := session.OpenWithContext(ctx)
```

Update presence:

```go
err := session.UpdatePresence("idle", "Processing queue")
```

Write a raw gateway frame:

```go
err := session.GatewayWrite(goclient.OPCodePresenceUpdate, goclient.PresenceUpdateRequest{
	Status: "online",
})
```

## Event Handlers

Handlers are functions with this shape:

```go
func(*goclient.Session, *goclient.MessageCreate)
```

Register a typed handler:

```go
session.AddHandler(func(s *goclient.Session, event *goclient.MessageCreate) {
	_, _ = s.ChannelMessageSend(context.Background(), event.Message.ChannelID, "received")
})
```

Register a handler that runs once:

```go
session.AddHandlerOnce(func(_ *goclient.Session, ready *goclient.Ready) {
	fmt.Println("ready", ready.SessionID)
})
```

Register a raw dispatch handler:

```go
session.AddHandler(func(_ *goclient.Session, event *goclient.Event) {
	fmt.Printf("event type=%v raw=%s\n", event.Type, event.RawData)
})
```

`AddHandler` returns a function that removes the handler:

```go
remove := session.AddHandler(handleMessage)
remove()
```

## Error Handling

Non-2xx REST responses return `*goclient.RESTError`:

```go
message, err := session.ChannelMessageSend(ctx, channelID, "hello")
if err != nil {
	var restErr *goclient.RESTError
	if errors.As(err, &restErr) {
		log.Printf("status=%d body=%s", restErr.StatusCode, restErr.Body)
	}
	return err
}
_ = message
```

Gateway open errors include:

| Error | Meaning |
|-------|---------|
| `ErrNilContext` | Nil context passed to a blocking call |
| `ErrWSAlreadyOpen` | Session already has a gateway connection |
| `ErrWSNotFound` | Gateway write attempted without an open connection |
| `ErrWSShardInvalid` | `ShardCount <= 0` |
| `ErrWSShardBounds` | `ShardID < 0` or `ShardID >= ShardCount` |

## Examples

From the client module:

```powershell
cd clients\bot\goclient
go run .\examples\pingpong -t gcb_your_token
go run .\examples\send_message -t gcb_your_token -channel 2230469276416868353 -content "hello"
go run .\examples\presence -t gcb_your_token -status online -text "building things"
```

All examples accept `-endpoint`.

## Tests

The client test suite covers REST route construction, authorization headers, event dispatch, endpoint defaults, and race-sensitive gateway handler behavior:

```powershell
cd clients\bot\goclient
go test ./...
go test -race ./...
go vet ./...
golangci-lint run
```
