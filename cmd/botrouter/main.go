package main

import (
	"log/slog"
	"os"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

func main() {
	obs, err := observability.Init("gochat-bot-router")
	if err != nil {
		slog.Error("unable to initialize observability", "error", err)
	}
	logger := obs.Logger()
	shut := shutter.NewShutter(logger)
	defer shut.Down()
	shut.Up(obs)

	app, err := NewApp(shut, logger)
	if err != nil {
		logger.Error("unable to initialize bot router", "error", err)
		os.Exit(1)
	}
	shut.Up(app)
	app.Start()
}
