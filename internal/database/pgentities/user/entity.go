package user

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type User interface {
	ModifyUser(ctx context.Context, userId int64, name *string, avatar *int64) error
	GetUserById(ctx context.Context, id int64) (model.User, error)
	GetUsersList(ctx context.Context, ids []int64) ([]model.User, error)
	CreateUser(ctx context.Context, id int64, name string) error
	SetUserAvatar(ctx context.Context, id, attachmentId int64) error
	SetUsername(ctx context.Context, id, name string) error
	SetUserBlocked(ctx context.Context, id int64, blocked bool) error
	SetUploadLimit(ctx context.Context, id int64, uploadLimit int64) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) User {
	return &Entity{c: c}
}
