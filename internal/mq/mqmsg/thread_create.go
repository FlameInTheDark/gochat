package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type CreateThread struct {
	GuildId *int64      `json:"guild_id"`
	Thread  dto.Channel `json:"thread"`
}

func (m *CreateThread) EventType() *EventType {
	e := EventTypeThreadCreate
	return &e
}

func (m *CreateThread) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *CreateThread) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
