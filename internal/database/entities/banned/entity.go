package banned

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Banned interface {
	BanUser(ctx context.Context, guildID, userID int64, reason *string) error
	UnbanUser(ctx context.Context, guildID, userID int64) error
	IsBanned(ctx context.Context, guildID, userID int64) (bool, error)
	GetGuildBans(ctx context.Context, guildID int64) ([]model.GuildBan, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) *Entity {
	return &Entity{c: c}
}
