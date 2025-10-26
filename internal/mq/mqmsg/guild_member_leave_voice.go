package mqmsg

import (
	"encoding/json"
)

type GuildMemberLeaveVoice struct {
	GuildId   int64 `json:"guild_id"`
	UserId    int64 `json:"user_id"`
	ChannelId int64 `json:"channel_id"`
}

func (m *GuildMemberLeaveVoice) EventType() *EventType {
	e := EventTypeGuildMemberLeaveVoice
	return &e
}

func (m *GuildMemberLeaveVoice) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *GuildMemberLeaveVoice) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
