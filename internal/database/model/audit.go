package model

import "time"

type Audit struct {
	GuildId   int64
	CreatedAt time.Time
	Changes   map[string]string
}
