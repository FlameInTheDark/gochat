package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type UpdateThread struct {
	GuildId *int64      `json:"guild_id"`
	Thread  dto.Channel `json:"thread"`
}

func (m *UpdateThread) EventType() *EventType {
	e := EventTypeThreadUpdate
	return &e
}

func (m *UpdateThread) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UpdateThread) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
