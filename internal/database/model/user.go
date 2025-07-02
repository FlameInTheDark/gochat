package model

import "time"

type User struct {
	Id          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Avatar      *int64    `json:"avatar" db:"avatar"`
	Blocked     bool      `json:"blocked" db:"blocked"`
	UploadLimit *int64    `json:"upload_limit" db:"upload_limit"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
