package main

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

func main() {
	obs, err := observability.Init("gochat-bot-ws")
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
