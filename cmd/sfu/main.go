package main

import (
	"log/slog"
	"os"

	"github.com/FlameInTheDark/gochat/cmd/sfu/config"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/shutter"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("unable to load config", "error", err)
		os.Exit(1)
	}

	obs, err := observability.Init("gochat-sfu",
		attribute.String("voice.region", cfg.Region),
		attribute.String("service.instance.id", cfg.ServiceID),
	)
	if err != nil {
		slog.Error("unable to initialize observability", "error", err)
	}
	logger := obs.Logger()
	shut := shutter.NewShutter(logger)
	defer shut.Down()
	shut.Up(obs)

	app := NewApp(shut, logger, cfg)
	shut.Up(app)
	app.Start()
}
