package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	recm "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/nats-io/nats.go"

	"github.com/FlameInTheDark/gochat/cmd/ws/auth"
	"github.com/FlameInTheDark/gochat/cmd/ws/config"
	"github.com/FlameInTheDark/gochat/cmd/ws/hub"
	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	authenticationrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/authentication"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/presence"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

type App struct {
	jwt      *auth.Auth
	natsConn *nats.Conn
	hub      *hub.Hub
	app      *fiber.App
	cdb      *db.CQLCon
	pg       *pgdb.DB
	cache    *kvs.Cache

	shut *shutter.Shut
	cfg  *config.Config
	log  *slog.Logger
	wsm  *observability.WSTelemetry
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) *App {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("unable to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	natsCon, err := nats.Connect(cfg.NATSConnString, nats.Compression(true))
	if err != nil {
		logger.Error("unable to connect to NATS", slog.String("error", err.Error()))
		os.Exit(1)
	}
	shut.UpFunc(natsCon.Close)

	dbcon, err := db.NewCQLCon(cfg.ClusterKeyspace, db.NewDBLogger(logger), cfg.Cluster...)
	if err != nil {
		logger.Error("unable to connect to cluster", slog.String("error", err.Error()))
		os.Exit(1)
	}
	shut.Up(dbcon)

	pg := pgdb.NewDB(logger)
	err = pg.Connect(cfg.PGDSN, pgdb.ConnectOptions{DriverName: cfg.PGDriver, MaxRetries: cfg.PGRetries})
	if err != nil {
		logger.Error("unable to connect to pg", slog.String("error", err.Error()))
		os.Exit(1)
	}
	shut.Up(pg)
	postgresProbeCtx, cancelPostgresProbe := context.WithCancel(context.Background())
	shut.UpFunc(cancelPostgresProbe)
	go pg.StartProbeLoop(postgresProbeCtx, 30*time.Second)

	// Cache (Redis)
	kv, err := kvs.New(cfg.CacheAddr)
	if err != nil {
		logger.Error("unable to connect to cache", slog.String("error", err.Error()))
		os.Exit(1)
	}
	shut.Up(kv)

	sessionChecker := helper.NewSessionVersionChecker(authenticationrepo.New(pg.Conn()), kv)
	jwtauth := auth.New(cfg.AuthSecret, "gochat", "api", sessionChecker)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(observability.RequestContextMiddleware())
	app.Use(observability.RequestLogger(logger))
	app.Use(observability.NewHTTPServerTelemetry("gochat-ws").Middleware())
	app.Use(recm.New())

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	wsHub := hub.New(natsCon, logger)

	a := &App{
		jwt:      jwtauth,
		natsConn: natsCon,
		hub:      wsHub,
		app:      app,
		cdb:      dbcon,
		pg:       pg,
		cache:    kv,
		cfg:      cfg,
		log:      logger,
		shut:     shut,
		wsm:      observability.NewWSTelemetry("gochat-ws"),
	}

	presenceReconcileCtx, cancelPresenceReconcile := context.WithCancel(context.Background())
	shut.UpFunc(cancelPresenceReconcile)
	go a.startPresenceReconciler(presenceReconcileCtx)

	return a
}

func (a *App) Start() {
	wscfg := websocket.Config{
		RecoverHandler: func(conn *websocket.Conn) {
			if err := recover(); err != nil {
				err := conn.WriteJSON(fiber.Map{"customError": "error occurred"})
				if err != nil {
					a.log.Error("failed to send error", slog.String("error", err.Error()))
				}
			}
		},
	}
	a.app.Get("/subscribe", websocket.New(a.wsHandler, wscfg))

	a.log.Info("Server starting", slog.String("addr", ":3100"))
	go func() {
		err := a.app.Listen(":3100")
		if err != nil {
			a.log.Error("failed to start app", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
}

func (a *App) Close() error {
	return a.app.ShutdownWithTimeout(time.Second * 30)
}

func (a *App) startPresenceReconciler(ctx context.Context) {
	if a == nil || a.cache == nil {
		return
	}
	store := presence.NewStore(a.cache)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			published, err := presence.ReconcileTouched(opCtx, store, a.natsConn, a.presenceTTLSeconds(), 1000)
			cancel()
			if err != nil {
				a.log.Warn("presence reconciliation failed", slog.String("error", err.Error()))
				continue
			}
			if published > 0 {
				a.log.Debug("presence reconciliation published corrections", slog.Int("count", published))
			}
		}
	}
}

func (a *App) presenceTTLSeconds() int64 {
	if a == nil || a.cfg == nil {
		return 60
	}
	ttl := a.cfg.HearthBeatTimeout * 2 / 1000
	if ttl < 1 {
		return 1
	}
	return ttl
}
