package mqmsg

import "encoding/json"

// VoiceRebind notifies clients in a voice channel to reconnect (route changed).
type VoiceRebind struct {
	Channel int64 `json:"channel"`
}

func (m *VoiceRebind) EventType() *EventType    { t := EventTypeRTCServerRebind; return &t }
func (m *VoiceRebind) Operation() OPCodeType    { return OPCodeRTC }
func (m *VoiceRebind) Marshal() ([]byte, error) { return json.Marshal(m) }
