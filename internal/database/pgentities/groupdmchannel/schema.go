package groupdmchannel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) JoinGroupDmChannelMany(ctx context.Context, channelId int64, users []int64) error {
	if len(users) == 0 {
		return nil
	}
	tx, err := e.c.Beginx()
	if err != nil {
		return err
	}

	for _, user := range users {
		q := squirrel.Insert("group_dm_channels").
			PlaceholderFormat(squirrel.Dollar).
			Columns("channel_id", "user_id").
			Values(channelId, user)
		raw, args, err := q.ToSql()
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("unable to create SQL query: %w", err)
		}
		_, err = tx.ExecContext(ctx, raw, args...)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("unable to add user to dm channel: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("unable to join group dm channel many: %w", err)
	}
	return nil
}

func (e *Entity) JoinGroupDmChannel(ctx context.Context, channelId, userId int64) error {
	q := squirrel.Insert("group_dm_channels").
		PlaceholderFormat(squirrel.Dollar).
		Columns("channel_id", "user_id").
		Values(channelId, userId)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("join group dm channel error: %w", err)
	}
	return nil
}

func (e *Entity) GetGroupDmChannel(ctx context.Context, channelId, userId int64) (model.GroupDMChannel, error) {
	var ch model.GroupDMChannel
	q := squirrel.Select("channel_id", "user_id").
		PlaceholderFormat(squirrel.Dollar).
		From("group_dm_channels").
		Where(
			squirrel.And{
				squirrel.Eq{"channel_id": channelId},
				squirrel.Eq{"user_id": userId},
			},
		).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return ch, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &ch, raw, args...)
	if err != nil {
		return model.GroupDMChannel{}, fmt.Errorf("get group dm channel error: %w", err)
	}
	return ch, nil
}

func (e *Entity) LeaveGroupDmChannel(ctx context.Context, channelId, userId int64) error {
	q := squirrel.Delete("group_dm_channels").
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
		return fmt.Errorf("leave group dm channel error: %w", err)
	}
	return nil
}

func (e *Entity) GetGroupDmParticipants(ctx context.Context, channelId int64) ([]model.GroupDMChannel, error) {
	var channels []model.GroupDMChannel
	q := squirrel.Select("channel_id", "user_id").
		PlaceholderFormat(squirrel.Dollar).
		From("group_dm_channels").
		Where(squirrel.Eq{"channel_id": channelId})
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &channels, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return channels, nil
	} else if err != nil {
		return nil, fmt.Errorf("get group dm participants error: %w", err)
	}
	return channels, nil
}

func (e *Entity) IsGroupDmParticipant(ctx context.Context, channelId int64, userId int64) (bool, error) {
	var count int64
	q := squirrel.Select("count(*)").
		PlaceholderFormat(squirrel.Dollar).
		From("group_dm_channels").
		Where(
			squirrel.And{
				squirrel.Eq{"channel_id": channelId},
				squirrel.Eq{"user_id": userId},
			},
		)
	raw, args, err := q.ToSql()
	if err != nil {
		return false, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &count, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("check group dm participant error: %w", err)
	}
	return count > 0, nil
}

func (e *Entity) GetUserGroupDmChannels(ctx context.Context, userId int64) ([]model.GroupDMChannel, error) {
	var items []model.GroupDMChannel
	q := squirrel.Select("channel_id", "user_id").
		PlaceholderFormat(squirrel.Dollar).
		From("group_dm_channels").
		Where(squirrel.Eq{"user_id": userId})
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &items, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return []model.GroupDMChannel{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("get group dm channels for user error: %w", err)
	}
	return items, nil
}
