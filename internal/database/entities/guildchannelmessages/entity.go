package guildchannelmessages

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
)

type GuildChannelMessages interface {
	GetChannelsMessages(ctx context.Context, guildId int64) (map[int64]int64, error)
	GetChannelMessage(ctx context.Context, guildId, channelId int64) (int64, error)
	SetChannelLastMessage(ctx context.Context, guildId, channelId, lastMessageId int64) error
	SetReadStateMany(ctx context.Context, guildId, values map[int64]int64) error
	GetChannelsMessagesForGuilds(ctx context.Context, guildIDs []int64) (map[int64]map[int64]int64, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) *Entity {
	return &Entity{c: c}
}
