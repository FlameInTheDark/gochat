package userrole

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) GetUserRoles(ctx context.Context, guildID, userId int64) ([]model.UserRole, error) {
	var roles []model.UserRole
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("user_roles").
		Where(
			squirrel.And{
				squirrel.Eq{"guild_id": guildID},
				squirrel.Eq{"user_id": userId},
			},
		)
	raw, args, err := q.ToSql()
	if err != nil {
		return roles, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &roles, raw, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return roles, nil
		}
		return nil, fmt.Errorf("unable to get user roles: %w", err)
	}
	return roles, nil
}

func (e *Entity) AddUserRole(ctx context.Context, guildID, userId, roleId int64) error {
	q := squirrel.Insert("user_roles").
		PlaceholderFormat(squirrel.Dollar).
		Columns("guild_id", "user_id", "role_id").
		Values(guildID, userId, roleId)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to add user role: %w", err)
	}
	return nil
}

func (e *Entity) RemoveUserRole(ctx context.Context, guildID, userId, roleId int64) error {
	q := squirrel.Delete("user_roles").
		PlaceholderFormat(squirrel.Dollar).
		Where(
			squirrel.And{
				squirrel.Eq{"guild_id": guildID},
				squirrel.Eq{"user_id": userId},
				squirrel.Eq{"role_id": roleId},
			},
		)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("unable to remove user role: %w", err)
	}
	return nil
}

// RemoveRoleAssignments removes role from all users within a guild
func (e *Entity) RemoveRoleAssignments(ctx context.Context, guildID, roleId int64) error {
	q := squirrel.Delete("user_roles").
		PlaceholderFormat(squirrel.Dollar).
		Where(
			squirrel.And{
				squirrel.Eq{"guild_id": guildID},
				squirrel.Eq{"role_id": roleId},
			},
		)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("unable to remove role assignments: %w", err)
	}
	return nil
}
