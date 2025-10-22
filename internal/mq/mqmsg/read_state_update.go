package mqmsg

import (
	"encoding/json"
)

type UpdateReadState struct {
	ChannelId int64 `json:"channel_id"`
	MessageId int64 `json:"message_id"`
}

func (m *UpdateReadState) EventType() *EventType {
	e := EventTypeUserUpdateReadState
	return &e
}

func (m *UpdateReadState) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UpdateReadState) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
