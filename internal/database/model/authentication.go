package model

import "time"

type Authentication struct {
	UserId       int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
