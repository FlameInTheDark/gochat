package dto

import "time"

type Member struct {
	User     User      `json:"user"`                                          // Guild member data
	Username *string   `json:"username" example:"FancyUserName"`              // Username in this guild
	Avatar   *int64    `json:"avatar" example:"2230469276416868352"`          // Avatar ID
	JoinAt   time.Time `json:"join_at"`                                       // Join date
	Roles    []int64   `json:"roles,omitempty" example:"2230469276416868352"` // List of assigned role IDs
}
