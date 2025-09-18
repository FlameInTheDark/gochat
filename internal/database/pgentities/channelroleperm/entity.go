package channelroleperm

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type ChannelRolePerm interface {
	GetChannelRolePermission(ctx context.Context, channelId, roleId int64) (model.ChannelRolesPermission, error)
	GetChannelRolePermissions(ctx context.Context, channelId int64) ([]model.ChannelRolesPermission, error)
	SetChannelRolePermission(ctx context.Context, channelId, roleId, accept, deny int64) error
	UpdateChannelRolePermission(ctx context.Context, channelId, roleId, accept, deny int64) error
	RemoveChannelRolePermission(ctx context.Context, channelId, roleId int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) ChannelRolePerm {
	return &Entity{c: c}
}
