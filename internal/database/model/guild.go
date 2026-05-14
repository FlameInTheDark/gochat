package model

import "time"

type Guild struct {
	Id             int64     `db:"id"`
	Name           string    `db:"name"`
	Description    string    `db:"description"`
	OwnerId        int64     `db:"owner_id"`
	Icon           *int64    `db:"icon"`
	Public         bool      `db:"public"`
	Permissions    int64     `db:"permissions"`
	CreatedAt      time.Time `db:"created_at"`
	SystemMessages *int64    `db:"system_messages"`
}

type GuildDiscoveryStats struct {
	GuildId      int64 `db:"guild_id"`
	MembersCount int64 `db:"members_count"`
}
