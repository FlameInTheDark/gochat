package member

import (
	"context"
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) AddMember(ctx context.Context, userID, guildID int64) error {
	q := squirrel.Insert("members").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "guild_id").
		Values(userID, guildID)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to add member: %w", err)
	}
	return nil
}

func (e *Entity) RemoveMember(ctx context.Context, userID, guildID int64) error {
	q := squirrel.Delete("members").
		PlaceholderFormat(squirrel.Dollar).
		Where(
			squirrel.And{
				squirrel.Eq{"user_id": userID},
				squirrel.Eq{"guild_id": guildID},
			},
		)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to remove member: %w", err)
	}
	return nil
}

func (e *Entity) GetMember(ctx context.Context, userId, guildId int64) (model.Member, error) {
	var m model.Member
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("members").
		Where(squirrel.And{squirrel.Eq{"user_id": userId}, squirrel.Eq{"guild_id": guildId}})

	sql, args, err := q.ToSql()
	if err != nil {
		return m, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &m, sql, args...)
	if err != nil {
		return m, fmt.Errorf("unable to get member: %w", err)
	}
	return m, nil
}

func (e *Entity) GetMembersList(ctx context.Context, guildId int64, ids []int64) ([]model.Member, error) {
	var members []model.Member
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("members").
		Where(
			squirrel.And{
				squirrel.Eq{"guild_id": guildId},
				squirrel.Eq{"user_id": ids},
			},
		)

	sql, args, err := q.ToSql()
	if err != nil {
		return members, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &members, sql, args...)
	if err != nil {
		return members, fmt.Errorf("unable to get members: %w", err)
	}
	return members, nil
}

func (e *Entity) GetGuildMembers(ctx context.Context, guildId int64) ([]model.Member, error) {
	var members []model.Member
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("members").
		Where(squirrel.Eq{"guild_id": guildId})

	sql, args, err := q.ToSql()
	if err != nil {
		return members, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &members, sql, args...)
	if err != nil {
		return members, fmt.Errorf("unable to get guild members: %w", err)
	}
	return members, nil
}

func (e *Entity) IsGuildMember(ctx context.Context, guildId, userId int64) (bool, error) {
	var count int
	q := squirrel.Select("count(*)").
		PlaceholderFormat(squirrel.Dollar).
		From("members").
		Where(
			squirrel.And{
				squirrel.Eq{"user_id": userId},
				squirrel.Eq{"guild_id": guildId},
			},
		)

	sql, args, err := q.ToSql()
	if err != nil {
		return false, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &count, sql, args...)
	if err != nil {
		return false, fmt.Errorf("unable to check if guild member exists: %w", err)
	}
	return count > 0, nil
}

func (e *Entity) GetUserGuilds(ctx context.Context, userId int64) ([]model.UserGuild, error) {
	var guilds []model.UserGuild
	q := squirrel.Select("user_id", "guild_id").
		PlaceholderFormat(squirrel.Dollar).
		From("members").
		Where(squirrel.Eq{"user_id": userId})

	sql, args, err := q.ToSql()
	if err != nil {
		return guilds, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &guilds, sql, args...)
	if err != nil {
		return guilds, fmt.Errorf("unable to get user guilds: %w", err)
	}
	return guilds, nil
}

func (e *Entity) SetTimeout(ctx context.Context, userId, guildId int64, timeout *time.Time) error {
	q := squirrel.Update("members").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.And{squirrel.Eq{"user_id": userId}, squirrel.Eq{"guild_id": guildId}}).
		Set("timeout", timeout)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to set timeout: %w", err)
	}
	return nil
}

func (e *Entity) CountGuildMembers(ctx context.Context, guildId int64) (int64, error) {
	var count int64
	q := squirrel.Select("count(*)").
		PlaceholderFormat(squirrel.Dollar).
		From("members").
		Where(squirrel.Eq{"guild_id": guildId})

	sql, args, err := q.ToSql()
	if err != nil {
		return 0, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err = e.c.GetContext(ctx, &count, sql, args...); err != nil {
		return 0, fmt.Errorf("unable to count guild members: %w", err)
	}
	return count, nil
}
