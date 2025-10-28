package model

type GuildChannel struct {
	GuildId   int64 `db:"guild_id"`
	ChannelId int64 `db:"channel_id"`
	Position  int   `db:"position"`
}

type GuildChannelUpdatePosition struct {
	GuildId   int64
	ChannelId int64
	Position  int
}
