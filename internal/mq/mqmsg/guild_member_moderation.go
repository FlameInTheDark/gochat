package mqmsg

import "encoding/json"

type GuildMemberModerationAction string

const (
	GuildMemberModerationKick  GuildMemberModerationAction = "kick"
	GuildMemberModerationBan   GuildMemberModerationAction = "ban"
	GuildMemberModerationUnban GuildMemberModerationAction = "unban"
)

type GuildMemberModeration struct {
	GuildId int64                       `json:"guild_id"`
	UserId  int64                       `json:"user_id"`
	ActorId int64                       `json:"actor_id"`
	Action  GuildMemberModerationAction `json:"action"`
	Reason  *string                     `json:"reason,omitempty"`
}

func (m *GuildMemberModeration) EventType() *EventType {
	e := EventTypeGuildMemberModeration
	return &e
}

func (m *GuildMemberModeration) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *GuildMemberModeration) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
