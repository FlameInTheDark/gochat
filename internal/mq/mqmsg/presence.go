package mqmsg

import "encoding/json"

type PresenceUpdate struct {
	UserID           int64             `json:"user_id"`
	Status           string            `json:"status"`
	CustomStatusText string            `json:"custom_status_text,omitempty"`
	Since            int64             `json:"since"`
	ClientStatus     map[string]string `json:"client_status,omitempty"`
	VoiceChannelID   *int64            `json:"voice_channel_id,omitempty"`
}

func (m *PresenceUpdate) EventType() *EventType    { return nil }
func (m *PresenceUpdate) Operation() OPCodeType    { return OPCodePresenceUpdate }
func (m *PresenceUpdate) Marshal() ([]byte, error) { return json.Marshal(m) }

type PresenceSubscription struct {
	Add    []int64 `json:"add,omitempty"`
	Remove []int64 `json:"remove,omitempty"`
	Set    []int64 `json:"set,omitempty"`
	Clear  bool    `json:"clear,omitempty"`
}
