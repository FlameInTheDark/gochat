package bot

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type Bot interface {
	CreateBot(ctx context.Context, bot model.Bot) error
	GetBot(ctx context.Context, botUserID int64) (model.Bot, error)
	GetBotsByIDs(ctx context.Context, botUserIDs []int64) ([]model.Bot, error)
	GetBotForOwner(ctx context.Context, ownerUserID, botUserID int64) (model.Bot, error)
	ListOwnerBots(ctx context.Context, ownerUserID int64) ([]model.Bot, error)
	UpdateBot(ctx context.Context, botUserID int64, description *string, public *bool, defaultPermissions *int64, disabled *bool) error
	DeleteBot(ctx context.Context, botUserID int64) error
	SearchPublicBots(ctx context.Context, query string, limit, offset uint64) ([]model.Bot, error)
	ListPublicEnabledBotIDs(ctx context.Context, limit uint64) ([]int64, error)
	SetBotTags(ctx context.Context, botUserID int64, tags []string) error
	GetTagsByBots(ctx context.Context, botUserIDs []int64) (map[int64][]string, error)
	GetInstallCounts(ctx context.Context, botUserIDs []int64) (map[int64]int64, error)

	CreateToken(ctx context.Context, token model.BotToken) error
	GetTokenByHash(ctx context.Context, tokenHash string) (model.BotToken, error)
	ListTokens(ctx context.Context, botUserID int64) ([]model.BotToken, error)
	RevokeToken(ctx context.Context, botUserID, tokenID int64) error
	TouchToken(ctx context.Context, tokenID int64) error

	CreateGrant(ctx context.Context, grant model.BotInstallGrant) error
	GetGrantByHash(ctx context.Context, tokenHash string) (model.BotInstallGrant, error)
	ListGrants(ctx context.Context, botUserID int64) ([]model.BotInstallGrant, error)
	RevokeGrant(ctx context.Context, botUserID, grantID int64) error
	UseGrant(ctx context.Context, grantID int64) error

	UpsertBotGuild(ctx context.Context, guild model.BotGuild) error
	GetBotGuild(ctx context.Context, botUserID, guildID int64) (model.BotGuild, error)
	ListBotGuilds(ctx context.Context, botUserID int64) ([]model.BotGuild, error)
	ListGuildBots(ctx context.Context, guildID int64) ([]model.BotGuild, error)
	DeleteBotGuild(ctx context.Context, botUserID, guildID int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) Bot {
	return &Entity{c: c}
}
