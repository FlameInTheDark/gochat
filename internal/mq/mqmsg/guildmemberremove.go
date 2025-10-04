package mqmsg

import (
	"encoding/json"
)

type RemoveGuildMember struct {
	GuildId int64 `json:"guild_id"`
	UserId  int64 `json:"user_id"`
}

func (m *RemoveGuildMember) EventType() *EventType {
	e := EventTypeGuildMemberRemove
	return &e
}

func (m *RemoveGuildMember) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *RemoveGuildMember) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
