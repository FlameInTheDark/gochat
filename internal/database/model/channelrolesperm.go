package model

type ChannelRolesPermission struct {
	ChannelId int64 `db:"channel_id"`
	RoleId    int64 `db:"role_id"`
	Accept    int64 `db:"accept"`
	Deny      int64 `db:"deny"`
}
