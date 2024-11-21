package model

type ChannelUserPermission struct {
	ChannelId int64
	UserId    int64
	Accept    int64
	Deny      int64
}
