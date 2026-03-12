package model

import "time"

type Channel struct {
	Id              int64       `db:"id"`
	Name            string      `db:"name"`
	Type            ChannelType `db:"type"`
	ParentID        *int64      `db:"parent_id"`
	CreatorID       *int64      `db:"creator_id"`
	Permissions     *int64      `db:"permissions"`
	Topic           *string     `db:"topic"`
	VoiceRegion     *string     `db:"voice_region"`
	Private         bool        `db:"private"`
	Closed          bool        `db:"closed"`
	LastMessage     int64       `db:"last_message"`
	MessageCount    int64       `db:"message_count"`
	MessagePosition int64       `db:"message_position"`
	CreatedAt       time.Time   `db:"created_at"`
}

type ChannelType int

const (
	ChannelTypeGuild         ChannelType = iota // Default text channel in guild
	ChannelTypeGuildVoice                       // Voice channel in guild
	ChannelTypeGuildCategory                    // Category channel in guild
	ChannelTypeDM                               // DM channel. Can't be created in Guild
	ChannelTypeGroupDM                          // Group DM channel. Can't be created in Guild'
	ChannelTypeThread                           // Thread channel
)
