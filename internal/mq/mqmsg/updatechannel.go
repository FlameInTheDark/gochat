package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type UpdateChannel struct {
	GuildId *int64      `json:"guild_id"`
	Channel dto.Channel `json:"channel"`
}

func (m *UpdateChannel) EventType() *EventType {
	e := EventTypeChannelUpdate
	return &e
}

func (m *UpdateChannel) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UpdateChannel) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
