package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type UpdateUserSettings struct {
	Settings model.UserSettingsData `json:"settings"`
}

func (m *UpdateUserSettings) EventType() *EventType {
	e := EventTypeUserUpdateSettings
	return &e
}

func (m *UpdateUserSettings) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UpdateUserSettings) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
