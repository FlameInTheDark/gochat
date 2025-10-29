package guildchannels

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type GuildChannels interface {
	AddChannel(ctx context.Context, guildID, channelID int64, channelName string, channelType model.ChannelType, parentID *int64, private bool, position int) error
	GetGuildChannel(ctx context.Context, guildID, channelID int64) (model.GuildChannel, error)
	GetGuildChannels(ctx context.Context, guildID int64) ([]model.GuildChannel, error)
	GetGuildByChannel(ctx context.Context, channelID int64) (model.GuildChannel, error)
	RemoveChannel(ctx context.Context, guildID, channelID int64) error
	SetGuildChannelPosition(ctx context.Context, updates []model.GuildChannelUpdatePosition) error
	ResetGuildChannelPositionBulk(ctx context.Context, chs []int64, guildId int64) error
	GetGuildsChannelsIDsMany(ctx context.Context, guilds []int64) ([]int64, error)
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) GuildChannels {
	return &Entity{c: c}
}
