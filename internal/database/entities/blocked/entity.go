package blocked

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
)

type Blocked interface {
	BlockUser(ctx context.Context, guildID, userID string) error
	UnblockUser(ctx context.Context, guildID, userID string) error
	IsBlocked(ctx context.Context, guildID, userID string) (bool, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) *Entity {
	return &Entity{c: c}
}
