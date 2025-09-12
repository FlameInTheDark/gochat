package main

import (
	"log/slog"
	"os"

	"github.com/FlameInTheDark/gochat/internal/shutter"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	shut := shutter.NewShutter(logger)
	defer shut.Down()

	app, err := NewApp(shut, logger)
	if err != nil {
		logger.Error("Unable to create app", slog.String("error", err.Error()))
		return
	}
	app.Start()
}
