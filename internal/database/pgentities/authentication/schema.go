package authentication

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) CreateAuthentication(ctx context.Context, userId int64, email, passwordHash string) error {
	q := squirrel.Insert("authentications").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "email", "password_hash").
		Values(userId, email, passwordHash)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to create authentication: %w", err)
	}
	return nil
}

func (e *Entity) RemoveAuthentication(ctx context.Context, userId int64) error {
	q := squirrel.Delete("authentications").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"user_id": userId})
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to remove authentication: %w", err)
	}
	return nil
}

func (e *Entity) GetAuthenticationByEmail(ctx context.Context, email string) (model.Authentication, error) {
	var a model.Authentication
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("authentications").
		Where(squirrel.Eq{"email": email}).
		Limit(1)
	sql, args, err := q.ToSql()
	if err != nil {
		return model.Authentication{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &a, sql, args...)
	if err != nil {
		return a, fmt.Errorf("unable to get authentication by email: %w", err)
	}
	return a, nil
}

func (e *Entity) GetAuthenticationByUserId(ctx context.Context, userId int64) (model.Authentication, error) {
	var a model.Authentication
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("authentications").
		Where(squirrel.Eq{"user_id": userId}).
		Limit(1)
	sql, args, err := q.ToSql()
	if err != nil {
		return model.Authentication{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &a, sql, args...)
	if err != nil {
		return a, fmt.Errorf("unable to get authentication by user id: %w", err)
	}
	return a, nil
}

func (e *Entity) SetPasswordHash(ctx context.Context, userId int64, hash string) error {
	q := squirrel.Update("authentications").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"user_id": userId}).
		Set("password_hash", hash)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to set password hash: %w", err)
	}
	return nil
}
