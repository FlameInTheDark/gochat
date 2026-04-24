package mqmsg

import (
	"encoding/json"

	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
)

type GuildMemberStartStream struct {
	GuildId   int64                   `json:"guild_id"`
	ChannelId int64                   `json:"channel_id"`
	UserId    int64                   `json:"user_id"`
	Stream    streammeta.ActiveStream `json:"stream"`
}

func (m *GuildMemberStartStream) EventType() *EventType {
	e := EventTypeGuildMemberStartStream
	return &e
}

func (m *GuildMemberStartStream) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *GuildMemberStartStream) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
