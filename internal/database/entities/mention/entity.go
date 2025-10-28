package mention

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Mention interface {
	AddMention(ctx context.Context, userId, channelId, messageId, authorId int64) error
	GetMentionsBefore(ctx context.Context, userId, channelId, messageId int64) ([]model.Mention, error)
	GetMentionsAfter(ctx context.Context, userId, channelId, messageId int64) ([]model.Mention, error)
	GetMentionsAfterMany(ctx context.Context, userId, messageId int64, channelIds []int64) (map[int64][]model.Mention, error)
	AddChannelMention(ctx context.Context, guildId, channelId, messageId, authorId int64, roleId *int64, mtype model.ChannelMentionType) error
	GetChannelMentionsBefore(ctx context.Context, channelId, messageId int64) ([]model.ChannelMention, error)
	GetChannelMentionsAfter(ctx context.Context, channelId, messageId int64) ([]model.ChannelMention, error)
	GetChannelMentionsAfterMany(ctx context.Context, messageId int64, channelIds []int64) (map[int64][]model.ChannelMention, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) Mention {
	return &Entity{c: c}
}
