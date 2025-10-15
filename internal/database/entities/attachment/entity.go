package attachment

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Attachment interface {
	CreateAttachment(ctx context.Context, id, channelId, fileSize int64, height, width int64, name, url, contentType string) error
	RemoveAttachment(ctx context.Context, id, channelId int64) error
	GetAttachment(ctx context.Context, id, channelId int64) (model.Attachment, error)
	DoneAttachment(ctx context.Context, id, channelId int64, contentType, url *string) error
	SelectAttachmentByIDs(ctx context.Context, ids []int64) ([]model.Attachment, error)
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) Attachment {
	return &Entity{c: c}
}
