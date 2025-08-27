package main

import (
	"log/slog"
	"os"

	"github.com/FlameInTheDark/gochat/internal/shutter"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create a shutter for graceful shutdown
	shut := shutter.NewShutter(logger)
	defer shut.Down()

	// Create app
	app, err := NewApp(shut, logger)
	if err != nil {
		logger.Error("Failed to create app", "error", err)
		os.Exit(1)
	}

	// Start app
	app.Start()
}
