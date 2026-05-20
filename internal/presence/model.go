package presence

import streammeta "github.com/FlameInTheDark/gochat/internal/stream"

type Presence struct {
	UserID           int64                    `json:"user_id"`
	Status           string                   `json:"status"`
	Since            int64                    `json:"since"`
	CustomStatusText string                   `json:"custom_status_text,omitempty"`
	ClientStatus     map[string]string        `json:"client_status,omitempty"`
	VoiceChannelID   *int64                   `json:"voice_channel_id,omitempty"`
	Mute             bool                     `json:"mute,omitempty"`
	Deafen           bool                     `json:"deafen,omitempty"`
	SelfVideo        bool                     `json:"self_video,omitempty"`
	ActiveStream     *streammeta.ActiveStream `json:"active_stream,omitempty"`
}

// SessionPresence represents a single device/session presence record.
// Values are stored in a Redis hash per user with field = sessionID.
type SessionPresence struct {
	SessionID        string `json:"session_id"`
	ConnectionID     string `json:"connection_id,omitempty"`
	Generation       int64  `json:"generation,omitempty"`
	Status           string `json:"status"`
	Platform         string `json:"platform,omitempty"`
	Since            int64  `json:"since"`
	UpdatedAt        int64  `json:"updated_at"`
	ExpiresAt        int64  `json:"expires_at"`
	CustomStatusText string `json:"custom_status_text,omitempty"`
	VoiceChannelID   *int64 `json:"voice_channel_id,omitempty"`
	Mute             bool   `json:"mute,omitempty"`
	Deafen           bool   `json:"deafen,omitempty"`
	SelfVideo        bool   `json:"self_video,omitempty"`
}

const (
	StatusOnline  = "online"
	StatusIdle    = "idle"
	StatusDND     = "dnd"
	StatusOffline = "offline"
)
