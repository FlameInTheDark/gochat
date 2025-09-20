package dto

import "time"

type GuildInvite struct {
	Id        int64     `json:"id"`
	Code      string    `json:"code"`
	GuildId   int64     `json:"guild_id"`
	AuthorId  int64     `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type InvitePreview struct {
	Id           int64     `json:"id"`
	Code         string    `json:"code"`
	Guild        Guild     `json:"guild"`
	AuthorId     int64     `json:"author_id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	MembersCount int       `json:"members_count"`
}
