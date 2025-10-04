package usersettings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

func (e *Entity) GetUserSettings(ctx context.Context, userId, version int64) (model.UserSettings, error) {
	var settings model.UserSettings
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("user_settings").
		Where(
			squirrel.And{
				squirrel.Eq{"user_id": userId},
				squirrel.Gt{"version": version},
			},
		)
	raw, args, err := q.ToSql()
	if err != nil {
		return settings, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.GetContext(ctx, &settings, raw, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return settings, nil
		}
		return settings, fmt.Errorf("unable to get user settings: %w", err)
	}
	return settings, nil
}

func (e *Entity) SetUserSettings(ctx context.Context, userId int64, settings model.UserSettingsData) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	q := squirrel.Insert("user_settings").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "settings", "version").
		Values(userId, string(data), 1).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET settings = EXCLUDED.settings, version = user_settings.version + 1")
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to set user settings: %w", err)
	}
	return nil
}
