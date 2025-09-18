package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/shutter"
	slogfiber "github.com/samber/slog-fiber"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	recm "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/nats-io/nats.go"

	"github.com/FlameInTheDark/gochat/cmd/ws/auth"
	"github.com/FlameInTheDark/gochat/cmd/ws/config"
)

type App struct {
	jwt      *auth.Auth
	natsConn *nats.Conn
	app      *fiber.App
	cdb      *db.CQLCon
	pg       *pgdb.DB

	shut *shutter.Shut
	cfg  *config.Config
	log  *slog.Logger
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) *App {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("unable to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	natsCon, err := nats.Connect(cfg.NatsConnString, nats.Compression(true))
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
	err = pg.Connect(cfg.PGDSN, cfg.PGRetries)
	if err != nil {
		logger.Error("unable to connect to pg", slog.String("error", err.Error()))
		os.Exit(1)
	}
	shut.Up(pg)

	jwtauth := auth.New(cfg.AuthSecret, "gochat", "api")

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	logMiddleware := slogfiber.NewWithFilters(
		logger,
		slogfiber.IgnorePath("/metrics"),
	)
	app.Use(logMiddleware)
	app.Use(recm.New())

	app.Use("/", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	return &App{
		jwt:      jwtauth,
		natsConn: natsCon,
		app:      app,
		cdb:      dbcon,
		pg:       pg,
		cfg:      cfg,
		log:      logger,
		shut:     shut,
	}
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
