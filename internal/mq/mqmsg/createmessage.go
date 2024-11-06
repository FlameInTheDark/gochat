package mqmsg

import (
	"encoding/json"
	"github.com/FlameInTheDark/gochat/internal/dto"
)

type CreateMessage struct {
	GuildId *int64      `json:"guild_id"`
	Message dto.Message `json:"message"`
}

func (m *CreateMessage) EventType() *EventType {
	e := EventTypeMessageCreate
	return &e
}

func (m *CreateMessage) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *CreateMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
