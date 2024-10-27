package messages

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type CreateMessage struct {
	GuildId int64         `json:"guild_id"`
	Message model.Message `json:"message"`
}

func (m CreateMessage) Type() EventType {
	return EventTypeMessageCreate
}

func (m CreateMessage) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m CreateMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
