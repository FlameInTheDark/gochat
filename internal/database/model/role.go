package model

type Role struct {
	Id          int64  `db:"id"`
	GuildId     int64  `db:"guild_id"`
	Name        string `db:"name"`
	Color       int    `db:"color"`
	Permissions int64  `db:"permissions"`
}
