package emoji

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

type Emoji interface {
	PruneExpired(ctx context.Context, guildID int64) error
	CountActiveGuildEmojis(ctx context.Context, guildID int64) (int64, error)
	CreatePlaceholder(ctx context.Context, emoji model.GuildEmoji) error
	ReusePendingPlaceholder(ctx context.Context, emoji model.GuildEmoji) (model.GuildEmoji, error)
	GetGuildEmoji(ctx context.Context, guildID, emojiID int64) (model.GuildEmoji, error)
	GetEmojiLookup(ctx context.Context, emojiID int64) (model.EmojiLookup, error)
	ListReadyGuildEmojis(ctx context.Context, guildID int64) ([]model.GuildEmoji, error)
	ListReadyGuildEmojisByGuilds(ctx context.Context, guildIDs []int64) ([]model.GuildEmoji, error)
	MarkReady(ctx context.Context, guildID, emojiID int64, animated bool, actualFileSize int64, width, height int64) (model.GuildEmoji, error)
	Rename(ctx context.Context, guildID, emojiID int64, name, normalized string) (model.GuildEmoji, error)
	Delete(ctx context.Context, guildID, emojiID int64) (model.GuildEmoji, error)
	DeleteGuildEmojis(ctx context.Context, guildID int64) ([]model.GuildEmoji, error)
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) Emoji {
	return &Entity{c: c}
}
