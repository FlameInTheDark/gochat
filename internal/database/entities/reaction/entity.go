package reaction

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Reaction interface {
	GetReactions(ctx context.Context, messageId int64) ([]model.Reaction, error)
	GetReactionsAfter(ctx context.Context, messageId, userId int64) ([]model.Reaction, error)
	AddReaction(ctx context.Context, messageId, userId, emoteId int64) error
	RemoveReaction(ctx context.Context, messageId, userId int64) error
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) *Entity {
	return &Entity{c: c}
}
