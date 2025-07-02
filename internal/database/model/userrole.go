package model

type UserRole struct {
	GuildId int64 `db:"guild_id"`
	UserId  int64 `db:"user_id"`
	RoleId  int64 `db:"role_id"`
}
