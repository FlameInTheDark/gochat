package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type UpdateGuildEmoji struct {
	Emoji dto.GuildEmoji `json:"emoji"`
}

func (m *UpdateGuildEmoji) EventType() *EventType {
	e := EventTypeGuildEmojiUpdate
	return &e
}

func (m *UpdateGuildEmoji) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UpdateGuildEmoji) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
