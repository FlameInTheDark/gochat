package main

import (
	"fmt"
	"log/slog"
	"strings"

	cfgpkg "github.com/FlameInTheDark/gochat/cmd/telemetrygateway/config"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/serviceauth"
	"github.com/FlameInTheDark/gochat/internal/shutter"
	"github.com/gofiber/fiber/v2"
	recm "github.com/gofiber/fiber/v2/middleware/recover"
)

type App struct {
	app   *fiber.App
	cfg   *cfgpkg.Config
	log   *slog.Logger
	proxy *telemetryProxy
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) (*App, error) {
	cfg, err := cfgpkg.LoadConfig(logger)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return nil, fmt.Errorf("jwt_secret is required")
	}

	proxy, err := newTelemetryProxy(cfg, logger, serviceauth.NewTokenManager(cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	fiberApp := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             cfg.BodyLimitBytes,
		ReadBufferSize:        8192,
		WriteBufferSize:       8192,
	})
	fiberApp.Get("/healthz", func(c *fiber.Ctx) error { return c.SendString("OK") })
	fiberApp.Use(observability.RequestContextMiddleware())
	rc := recm.ConfigDefault
	rc.EnableStackTrace = true
	fiberApp.Use(recm.New(rc))
	if cfg.Log {
		fiberApp.Use(observability.RequestLoggerWithLevel(logger, observability.ParseLogLevel(cfg.LogLevel)))
	}
	fiberApp.Use(observability.NewHTTPServerTelemetry("gochat-telemetry-gateway").Middleware())

	fiberApp.Post("/v1/traces", proxy.handle("traces"))
	fiberApp.Post("/v1/metrics", proxy.handle("metrics"))
	fiberApp.Post("/v1/logs", proxy.handle("logs"))

	return &App{
		app:   fiberApp,
		cfg:   cfg,
		log:   logger,
		proxy: proxy,
	}, nil
}

func (a *App) Start() {
	a.log.Info("Starting", slog.String("addr", a.cfg.ServerAddress))
	go func() {
		if err := a.app.Listen(a.cfg.ServerAddress); err != nil {
			a.log.Error("Error starting server", slog.String("error", err.Error()))
		}
	}()
}

func (a *App) Close() error { return a.app.Shutdown() }
