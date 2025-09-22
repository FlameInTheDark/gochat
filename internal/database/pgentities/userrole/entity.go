package userrole

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type UserRole interface {
	GetUserRoles(ctx context.Context, guildID, userId int64) ([]model.UserRole, error)
	AddUserRole(ctx context.Context, guildID, userId, roleId int64) error
	RemoveUserRole(ctx context.Context, guildID, userId, roleId int64) error
	RemoveRoleAssignments(ctx context.Context, guildID, roleId int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) UserRole {
	return &Entity{c: c}
}
