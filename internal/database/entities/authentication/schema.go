package authentication

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	createAuthentication      = `INSERT INTO gochat.authentications (user_id, email, password_hash, created_at) VALUES (?, ?, ?, toTimestamp(now()))`
	removeAuthentication      = `DELETE FROM gochat.authentications WHERE user_id=?`
	getAuthenticationByEmail  = `SELECT user_id, email, password_hash, created_at FROM gochat.authentications WHERE email = ?`
	getAuthenticationByUserId = `SELECT user_id, email, password_hash, created_at FROM gochat.authentications WHERE user_id = ?`
	setPasswordHash           = `UPDATE gochat.authentications SET password_hash = ? WHERE user_id = ?`
)

func (e *Entity) CreateAuthentication(ctx context.Context, user_id int64, email, password_hash string) error {
	err := e.c.Session().
		Query(createAuthentication).
		WithContext(ctx).
		Bind(user_id, email, password_hash).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create authentication: %w", err)
	}
	return nil
}

func (e *Entity) RemoveAuthentication(ctx context.Context, userId int64) error {
	err := e.c.Session().
		Query(removeAuthentication).
		WithContext(ctx).
		Bind(userId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove authentication: %w", err)
	}
	return nil
}

func (e *Entity) GetAuthenticationByEmail(ctx context.Context, email string) (model.Authentication, error) {
	var a model.Authentication
	err := e.c.Session().
		Query(getAuthenticationByEmail).
		WithContext(ctx).
		Bind(email).
		Scan(&a.UserId, &a.Email, &a.PasswordHash, &a.CreatedAt)
	if err != nil {
		return a, fmt.Errorf("unable to get authentication by email: %w", err)
	}
	return a, nil
}

func (e *Entity) GetAuthenticationByUserId(ctx context.Context, userId int64) (model.Authentication, error) {
	var a model.Authentication
	err := e.c.Session().
		Query(getAuthenticationByUserId).
		WithContext(ctx).
		Bind(userId).
		Scan(&a.UserId, &a.Email, &a.PasswordHash, &a.CreatedAt)
	if err != nil {
		return a, fmt.Errorf("unable to get authentication by user id: %w", err)
	}
	return a, nil
}

func (e *Entity) SetPasswordHash(ctx context.Context, userId int64, hash string) error {
	err := e.c.Session().
		Query(setPasswordHash).
		WithContext(ctx).
		Bind(hash, userId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set password hash: %w", err)
	}
	return nil
}
