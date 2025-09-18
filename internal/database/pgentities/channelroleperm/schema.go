package channelroleperm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) GetChannelRolePermission(ctx context.Context, channelId, roleId int64) (model.ChannelRolesPermission, error) {
	var r model.ChannelRolesPermission
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("channel_roles_permissions").
		Where(
			squirrel.And{
				squirrel.Eq{"channel_id": channelId},
				squirrel.Eq{"role_id": roleId},
			},
		).Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return r, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &r, raw, args...)
	if err != nil {
		return r, fmt.Errorf("unable to get channel role: %w", err)
	}
	return r, nil
}

func (e *Entity) GetChannelRolePermissions(ctx context.Context, channelId int64) ([]model.ChannelRolesPermission, error) {
	var roles []model.ChannelRolesPermission
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("channel_roles_permissions").
		Where(squirrel.Eq{"channel_id": channelId})
	raw, args, err := q.ToSql()
	if err != nil {
		return roles, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &roles, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return roles, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get channel role permissions: %w", err)
	}
	return roles, nil
}

func (e *Entity) SetChannelRolePermission(ctx context.Context, channelId, roleId, accept, deny int64) error {
	q := squirrel.Insert("channel_roles_permissions").
		PlaceholderFormat(squirrel.Dollar).
		Columns("channel_id", "role_id", "accept", "deny").
		Values(channelId, roleId, accept, deny)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to set channel role permission: %w", err)
	}
	return nil
}

func (e *Entity) UpdateChannelRolePermission(ctx context.Context, channelId, roleId, accept, deny int64) error {
	q := squirrel.Update("channel_roles_permissions").
		PlaceholderFormat(squirrel.Dollar).
		Where(
			squirrel.And{
				squirrel.Eq{"channel_id": channelId},
				squirrel.Eq{"role_id": roleId},
			},
		).
		Set("accept", accept).
		Set("deny", deny)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to update channel role permission: %w", err)
	}
	return nil
}

func (e *Entity) RemoveChannelRolePermission(ctx context.Context, channelId, roleId int64) error {
	q := squirrel.Delete("channel_roles_permissions").
		PlaceholderFormat(squirrel.Dollar).
		Where(
			squirrel.And{
				squirrel.Eq{"channel_id": channelId},
				squirrel.Eq{"role_id": roleId},
			},
		)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to remove channel role permission: %w", err)
	}
	return nil
}
