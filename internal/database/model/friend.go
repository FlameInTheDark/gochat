package model

import "time"

type Friend struct {
	UserID    int64
	FriendID  int64
	CreatedAt time.Time
}
