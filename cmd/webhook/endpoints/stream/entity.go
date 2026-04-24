package stream

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/presence"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/serviceauth"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
	"github.com/gofiber/fiber/v2"
	natsio "github.com/nats-io/nats.go"
)

const entityName = "stream"

type entity struct {
	name   string
	log    *slog.Logger
	disco  discovery.Manager
	tokens *serviceauth.TokenManager
	cache  cache.Cache
	mqt    mq.SendTransporter
	pstore *presence.Store
	nats   *natsio.Conn
}

func New(log *slog.Logger, disco discovery.Manager, tokens *serviceauth.TokenManager, cache cache.Cache, mqt mq.SendTransporter, pstore *presence.Store, nats *natsio.Conn) server.Entity {
	return &entity{
		name:   entityName,
		log:    log,
		disco:  disco,
		tokens: tokens,
		cache:  cache,
		mqt:    mqt,
		pstore: pstore,
		nats:   nats,
	}
}

func (e *entity) Name() string { return e.name }

func (e *entity) Init(router fiber.Router) {
	router.Post("/heartbeat", e.Heartbeat)
	router.Post("/start", e.Start)
	router.Post("/stop", e.Stop)
	router.Post("/alive", e.Alive)
}
