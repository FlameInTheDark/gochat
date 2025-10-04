package model

import "github.com/lib/pq"

type ChannelRolesPermission struct {
	ChannelId int64 `db:"channel_id"`
	RoleId    int64 `db:"role_id"`
	Accept    int64 `db:"accept"`
	Deny      int64 `db:"deny"`
}

type ChannelRoles struct {
	ChannelId int64         `db:"channel_id"`
	Roles     pq.Int64Array `db:"roles"`
}
