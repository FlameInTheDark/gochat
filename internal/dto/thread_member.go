package dto

import "time"

type ThreadMember struct {
	UserId        int64     `json:"user_id" example:"2230469276416868352"`
	JoinTimestamp time.Time `json:"join_timestamp"`
	Flags         int       `json:"flags"`
}
