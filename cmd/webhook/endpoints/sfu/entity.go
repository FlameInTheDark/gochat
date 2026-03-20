package sfu

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/cmd/webhook/auth"
	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
	"github.com/gofiber/fiber/v2"
)

const entityName = "sfu"

type entity struct {
	name   string
	log    *slog.Logger
	disco  discovery.Manager
	tokens *auth.TokenManager
	cache  cache.Cache
	mqt    mq.SendTransporter
}

func New(log *slog.Logger, disco discovery.Manager, tokens *auth.TokenManager, cache cache.Cache, mqt mq.SendTransporter) server.Entity {
	return &entity{
		name:   entityName,
		log:    log,
		disco:  disco,
		tokens: tokens,
		cache:  cache,
		mqt:    mqt,
	}
}

func (e *entity) Name() string { return e.name }

func (e *entity) Init(router fiber.Router) {
	router.Post("/heartbeat", e.Heartbeat)
	router.Post("/voice/join", e.ChannelUserJoin)
	router.Post("/voice/leave", e.ChannelUserLeave)
	router.Post("/channel/alive", e.ChannelAlive)
}
