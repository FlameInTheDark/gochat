package mqmsg

import (
	"encoding/json"
)

type GuildChannelMessage struct {
	GuildId   *int64 `json:"guild_id"`
	ChannelId int64  `json:"channel_id"`
	MessageId int64  `json:"message_id"`
}

func (m *GuildChannelMessage) EventType() *EventType {
	e := EventTypeGuildChannelMessage
	return &e
}

func (m *GuildChannelMessage) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *GuildChannelMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
