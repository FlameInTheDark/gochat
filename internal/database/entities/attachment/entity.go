package attachment

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Attachment interface {
	// CreateAttachment creates a placeholder attachment row with TTL and done=false
	// Width/Height/URL/ContentType are filled on finalize; authorId is enforced during upload
	CreateAttachment(ctx context.Context, id, channelId, authorId, ttlSeconds, fileSize int64, name string) error
	RemoveAttachment(ctx context.Context, id, channelId int64) error
	GetAttachment(ctx context.Context, id, channelId int64) (model.Attachment, error)
	// DoneAttachment finalizes the attachment, clears TTL, and sets metadata and URLs
	DoneAttachment(ctx context.Context, id, channelId int64, contentType, url, previewURL *string, height, width, fileSize *int64, name *string, authorId *int64) error
	SelectAttachmentByIDs(ctx context.Context, ids []int64) ([]model.Attachment, error)
	// UpdateFileSize updates only the filesize column for an attachment
	UpdateFileSize(ctx context.Context, id, channelId int64, fileSize int64) error
	// ListDoneZeroSize returns attachments marked done with missing/zero filesize
	ListDoneZeroSize(ctx context.Context) ([]model.Attachment, error)
	// UpdateName updates attachment name
	UpdateName(ctx context.Context, id, channelId int64, name string) error
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) Attachment {
	return &Entity{c: c}
}
