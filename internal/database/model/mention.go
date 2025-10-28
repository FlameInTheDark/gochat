package model

type Mention struct {
	UserId    int64
	ChannelId int64
	MessageId int64
	AuthorId  int64
}

type ChannelMention struct {
	GuildId   int64
	ChannelId int64
	MessageId int64
	AuthorId  int64
	RoleId    *int64
	Type      int
}

type ChannelMentionType int

const (
	ChannelMentionUser ChannelMentionType = iota
	ChannelMentionRole
	ChannelMentionEveryone
	ChannelMentionHere
)
