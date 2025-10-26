package mqmsg

import (
	"encoding/json"
)

type GuildMemberJoinVoice struct {
	GuildId   int64 `json:"guild_id"`
	UserId    int64 `json:"user_id"`
	ChannelId int64 `json:"channel_id"`
}

func (m *GuildMemberJoinVoice) EventType() *EventType {
	e := EventTypeGuildMemberJoinVoice
	return &e
}

func (m *GuildMemberJoinVoice) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *GuildMemberJoinVoice) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
