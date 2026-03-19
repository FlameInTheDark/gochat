package mqmsg

import "encoding/json"

type DeleteThread struct {
	GuildId  *int64 `json:"guild_id"`
	ThreadId int64  `json:"thread_id"`
}

func (m *DeleteThread) EventType() *EventType {
	e := EventTypeThreadDelete
	return &e
}

func (m *DeleteThread) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *DeleteThread) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
