package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FlameInTheDark/gochat/cmd/botws/config"
	botwshandler "github.com/FlameInTheDark/gochat/cmd/botws/handler"
	"github.com/FlameInTheDark/gochat/cmd/ws/hub"
	"github.com/FlameInTheDark/gochat/internal/botauth"
	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	botrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/bot"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/dmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/groupdmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/presence"
	"github.com/FlameInTheDark/gochat/internal/shutter"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	recm "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type App struct {
	natsConn         *nats.Conn
	presenceNATSConn *nats.Conn
	hub              *hub.Hub
	app              *fiber.App
	pg               *pgdb.DB
	cache            *kvs.Cache
	cfg              *config.Config
	log              *slog.Logger
	bot              botrepo.Bot
	user             user.User
	dm               dmchannel.DmChannel
	gdm              groupdmchannel.GroupDMChannel
	perm             rolecheck.RoleCheck
	handler          *botwshandler.Gateway
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) *App {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("unable to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	natsCon, err := nats.Connect(cfg.BotNATSConnString, nats.Compression(true))
	if err != nil {
		logger.Error("unable to connect to Bot NATS", slog.String("error", err.Error()))
		os.Exit(1)
	}
	shut.UpFunc(natsCon.Close)

	presenceNATSCon, err := nats.Connect(cfg.NATSConnString, nats.Compression(true))
	if err != nil {
		logger.Error("unable to connect to NATS", slog.String("error", err.Error()))
		os.Exit(1)
	}
	shut.UpFunc(presenceNATSCon.Close)

	pg := pgdb.NewDB(logger)
	if err := pg.Connect(cfg.PGDSN, pgdb.ConnectOptions{DriverName: cfg.PGDriver, MaxRetries: cfg.PGRetries}); err != nil {
		logger.Error("unable to connect to pg", slog.String("error", err.Error()))
		os.Exit(1)
	}
	shut.Up(pg)
	postgresProbeCtx, cancelPostgresProbe := context.WithCancel(context.Background())
	shut.UpFunc(cancelPostgresProbe)
	go pg.StartProbeLoop(postgresProbeCtx, 30*time.Second)

	kv, err := kvs.New(cfg.CacheAddr)
	if err != nil {
		logger.Error("unable to connect to cache", slog.String("error", err.Error()))
		os.Exit(1)
	}
	shut.Up(kv)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/healthz", func(c *fiber.Ctx) error { return c.SendString("OK") })
	app.Use(observability.RequestContextMiddleware())
	app.Use(observability.RequestLogger(logger))
	app.Use(observability.NewHTTPServerTelemetry("gochat-bot-ws").Middleware())
	app.Use(recm.New())
	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/healthz" || websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	bot := botrepo.New(pg.Conn())
	dm := dmchannel.New(pg.Conn())
	gdm := groupdmchannel.New(pg.Conn())
	perm := rolecheck.New(pg)
	wsHub := hub.New(natsCon, logger)
	instanceID := uuid.NewString()

	return &App{
		natsConn:         natsCon,
		presenceNATSConn: presenceNATSCon,
		hub:              wsHub,
		app:              app,
		pg:               pg,
		cache:            kv,
		cfg:              cfg,
		log:              logger,
		bot:              bot,
		user:             user.New(pg.Conn()),
		dm:               dm,
		gdm:              gdm,
		perm:             perm,
		handler:          botwshandler.New(wsHub, kv, presence.NewStore(kv), presenceNATSCon, instanceID, bot, dm, gdm, perm, cfg.HearthBeatTimeout, logger),
	}
}

func (a *App) Start() {
	a.app.Get("/bot/ws", a.authenticateUpgrade, websocket.New(a.handler.WSHandler, websocket.Config{}))
	host := a.cfg.Host
	if host == "" {
		host = ":3101"
	}
	a.log.Info("Bot WebSocket server starting", slog.String("addr", host), slog.String("path", "/bot/ws"))
	go func() {
		if err := a.app.Listen(host); err != nil {
			a.log.Error("failed to start bot ws app", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
}

func (a *App) Close() error {
	return a.app.ShutdownWithTimeout(30 * time.Second)
}

func (a *App) authenticateUpgrade(c *fiber.Ctx) error {
	principal, err := botauth.Authenticate(c.UserContext(), a.bot, a.user, c.Get("Authorization"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}
	c.Locals("bot_principal", principal)
	return c.Next()
}
