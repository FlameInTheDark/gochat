package main

import (
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/FlameInTheDark/gochat/cmd/attachments/config"
	attachments "github.com/FlameInTheDark/gochat/cmd/attachments/endpoints/attachments"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/s3"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

type App struct {
	server *server.Server
	db     *db.CQLCon
	logger *slog.Logger

	addr string
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) (*App, error) {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		return nil, err
	}

	// Database connection
	database, err := db.NewCQLCon(cfg.ClusterKeyspace, db.NewDBLogger(logger), cfg.Cluster...)
	if err != nil {
		return nil, err
	}
	shut.Up(database)

	storage, err := s3.NewClient(cfg.S3Endpoint, cfg.S3AccessKeyID, cfg.S3SecretAccessKey, cfg.S3Region, cfg.S3Bucket, cfg.S3UseSSL)
	if err != nil {
		return nil, err
	}

	// Compute public base URL for objects
	publicBase := strings.TrimRight(cfg.S3ExternalURL, "/")
	if publicBase == "" {
		endp := cfg.S3Endpoint
		low := strings.ToLower(endp)
		if !strings.HasPrefix(low, "http://") && !strings.HasPrefix(low, "https://") {
			if cfg.S3UseSSL {
				endp = "https://" + endp
			} else {
				endp = "http://" + endp
			}
		}
		endp = strings.TrimRight(endp, "/")
		publicBase = endp + "/" + strings.Trim(cfg.S3Bucket, "/")
	}

	// HTTP Server
	s := server.NewServer()
	shut.Up(s)

	// HTTP Middlewares
	s.WithCORS()
	s.WithMetrics("gochat-attachments")
	s.AuthMiddleware(cfg.AuthSecret)
	s.Use(helper.RequireTokenType("access", "api"))

	// HTTP Router
	s.Register(
		"/api/v1",
		attachments.New(database, storage, publicBase, logger),
	)

	return &App{
		server: s,
		db:     database,
		logger: logger,
		addr:   cfg.ServerAddress,
	}, nil
}

func (app *App) Start() {
	app.logger.Info("Starting")
	go func() {
		err := app.server.Start(app.addr)
		if err != nil {
			app.logger.Error("Error starting server", "error", err)
			os.Exit(1)
		}
	}()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
}
