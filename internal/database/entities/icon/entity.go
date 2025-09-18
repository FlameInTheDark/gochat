package icon

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Icon interface {
	CreateIcon(ctx context.Context, id, guildId int64, object string) error
	RemoveIcon(ctx context.Context, id int64) error
	GetIcon(ctx context.Context, id int64) (model.Icon, error)
	GetIconsByUserId(ctx context.Context, userId int64) ([]model.Icon, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) Icon {
	return &Entity{c: c}
}
