# Quick Start

This guide starts a small bot that responds to `ping` with `Pong!`.

## Prerequisites

- Go 1.22 or newer.
- A GoChat bot token that starts with `gcb_`.
- The bot installed in at least one guild where it can view and send messages.

## Run The Example Bot

From the repository root:

```powershell
cd clients\bot\goclient
go run .\examples\pingpong -t gcb_your_token
```

The client defaults to the hosted GoChat service:

```text
https://gochat.anticode.dev
```

Use another deployment with `-endpoint`:

```powershell
go run .\examples\pingpong -t gcb_your_token -endpoint http://127.0.0.1:3102
```

If REST and gateway are served from different origins, configure them in code with `WithAPIEndpoint` and `WithGatewayEndpoint`.

## Minimal Bot

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/FlameInTheDark/gochat/clients/bot/goclient"
)

func main() {
	session, err := goclient.New(os.Getenv("GOCHAT_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	var botID int64
	session.AddHandler(func(_ *goclient.Session, ready *goclient.Ready) {
		botID = ready.Bot.ID
		fmt.Println("ready as", ready.Bot.Name)
	})

	session.AddHandler(func(s *goclient.Session, event *goclient.MessageCreate) {
		msg := event.Message
		if msg.Author.ID == botID {
			return
		}
		if strings.TrimSpace(strings.ToLower(msg.Content)) == "ping" {
			_, _ = s.ChannelMessageSend(context.Background(), msg.ChannelID, "Pong!")
		}
	})

	if err := session.Open(); err != nil {
		panic(err)
	}
	defer session.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
```

## Send A Message Without A Gateway

Use REST only when the bot does not need realtime events:

```powershell
cd clients\bot\goclient
go run .\examples\send_message -t gcb_your_token -channel 2230469276416868353 -content "hello"
```

Equivalent Go:

```go
session, _ := goclient.New("gcb_your_token")
message, err := session.ChannelMessageSend(ctx, 2230469276416868353, "hello")
```

## Update Presence

```powershell
cd clients\bot\goclient
go run .\examples\presence -t gcb_your_token -status idle -text "Processing queue"
```

Allowed statuses are `online`, `idle`, `dnd`, and `offline`. Custom status text is limited to 255 Unicode characters.

## Next Steps

- Use [Authentication](authentication.md) before deploying.
- Use [REST API](rest.md) for message, channel, guild, and reaction operations.
- Use [Gateway](gateway.md) and [Events](events.md) to handle realtime events reliably.
