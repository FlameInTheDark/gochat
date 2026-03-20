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
	obs, err := observability.Init("gochat-indexer")
	if err != nil {
		slog.Error("unable to initialize observability", "error", err)
	}
	logger := obs.Logger()
	shut := shutter.NewShutter(logger)
	defer shut.Down()
	shut.Up(obs)
	app, err := NewApp(logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	shut.Up(app)
	err = app.Start()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Info("Service started")

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
}
