package message

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Message interface {
	CreateMessage(ctx context.Context, id, channel_id, user_id int64, content string, attachments []int64) error
	UpdateMessage(ctx context.Context, id, channel_id int64, content string) error
	DeleteMessage(ctx context.Context, id, channelId int64) error
	DeleteChannelMessages(ctx context.Context, channel_id, lastId int64) error
	GetMessage(ctx context.Context, id, channelId int64) (model.Message, error)
	GetMessagesBefore(ctx context.Context, channelId, msgId int64, limit int) ([]model.Message, []int64, error)
	GetMessagesAfter(ctx context.Context, channelId, msgId, lastChannelMessage int64, limit int) ([]model.Message, []int64, error)
	GetMessagesList(ctx context.Context, msgIds []int64) ([]model.Message, error)
	GetChannelMessagesByIDs(ctx context.Context, channelId int64, ids []int64) ([]model.Message, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) Message {
	return &Entity{c: c}
}
