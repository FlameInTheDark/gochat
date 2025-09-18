package member

import (
	"context"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type Member interface {
	AddMember(ctx context.Context, userID, guildID int64) error
	RemoveMember(ctx context.Context, userID, guildID int64) error
	GetMember(ctx context.Context, userId, guildId int64) (model.Member, error)
	GetMembersList(ctx context.Context, guildId int64, ids []int64) ([]model.Member, error)
	GetGuildMembers(ctx context.Context, guildId int64) ([]model.Member, error)
	IsGuildMember(ctx context.Context, guildId, userId int64) (bool, error)
	GetUserGuilds(ctx context.Context, userId int64) ([]model.UserGuild, error)
	SetTimeout(ctx context.Context, userId, guildId int64, timeout *time.Time) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) Member {
	return &Entity{c: c}
}
