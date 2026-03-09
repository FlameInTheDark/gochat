package mqmsg

import "encoding/json"

// VoiceStateUpdate is sent when a user updates their mute/deafen status in a voice channel.
// This is broadcast to guild members so they can see who is muted/deafened.
type VoiceStateUpdate struct {
	GuildId   int64 `json:"guild_id"`
	UserId    int64 `json:"user_id"`
	ChannelId int64 `json:"channel_id"`
	Mute      bool  `json:"mute"`
	Deafen    bool  `json:"deafen"`
}

func (m *VoiceStateUpdate) EventType() *EventType {
	e := EventTypeVoiceStateUpdate
	return &e
}

func (m *VoiceStateUpdate) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *VoiceStateUpdate) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
