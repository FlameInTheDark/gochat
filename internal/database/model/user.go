package model

import "time"

type User struct {
	Id           int64     `json:"id"`
	Name         string    `json:"name"`
	Determinator string    `json:"determinator"`
	Avatar       *int64    `json:"avatar"`
	Blocked      bool      `json:"blocked"`
	CreatedAt    time.Time `json:"created_at"`
}
