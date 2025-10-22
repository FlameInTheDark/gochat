package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type CreateChannel struct {
	GuildId *int64      `json:"guild_id"`
	Channel dto.Channel `json:"channel"`
}

func (m *CreateChannel) EventType() *EventType {
	e := EventTypeChannelCreate
	return &e
}

func (m *CreateChannel) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *CreateChannel) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
