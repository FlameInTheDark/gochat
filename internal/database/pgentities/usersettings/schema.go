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

func (e *Entity) SetUserAppearance(ctx context.Context, userId int64, a model.UserSettingsAppearance) error {
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}

	q := squirrel.Insert("user_settings").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "settings", "version").
		Values(userId, squirrel.Expr("jsonb_build_object('appearance', ?::jsonb)", string(data)), 1).
		Suffix(`
			ON CONFLICT (user_id) DO UPDATE
			SET settings = jsonb_set(
					COALESCE(user_settings.settings, '{}'::jsonb),
					'{appearance}',
					EXCLUDED.settings->'appearance',
					true
				),
			    version = user_settings.version + 1`)

	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err := e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to set user appearance: %w", err)
	}
	return nil
}

func (e *Entity) SetUserSelectedGuild(ctx context.Context, userId, guildId int64) error {
	q := squirrel.
		Insert("user_settings").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "settings", "version").
		Values(
			userId,
			squirrel.Expr("jsonb_build_object('selected_guild', to_jsonb(?::bigint))", guildId),
			1,
		).
		Suffix(`
			ON CONFLICT (user_id) DO UPDATE
			SET settings = jsonb_set(
							 COALESCE(user_settings.settings, '{}'::jsonb),
							 '{selected_guild}',
							 to_jsonb(?::bigint),
							 true
						   ),
				version  = user_settings.version + 1`, guildId)

	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to build SQL: %w", err)
	}
	if _, err := e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to set selected_guild: %w", err)
	}
	return nil
}

func (e *Entity) SetUserSelectedChannel(ctx context.Context, userId, guildId, channelId int64) error {
	q := squirrel.
		Insert("user_settings").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "settings", "version").
		Values(
			userId,
			squirrel.Expr(
				"jsonb_build_object('guilds', jsonb_build_array(jsonb_build_object('guild_id', to_jsonb(?::bigint), 'selected_channel', to_jsonb(?::bigint))))",
				guildId, channelId,
			),
			1,
		).
		Suffix(`
			ON CONFLICT (user_id) DO UPDATE
			SET settings = (
			  WITH updated AS (
				SELECT CASE
				  WHEN EXISTS (
					SELECT 1
					FROM jsonb_array_elements(COALESCE(user_settings.settings->'guilds', '[]'::jsonb)) el
					WHERE (el->>'guild_id')::bigint = ?
				  )
				  THEN (
					SELECT jsonb_agg(
							 CASE WHEN (el->>'guild_id')::bigint = ?
								  THEN jsonb_set(el, '{selected_channel}', to_jsonb(?::bigint), true)
								  ELSE el
							 END
						   )
					FROM jsonb_array_elements(COALESCE(user_settings.settings->'guilds', '[]'::jsonb)) el
				  )
				  ELSE COALESCE(user_settings.settings->'guilds', '[]'::jsonb)
					   || jsonb_build_array(jsonb_build_object(
							'guild_id', to_jsonb(?::bigint),
							'selected_channel', to_jsonb(?::bigint)
						  ))
				END AS guilds
			  )
			  SELECT jsonb_set(
					   COALESCE(user_settings.settings, '{}'::jsonb),
					   '{guilds}',
					   updated.guilds,
					   true
					 )
			  FROM updated
			),
			version = user_settings.version + 1`,
			guildId, guildId, channelId, guildId, channelId)

	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to build SQL: %w", err)
	}
	if _, err := e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to set selected_channel: %w", err)
	}
	return nil
}
