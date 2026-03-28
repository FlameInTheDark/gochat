package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

func main() {
	obs, err := observability.Init("gochat-telemetry-gateway")
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
		os.Exit(1)
	}

	shut.Up(app)
	app.Start()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
