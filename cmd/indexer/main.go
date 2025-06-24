package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/FlameInTheDark/gochat/internal/shutter"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(slog.String("service", "indexer"))
	shut := shutter.NewShutter(logger)
	defer shut.Down()
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
