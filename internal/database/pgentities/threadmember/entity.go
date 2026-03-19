package threadmember

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type ThreadMember interface {
	AddThreadMember(ctx context.Context, threadID, userID int64) (model.ThreadMember, error)
	RemoveThreadMember(ctx context.Context, threadID, userID int64) error
	RemoveThreadMembers(ctx context.Context, threadID int64) error
	GetThreadMember(ctx context.Context, threadID, userID int64) (model.ThreadMember, error)
	GetThreadMembers(ctx context.Context, threadID int64) ([]model.ThreadMember, error)
	GetThreadMembersBulk(ctx context.Context, threadIDs []int64) ([]model.ThreadMember, error)
	GetThreadMembersByUser(ctx context.Context, userID int64, threadIDs []int64) ([]model.ThreadMember, error)
	GetUserThreadMembers(ctx context.Context, userID int64) ([]model.ThreadMember, error)
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) ThreadMember {
	return &Entity{c: c}
}
