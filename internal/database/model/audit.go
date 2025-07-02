package model

import "time"

type Audit struct {
	GuildId   int64             `db:"guild_id"`
	CreatedAt time.Time         `db:"created_at"`
	Changes   map[string]string `db:"changes"`
}
