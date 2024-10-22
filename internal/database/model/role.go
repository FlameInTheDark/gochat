package model

type Role struct {
	Id          int64
	GuildId     int64
	Name        string
	Color       int
	Permissions int64
}
