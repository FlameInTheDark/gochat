package mqmsg

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

type UpdateGuildRole struct {
	GuildId int64    `json:"guild_id"`
	Role    dto.Role `json:"role"`
}

func (m *UpdateGuildRole) EventType() *EventType {
	e := EventTypeGuildRoleUpdate
	return &e
}

func (m *UpdateGuildRole) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UpdateGuildRole) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
