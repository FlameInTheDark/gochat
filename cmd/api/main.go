package main

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

func main() {
	obs, err := observability.Init("gochat-api")
	if err != nil {
		slog.Error("unable to initialize observability", slog.String("error", err.Error()))
	}
	logger := obs.Logger()
	shut := shutter.NewShutter(logger)
	defer shut.Down()
	shut.Up(obs)

	app, err := NewApp(shut, logger)
	if err != nil {
		logger.Error("Unable to create app", slog.String("error", err.Error()))
		return
	}
	app.Start()
}
