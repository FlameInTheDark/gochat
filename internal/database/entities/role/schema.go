package role

import (
	"context"
	"fmt"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	getRoleByID        = `SELECT id, guild_id, name, color, permissions FROM gochat.roles WHERE id = ?`
	getGuildRoles      = `SELECT id, guild_id, name, color, permissions FROM gochat.roles WHERE guild_id = ?`
	getRolesBulk       = `SELECT id, guild_id, name, color, permissions FROM gochat.roles WHERE guild_id IN ?`
	createRole         = `INSERT INTO gochat.roles (id, guild_id, name, color, permissions) VALUES (?, ?, ?, ?, ?)`
	removeRole         = `DELETE FROM gochat.roles WHERE id = ?`
	setRoleColor       = `UPDATE gochat.roles SET color = ? WHERE id = ?`
	setRoleName        = `UPDATE gochat.roles SET name = ? WHERE id = ?`
	setRolePermissions = `UPDATE gochat.roles SET permissions = ? WHERE id = ?`
)

func (e *Entity) GetRoleByID(ctx context.Context, id int64) (model.Role, error) {
	var r model.Role
	err := e.c.Session().
		Query(getRoleByID).
		WithContext(ctx).
		Bind(id).
		Scan(&r.Id, &r.GuildId, &r.Name, &r.Color, &r.Permissions)
	if err != nil {
		return r, fmt.Errorf("unable to get role by id: %w", err)
	}
	return r, nil
}

func (e *Entity) GetGuildRoles(ctx context.Context, guildId int64) ([]model.Role, error) {
	var roles []model.Role
	iter := e.c.Session().
		Query(getGuildRoles).
		WithContext(ctx).
		Bind(guildId).
		Iter()
	var r model.Role
	for iter.Scan(&r.Id, &r.GuildId, &r.Name, &r.Color, &r.Permissions) {
		roles = append(roles, r)
	}
	err := iter.Close()
	if err != nil {
		return roles, fmt.Errorf("unable to get roles for guild %d: %w", guildId, err)
	}
	return roles, nil
}

func (e *Entity) GetRolesBulk(ctx context.Context, ids []int64) ([]model.Role, error) {
	var roles []model.Role
	iter := e.c.Session().
		Query(getRolesBulk).
		WithContext(ctx).
		Bind(ids).
		Iter()
	var r model.Role
	for iter.Scan(&r.Id, &r.GuildId, &r.Name, &r.Color, &r.Permissions) {
		roles = append(roles, r)
	}
	err := iter.Close()
	if err != nil {
		return roles, fmt.Errorf("unable to get roles for guild %d: %w", ids, err)
	}
	return roles, nil
}

func (e *Entity) CreateRole(ctx context.Context, id, guildId int64, name string, color int, permissions int64) error {
	err := e.c.Session().
		Query(createRole).
		WithContext(ctx).
		Bind(id, guildId, name, color, permissions).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create role for guild %d: %w", guildId, err)
	}
	return nil
}

func (e *Entity) RemoveRole(ctx context.Context, id int64) error {
	err := e.c.Session().
		Query(removeRole).
		WithContext(ctx).
		Bind(id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove role for guild %d: %w", id, err)
	}
	return nil
}

func (e *Entity) SetRoleColor(ctx context.Context, id int64, color int) error {
	err := e.c.Session().
		Query(setRoleColor).
		WithContext(ctx).
		Bind(id, color).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set role color for guild %d: %w", id, err)
	}
	return nil
}

func (e *Entity) SetRoleName(ctx context.Context, id int64, name string) error {
	err := e.c.Session().
		Query(setRoleName).
		WithContext(ctx).
		Bind(id, name).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set role name for guild %d: %w", id, err)
	}
	return nil
}

func (e *Entity) SetRolePermissions(ctx context.Context, id int64, permissions int64) error {
	err := e.c.Session().
		Query(setRolePermissions).
		WithContext(ctx).
		Bind(id, permissions).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set role permissions for guild %d: %w", id, err)
	}
	return nil
}
