package presence

type Presence struct {
	UserID           int64  `json:"user_id"`
	Status           string `json:"status"`
	Since            int64  `json:"since"`
	CustomStatusText string `json:"custom_status_text,omitempty"`
	VoiceChannelID   *int64 `json:"voice_channel_id,omitempty"`
}

// SessionPresence represents a single device/session presence record.
// Values are stored in a Redis hash per user with field = sessionID.
type SessionPresence struct {
	SessionID        string `json:"session_id"`
	Status           string `json:"status"`
	Platform         string `json:"platform,omitempty"`
	Since            int64  `json:"since"`
	UpdatedAt        int64  `json:"updated_at"`
	ExpiresAt        int64  `json:"expires_at"`
	CustomStatusText string `json:"custom_status_text,omitempty"`
	VoiceChannelID   *int64 `json:"voice_channel_id,omitempty"`
}

const (
	StatusOnline  = "online"
	StatusIdle    = "idle"
	StatusDND     = "dnd"
	StatusOffline = "offline"
)
