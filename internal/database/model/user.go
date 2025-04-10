package model

import "time"

type User struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Avatar      *int64    `json:"avatar"`
	Blocked     bool      `json:"blocked"`
	UploadLimit *int64    `json:"upload_limit"`
	CreatedAt   time.Time `json:"created_at"`
}
