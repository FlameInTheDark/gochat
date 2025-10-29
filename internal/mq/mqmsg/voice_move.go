package mqmsg

import "encoding/json"

// VoiceMove notifies a user to move to another voice channel and provides SFU connection info.
type VoiceMove struct {
	UserID   int64  `json:"-"`
	Channel  int64  `json:"channel"`
	SFUURL   string `json:"sfu_url"`
	SFUToken string `json:"sfu_token"`
}

func (m *VoiceMove) EventType() *EventType    { t := EventTypeRTCMoved; return &t }
func (m *VoiceMove) Operation() OPCodeType    { return OPCodeRTC }
func (m *VoiceMove) Marshal() ([]byte, error) { return json.Marshal(m) }
