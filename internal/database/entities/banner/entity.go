package banner

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Banner interface {
	CreateBanner(ctx context.Context, id, userId, ttlSeconds, fileSize int64) error
	GetBanner(ctx context.Context, id, userId int64) (model.Banner, error)
	DoneBanner(ctx context.Context, id, userId int64, contentType, url *string, height, width, fileSize *int64) error
	RemoveBanner(ctx context.Context, id, userId int64) error
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) Banner {
	return &Entity{c: c}
}
