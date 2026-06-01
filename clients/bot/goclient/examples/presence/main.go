package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/FlameInTheDark/gochat/clients/bot/goclient"
)

var (
	token    string
	endpoint string
	status   string
	text     string
)

func init() {
	flag.StringVar(&token, "t", os.Getenv("GOCHAT_BOT_TOKEN"), "GoChat bot token")
	flag.StringVar(&endpoint, "endpoint", "", "GoChat service URL")
	flag.StringVar(&status, "status", "online", "presence status: online, idle, dnd, or offline")
	flag.StringVar(&text, "text", "", "custom status text")
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

	session.AddHandlerOnce(func(session *goclient.Session, ready *goclient.Ready) {
		if err := session.UpdatePresence(status, text); err != nil {
			fmt.Println("error updating presence:", err)
			return
		}
		fmt.Printf("ready as %s; presence set to %q\n", ready.Bot.Name, status)
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
