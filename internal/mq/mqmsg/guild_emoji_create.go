package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type CreateGuildEmoji struct {
	Emoji dto.GuildEmoji `json:"emoji"`
}

func (m *CreateGuildEmoji) EventType() *EventType {
	e := EventTypeGuildEmojiCreate
	return &e
}

func (m *CreateGuildEmoji) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *CreateGuildEmoji) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
