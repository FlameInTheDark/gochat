package main

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

//var rabbitConn *amqp.Connection

type Message struct {
	ID          int64        `json:"id"`
	ChannelID   int64        `json:"channel_id"`
	AuthorID    Author       `json:"author_id"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments"`
}

type Author struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Discriminator string `json:"discriminator"`
}

type Attachment struct {
	ContentType string `json:"content_type"`
	Filename    string `json:"filename"`
	Height      *int   `json:"height"`
	Width       *int   `json:"width"`
	URL         string `json:"url"`
	Size        int    `json:"size"`
}

func main() {
	obs, err := observability.Init("gochat-ws")
	if err != nil {
		slog.Error("unable to initialize observability", "error", err)
	}
	logger := obs.Logger()
	shut := shutter.NewShutter(logger)
	defer shut.Down()
	shut.Up(obs)

	app := NewApp(shut, logger)
	shut.Up(app)

	app.Start()
}
