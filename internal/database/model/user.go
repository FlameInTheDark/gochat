package model

import "time"

type User struct {
	Id          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Bio         *string   `json:"bio" db:"bio"`
	BannerColor *int      `json:"banner_color" db:"banner_color"`
	PanelColor  *int      `json:"panel_color" db:"panel_color"`
	Avatar      *int64    `json:"avatar" db:"avatar"`
	Banner      *int64    `json:"banner" db:"banner"`
	Blocked     bool      `json:"blocked" db:"blocked"`
	UploadLimit *int64    `json:"upload_limit" db:"upload_limit"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
