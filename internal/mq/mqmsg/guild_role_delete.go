package mqmsg

import (
	"encoding/json"
)

type DeleteGuildRole struct {
	GuildId int64 `json:"guild_id"`
	RoleId  int64 `json:"role_id"`
}

func (m *DeleteGuildRole) EventType() *EventType {
	e := EventTypeGuildRoleDelete
	return &e
}

func (m *DeleteGuildRole) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *DeleteGuildRole) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
