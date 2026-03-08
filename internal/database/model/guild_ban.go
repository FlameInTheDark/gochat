package model

type GuildBan struct {
	GuildId int64
	UserId  int64
	Reason  *string
}
