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
	qUser := squirrel.Insert("dm_channels").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "participant_id", "channel_id").
		Values(userId, participantId, channelId)
	rawUser, argsUser, err := qUser.ToSql()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = tx.ExecContext(ctx, rawUser, argsUser)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to create SQL query: %w", err)
	}

	qPart := squirrel.Insert("dm_channels").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "participant_id", "channel_id").
		Values(participantId, userId, channelId)
	rawPart, argsPart, err := qPart.ToSql()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = tx.ExecContext(ctx, rawPart, argsPart)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("unable to create SQL query: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("unable to create dm channel: %w", err)
	}
	return nil
}

func (e *Entity) IsDmChannelParticipant(ctx context.Context, channelId, userId int64) (bool, error) {
	var count int64
	q := squirrel.Select("count(*)").
		PlaceholderFormat(squirrel.Dollar).
		From("dm_channels").
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
		return false, fmt.Errorf("check dm participant error: %w", err)
	}
	return count > 0, nil
}
