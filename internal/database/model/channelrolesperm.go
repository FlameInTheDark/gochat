package model

type ChannelRolesPermission struct {
	ChannelId int64
	RoleId    int64
	Accept    int64
	Deny      int64
}
