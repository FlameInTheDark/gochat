package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FlameInTheDark/gochat/cmd/auth/config"
	"github.com/FlameInTheDark/gochat/cmd/auth/endpoints/auth"
	"github.com/FlameInTheDark/gochat/internal/cache/vkc"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mailer"
	"github.com/FlameInTheDark/gochat/internal/mailer/providers/logmailer"
	"github.com/FlameInTheDark/gochat/internal/mailer/providers/sendpulse"
	"github.com/FlameInTheDark/gochat/internal/mailer/providers/smtp"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

type App struct {
	server *server.Server
	logger *slog.Logger
	addr   string
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) (*App, error) {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		return nil, err
	}

	pg := pgdb.NewDB(logger)
	err = pg.Connect(cfg.PGDSN, cfg.PGRetries)
	if err != nil {
		return nil, err
	}
	shut.Up(pg)

	cache, err := vkc.New(cfg.KeyDB)
	if err != nil {
		return nil, err
	}
	shut.Up(cache)

	// Email notifier
	tmpl, err := mailer.NewEmailTemplate(cfg.EmailTemplate, cfg.PasswordResetTemplate, cfg.BaseUrl, cfg.AppName, time.Now().Year())
	if err != nil {
		return nil, err
	}
	var provider mailer.Provider
	switch cfg.EmailProvider {
	case "log":
		provider = logmailer.New(logger)
	case "sendpulse":
		provider = sendpulse.New(cfg.SendpulseUserId, cfg.SendpulseSecret)
	case "smtp":
		provider = smtp.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPUseTLS)
	default:
		provider = logmailer.New(logger)
	}
	m := mailer.NewMailer(provider, tmpl, mailer.User{Email: cfg.EmailSource, Name: cfg.EmailName})

	// ID generator setup
	idgen.New(0)

	// HTTP Server
	s := server.NewServer()
	shut.Up(s)

	s.WithCache(cache)

	// HTTP Middlewares
	if cfg.Swagger {
		s.WithSwagger("auth")
	}
	if cfg.ApiLog {
		s.WithLogger(logger)
	}
	s.WithCORS()
	s.WithMetrics()
	s.WithIdempotency(cache.Client(), cfg.IdempotencyStorageLifetime)
	s.RateLimitMiddleware(cfg.RateLimitRequests, cfg.RateLimitTime)
	s.AuthMiddleware(cfg.AuthSecret)

	// HTTP Router
	s.Register(
		"/api/v1",
		auth.New(pg, m, cfg.AuthSecret, logger, helper.RequireTokenType("refresh", "refresh")),
	)

	return &App{
		server: s,
		logger: logger,
		addr:   cfg.ServerAddress,
	}, nil
}

func (app *App) Start() {
	app.logger.Info("Starting Auth Service")
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
