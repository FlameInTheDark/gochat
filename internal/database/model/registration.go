package model

import "time"

type Registration struct {
	UserId            int64
	Email             string
	ConfirmationToken string
	CreatedAt         time.Time
}
