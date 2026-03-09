package message

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Message interface {
	CreateMessage(ctx context.Context, id, channelID, userID int64, content string, attachments []int64, embedsJSON, autoEmbedsJSON string) error
	CreateSystemMessage(ctx context.Context, id, channelId, userId int64, content string, msgType model.MessageType) error
	UpdateMessage(ctx context.Context, id, channelID int64, content, embedsJSON, autoEmbedsJSON string, flags int) error
	UpdateGeneratedEmbeds(ctx context.Context, id, channelID int64, autoEmbedsJSON string) error
	DeleteMessage(ctx context.Context, id, channelId int64) error
	DeleteChannelMessages(ctx context.Context, channelID, lastId int64) error
	GetMessage(ctx context.Context, id, channelId int64) (model.Message, error)
	GetMessagesBefore(ctx context.Context, channelId, msgId int64, limit int) ([]model.Message, []int64, error)
	GetMessagesAfter(ctx context.Context, channelId, msgId, lastChannelMessage int64, limit int) ([]model.Message, []int64, error)
	GetMessagesAround(ctx context.Context, channelId, msgId, lastChannelMessage int64, limit int) ([]model.Message, []int64, error)
	GetMessagesList(ctx context.Context, msgIds []int64) ([]model.Message, error)
	GetChannelMessagesByIDs(ctx context.Context, channelId int64, ids []int64) ([]model.Message, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) Message {
	return &Entity{c: c}
}
