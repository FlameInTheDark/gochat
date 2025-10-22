package mqmsg

// PresenceUpdateRequest represents the client → WS presence update payload.
// It is intentionally minimal and distinct from the server-dispatched PresenceUpdate event.
type PresenceUpdateRequest struct {
	Status           string `json:"status"`
	Platform         string `json:"platform,omitempty"`
	CustomStatusText string `json:"custom_status_text,omitempty"`
	VoiceChannelID   *int64 `json:"voice_channel_id,omitempty"`
}
