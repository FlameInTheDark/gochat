package main

import (
	"github.com/FlameInTheDark/gochat/internal/mq"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FlameInTheDark/gochat/cmd/api/config"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/auth"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/message"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/user"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/webhook"
	"github.com/FlameInTheDark/gochat/internal/cache/vkc"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mailer"
	"github.com/FlameInTheDark/gochat/internal/mailer/providers/logmailer"
	"github.com/FlameInTheDark/gochat/internal/mailer/providers/sendpulse"
	"github.com/FlameInTheDark/gochat/internal/s3"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/shut"
)

type App struct {
	server *server.Server
	db     *db.CQLCon
	mailer *mailer.Mailer
	logger *slog.Logger
}

func NewApp(sh *shut.Shut, logger *slog.Logger) (*App, error) {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		return nil, err
	}

	// Database connection
	database, err := db.NewCQLCon(cfg.ClusterKeyspace, db.NewDBLogger(logger), cfg.Cluster...)
	if err != nil {
		return nil, err
	}
	sh.Up(database)

	queue, err := mq.New(cfg.RabbitMQHost, cfg.RabbitMQPort, cfg.RabbitMQUsername, cfg.RabbitMQPassword)
	if err != nil {
		return nil, err
	}
	sh.Up(queue)

	cache, err := vkc.New(cfg.KeyDB)
	if err != nil {
		return nil, err
	}
	sh.Up(cache)

	storage, err := s3.NewClient(cfg.S3Endpoint, cfg.S3AccessKeyID, cfg.S3SecretAccessKey, cfg.S3UseSSL)
	if err != nil {
		return nil, err
	}

	// Email notifier
	tmpl, err := mailer.NewEmailTemplate(cfg.EmailTemplate, cfg.BaseUrl, cfg.AppName, time.Now().Year())
	if err != nil {
		return nil, err
	}
	var provider mailer.Provider
	switch cfg.EmailProvider {
	case "log":
		provider = logmailer.New(logger)
	case "sendpulse":
		provider = sendpulse.New(cfg.SendpulseUserId, cfg.SendpulseSecret)
	default:
		provider = logmailer.New(logger)
	}
	m := mailer.NewMailer(provider, tmpl, mailer.User{Email: cfg.EmailSource, Name: cfg.EmailName})

	// ID generator setup
	idgen.New(0)

	// HTTP Server
	s := server.NewServer()
	sh.Up(s)

	s.WithCache(cache)

	// HTTP Middlewares
	if cfg.Swagger {
		s.WithSwagger("api")
	}
	if cfg.ApiLog {
		s.WithLogger()
	}
	s.AuthMiddleware(cfg.AuthSecret)
	s.RateLimitMiddleware(cfg.RateLimitRequests, cfg.RateLimitTime)

	// HTTP Router
	s.Register(
		"/api/v1",
		auth.New(database, m, cfg.AuthSecret, logger),
		user.New(database, logger),
		message.New(database, storage, queue, cfg.UploadLimit, logger),
		webhook.New(database, storage, logger),
	)

	return &App{
		server: s,
		db:     database,
		mailer: m,
		logger: logger,
	}, nil
}

func (app *App) Start(addr string) {
	app.logger.Info("Starting")
	go app.server.Start(addr)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
}
