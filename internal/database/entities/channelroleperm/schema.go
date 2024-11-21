package channelroleperm

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocql/gocql"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	getChannelRolePermission     = `SELECT channel_id, role_id, accept, deny FROM gochat.channel_role_permissions WHERE channel_id = ? AND role_id = ?`
	getChannelRolePermissions    = `SELECT channel_id, role_id, accept, deny FROM gochat.channel_role_permissions WHERE channel_id = ?`
	setChannelRolePermission     = `INSERT INTO gochat.channel_role_permissions (channel_id, role_id, accept, deny) VALUES (?, ?, ?, ?)`
	updateChannelRolePermissions = `UPDATE gochat.channel_role_permissions SET accept = ?, deny = ? WHERE channel_id = ? AND role_id = ?`
	removeChannelRolePermission  = `DELETE FROM gochat.channel_role_permissions WHERE channel_id = ? AND role_id = ?`
)

func (e *Entity) GetChannelRolePermission(ctx context.Context, channelId, roleId int64) (model.ChannelRolesPermission, error) {
	var r model.ChannelRolesPermission
	err := e.c.Session().
		Query(getChannelRolePermission).
		WithContext(ctx).
		Bind(channelId, roleId).
		Scan(&r.ChannelId, &r.RoleId, &r.Accept, &r.Deny)
	if err != nil {
		return r, fmt.Errorf("unable to get channel role: %w", err)
	}
	return r, nil
}

func (e *Entity) GetChannelRolePermissions(ctx context.Context, channelId int64) ([]model.ChannelRolesPermission, error) {
	var roles []model.ChannelRolesPermission
	iter := e.c.Session().
		Query(getChannelRolePermissions).
		WithContext(ctx).
		Bind(channelId).
		Iter()
	var r model.ChannelRolesPermission
	for iter.Scan(&r.ChannelId, &r.RoleId, &r.Accept, &r.Deny) {
		roles = append(roles, r)
	}
	err := iter.Close()
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, fmt.Errorf("unable to get channel role permissions: %w", err)
	}
	return roles, nil
}

func (e *Entity) SetChannelRolePermission(ctx context.Context, channelId, roleId, accept, deny int64) error {
	err := e.c.Session().
		Query(setChannelRolePermission).
		WithContext(ctx).
		Bind(channelId, roleId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set channel role permission: %w", err)
	}
	return nil
}

func (e *Entity) UpdateChannelRolePermission(ctx context.Context, channelId, roleId, accept, deny int64) error {
	err := e.c.Session().
		Query(updateChannelRolePermissions).
		WithContext(ctx).
		Bind(accept, deny, channelId, roleId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to update channel role permission: %w", err)
	}
	return nil
}

func (e *Entity) RemoveChannelRolePermission(ctx context.Context, channelId, roleId int64) error {
	err := e.c.Session().
		Query(removeChannelRolePermission).
		WithContext(ctx).
		Bind(channelId, roleId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove channel role permission: %w", err)
	}
	return nil
}
