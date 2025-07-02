package model

import "time"

type Registration struct {
	UserId            int64     `db:"user_id"`
	Email             string    `db:"email"`
	ConfirmationToken string    `db:"confirmation_token"`
	CreatedAt         time.Time `db:"created_at"`
}
