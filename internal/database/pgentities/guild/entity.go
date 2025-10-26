package guild

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type Guild interface {
	GetGuildById(ctx context.Context, id int64) (model.Guild, error)
	CreateGuild(ctx context.Context, id int64, name string, ownerId, permissions int64) error
	DeleteGuild(ctx context.Context, id int64) error
	SetGuildIcon(ctx context.Context, id, icon int64) error
	SetGuildPublic(ctx context.Context, id int64, public bool) error
	ChangeGuildOwner(ctx context.Context, id, ownerId int64) error
	GetGuildsList(ctx context.Context, ids []int64) ([]model.Guild, error)
	SetGuildPermissions(ctx context.Context, id int64, permissions int64) error
	UpdateGuild(ctx context.Context, id int64, name *string, icon *int64, public *bool, permissions *int64) error
	SetSystemMessagesChannel(ctx context.Context, id int64, channelId *int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) Guild {
	return &Entity{c: c}
}
