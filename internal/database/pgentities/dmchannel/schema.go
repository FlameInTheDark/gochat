package dmchannel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) GetDmChannel(ctx context.Context, userId, participantId int64) (model.DMChannel, error) {
	var ch model.DMChannel
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("dm_channels").
		Where(
			squirrel.And{
				squirrel.Eq{"user_id": userId},
				squirrel.Eq{"participant_id": participantId},
			},
		)
	raw, args, err := q.ToSql()
	if err != nil {
		return ch, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &ch, raw, args...)
	if err != nil {
		return model.DMChannel{}, fmt.Errorf("unable to get dm channel: %w", err)
	}
	return ch, nil
}

func (e *Entity) CreateDmChannel(ctx context.Context, userId, participantId, channelId int64) error {
	tx, err := e.c.Beginx()
	if err != nil {
		return fmt.Errorf("unable to start transaction: %w", err)
	}

	// Batch both symmetric rows into a single multi-row INSERT
	q := squirrel.Insert("dm_channels").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "participant_id", "channel_id").
		Values(userId, participantId, channelId).
		Values(participantId, userId, channelId).
		Suffix("ON CONFLICT (channel_id, user_id) DO NOTHING")
	raw, args, err := q.ToSql()
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = tx.ExecContext(ctx, raw, args...)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("unable to create dm channel: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("unable to commit dm channel: %w", err)
	}
	return nil
}

func (e *Entity) IsDmChannelParticipant(ctx context.Context, channelId, userId int64) (bool, error) {
	var exists bool
	raw := "SELECT EXISTS(SELECT 1 FROM dm_channels WHERE channel_id = $1 AND user_id = $2)"
	err := e.c.GetContext(ctx, &exists, raw, channelId, userId)
	if err != nil {
		return false, fmt.Errorf("check dm participant error: %w", err)
	}
	return exists, nil
}

func (e *Entity) GetUserDmChannels(ctx context.Context, userId int64) ([]model.DMChannel, error) {
	var chs []model.DMChannel
	q := squirrel.Select("user_id", "participant_id", "channel_id").
		PlaceholderFormat(squirrel.Dollar).
		From("dm_channels").
		Where(squirrel.Eq{"user_id": userId})
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &chs, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return []model.DMChannel{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get user dm channels: %w", err)
	}
	return chs, nil
}

func (e *Entity) GetDmChannelByChannelId(ctx context.Context, channelId int64) ([]model.DMChannel, error) {
	var chs []model.DMChannel
	q := squirrel.Select("user_id", "participant_id", "channel_id").
		PlaceholderFormat(squirrel.Dollar).
		From("dm_channels").
		Where(squirrel.Eq{"channel_id": channelId})
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.SelectContext(ctx, &chs, raw, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.DMChannel{}, nil
		}
		return nil, fmt.Errorf("unable to get dm channel by channel id: %w", err)
	}
	return chs, nil
}
