package registration

import (
	"context"
	"fmt"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"log"
)

const (
	getRegistrationByUserId = `SELECT user_id, email, confirmation_token FROM gochat.registrations WHERE user_id = ?;`
	getRegistrationByEmail  = `SELECT user_id, email, confirmation_token, created_at FROM gochat.registrations WHERE email = ?;`
	createRegistration      = `INSERT INTO gochat.registrations (user_id, email, confirmation_token, created_at) VALUES (?, ?, ?, toTimestamp(now()));`
	setRegistrationToken    = `UPDATE gochat.registrations SET confirmation_token = ? WHERE user_id = ?;`
	removeRegistration      = `DELETE FROM gochat.registrations WHERE user_id = ?;`
)

func (e *Entity) GetRegistrationByUserId(ctx context.Context, userId int64) (model.Registration, error) {
	var r model.Registration
	err := e.c.Session().
		Query(getRegistrationByUserId).
		WithContext(ctx).
		Bind(userId).
		Scan(&r.UserId, &r.Email, &r.ConfirmationToken)
	if err != nil {
		return r, fmt.Errorf("unable to get registration by id: %w", err)
	}
	return r, nil
}

func (e *Entity) GetRegistrationByEmail(ctx context.Context, email string) (model.Registration, error) {
	var r model.Registration
	err := e.c.Session().
		Query(getRegistrationByEmail).
		WithContext(ctx).
		Bind(email).
		Scan(&r.UserId, &r.Email, &r.ConfirmationToken, &r.CreatedAt)
	if err != nil {
		return r, fmt.Errorf("unable to get registration by email: %w", err)
	}
	return r, nil
}

func (e *Entity) CreateRegistration(ctx context.Context, userId int64, email string, confirmation string) error {
	log.Println("inside create reg", email)
	err := e.c.Session().
		Query(createRegistration).
		WithContext(ctx).
		Bind(userId, email, confirmation).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create registration: %w", err)
	}
	return nil
}

func (e *Entity) RemoveRegistration(ctx context.Context, userId int64) error {
	err := e.c.Session().
		Query(removeRegistration).
		WithContext(ctx).
		Bind(userId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove registration: %w", err)
	}
	return nil
}

func (e *Entity) SetRegistrationToken(ctx context.Context, userId int64, token string) error {
	err := e.c.Session().
		Query(setRegistrationToken).
		WithContext(ctx).
		Bind(token, userId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set registration token: %w", err)
	}
	return nil
}
