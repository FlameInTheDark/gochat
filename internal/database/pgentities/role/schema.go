package role

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) GetRoleByID(ctx context.Context, id int64) (model.Role, error) {
	var r model.Role
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("roles").
		Where(squirrel.Eq{"id": id}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return model.Role{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &r, raw, args...)
	if err != nil {
		return r, fmt.Errorf("unable to get role by id: %w", err)
	}
	return r, nil
}

func (e *Entity) GetGuildRoles(ctx context.Context, guildId int64) ([]model.Role, error) {
	var roles []model.Role
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("roles").
		Where(squirrel.Eq{"guild_id": guildId})
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &roles, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return roles, nil
	} else if err != nil {
		return roles, fmt.Errorf("unable to get roles for guild %d: %w", guildId, err)
	}
	return roles, nil
}

func (e *Entity) GetRolesBulk(ctx context.Context, ids []int64) ([]model.Role, error) {
	var roles []model.Role
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("roles").
		Where(squirrel.Eq{"id": ids})
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &roles, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return roles, nil
	} else if err != nil {
		return roles, fmt.Errorf("unable to get roles for guild %d: %w", ids, err)
	}
	return roles, nil
}

func (e *Entity) CreateRole(ctx context.Context, id, guildId int64, name string, color int, permissions int64) error {
	q := squirrel.Insert("roles").
		PlaceholderFormat(squirrel.Dollar).
		Columns("id", "guild_id", "name", "color", "permissions").
		Values(id, guildId, name, color, permissions)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to create role for guild %d: %w", guildId, err)
	}
	return nil
}

func (e *Entity) RemoveRole(ctx context.Context, id int64) error {
	q := squirrel.Delete("roles").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to remove role for guild %d: %w", id, err)
	}
	return nil
}

func (e *Entity) SetRoleColor(ctx context.Context, id int64, color int) error {
	q := squirrel.Update("roles").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("color", color)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to set role color for guild %d: %w", id, err)
	}
	return nil
}

func (e *Entity) SetRoleName(ctx context.Context, id int64, name string) error {
	q := squirrel.Update("roles").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("name", name)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to set role name for guild %d: %w", id, err)
	}
	return nil
}

func (e *Entity) SetRolePermissions(ctx context.Context, id int64, permissions int64) error {
	q := squirrel.Update("roles").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": id}).
		Set("permissions", permissions)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to set role permissions for guild %d: %w", id, err)
	}
	return nil
}
