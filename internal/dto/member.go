package dto

import "time"

type Member struct {
	User     User      `json:"user_id"`
	Username *string   `json:"username"`
	Avatar   *int64    `json:"avatar"`
	JoinAt   time.Time `json:"join_at"`
	Roles    []int64   `json:"roles"`
}
