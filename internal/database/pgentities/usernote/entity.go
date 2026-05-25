package usernote

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type UserNote interface {
	GetNote(ctx context.Context, ownerUserId, targetUserId int64) (model.UserNote, error)
	UpsertNote(ctx context.Context, ownerUserId, targetUserId int64, note string) error
	DeleteNote(ctx context.Context, ownerUserId, targetUserId int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) UserNote {
	return &Entity{c: c}
}
