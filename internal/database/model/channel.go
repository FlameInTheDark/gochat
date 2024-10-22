package model

import "time"

type Channel struct {
	Id          int64
	Name        string
	Type        string
	ParentID    int64
	GuildId     int64
	Permissions int64
	CreatedAt   time.Time
}
