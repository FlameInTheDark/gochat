package friend

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type Friend interface {
	AddFriend(ctx context.Context, userID, friendID int64) error
	RemoveFriend(ctx context.Context, userID, friendID int64) error
	GetFriends(ctx context.Context, userID int64) ([]model.Friend, error)
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) *Entity {
	return &Entity{c: c}
}
