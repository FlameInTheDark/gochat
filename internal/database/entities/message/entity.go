package message

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Message interface {
	CreateMessage(ctx context.Context, id, channelID, userID int64, content string, attachments []int64, embedsJSON, autoEmbedsJSON string, position int64) error
	CreateMessageWithMeta(ctx context.Context, id, channelID, userID int64, content string, attachments []int64, embedsJSON, autoEmbedsJSON string, flags int, msgType model.MessageType, referenceChannel, reference, thread, position int64) error
	CreateSystemMessage(ctx context.Context, id, channelId, userId int64, content string, msgType model.MessageType, position int64) error
	CreateThreadCreatedMessageRef(ctx context.Context, threadID, channelID, messageID int64) error
	ClaimThread(ctx context.Context, channelID, messageID, threadID int64) (bool, int64, error)
	DeleteThreadCreatedMessageRef(ctx context.Context, threadID int64) error
	ReleaseThreadClaim(ctx context.Context, channelID, messageID int64) error
	SetThread(ctx context.Context, id, channelID, threadID int64) error
	UpdateMessageContent(ctx context.Context, id, channelID int64, content string) error
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
	GetThreadCreatedMessageRef(ctx context.Context, threadID int64) (int64, int64, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) Message {
	return &Entity{c: c}
}
