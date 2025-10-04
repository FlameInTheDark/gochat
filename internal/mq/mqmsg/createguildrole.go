package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type CreateGuildRole struct {
	Role dto.Role `json:"role"`
}

func (m *CreateGuildRole) EventType() *EventType {
	e := EventTypeGuildRoleCreate
	return &e
}

func (m *CreateGuildRole) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *CreateGuildRole) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
