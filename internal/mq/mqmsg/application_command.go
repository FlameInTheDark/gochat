package mqmsg

import (
	"encoding/json"

	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
	"github.com/FlameInTheDark/gochat/internal/dto"
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

type ApplicationCommandInteractionStatus struct {
	InteractionID int64                           `json:"interaction_id"`
	ApplicationID int64                           `json:"application_id"`
	CommandID     int64                           `json:"command_id"`
	CommandName   string                          `json:"command_name,omitempty"`
	ChannelID     int64                           `json:"channel_id"`
	GuildID       *int64                          `json:"guild_id,omitempty"`
	UserID        int64                           `json:"user_id"`
	State         string                          `json:"state"`
	Response      *appcmd.InteractionResponseData `json:"response,omitempty"`
	Message       *dto.Message                    `json:"message,omitempty"`
}

func (m *ApplicationCommandInteractionStatus) EventType() *EventType {
	t := EventTypeApplicationCommandInteractionStatus
	return &t
}

func (m *ApplicationCommandInteractionStatus) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *ApplicationCommandInteractionStatus) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
