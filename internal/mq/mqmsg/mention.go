package mqmsg

import (
	"encoding/json"
)

type Mention struct {
	GuildId   *int64 `json:"guild_id"`
	ChannelId int64  `json:"channel_id"`
	MessageId int64  `json:"message_id"`
	AuthorId  int64  `json:"author_id"`
	Type      int    `json:"type"`
}

func (m *Mention) EventType() *EventType {
	e := EventTypeMention
	return &e
}

func (m *Mention) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *Mention) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
