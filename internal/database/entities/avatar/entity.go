package avatar

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Avatar interface {
	CreateAvatar(ctx context.Context, id, userId int64, object string) error
	RemoveAvatar(ctx context.Context, id int64) error
	GetAvatar(ctx context.Context, id int64) (model.Avatar, error)
	GetAvatarsByUserId(ctx context.Context, userId int64) ([]model.Avatar, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) *Entity {
	return &Entity{c: c}
}
