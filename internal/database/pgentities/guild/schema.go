package guild

import (
	"context"
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) GetGuildById(ctx context.Context, id int64) (model.Guild, error) {
	var g model.Guild

	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("guilds").
		Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return model.Guild{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &g, sql, args...)
	if err != nil {
		return g, fmt.Errorf("unable to get guild by id: %w", err)
	}
	return g, nil
}

func (e *Entity) CreateGuild(ctx context.Context, id int64, name string, ownerId, permissions int64) error {
	q := squirrel.Insert("guilds").
		PlaceholderFormat(squirrel.Dollar).
		Columns("id", "name", "owner_id", "permissions", "created_at").
		Values(id, name, ownerId, permissions, time.Now())

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to create guild: %w", err)
	}
	return nil
}

func (e *Entity) DeleteGuild(ctx context.Context, id int64) error {
	q := squirrel.Delete("guilds").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to delete guild: %w", err)
	}
	return nil
}

func (e *Entity) SetGuildIcon(ctx context.Context, id, icon int64) error {
	q := squirrel.Update("guilds").
		PlaceholderFormat(squirrel.Dollar).
		Set("icon", icon).
		Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to set icon: %w", err)
	}
	return nil
}

func (e *Entity) SetGuildPublic(ctx context.Context, id int64, public bool) error {
	q := squirrel.Update("guilds").
		PlaceholderFormat(squirrel.Dollar).
		Set("public", public).
		Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to set public: %w", err)
	}
	return nil
}

func (e *Entity) ChangeGuildOwner(ctx context.Context, id, ownerId int64) error {
	q := squirrel.Update("guilds").
		PlaceholderFormat(squirrel.Dollar).
		Set("owner_id", ownerId).
		Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to change owner id: %w", err)
	}
	return nil
}

func (e *Entity) GetGuildsList(ctx context.Context, ids []int64) ([]model.Guild, error) {
	var gs []model.Guild
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("guilds").
		Where(squirrel.Eq{"id": ids})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}

	err = e.c.SelectContext(ctx, &gs, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to get guilds: %w", err)
	}
	return gs, nil
}

func (e *Entity) SetGuildPermissions(ctx context.Context, id int64, permissions int64) error {
	q := squirrel.Update("guilds").
		PlaceholderFormat(squirrel.Dollar).
		Set("permissions", permissions).
		Where(squirrel.Eq{"id": id})

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to set guild permissions: %w", err)
	}
	return nil
}

func (e *Entity) UpdateGuild(ctx context.Context, id int64, name *string, icon *int64, public *bool, permissions *int64) error {
	q := squirrel.Update("guilds").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id})
	if name != nil {
		q = q.Set("name", *name)
	}
	if icon != nil {
		if *icon == 0 {
			q = q.Set("icon", nil)
		} else {
			q = q.Set("icon", *icon)
		}
	}
	if public != nil {
		q = q.Set("public", *public)
	}
	if permissions != nil {
		q = q.Set("permissions", *permissions)
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to update guild: %w", err)
	}
	return nil
}

func (e *Entity) SetSystemMessagesChannel(ctx context.Context, id int64, channelId *int64) error {
	q := squirrel.Update("guilds").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id})
	if channelId != nil {
		q = q.Set("system_messages", *channelId)
	} else {
		q = q.Set("system_messages", nil)
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to set system messages channel: %w", err)
	}
	return nil
}
