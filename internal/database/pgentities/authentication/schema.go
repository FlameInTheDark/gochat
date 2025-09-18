package authentication

import (
	"context"
	"fmt"
	"time"

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

func (e *Entity) CreateRecovery(ctx context.Context, userId int64, token string, expires time.Time) error {
	q := squirrel.Insert("recoveries").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "token", "expires_at").
		Values(userId, token, expires)
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to create recovery: %w", err)
	}
	return nil
}

func (e *Entity) RemoveRecovery(ctx context.Context, userId int64) error {
	q := squirrel.Delete("recoveries").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"user_id": userId})
	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("unable to remove recovery: %w", err)
	}
	return nil
}

func (e *Entity) GetRecoveryByUserId(ctx context.Context, userId int64) (model.Recovery, error) {
	var r model.Recovery
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("recoveries").
		Where(squirrel.Eq{"user_id": userId}).
		Limit(1)
	sql, args, err := q.ToSql()
	if err != nil {
		return model.Recovery{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &r, sql, args...)
	if err != nil {
		return r, fmt.Errorf("unable to get recovery by user id: %w", err)
	}
	return r, nil
}

func (e *Entity) GetRecovery(ctx context.Context, userId int64, token string) (model.Recovery, error) {
	var r model.Recovery
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("recoveries").
		Where(squirrel.And{squirrel.Eq{"token": token}, squirrel.Eq{"expired": false}, squirrel.Eq{"user_id": userId}}).
		Limit(1)
	sql, args, err := q.ToSql()
	if err != nil {
		return model.Recovery{}, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &r, sql, args...)
	if err != nil {
		return r, fmt.Errorf("unable to get recovery by token: %w", err)
	}
	return r, nil
}
