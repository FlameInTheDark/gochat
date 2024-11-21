package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type DeleteChannel struct {
	GuildId     *int64 `json:"guild_id"`
	ChannelType model.ChannelType
	ChannelId   int64
}

func (m *DeleteChannel) EventType() *EventType {
	e := EventTypeChannelDelete
	return &e
}

func (m *DeleteChannel) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *DeleteChannel) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
