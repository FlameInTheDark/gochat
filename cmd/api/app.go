package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FlameInTheDark/gochat/cmd/api/config"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/guild"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/message"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/search"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/user"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/voice"
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
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
	"github.com/gofiber/fiber/v2"
	natsio "github.com/nats-io/nats.go"
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

	logger.Info("Connecting to ScyllaDB")
	database, err := db.NewCQLCon(cfg.ClusterKeyspace, db.NewDBLogger(logger), cfg.Cluster...)
	if err != nil {
		return nil, err
	}
	shut.Up(database)

	logger.Info("Connecting to PostgreSQL")
	pg := pgdb.NewDB(logger)
	err = pg.Connect(cfg.PGDSN, cfg.PGRetries)
	if err != nil {
		return nil, err
	}
	shut.Up(pg)

	var storage *s3.Client
	if cfg.S3Endpoint != "" {
		logger.Info("Connecting to S3")
		storage, err = s3.NewClient(cfg.S3Endpoint, cfg.S3AccessKeyID, cfg.S3SecretAccessKey, cfg.S3Region, cfg.S3Bucket, cfg.S3UseSSL)
		if err != nil {
			return nil, err
		}
	}

	logger.Info("Connecting to NATS")
	var qt mq.SendTransporter
	nt, err := nats.New(cfg.NatsConnString)
	if err != nil {
		return nil, err
	}
	shut.Up(nt)
	qt = nt

	logger.Info("Connecting to Indexer NATS")
	imq, err := indexmq.NewIndexMQ(cfg.IndexerNatsConnString)
	if err != nil {
		return nil, err
	}

	logger.Info("Connecting to KeyDB")
	cache, err := kvs.New(cfg.KeyDB)
	if err != nil {
		return nil, err
	}
	shut.Up(cache)

	logger.Info("Connecting to NATS for SFU occupancy updates")
	if cfg.NatsConnString != "" {
		if occNc, err := natsio.Connect(cfg.NatsConnString, natsio.Compression(true)); err == nil {
			shut.UpFunc(func() { _ = occNc.Drain() })
			_, _ = occNc.Subscribe("voice.occ", func(m *natsio.Msg) {
				type occ struct {
					Channel int64 `json:"channel"`
					Delta   int   `json:"delta"`
				}
				var o occ
				if err := json.Unmarshal(m.Data, &o); err != nil || o.Channel <= 0 || (o.Delta != 1 && o.Delta != -1) {
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				key := fmt.Sprintf("voice:occ:%d", o.Channel)
				if o.Delta == 1 {
					_, _ = cache.Incr(ctx, key)
				} else {
					if n, err := cache.GetInt64(ctx, key); err == nil {
						if n > 0 {
							n--
						}
						_ = cache.SetInt64(ctx, key, n)
						if n <= 0 {
							_ = cache.Delete(ctx, fmt.Sprintf("voice:route:%d", o.Channel))
						}
					}
				}
			})
		}
	}

	logger.Info("Connecting to OpenSearch")
	searchService, err := msgsearch.NewSearch(cfg.OSAddresses, cfg.OSInsecureSkipVerify, cfg.OSUsername, cfg.OSPassword)
	if err != nil {
		return nil, err
	}

	logger.Info("Connecting to Etcd")
	disco, err := discovery.NewManager(cfg.EtcdEndpoints, cfg.EtcdPrefix, cfg.EtcdUsername, cfg.EtcdPassword)
	if err != nil {
		return nil, err
	}

	idgen.New(0)

	logger.Info("Registering HTTP server")
	s := server.NewServer()
	shut.Up(s)

	s.WithCache(cache)
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
	s.RateLimitPipedMiddleware(cfg.RateLimitRequests, cfg.RateLimitTime)
	s.Use(helper.RequireTokenType("access", "api"))
	s.Use(func(c *fiber.Ctx) error {
		c.Locals("base_url", cfg.BaseUrl)
		return c.Next()
	})

	s.Register(
		"/api/v1",
		user.New(database, pg, qt, cache, cfg.AttachmentTTLMinutes*60, logger),
		message.New(database, pg, qt, imq, cfg.UploadLimit, cfg.AttachmentTTLMinutes*60, cache, logger),
		guild.New(database, pg, qt, imq, cache, storage, cfg.AttachmentTTLMinutes*60, cfg.AuthSecret, cfg.VoiceDefaultRegion, disco, extractRegionIDs(cfg.VoiceRegions), logger),
		voice.New(convertRegions(cfg.VoiceRegions), logger),
		search.New(database, pg, searchService, logger),
	)

	return &App{server: s, db: database, logger: logger, addr: cfg.ServerAddress}, nil
}

func extractRegionIDs(v []config.VoiceRegion) []string {
	if len(v) == 0 {
		return nil
	}
	out := make([]string, 0, len(v))
	for _, r := range v {
		if r.ID != "" {
			out = append(out, r.ID)
		}
	}
	return out
}

func convertRegions(v []config.VoiceRegion) []voice.Region {
	if len(v) == 0 {
		return nil
	}
	out := make([]voice.Region, 0, len(v))
	for _, r := range v {
		if r.ID == "" {
			continue
		}
		out = append(out, voice.Region{ID: r.ID, Name: r.Name})
	}
	return out
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
