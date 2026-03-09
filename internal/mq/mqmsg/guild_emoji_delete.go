package mqmsg

import "encoding/json"

type DeleteGuildEmoji struct {
	GuildId int64 `json:"guild_id,string"`
	EmojiId int64 `json:"emoji_id,string"`
}

func (m *DeleteGuildEmoji) EventType() *EventType {
	e := EventTypeGuildEmojiDelete
	return &e
}

func (m *DeleteGuildEmoji) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *DeleteGuildEmoji) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
