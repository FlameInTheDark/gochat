package channeluserperm

import (
	"context"
	"fmt"
	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	getChannelUserPermission    = `SELECT channel_id, user_id, permissions FROM gochat.channel_user_permissions WHERE channel_id = ? AND user_id = ?`
	getChannelUserPermissions   = `SELECT channel_id, user_id, permissions FROM gochat.channel_user_permissions WHERE channel_id = ?`
	setChannelUserPermission    = `INSERT INTO gochat.channel_user_permissions (channel_id, user_id, permissions) VALUES (?, ?, ?)`
	removeChannelUserPermission = `DELETE FROM gochat.channel_user_permissions WHERE channel_id = ? AND user_id = ?`
)

func (e *Entity) GetUserChannelPermission(ctx context.Context, channelId, userId int64, permissions int) (model.ChannelUserPermission, error) {
	var perm model.ChannelUserPermission
	err := e.c.Session().
		Query(getChannelUserPermission).
		WithContext(ctx).
		Bind(channelId, userId).
		Scan(&perm.ChannelId, &perm.UserId, &perm.Permissions)
	if err != nil {
		return perm, fmt.Errorf("unable to get user channel permission: %w", err)
	}
	return perm, nil
}

func (e *Entity) GetUserChannelPermissions(ctx context.Context, channelId, userId int64) ([]model.ChannelUserPermission, error) {
	var perm []model.ChannelUserPermission
	iter := e.c.Session().
		Query(getChannelUserPermissions).
		WithContext(ctx).
		Bind(channelId, userId).
		Iter()
	var p model.ChannelUserPermission
	for iter.Scan(&p.ChannelId, &p.UserId, &p.Permissions) {
		perm = append(perm, p)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get user channel permissions: %w", err)
	}
	return perm, nil
}

func (e *Entity) CreateUserChannelPermission(ctx context.Context, channelId, userId int64, permissions int) error {
	err := e.c.Session().
		Query(setChannelUserPermission).
		WithContext(ctx).
		Bind(channelId, userId, permissions).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create user channel permission: %w", err)
	}
	return nil
}

func (e *Entity) RemoveUserChannelPermission(ctx context.Context, channelId, userId int64) error {
	err := e.c.Session().
		Query(removeChannelUserPermission).
		WithContext(ctx).
		Bind(channelId, userId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove user channel permission: %w", err)
	}
	return nil
}
