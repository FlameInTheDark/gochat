package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FlameInTheDark/gochat/cmd/botapi/config"
	botguild "github.com/FlameInTheDark/gochat/cmd/botapi/endpoints/guild"
	botmessage "github.com/FlameInTheDark/gochat/cmd/botapi/endpoints/message"
	botuser "github.com/FlameInTheDark/gochat/cmd/botapi/endpoints/user"
	"github.com/FlameInTheDark/gochat/internal/botauth"
	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	botrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/bot"
	userrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/botnats"
	"github.com/FlameInTheDark/gochat/internal/mq/nats"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

type App struct {
	server *server.Server
	db     *db.CQLCon
	logger *slog.Logger
	addr   string
}

const startupProbeTimeout = 3 * time.Second
const startupProbeMaxDelay = 5 * time.Second

func retryStartupProbe(ctx context.Context, logger *slog.Logger, dependency string, maxRetries int, probe func(context.Context) error) error {
	delay := time.Second
	for attempt := 1; ; attempt++ {
		probeCtx, cancel := context.WithTimeout(ctx, startupProbeTimeout)
		err := probe(probeCtx)
		cancel()
		if err == nil {
			return nil
		}
		if logger != nil {
			logger.Warn("startup dependency probe failed; retrying", slog.String("dependency", dependency), slog.Int("attempt", attempt), slog.String("error", err.Error()))
		}
		if maxRetries > 0 && attempt >= maxRetries {
			return fmt.Errorf("probe %s after %d attempts: %w", dependency, attempt, err)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("probe %s canceled: %w", dependency, ctx.Err())
		case <-time.After(delay):
		}
		delay *= 2
		if delay > startupProbeMaxDelay {
			delay = startupProbeMaxDelay
		}
	}
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
	if err = pg.Connect(cfg.PGDSN, pgdb.ConnectOptions{
		DriverName:   cfg.PGDriver,
		MaxRetries:   cfg.PGRetries,
		QueryLog:     cfg.PGQueryLog,
		MaxOpenConns: cfg.PGMaxOpenConns,
		MaxIdleConns: cfg.PGMaxIdleConns,
	}); err != nil {
		return nil, err
	}
	shut.Up(pg)
	postgresProbeCtx, cancelPostgresProbe := context.WithCancel(context.Background())
	shut.UpFunc(cancelPostgresProbe)
	go pg.StartProbeLoop(postgresProbeCtx, 30*time.Second)

	logger.Info("Connecting to NATS")
	nt, err := nats.New(cfg.NATSConnString)
	if err != nil {
		return nil, err
	}
	shut.Up(nt)

	logger.Info("Connecting to Bot NATS")
	botNt, err := botnats.New(cfg.BotNATSConnString, cfg.BotEventPartitions)
	if err != nil {
		return nil, err
	}
	shut.Up(botNt)
	qt := mq.NewFanoutTransporter(nt, botNt)

	logger.Info("Connecting to KeyDB")
	cache, err := kvs.New(cfg.KeyDB, kvs.Options{
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
	})
	if err != nil {
		return nil, err
	}
	shut.Up(cache)

	baseCtx := context.Background()
	if err := retryStartupProbe(baseCtx, logger, "postgres", cfg.PGRetries, func(ctx context.Context) error {
		return pg.Conn().PingContext(ctx)
	}); err != nil {
		return nil, err
	}
	if err := retryStartupProbe(baseCtx, logger, "scylla", cfg.PGRetries, database.Ping); err != nil {
		return nil, err
	}
	if err := retryStartupProbe(baseCtx, logger, "redis", cfg.PGRetries, cache.Ping); err != nil {
		return nil, err
	}
	if err := retryStartupProbe(baseCtx, logger, "nats", cfg.PGRetries, nt.Ping); err != nil {
		return nil, err
	}
	if err := retryStartupProbe(baseCtx, logger, "bot_nats", cfg.PGRetries, botNt.Ping); err != nil {
		return nil, err
	}

	s := server.NewServer()
	shut.Up(s)
	s.WithCache(cache)
	s.WithLoggerLevel(logger, observability.ParseLogLevel(cfg.LogLevel))
	s.WithCORS()
	s.WithCompression()
	s.WithMetrics("gochat-bot-api")
	s.WithIdempotency(cache.Client(), cfg.IdempotencyStorageLifetime)
	s.RateLimitPipedMiddleware(cfg.RateLimitRequests, cfg.RateLimitTime)
	s.Use(botauth.Middleware(botrepo.New(pg.Conn()), userrepo.New(pg.Conn())))
	s.Register(
		"/bot/api/v1",
		botuser.New(logger),
		botguild.New(pg, logger),
		botmessage.New(database, pg, qt, logger),
	)

	return &App{server: s, db: database, logger: logger, addr: cfg.ServerAddress}, nil
}

func (app *App) Start() {
	app.logger.Info("Starting bot api", "addr", app.addr)
	go func() {
		if err := app.server.Start(app.addr); err != nil {
			app.logger.Error("Error starting bot api", "error", err)
			os.Exit(1)
		}
	}()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
}

func (app *App) Close() error {
	return app.server.Close()
}
