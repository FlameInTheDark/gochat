package channel

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type Channel interface {
	GetChannel(ctx context.Context, id int64) (model.Channel, error)
	GetChannelsBulk(ctx context.Context, ids []int64) ([]model.Channel, error)
	GetChannelThreads(ctx context.Context, channelId int64) ([]model.Channel, error)
	CreateChannel(ctx context.Context, id int64, name string, channelType model.ChannelType, parent *int64, permissions *int64, private bool) error
	DeleteChannel(ctx context.Context, id int64) error
	RenameChannel(ctx context.Context, id int64, newName string) error
	SetChannelPermissions(ctx context.Context, id int64, permissions int) error
	SetChannelPrivate(ctx context.Context, id int64, private bool) error
	SetChannelTopic(ctx context.Context, id int64, topic *string) error
	SetChannelParent(ctx context.Context, id int64, parent *int64) error
	SetChannelParentBulk(ctx context.Context, id []int64, parent *int64) error
	SetLastMessage(ctx context.Context, id, lastMessage int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) Channel {
	return &Entity{c: c}
}
