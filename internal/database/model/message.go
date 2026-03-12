package model

import "time"

type Message struct {
	Id               int64
	ChannelId        int64
	UserId           int64
	Content          string
	Position         int64
	Attachments      []int64
	EmbedsJSON       *string
	AutoEmbedsJSON   *string
	Flags            *int
	Type             int
	ReferenceChannel int64
	Reference        int64
	Thread           int64
	EditedAt         *time.Time
}

type MessageType int

const (
	MessageTypeChat MessageType = iota
	MessageTypeReply
	MessageTypeJoin
	MessageTypeThreadCreated
	MessageTypeThreadInitial
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

func IsEditableMessageType(msgType MessageType) bool {
	switch msgType {
	case MessageTypeChat, MessageTypeReply:
		return true
	default:
		return false
	}
}
