package groupdmchannel

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type GroupDMChannel interface {
	JoinGroupDmChannelMany(ctx context.Context, channelId int64, users []int64) error
	JoinGroupDmChannel(ctx context.Context, channelId, userId int64) error
	GetGroupDmChannel(ctx context.Context, channelId, userId int64) (model.GroupDMChannel, error)
	LeaveGroupDmChannel(ctx context.Context, channelId, userId int64) error
	GetGroupDmParticipants(ctx context.Context, channelId int64) ([]model.GroupDMChannel, error)
	IsGroupDmParticipant(ctx context.Context, channelId int64, userId int64) (bool, error)
	GetUserGroupDmChannels(ctx context.Context, userId int64) ([]model.GroupDMChannel, error)
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) *Entity {
	return &Entity{c: c}
}
