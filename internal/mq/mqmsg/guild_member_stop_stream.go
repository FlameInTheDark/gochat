package mqmsg

import "encoding/json"

type GuildMemberStopStream struct {
	GuildId   int64  `json:"guild_id"`
	ChannelId int64  `json:"channel_id"`
	UserId    int64  `json:"user_id"`
	StreamId  int64  `json:"stream_id"`
	Reason    string `json:"reason,omitempty"`
}

func (m *GuildMemberStopStream) EventType() *EventType {
	e := EventTypeGuildMemberStopStream
	return &e
}

func (m *GuildMemberStopStream) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *GuildMemberStopStream) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
