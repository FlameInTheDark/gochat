package mqmsg

import (
	"encoding/json"
)

type AddGuildMemberRole struct {
	GuildId int64 `json:"guild_id"`
	RoleId  int64 `json:"role_id"`
	UserId  int64 `json:"user_id"`
}

func (m *AddGuildMemberRole) EventType() *EventType {
	e := EventTypeGuildMemberAddRole
	return &e
}

func (m *AddGuildMemberRole) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *AddGuildMemberRole) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
