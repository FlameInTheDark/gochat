package stream

import validation "github.com/go-ozzo/ozzo-validation/v4"

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

type StartRequest struct {
	StreamID           int64  `json:"stream_id"`
	ChannelID          int64  `json:"channel_id"`
	GuildID            int64  `json:"guild_id"`
	OwnerUserID        int64  `json:"owner_user_id"`
	SourceType         string `json:"source_type"`
	AudioMode          string `json:"audio_mode"`
	StartedAt          int64  `json:"started_at,omitempty"`
	RouteID            string `json:"route_id"`
	RouteURL           string `json:"route_url"`
	Region             string `json:"region,omitempty"`
	PublisherSessionID string `json:"publisher_session_id,omitempty"`
}

func (r StartRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.StreamID, validation.Required),
		validation.Field(&r.ChannelID, validation.Required),
		validation.Field(&r.GuildID, validation.Required),
		validation.Field(&r.OwnerUserID, validation.Required),
		validation.Field(&r.SourceType, validation.Required),
		validation.Field(&r.AudioMode, validation.Required),
		validation.Field(&r.RouteID, validation.Required),
		validation.Field(&r.RouteURL, validation.Required),
	)
}

type StopRequest struct {
	StreamID           int64  `json:"stream_id"`
	ChannelID          int64  `json:"channel_id"`
	GuildID            int64  `json:"guild_id"`
	OwnerUserID        int64  `json:"owner_user_id"`
	PublisherSessionID string `json:"publisher_session_id,omitempty"`
	Reason             string `json:"reason,omitempty"`
}

func (r StopRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.StreamID, validation.Required),
		validation.Field(&r.ChannelID, validation.Required),
		validation.Field(&r.GuildID, validation.Required),
		validation.Field(&r.OwnerUserID, validation.Required),
	)
}

type AliveRequest struct {
	StreamID    int64  `json:"stream_id"`
	ChannelID   int64  `json:"channel_id"`
	OwnerUserID int64  `json:"owner_user_id,omitempty"`
	RouteID     string `json:"route_id,omitempty"`
	RouteURL    string `json:"route_url,omitempty"`
	Region      string `json:"region,omitempty"`
}

func (r AliveRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.StreamID, validation.Required),
	)
}
