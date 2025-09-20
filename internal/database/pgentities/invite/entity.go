package invite

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type Invite interface {
	// CreateInvite inserts both guild_invites and guild_invite_codes in one transaction
	CreateInvite(ctx context.Context, code string, inviteID, guildID, authorID int64, expiresAt int64) (model.GuildInvite, error)
	// GetGuildInvites returns active invites for a guild joined with codes
	GetGuildInvites(ctx context.Context, guildID int64) ([]model.GuildInvite, error)
	// DeleteInviteByCode deletes invite by guild and code (removes mapping too)
	DeleteInviteByCode(ctx context.Context, guildID int64, code string) error
	// DeleteInviteByID deletes invite by guild and invite_id
	DeleteInviteByID(ctx context.Context, guildID, inviteID int64) error
	// FetchInvite uses DB function fetch_guild_invite to resolve code and return a valid invite (or no rows)
	FetchInvite(ctx context.Context, code string) (model.GuildInvite, error)
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) Invite {
	return &Entity{c: c}
}
