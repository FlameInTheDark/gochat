package model

import "time"

type Friend struct {
	UserID    int64     `db:"user_id"`
	FriendID  int64     `db:"friend_id"`
	CreatedAt time.Time `db:"created_at"`
}
