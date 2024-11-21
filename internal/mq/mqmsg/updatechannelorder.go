package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type UpdateChannelList struct {
	GuildId  *int64             `json:"guild_id"`
	Channels []dto.ChannelOrder `json:"channels"`
}

func (m *UpdateChannelList) EventType() *EventType {
	e := EventTypeChannelOrderUpdate
	return &e
}

func (m *UpdateChannelList) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UpdateChannelList) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
