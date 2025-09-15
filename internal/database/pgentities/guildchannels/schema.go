package guildchannels

import (
	"context"
	"fmt"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) AddChannel(ctx context.Context, guildID, channelID int64, position int) error {
	q := squirrel.Insert("guild_channels").
		PlaceholderFormat(squirrel.Dollar).
		Columns("guild_id", "channel_id", "position").
		Values(guildID, channelID, position)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to add channel: %w", err)
	}
	return nil
}

func (e *Entity) GetGuildChannel(ctx context.Context, guildID, channelID int64) (model.GuildChannel, error) {
	var ch model.GuildChannel
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("guild_channels").
		Where(
			squirrel.And{
				squirrel.Eq{"guild_id": guildID},
				squirrel.Eq{"channel_id": channelID},
			},
		)
	sql, args, err := q.ToSql()
	if err != nil {
		return model.GuildChannel{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &ch, sql, args...)
	if err != nil {
		return ch, fmt.Errorf("unable to get guild channel: %w", err)
	}
	return ch, nil
}

func (e *Entity) GetGuildChannels(ctx context.Context, guildID int64) ([]model.GuildChannel, error) {
	var chans []model.GuildChannel
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("guild_channels").
		Where(squirrel.Eq{"guild_id": guildID})
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &chans, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to get guild channels: %w", err)
	}
	return chans, nil
}

func (e *Entity) GetGuildByChannel(ctx context.Context, channelID int64) (model.GuildChannel, error) {
	var ch model.GuildChannel
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("guild_channels").
		Where(squirrel.Eq{"channel_id": channelID}).
		Limit(1)
	sql, args, err := q.ToSql()
	if err != nil {
		return model.GuildChannel{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &ch, sql, args...)
	if err != nil {
		return model.GuildChannel{}, fmt.Errorf("unable to get guild by channel: %w", err)
	}
	return ch, nil
}

func (e *Entity) RemoveChannel(ctx context.Context, guildID, channelID string) error {
	q := squirrel.Delete("guild_channels").
		Where(
			squirrel.And{
				squirrel.Eq{"guild_id": guildID},
				squirrel.Eq{"channel_id": channelID},
			},
		)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to remove channel: %w", err)
	}
	return nil
}

func (e *Entity) SetGuildChannelPosition(ctx context.Context, updates []model.GuildChannelUpdatePosition) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := e.c.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// Build a single UPDATE using a VALUES list to map channel_id -> position
	// This avoids any potential placeholder/type mismatches with IN clauses
	guildID := updates[0].GuildId
	values := make([]string, 0, len(updates))
	args := make([]interface{}, 0, 1+len(updates)*2)

	// $1 is reserved for guildID
	args = append(args, guildID)
	argIdx := 2
	for _, u := range updates {
		// Each pair adds two placeholders with explicit type casts
		// to ensure Postgres infers correct column types in VALUES
		values = append(values, fmt.Sprintf("($%d::bigint,$%d::int)", argIdx, argIdx+1))
		args = append(args, u.ChannelId, u.Position)
		argIdx += 2
	}

	query := fmt.Sprintf(
		"UPDATE guild_channels AS gc SET position = v.position FROM (VALUES %s) AS v(channel_id, position) WHERE gc.guild_id = $1 AND gc.channel_id = v.channel_id",
		strings.Join(values, ","),
	)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	return nil
}

func (e *Entity) ResetGuildChannelPositionBulk(ctx context.Context, chs []int64, guildId int64) error {
	if len(chs) == 0 {
		return nil
	}
	q := squirrel.Update("guild_channels").
		PlaceholderFormat(squirrel.Dollar).
		Where(
			squirrel.And{
				squirrel.Eq{"guild_id": guildId},
				squirrel.Eq{"channel_id": chs},
			},
		).
		Set("position", 0)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to reset guild channel position bulk: %w", err)
	}
	return nil
}
