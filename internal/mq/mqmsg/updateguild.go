package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type UpdateGuild struct {
	Guild dto.Guild `json:"guild"`
}

func (m *UpdateGuild) EventType() *EventType {
	e := EventTypeGuildUpdate
	return &e
}

func (m *UpdateGuild) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UpdateGuild) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
