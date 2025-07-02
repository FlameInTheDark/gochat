package model

import "time"

type Member struct {
	UserId   int64     `db:"user_id"`
	GuildId  int64     `db:"guild_id"`
	Username *string   `db:"username"`
	Avatar   *int64    `db:"avatar"`
	JoinAt   time.Time `db:"join_at"`
	Timeout  time.Time `db:"timeout"`
}

type UserGuild struct {
	UserId  int64 `db:"user_id"`
	GuildId int64 `db:"guild_id"`
}
