package mqmsg

import (
	"encoding/json"

	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
)

type ApplicationCommandInteractionCreate struct {
	Interaction appcmd.Interaction `json:"interaction"`
}

func (m *ApplicationCommandInteractionCreate) EventType() *EventType {
	t := EventTypeApplicationCommandInteractionCreate
	return &t
}

func (m *ApplicationCommandInteractionCreate) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *ApplicationCommandInteractionCreate) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
