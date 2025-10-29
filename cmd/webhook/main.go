package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/FlameInTheDark/gochat/internal/shutter"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	shut := shutter.NewShutter(logger)
	defer shut.Down()

	app, err := NewApp(shut, logger)
	if err != nil {
		logger.Error("Unable to create webhook app", slog.String("error", err.Error()))
		return
	}
	shut.Up(app)
	app.Start()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
