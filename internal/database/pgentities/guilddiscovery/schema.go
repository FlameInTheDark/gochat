package guilddiscovery

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) SetGuildDiscovery(ctx context.Context, guildID int64, public bool, description string, tags []string) error {
	tx, err := e.c.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin guild discovery transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	q := squirrel.Update("guilds").
		PlaceholderFormat(squirrel.Dollar).
		Set("public", public).
		Set("description", description).
		Where(squirrel.Eq{"id": guildID})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("create guild discovery update SQL: %w", err)
	}
	if _, err = tx.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("update guild discovery metadata: %w", err)
	}

	del := squirrel.Delete("guild_tags").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"guild_id": guildID})
	sql, args, err = del.ToSql()
	if err != nil {
		return fmt.Errorf("create guild tags delete SQL: %w", err)
	}
	if _, err = tx.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("delete guild tags: %w", err)
	}

	if len(tags) > 0 {
		ins := squirrel.Insert("guild_tags").
			PlaceholderFormat(squirrel.Dollar).
			Columns("guild_id", "tag")
		for _, tag := range tags {
			ins = ins.Values(guildID, tag)
		}
		sql, args, err = ins.ToSql()
		if err != nil {
			return fmt.Errorf("create guild tags insert SQL: %w", err)
		}
		if _, err = tx.ExecContext(ctx, sql, args...); err != nil {
			return fmt.Errorf("insert guild tags: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit guild discovery transaction: %w", err)
	}
	return nil
}

func (e *Entity) GetTagsByGuilds(ctx context.Context, guildIDs []int64) (map[int64][]string, error) {
	result := make(map[int64][]string, len(guildIDs))
	if len(guildIDs) == 0 {
		return result, nil
	}

	type row struct {
		GuildID int64  `db:"guild_id"`
		Tag     string `db:"tag"`
	}
	var rows []row
	q := squirrel.Select("guild_id", "tag").
		PlaceholderFormat(squirrel.Dollar).
		From("guild_tags").
		Where(squirrel.Eq{"guild_id": guildIDs}).
		OrderBy("tag ASC")

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("create guild tags select SQL: %w", err)
	}
	if err = e.c.SelectContext(ctx, &rows, sql, args...); err != nil {
		return nil, fmt.Errorf("select guild tags: %w", err)
	}
	for _, r := range rows {
		result[r.GuildID] = append(result[r.GuildID], r.Tag)
	}
	return result, nil
}

func (e *Entity) GetStatsByGuilds(ctx context.Context, guildIDs []int64) (map[int64]model.GuildDiscoveryStats, error) {
	result := make(map[int64]model.GuildDiscoveryStats, len(guildIDs))
	if len(guildIDs) == 0 {
		return result, nil
	}

	var stats []model.GuildDiscoveryStats
	q := squirrel.Select("guild_id", "members_count").
		PlaceholderFormat(squirrel.Dollar).
		From("guild_discovery_stats").
		Where(squirrel.Eq{"guild_id": guildIDs})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("create guild stats select SQL: %w", err)
	}
	if err = e.c.SelectContext(ctx, &stats, sql, args...); err != nil {
		return nil, fmt.Errorf("select guild stats: %w", err)
	}
	for _, stat := range stats {
		result[stat.GuildId] = stat
	}
	return result, nil
}

func (e *Entity) AdjustMembersCount(ctx context.Context, guildID, delta int64) error {
	_, err := e.c.ExecContext(ctx, `
INSERT INTO guild_discovery_stats (guild_id, members_count)
VALUES ($1, GREATEST($2, 0))
ON CONFLICT (guild_id) DO UPDATE
SET members_count = GREATEST(guild_discovery_stats.members_count + $2, 0)
`, guildID, delta)
	if err != nil {
		return fmt.Errorf("adjust guild discovery member count: %w", err)
	}
	return nil
}

func (e *Entity) RecountMembers(ctx context.Context, guildID int64) error {
	_, err := e.c.ExecContext(ctx, `
INSERT INTO guild_discovery_stats (guild_id, members_count)
SELECT $1, COUNT(*) FROM members WHERE guild_id = $1
ON CONFLICT (guild_id) DO UPDATE SET members_count = EXCLUDED.members_count
`, guildID)
	if err != nil {
		return fmt.Errorf("recount guild discovery member count: %w", err)
	}
	return nil
}
