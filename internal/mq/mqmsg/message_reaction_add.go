package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type MessageReactionAdd struct {
	GuildId   *int64              `json:"guild_id"`
	ChannelId int64               `json:"channel_id"`
	MessageId int64               `json:"message_id"`
	Reaction  dto.MessageReaction `json:"reaction"`
}

func (m *MessageReactionAdd) EventType() *EventType {
	e := EventTypeMessageReactionAdd
	return &e
}

func (m *MessageReactionAdd) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *MessageReactionAdd) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
