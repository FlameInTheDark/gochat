package attachments

import (
	"log/slog"

	"github.com/FlameInTheDark/gochat/cmd/webhook/auth"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/gofiber/fiber/v2"
)

const entityName = "attachments"

type entity struct {
	name   string
	log    *slog.Logger
	tokens *auth.TokenManager
	att    attachment.Attachment
}

func New(log *slog.Logger, att attachment.Attachment, tokens *auth.TokenManager) server.Entity {
	return &entity{
		name:   entityName,
		log:    log,
		att:    att,
		tokens: tokens,
	}
}

func (e *entity) Name() string { return e.name }

func (e *entity) Init(router fiber.Router) {
	router.Post("/finalize", e.Finalize)
}
