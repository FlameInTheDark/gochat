package guilddiscovery

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type GuildDiscovery interface {
	SetGuildDiscovery(ctx context.Context, guildID int64, public bool, description string, tags []string) error
	GetTagsByGuilds(ctx context.Context, guildIDs []int64) (map[int64][]string, error)
	GetStatsByGuilds(ctx context.Context, guildIDs []int64) (map[int64]model.GuildDiscoveryStats, error)
	AdjustMembersCount(ctx context.Context, guildID, delta int64) error
	RecountMembers(ctx context.Context, guildID int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) GuildDiscovery {
	return &Entity{c: c}
}
