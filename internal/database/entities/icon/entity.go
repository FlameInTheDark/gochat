package icon

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Icon interface {
	CreateIcon(ctx context.Context, id, guildId, ttlSeconds, fileSize int64) error
	DoneIcon(ctx context.Context, id, guildId int64, contentType, url *string, height, width, fileSize *int64) error
	RemoveIcon(ctx context.Context, id, guildId int64) error
	GetIcon(ctx context.Context, id, guildId int64) (model.Icon, error)
	GetIconsByGuildId(ctx context.Context, guildId int64) ([]model.Icon, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) Icon {
	return &Entity{c: c}
}
