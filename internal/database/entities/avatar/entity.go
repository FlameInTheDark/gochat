package avatar

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Avatar interface {
	CreateAvatar(ctx context.Context, id, userId, ttlSeconds, fileSize int64) error
	GetAvatar(ctx context.Context, id, userId int64) (model.Avatar, error)
	DoneAvatar(ctx context.Context, id, userId int64, contentType, url *string, height, width, fileSize *int64) error
	RemoveAvatar(ctx context.Context, id, userId int64) error
	GetAvatarsByUserId(ctx context.Context, userId int64) ([]model.Avatar, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) Avatar {
	return &Entity{c: c}
}
