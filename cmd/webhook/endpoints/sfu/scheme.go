package sfu

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type HeartbeatRequest struct {
	ID     string `json:"id"`
	Region string `json:"region"`
	URL    string `json:"url"`
	Load   int64  `json:"load"`
}

func (r HeartbeatRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required),
		validation.Field(&r.Region, validation.Required),
		validation.Field(&r.URL, validation.Required),
	)
}

type ChannelUserJoin struct {
	ChannelId int64  `json:"channel_id"`
	UserId    int64  `json:"user_id"`
	GuildId   *int64 `json:"guild_id,omitempty"`
}

func (r ChannelUserJoin) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ChannelId, validation.Required),
		validation.Field(&r.UserId, validation.Required),
	)
}

type ChannelUserLeave struct {
	ChannelId int64  `json:"channel_id"`
	UserId    int64  `json:"user_id"`
	GuildId   *int64 `json:"guild_id,omitempty"`
}

func (r ChannelUserLeave) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ChannelId, validation.Required),
		validation.Field(&r.UserId, validation.Required),
	)
}

type ChannelAlive struct {
	ChannelId int64  `json:"channel_id"`
	GuildId   *int64 `json:"guild_id,omitempty"`
}

func (r ChannelAlive) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ChannelId, validation.Required),
	)
}
