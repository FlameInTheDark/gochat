package readstates

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/db"
)

type ReadStates interface {
	GetReadStates(ctx context.Context, userId int64) (map[int64]int64, error)
	GetReadState(ctx context.Context, userId, channelId int64) (int64, error)
	SetReadState(ctx context.Context, userId, channelId, lastMessageId int64) error
	SetReadStateMany(ctx context.Context, userId, values map[int64]int64) error
}

type Entity struct {
	c *db.CQLCon
}

func New(c *db.CQLCon) *Entity {
	return &Entity{c: c}
}
