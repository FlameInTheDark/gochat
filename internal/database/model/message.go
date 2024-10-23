package model

import "time"

type Message struct {
	Id          int64
	ChannelId   int64
	UserId      int64
	Content     string
	Attachments []int64
	Type        int
	Reference   int64
	Thread      int64
	EditedAt    *time.Time
}

type MessageType int

const (
	MessageTypeSystem MessageType = iota
	MessageTypeChat
	MessageTypeReply
)
