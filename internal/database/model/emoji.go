package model

import "time"

type GuildEmoji struct {
	GuildId          int64     `db:"guild_id"`
	Id               int64     `db:"id"`
	Name             string    `db:"name"`
	NameNormalized   string    `db:"name_normalized"`
	CreatorId        int64     `db:"creator_id"`
	Done             bool      `db:"done"`
	Animated         bool      `db:"animated"`
	DeclaredFileSize int64     `db:"declared_file_size"`
	ActualFileSize   *int64    `db:"actual_file_size"`
	ContentType      *string   `db:"content_type"`
	Width            *int64    `db:"width"`
	Height           *int64    `db:"height"`
	UploadExpiresAt  time.Time `db:"upload_expires_at"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

type EmojiLookup struct {
	Id        int64     `db:"id"`
	GuildId   int64     `db:"guild_id"`
	Name      string    `db:"name"`
	Done      bool      `db:"done"`
	Animated  bool      `db:"animated"`
	Width     *int64    `db:"width"`
	Height    *int64    `db:"height"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
