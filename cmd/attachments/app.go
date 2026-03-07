package main

import (
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/FlameInTheDark/gochat/cmd/attachments/config"
	attachments "github.com/FlameInTheDark/gochat/cmd/attachments/endpoints/attachments"
	avatars "github.com/FlameInTheDark/gochat/cmd/attachments/endpoints/avatars"
	publicemoji "github.com/FlameInTheDark/gochat/cmd/attachments/endpoints/emoji"
	emojis "github.com/FlameInTheDark/gochat/cmd/attachments/endpoints/emojis"
	icons "github.com/FlameInTheDark/gochat/cmd/attachments/endpoints/icons"
	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/nats"
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

	database, err := db.NewCQLCon(cfg.ClusterKeyspace, db.NewDBLogger(logger), cfg.Cluster...)
	if err != nil {
		return nil, err
	}
	shut.Up(database)

	storage, err := s3.NewClient(cfg.S3Endpoint, cfg.S3AccessKeyID, cfg.S3SecretAccessKey, cfg.S3Region, cfg.S3Bucket, cfg.S3UseSSL)
	if err != nil {
		return nil, err
	}

	pg := pgdb.NewDB(logger)
	if err := pg.Connect(cfg.PGDSN, cfg.PGRetries); err != nil {
		return nil, err
	}
	shut.Up(pg)

	cache, err := kvs.New(cfg.KeyDB)
	if err != nil {
		return nil, err
	}
	shut.Up(cache)

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

	s := server.NewServer()
	shut.Up(s)

	s.WithCORS()
	s.WithMetrics("gochat-attachments")
	s.AuthMiddleware(cfg.AuthSecret)
	s.Use(helper.RequireTokenType("access", "api"))

	nt, err := nats.New(cfg.NatsConnString)
	if err != nil {
		return nil, err
	}
	shut.Up(nt)

	s.Register(
		"/api/v1/upload",
		attachments.New(database, pg, storage, publicBase, logger),
		avatars.New(database, pg, storage, publicBase, nt, logger),
		emojis.New(pg, storage, cache, nt, publicBase, logger),
		icons.New(database, pg, storage, publicBase, nt, logger),
	)
	s.Register("", publicemoji.New(publicBase, logger))

	return &App{server: s, db: database, logger: logger, addr: cfg.ServerAddress}, nil
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
