package attachment

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Attachment interface {
	CreateAttachment(ctx context.Context, id, channelId, authorId, ttlSeconds, fileSize int64, name string) error
	RemoveAttachment(ctx context.Context, id, channelId int64) error
	GetAttachment(ctx context.Context, id, channelId int64) (model.Attachment, error)
	DoneAttachment(ctx context.Context, id, channelId int64, contentType, url, previewURL *string, height, width, fileSize *int64, name *string, authorId *int64) error
	SelectAttachmentsByChannel(ctx context.Context, channelId int64, ids []int64) ([]model.Attachment, error)
	UpdateFileSize(ctx context.Context, id, channelId int64, fileSize int64) error
	ListDoneZeroSize(ctx context.Context) ([]model.Attachment, error)
	UpdateName(ctx context.Context, id, channelId int64, name string) error
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) Attachment {
	return &Entity{c: c}
}
