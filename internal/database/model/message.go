package model

import "time"

type Message struct {
	Id          int64
	ChannelId   int64
	UserId      int64
	Content     string
	Attachments []int64
	UpdatedAt   *time.Time
}
