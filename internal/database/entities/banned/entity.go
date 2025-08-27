package banned

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
)

type Banned interface {
	BanUser(ctx context.Context, guildID, userID string) error
	UnbanUser(ctx context.Context, guildID, userID string) error
	IsBanned(ctx context.Context, guildID, userID string) (bool, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) *Entity {
	return &Entity{c: c}
}
