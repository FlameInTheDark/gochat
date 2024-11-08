package model

import "time"

type Channel struct {
	Id          int64
	Name        string
	Type        ChannelType
	ParentID    *int64
	Permissions int64
	Private     bool
	CreatedAt   time.Time
}

type ChannelType int

const (
	ChannelTypeGuild ChannelType = iota
	ChannelTypeGuildVoice
	ChannelTypeDM
	ChannelTypeGroupDM
	ChannelTypeThread
)
