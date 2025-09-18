package discriminator

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type Discriminator interface {
	CreateDiscriminator(ctx context.Context, userId int64, discriminator string) error
	GetDiscriminatorByUserId(ctx context.Context, userId int64) (model.Discriminator, error)
	GetUserIdByDiscriminator(ctx context.Context, discriminator string) (model.Discriminator, error)
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) Discriminator {
	return &Entity{c: c}
}
