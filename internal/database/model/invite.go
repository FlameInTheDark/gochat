package model

import "time"

// GuildInvite represents an invite joined with its code and metadata
type GuildInvite struct {
	InviteCode string    `db:"invite_code"`
	InviteId   int64     `db:"invite_id"`
	GuildId    int64     `db:"guild_id"`
	AuthorId   int64     `db:"author_id"`
	CreatedAt  time.Time `db:"created_at"`
	ExpiresAt  time.Time `db:"expires_at"`
}
