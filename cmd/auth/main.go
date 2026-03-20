package main

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

func main() {
	obs, err := observability.Init("gochat-auth")
	if err != nil {
		slog.Error("unable to initialize observability", "error", err)
	}
	logger := obs.Logger()

	shut := shutter.NewShutter(logger)
	defer shut.Down()
	shut.Up(obs)

	app, err := NewApp(shut, logger)
	if err != nil {
		logger.Error("Failed to create app", "error", err)
		return
	}

	app.Start()
}
