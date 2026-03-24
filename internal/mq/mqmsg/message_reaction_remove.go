package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type MessageReactionRemove struct {
	GuildId   *int64              `json:"guild_id"`
	ChannelId int64               `json:"channel_id"`
	MessageId int64               `json:"message_id"`
	Reaction  dto.MessageReaction `json:"reaction"`
}

func (m *MessageReactionRemove) EventType() *EventType {
	e := EventTypeMessageReactionRemove
	return &e
}

func (m *MessageReactionRemove) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *MessageReactionRemove) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
