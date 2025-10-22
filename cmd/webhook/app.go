package main

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/cmd/webhook/auth"
	cfgpkg "github.com/FlameInTheDark/gochat/cmd/webhook/config"
	attentity "github.com/FlameInTheDark/gochat/cmd/webhook/endpoints/attachments"
	sfuentity "github.com/FlameInTheDark/gochat/cmd/webhook/endpoints/sfu"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/shutter"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
)

type App struct {
	server *server.Server
	cfg    *cfgpkg.Config
	log    *slog.Logger
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) (*App, error) {
	cfg, err := cfgpkg.LoadConfig(logger)
	if err != nil {
		return nil, err
	}

	// Dependencies
	tokens := auth.NewTokenManager(cfg.JWTSecret)
	disco, err := discovery.NewManager(cfg.EtcdEndpoints, cfg.EtcdPrefix, cfg.EtcdUsername, cfg.EtcdPassword)
	if err != nil {
		// If discovery is core to this service, fail fast
		return nil, err
	}

	var att attachment.Attachment
	if len(cfg.Cluster) > 0 {
		cql, err := db.NewCQLCon(cfg.ClusterKeyspace, db.NewDBLogger(logger), cfg.Cluster...)
		if err != nil {
			return nil, err
		}
		shut.Up(cql)
		att = attachment.New(cql)
	}

	// HTTP server
	s := server.NewServer()
	shut.Up(s)

	if cfg.Log {
		s.WithLogger(logger)
	}
	s.WithCORS()
	s.WithMetrics("gochat-webhook")
	if cfg.Swagger {
		s.WithSwagger("webhook")
	}

	// Register endpoints under /webhook
	s.Register(
		"/api/v1/webhook",
		sfuentity.New(logger, disco, tokens),
		attentity.New(logger, att, tokens),
	)

	return &App{server: s, cfg: cfg, log: logger}, nil
}

func (a *App) Start() {
	a.log.Info("Starting", slog.String("addr", a.cfg.ServerAddress))
	go func() {
		if err := a.server.Start(a.cfg.ServerAddress); err != nil {
			a.log.Error("Error starting server", slog.String("error", err.Error()))
		}
	}()
}

func (a *App) Close() error { return a.server.Close() }
