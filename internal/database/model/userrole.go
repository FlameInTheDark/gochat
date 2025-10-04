package model

import "github.com/lib/pq"

type UserRole struct {
	GuildId int64 `db:"guild_id"`
	UserId  int64 `db:"user_id"`
	RoleId  int64 `db:"role_id"`
}

type UserRoles struct {
	UserId int64         `db:"user_id"`
	Roles  pq.Int64Array `db:"roles"`
}
