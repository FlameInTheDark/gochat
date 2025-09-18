package main

import (
	"log/slog"
	"os"

	"github.com/FlameInTheDark/gochat/internal/shutter"
	"github.com/nats-io/nats.go"
)

//var rabbitConn *amqp.Connection

var natsConn *nats.Conn

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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	shut := shutter.NewShutter(logger)
	defer shut.Down()

	app := NewApp(shut, logger)
	shut.Up(app)

	app.Start()
}
