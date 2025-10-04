package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type AddGuildMember struct {
	GuildId int64      `json:"guild_id"`
	UserId  int64      `json:"user_id"`
	Member  dto.Member `json:"member"`
}

func (m *AddGuildMember) EventType() *EventType {
	e := EventTypeGuildMemberAdd
	return &e
}

func (m *AddGuildMember) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *AddGuildMember) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
