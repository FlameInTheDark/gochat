package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type UpdateGuildMember struct {
	GuildId int64      `json:"guild_id"`
	Member  dto.Member `json:"member"`
}

func (m *UpdateGuildMember) EventType() *EventType {
	e := EventTypeGuildMemberUpdate
	return &e
}

func (m *UpdateGuildMember) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UpdateGuildMember) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
