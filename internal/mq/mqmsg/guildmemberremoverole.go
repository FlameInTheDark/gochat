package mqmsg

import (
	"encoding/json"
)

type RemoveGuildMemberRole struct {
	GuildId int64 `json:"guild_id"`
	RoleId  int64 `json:"role_id"`
	UserId  int64 `json:"user_id"`
}

func (m *RemoveGuildMemberRole) EventType() *EventType {
	e := EventTypeGuildMemberRemoveRole
	return &e
}

func (m *RemoveGuildMemberRole) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *RemoveGuildMemberRole) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
