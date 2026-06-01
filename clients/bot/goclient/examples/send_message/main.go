package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/FlameInTheDark/gochat/clients/bot/goclient"
)

var (
	token     string
	endpoint  string
	channelID int64
	content   string
)

func init() {
	flag.StringVar(&token, "t", os.Getenv("GOCHAT_BOT_TOKEN"), "GoChat bot token")
	flag.StringVar(&endpoint, "endpoint", "", "GoChat service URL")
	flag.Int64Var(&channelID, "channel", 0, "channel ID")
	flag.StringVar(&content, "content", "hello from GoChat", "message content")
	flag.Parse()
}

func main() {
	if token == "" {
		fmt.Println("missing bot token; pass -t or set GOCHAT_BOT_TOKEN")
		return
	}
	if channelID == 0 {
		fmt.Println("missing channel ID; pass -channel")
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	message, err := session.ChannelMessageSend(ctx, channelID, content)
	if err != nil {
		fmt.Println("error sending message:", err)
		return
	}
	fmt.Printf("sent message %d to channel %d\n", message.ID, message.ChannelID)
}
