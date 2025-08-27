package authentication

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type Authentication interface {
	CreateAuthentication(ctx context.Context, userId int64, email, passwordHash string) error
	RemoveAuthentication(ctx context.Context, userId int64) error
	GetAuthenticationByEmail(ctx context.Context, email string) (model.Authentication, error)
	GetAuthenticationByUserId(ctx context.Context, userId int64) (model.Authentication, error)
	SetPasswordHash(ctx context.Context, userId int64, hash string) error
	CreateRecovery(ctx context.Context, userId int64, email, token string) error
	RemoveRecovery(ctx context.Context, userId int64) error
	GetRecoveryByUserId(ctx context.Context, userId int64) (model.Recovery, error)
	GetRecovery(ctx context.Context, userId int64, token string) (model.Recovery, error)
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) Authentication {
	return &Entity{c: c}
}
