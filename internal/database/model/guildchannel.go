package model

type GuildChannel struct {
	GuildId   int64
	ChannelId int64
	Position  int
}

type GuildChannelUpdatePosition struct {
	GuildId   int64
	ChannelId int64
	Position  int
}
