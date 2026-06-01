package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/FlameInTheDark/gochat/clients/bot/goclient"
)

var (
	token    string
	endpoint string
)

func init() {
	flag.StringVar(&token, "t", os.Getenv("GOCHAT_BOT_TOKEN"), "GoChat bot token")
	flag.StringVar(&endpoint, "endpoint", "", "GoChat service URL")
	flag.Parse()
}

func main() {
	if token == "" {
		fmt.Println("missing bot token; pass -t or set GOCHAT_BOT_TOKEN")
		return
	}

	options := []goclient.Option{}
	if endpoint != "" {
		options = append(options, goclient.WithEndpoint(endpoint))
	}
	session, err := goclient.New(token, options...)
	if err != nil {
		fmt.Println("error creating GoChat session:", err)
		return
	}

	var botUserID int64
	session.AddHandler(func(_ *goclient.Session, ready *goclient.Ready) {
		botUserID = ready.Bot.ID
		fmt.Printf("ready as %s on shard %d/%d\n", ready.Bot.Name, ready.ShardID, ready.ShardCount)
	})

	session.AddHandler(func(session *goclient.Session, event *goclient.MessageCreate) {
		message := event.Message
		if message.Author.IsBot || (botUserID != 0 && message.Author.ID == botUserID) {
			return
		}
		switch strings.ToLower(strings.TrimSpace(message.Content)) {
		case "ping":
			_, _ = session.ChannelMessageSend(context.Background(), message.ChannelID, "Pong!")
		case "pong":
			_, _ = session.ChannelMessageSend(context.Background(), message.ChannelID, "Ping!")
		}
	})

	if err := session.Open(); err != nil {
		fmt.Println("error opening gateway:", err)
		return
	}
	defer func() { _ = session.Close() }()

	fmt.Println("bot is running; press Ctrl+C to exit")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
