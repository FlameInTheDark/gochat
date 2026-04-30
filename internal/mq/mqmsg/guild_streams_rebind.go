package mqmsg

import "encoding/json"

type GuildStreamsRebind struct {
	GuildId   int64   `json:"guild_id"`
	ChannelId int64   `json:"channel_id"`
	StreamIds []int64 `json:"stream_ids"`
	JitterMs  int     `json:"jitter_ms,omitempty"`
}

func (m *GuildStreamsRebind) EventType() *EventType {
	e := EventTypeGuildStreamsRebind
	return &e
}

func (m *GuildStreamsRebind) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *GuildStreamsRebind) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
