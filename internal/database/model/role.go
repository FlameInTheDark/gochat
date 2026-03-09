package model

type Role struct {
	Id          int64  `db:"id"`
	GuildId     int64  `db:"guild_id"`
	Name        string `db:"name"`
	Color       int    `db:"color"`
	Permissions int64  `db:"permissions"`
	Position    int    `db:"position"`
}

type RoleUpdatePosition struct {
	GuildId  int64
	RoleId   int64
	Position int
}
