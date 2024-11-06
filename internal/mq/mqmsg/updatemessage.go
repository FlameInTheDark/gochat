package mqmsg

import (
	"encoding/json"
	"github.com/FlameInTheDark/gochat/internal/dto"
)

type UpdateMessage struct {
	GuildId *int64      `json:"guild_id"`
	Message dto.Message `json:"message"`
}

func (m *UpdateMessage) EventType() *EventType {
	e := EventTypeMessageUpdate
	return &e
}

func (m *UpdateMessage) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UpdateMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
