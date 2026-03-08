package model

import "time"

type Message struct {
	Id             int64
	ChannelId      int64
	UserId         int64
	Content        string
	Attachments    []int64
	EmbedsJSON     *string
	AutoEmbedsJSON *string
	Flags          *int
	Type           int
	Reference      int64
	Thread         int64
	EditedAt       *time.Time
}

type MessageType int

const (
	MessageTypeChat MessageType = iota
	MessageTypeReply
	MessageTypeJoin
)

const (
	MessageFlagSuppressEmbeds = 1 << 2
	MessageFlagBannedAuthor   = 1 << 3
)

func NormalizeMessageFlags(flags *int) int {
	if flags == nil {
		return 0
	}
	return *flags
}

func HasMessageFlag(flags, flag int) bool {
	return flags&flag == flag
}
