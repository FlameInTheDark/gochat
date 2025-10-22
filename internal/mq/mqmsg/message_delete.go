package mqmsg

import (
	"encoding/json"
)

type DeleteMessage struct {
	GuildId   *int64 `json:"guild_id"`
	ChannelId int64  `json:"channel_id"`
	MessageId int64  `json:"message_id"`
}

func (m *DeleteMessage) EventType() *EventType {
	e := EventTypeMessageDelete
	return &e
}

func (m *DeleteMessage) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *DeleteMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
