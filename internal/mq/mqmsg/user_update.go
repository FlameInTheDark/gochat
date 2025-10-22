package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

// UpdateUser is emitted when user's public profile changes (name, avatar).
type UpdateUser struct {
	User dto.User `json:"user"`
}

func (m *UpdateUser) EventType() *EventType {
	e := EventTypeUserUpdate
	return &e
}

func (m *UpdateUser) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UpdateUser) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
