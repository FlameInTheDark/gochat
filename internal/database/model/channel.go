package model

import "time"

type Channel struct {
	Id          int64
	Name        string
	Type        int
	ParentID    int64
	GuildId     int64
	Permissions int64
	CreatedAt   time.Time
}

type ChannelType int

const (
	ChannelTypeGuild ChannelType = iota
	ChannelTypeGuildVoice
	ChannelTypeDM
	ChannelTypeThread
)
