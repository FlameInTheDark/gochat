package model

type ChannelUserPermission struct {
	ChannelId int64 `db:"channel_id"`
	UserId    int64 `db:"user_id"`
	Accept    int64 `db:"accept"`
	Deny      int64 `db:"deny"`
}
