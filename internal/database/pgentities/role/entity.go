package role

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type Role interface {
	GetRoleByID(ctx context.Context, id int64) (model.Role, error)
	GetGuildRoles(ctx context.Context, guildId int64) ([]model.Role, error)
	GetRolesBulk(ctx context.Context, ids []int64) ([]model.Role, error)
	CreateRole(ctx context.Context, id, guildId int64, name string, color int, permissions int64) error
	RemoveRole(ctx context.Context, id int64) error
	SetRoleColor(ctx context.Context, id int64, color int) error
	SetRoleName(ctx context.Context, id int64, name string) error
	SetRolePermissions(ctx context.Context, id int64, permissions int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) Role {
	return &Entity{c: c}
}
