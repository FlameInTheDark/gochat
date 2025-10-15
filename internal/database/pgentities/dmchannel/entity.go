package dmchannel

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type DmChannel interface {
	GetDmChannel(ctx context.Context, userId, participantId int64) (model.DMChannel, error)
	CreateDmChannel(ctx context.Context, userId, participantId, channelId int64) error
	IsDmChannelParticipant(ctx context.Context, channelId, userId int64) (bool, error)
	GetUserDmChannels(ctx context.Context, userId int64) ([]model.DMChannel, error)
	GetDmChannelByChannelId(ctx context.Context, channelId int64) ([]model.DMChannel, error)
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) *Entity {
	return &Entity{c: c}
}
