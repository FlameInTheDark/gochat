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

	app := NewApp(shut, logger)
	shut.Up(app)
	app.Start()
}
