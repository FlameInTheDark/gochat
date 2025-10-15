package main

import (
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/FlameInTheDark/gochat/cmd/api/config"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/guild"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/message"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/search"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/user"
	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/indexmq"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/nats"
	"github.com/FlameInTheDark/gochat/internal/msgsearch"
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

	pg := pgdb.NewDB(logger)
	err = pg.Connect(cfg.PGDSN, cfg.PGRetries)
	if err != nil {
		return nil, err
	}
	shut.Up(pg)

	var qt mq.SendTransporter
	nt, err := nats.New(cfg.NatsConnString)
	if err != nil {
		return nil, err
	}
	shut.Up(nt)
	qt = nt

	imq, err := indexmq.NewIndexMQ(cfg.IndexerNatsConnString)
	if err != nil {
		return nil, err
	}

	cache, err := kvs.New(cfg.KeyDB)
	if err != nil {
		return nil, err
	}
	shut.Up(cache)

	storage, err := s3.NewClient(cfg.S3Endpoint, cfg.S3AccessKeyID, cfg.S3SecretAccessKey, cfg.S3Region, cfg.S3Bucket, cfg.S3UseSSL)
	if err != nil {
		return nil, err
	}

	// Initialize message search service
	searchService, err := msgsearch.NewSearch(cfg.OSAddresses, cfg.OSInsecureSkipVerify, cfg.OSUsername, cfg.OSPassword)
	if err != nil {
		return nil, err
	}

	// ID generator setup
	idgen.New(0)

	// HTTP Server
	s := server.NewServer()
	shut.Up(s)

	s.WithCache(cache)

	// HTTP Middlewares
	if cfg.Swagger {
		s.WithSwagger("api")
	}
	if cfg.ApiLog {
		s.WithLogger(logger)
	}
	s.WithCORS()
	s.WithMetrics("gochat-api")
	s.WithIdempotency(cache.Client(), cfg.IdempotencyStorageLifetime)
	s.AuthMiddleware(cfg.AuthSecret)
	//s.RateLimitMiddleware(cfg.RateLimitRequests, cfg.RateLimitTime)
	s.RateLimitPipedMiddleware(cfg.RateLimitRequests, cfg.RateLimitTime)
	s.Use(helper.RequireTokenType("access", "api"))

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

	// HTTP Router
	s.Register(
		"/api/v1",
		user.New(database, pg, qt, logger),
		message.New(database, pg, storage, qt, imq, cfg.UploadLimit, publicBase, logger),
		guild.New(database, pg, qt, cache, logger),
		search.New(database, pg, searchService, logger),
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
