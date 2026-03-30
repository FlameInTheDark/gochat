package model

import "time"

type Authentication struct {
	UserId         int64     `db:"user_id"`
	Email          string    `db:"email"`
	PasswordHash   string    `db:"password_hash"`
	SessionVersion int64     `db:"session_version"`
	CreatedAt      time.Time `db:"created_at"`
}

type Recovery struct {
	UserId    int64     `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
