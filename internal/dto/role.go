package dto

type Role struct {
	Id          int64  `json:"id"`
	GuildId     int64  `json:"guild_id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Permissions int64  `json:"permissions"`
}
