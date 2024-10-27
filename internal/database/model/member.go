package model

import "time"

type Member struct {
	UserId   int64
	GuildId  int64
	Username *string
	Avatar   *int64
	JoinAt   time.Time
}

type UserGuild struct {
	UserId  int64
	GuildId int64
}
