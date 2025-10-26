package mqmsg

import (
	"encoding/json"
)

type ChannelUserTyping struct {
	ChannelId int64 `json:"channel_id"`
	UserId    int64 `json:"user_id"`
}

func (m *ChannelUserTyping) EventType() *EventType {
	e := EventTypeChannelUserTyping
	return &e
}

func (m *ChannelUserTyping) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *ChannelUserTyping) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
