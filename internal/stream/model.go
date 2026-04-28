package stream

const (
	RolePublisher = "publisher"
	RoleViewer    = "viewer"

	SourceTypeScreen      = "screen"
	SourceTypeApplication = "application"

	AudioModeDesktop     = "desktop"
	AudioModeApplication = "application"
	AudioModeNone        = "none"
)

// ActiveStream is the user-facing metadata shared across API, presence, and
// guild events for an active stream in a voice channel.
type ActiveStream struct {
	ID         int64  `json:"id"`
	ChannelID  int64  `json:"channel_id"`
	SourceType string `json:"source_type"`
	AudioMode  string `json:"audio_mode"`
	StartedAt  int64  `json:"started_at"`
}

// Metadata is the backend representation of an active stream route/state.
type Metadata struct {
	ActiveStream
	GuildID      int64  `json:"guild_id"`
	OwnerUserID  int64  `json:"owner_user_id"`
	Region       string `json:"region,omitempty"`
	RouteID      string `json:"route_id,omitempty"`
	RouteURL     string `json:"route_url,omitempty"`
	PublisherSID string `json:"publisher_session_id,omitempty"`
}

// RouteBinding pins a stream to a specific stream service instance.
type RouteBinding struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Region string `json:"region,omitempty"`
}
