package main

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

func main() {
	obs, err := observability.Init("gochat-bot-api")
	if err != nil {
		slog.Error("unable to initialize observability", "error", err)
	}
	logger := obs.Logger()
	shut := shutter.NewShutter(logger)
	defer shut.Down()
	shut.Up(obs)

	app, err := NewApp(shut, logger)
	if err != nil {
		logger.Error("unable to initialize bot api", "error", err)
		return
	}
	shut.Up(app)
	app.Start()
}
