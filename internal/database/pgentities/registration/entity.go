package registration

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type Registration interface {
	GetRegistrationByUserId(ctx context.Context, userId int64) (model.Registration, error)
	GetRegistrationByEmail(ctx context.Context, email string) (model.Registration, error)
	CreateRegistration(ctx context.Context, userId int64, email string, confirmation string) error
	RemoveRegistration(ctx context.Context, userId int64) error
	SetRegistrationToken(ctx context.Context, userId int64, token string) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) *Entity {
	return &Entity{c: c}
}
