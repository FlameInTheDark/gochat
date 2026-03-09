package role

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
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
		Where(squirrel.Eq{"guild_id": guildId}).
		OrderBy("position ASC", "id ASC")
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

func (e *Entity) GetRolesBulk(ctx context.Context, guildID int64, ids []int64) ([]model.Role, error) {
	var roles []model.Role
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("roles").
		Where(squirrel.And{
			squirrel.Eq{"guild_id": guildID},
			squirrel.Eq{"id": ids},
		}).
		OrderBy("position ASC", "id ASC")
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &roles, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return roles, nil
	} else if err != nil {
		return roles, fmt.Errorf("unable to get roles for guild %d: %w", guildID, err)
	}
	return roles, nil
}

func (e *Entity) CreateRole(ctx context.Context, id, guildId int64, name string, color int, permissions int64) error {
	const query = `
INSERT INTO roles (id, guild_id, name, color, permissions, position)
SELECT $1, $2, $3, $4, $5, COALESCE(MAX(position), -1) + 1
FROM roles
WHERE guild_id = $2
`
	_, err := e.c.ExecContext(ctx, query, id, guildId, name, color, permissions)
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

func (e *Entity) SetRolePosition(ctx context.Context, updates []model.RoleUpdatePosition) (err error) {
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

	guildID := updates[0].GuildId
	roleIDs := make([]int64, 0, len(updates))
	positions := make([]int32, 0, len(updates))
	for _, u := range updates {
		roleIDs = append(roleIDs, u.RoleId)
		positions = append(positions, int32(u.Position))
	}

	q := squirrel.
		Update("roles AS r").
		PlaceholderFormat(squirrel.Dollar).
		Prefix(
			"WITH v(id, position) AS (SELECT * FROM unnest(?::bigint[], ?::int[]))",
			pq.Array(roleIDs), pq.Array(positions),
		).
		Set("position", squirrel.Expr("v.position")).
		From("v").
		Where(squirrel.Eq{"r.guild_id": guildID}).
		Where(squirrel.Expr("r.id = v.id"))

	raw, args, buildErr := q.ToSql()
	if buildErr != nil {
		return fmt.Errorf("build update: %w", buildErr)
	}
	if _, err = tx.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}
	return nil
}
