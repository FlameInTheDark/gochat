package mqmsg

import "encoding/json"

// VoiceRegionChanging is sent to guild members before a voice region migration begins.
// Clients should wait DelayMs milliseconds before calling JoinVoice to reconnect.
type VoiceRegionChanging struct {
	ChannelId int64  `json:"channel_id"`
	Region    string `json:"region"`
	DelayMs   int    `json:"delay_ms"`
}

func (m *VoiceRegionChanging) EventType() *EventType {
	t := EventTypeGuildVoiceRegionChanging
	return &t
}
func (m *VoiceRegionChanging) Operation() OPCodeType    { return OpCodeDispatch }
func (m *VoiceRegionChanging) Marshal() ([]byte, error) { return json.Marshal(m) }
