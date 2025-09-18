package channeluserperm

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type ChannelUserPerm interface {
	GetUserChannelPermission(ctx context.Context, channelId, userId int64) (model.ChannelUserPermission, error)
	GetUserChannelPermissions(ctx context.Context, channelId int64) ([]model.ChannelUserPermission, error)
	CreateUserChannelPermission(ctx context.Context, channelId, userId int64, accept, deny int64) error
	UpdateChannelUserPermission(ctx context.Context, channelId, userId int64, accept, deny int64) error
	RemoveUserChannelPermission(ctx context.Context, channelId, userId int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) ChannelUserPerm {
	return &Entity{c: c}
}
