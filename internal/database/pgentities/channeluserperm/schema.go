package channeluserperm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) GetUserChannelPermission(ctx context.Context, channelId, userId int64) (model.ChannelUserPermission, error) {
	var perm model.ChannelUserPermission
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("channel_user_permissions").
		Where(
			squirrel.And{
				squirrel.Eq{"channel_id": channelId},
				squirrel.Eq{"user_id": userId},
			},
		).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return model.ChannelUserPermission{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &perm, raw, args...)
	if err != nil {
		return perm, fmt.Errorf("unable to get user channel permission: %w", err)
	}
	return perm, nil
}

func (e *Entity) GetUserChannelPermissions(ctx context.Context, channelId int64) ([]model.ChannelUserPermission, error) {
	var perm []model.ChannelUserPermission
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("channel_user_permissions").
		Where(squirrel.Eq{"channel_id": channelId})
	raw, args, err := q.ToSql()
	if err != nil {
		return perm, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &perm, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return perm, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get user channel permissions: %w", err)
	}
	return perm, nil
}

func (e *Entity) CreateUserChannelPermission(ctx context.Context, channelId, userId int64, accept, deny int64) error {
	q := squirrel.Insert("channel_user_permissions").
		PlaceholderFormat(squirrel.Dollar).
		Columns("channel_id", "user_id", "accept", "deny").
		Values(channelId, userId, accept, deny)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to create user channel permission: %w", err)
	}
	return nil
}

func (e *Entity) UpdateChannelUserPermission(ctx context.Context, channelId, userId int64, accept, deny int64) error {
	q := squirrel.Update("channel_user_permissions").
		PlaceholderFormat(squirrel.Dollar).
		Where(
			squirrel.And{
				squirrel.Eq{"channel_id": channelId},
				squirrel.Eq{"user_id": userId},
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
		return fmt.Errorf("unable to update user channel permission: %w", err)
	}
	return nil
}

func (e *Entity) RemoveUserChannelPermission(ctx context.Context, channelId, userId int64) error {
	q := squirrel.Delete("channel_user_permissions").
		PlaceholderFormat(squirrel.Dollar).
		Where(
			squirrel.And{
				squirrel.Eq{"channel_id": channelId},
				squirrel.Eq{"user_id": userId},
			},
		)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to remove user channel permission: %w", err)
	}
	return nil
}
