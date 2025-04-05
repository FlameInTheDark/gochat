package model

import "time"

type Channel struct {
	Id          int64
	Name        string
	Type        ChannelType
	ParentID    *int64
	Permissions *int64
	Topic       *string
	Private     bool
	LastMessage int64
	CreatedAt   time.Time
}

type ChannelType int

const (
	ChannelTypeGuild ChannelType = iota
	ChannelTypeGuildVoice
	ChannelTypeGuildCategory
	ChannelTypeDM
	ChannelTypeGroupDM
	ChannelTypeThread
)
