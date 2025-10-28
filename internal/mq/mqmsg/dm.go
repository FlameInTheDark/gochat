package mqmsg

import "encoding/json"

// DMMessage notifies a user about a new direct message in a DM channel
type DMMessage struct {
	ChannelId int64     `json:"channel_id"`
	MessageId int64     `json:"message_id"`
	From      UserBrief `json:"from"`
}

func (m *DMMessage) EventType() *EventType {
	e := EventTypeUserDMMessage
	return &e
}

func (m *DMMessage) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *DMMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
