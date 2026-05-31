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
	Flags       int64     `json:"flags" db:"flags"`
	UploadLimit *int64    `json:"upload_limit" db:"upload_limit"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

const (
	UserFlagBot int64 = 1 << iota
)

func (u User) IsBot() bool {
	return u.Flags&UserFlagBot != 0
}
