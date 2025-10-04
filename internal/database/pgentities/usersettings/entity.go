package usersettings

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type UserSettings interface {
	GetUserSettings(ctx context.Context, userId, version int64) (model.UserSettings, error)
	SetUserSettings(ctx context.Context, userId int64, settings model.UserSettingsData) error
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) UserSettings {
	return &Entity{c: c}
}
