package userrole

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocql/gocql"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	getUserRoles   = `SELECT guild_id, user_id, role_id FROM gochat.user_roles WHERE guild_id = ? AND user_id = ?`
	addUserRole    = `INSERT INTO gochat.user_roles (guild_id, user_id, role_id) VALUES(?, ?, ?)`
	removeUserRole = `DELETE FROM gochat.user_roles WHERE guild_id = ? AND user_id = ? AND role_id = ?`
)

func (e *Entity) GetUserRoles(ctx context.Context, guildID, userId int64) ([]model.UserRole, error) {
	var roles []model.UserRole
	iter := e.c.Session().
		Query(getUserRoles).
		WithContext(ctx).
		Bind(guildID, userId).
		Iter()
	var r model.UserRole
	for iter.Scan(&r.GuildId, &r.UserId, &r.RoleId) {
		roles = append(roles, r)
	}
	err := iter.Close()
	if errors.Is(err, gocql.ErrNotFound) {
		return roles, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get user roles: %w", err)
	}
	return roles, nil
}

func (e *Entity) AddUserRole(ctx context.Context, guildID, userId, roleId int64) error {
	err := e.c.Session().
		Query(addUserRole).
		WithContext(ctx).
		Bind(guildID, userId, roleId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to add user role: %w", err)
	}
	return nil
}

func (e *Entity) RemoveUserRole(ctx context.Context, guildID, userId, roleId int64) error {
	err := e.c.Session().
		Query(removeUserRole).
		WithContext(ctx).
		Bind(guildID, userId, roleId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove user role: %w", err)
	}
	return nil
}
