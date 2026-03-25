package reaction

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Reaction interface {
	GetUserReaction(ctx context.Context, messageId, userId int64, bucketKey string) (model.Reaction, error)
	GetUserReactions(ctx context.Context, messageId, userId int64) ([]model.Reaction, error)
	ListBucketReactions(ctx context.Context, messageId int64, bucketKey string, after *int64, limit int) ([]model.Reaction, error)
	ListMessageSummaries(ctx context.Context, messageId int64) ([]model.ReactionSummary, error)
	UpsertReaction(ctx context.Context, reaction model.Reaction) error
	DeleteReaction(ctx context.Context, reaction model.Reaction) error
	SetSummary(ctx context.Context, summary model.ReactionSummary) error
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) *Entity {
	return &Entity{c: c}
}
