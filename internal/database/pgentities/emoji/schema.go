package emoji

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	emojiutil "github.com/FlameInTheDark/gochat/internal/emoji"
)

var (
	ErrEmojiNotFound      = errors.New("emoji not found")
	ErrEmojiNameTaken     = errors.New("emoji name already exists")
	ErrEmojiQuotaExceeded = errors.New("emoji quota exceeded")
	ErrEmojiUploadExpired = errors.New("emoji upload expired")
)

func (e *Entity) PruneExpired(ctx context.Context, guildID int64) error {
	expired, err := e.listExpired(ctx, guildID)
	if err != nil {
		return err
	}
	if len(expired) == 0 {
		return nil
	}
	return e.deleteByRows(ctx, expired)
}

func (e *Entity) CountActiveGuildEmojis(ctx context.Context, guildID int64) (int64, error) {
	var count int64
	query := `
		SELECT count(*)
		FROM guild_emojis
		WHERE guild_id = $1
		  AND (done = TRUE OR upload_expires_at > now())`
	if err := e.c.GetContext(ctx, &count, query, guildID); err != nil {
		return 0, fmt.Errorf("count active guild emojis: %w", err)
	}
	return count, nil
}

func (e *Entity) CreatePlaceholder(ctx context.Context, emoji model.GuildEmoji) (err error) {
	tx, err := e.c.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create emoji placeholder tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	now := time.Now().UTC()
	q := squirrel.Insert("guild_emojis").
		PlaceholderFormat(squirrel.Dollar).
		Columns(
			"guild_id", "id", "name", "name_normalized", "creator_id", "done", "animated",
			"declared_file_size", "actual_file_size", "content_type", "width", "height",
			"upload_expires_at", "created_at", "updated_at",
		).
		Values(
			emoji.GuildId, emoji.Id, emoji.Name, emoji.NameNormalized, emoji.CreatorId, false, false,
			emoji.DeclaredFileSize, nil, nil, nil, nil,
			emoji.UploadExpiresAt, now, now,
		)
	sqlText, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("build create guild emoji placeholder query: %w", err)
	}
	if _, err = tx.ExecContext(ctx, sqlText, args...); err != nil {
		return mapSQLError("insert guild emoji", err)
	}

	q = squirrel.Insert("emoji_lookup").
		PlaceholderFormat(squirrel.Dollar).
		Columns("id", "guild_id", "name", "done", "animated", "width", "height", "created_at", "updated_at").
		Values(emoji.Id, emoji.GuildId, emoji.Name, false, false, nil, nil, now, now)
	sqlText, args, err = q.ToSql()
	if err != nil {
		return fmt.Errorf("build create emoji lookup query: %w", err)
	}
	if _, err = tx.ExecContext(ctx, sqlText, args...); err != nil {
		return fmt.Errorf("insert emoji lookup: %w", err)
	}
	return nil
}

func (e *Entity) GetGuildEmoji(ctx context.Context, guildID, emojiID int64) (model.GuildEmoji, error) {
	var emoji model.GuildEmoji
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("guild_emojis").
		Where(squirrel.Eq{"guild_id": guildID, "id": emojiID})
	sqlText, args, err := q.ToSql()
	if err != nil {
		return emoji, fmt.Errorf("build get guild emoji query: %w", err)
	}
	if err = e.c.GetContext(ctx, &emoji, sqlText, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emoji, ErrEmojiNotFound
		}
		return emoji, fmt.Errorf("get guild emoji: %w", err)
	}
	return emoji, nil
}

func (e *Entity) GetEmojiLookup(ctx context.Context, emojiID int64) (model.EmojiLookup, error) {
	var emoji model.EmojiLookup
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("emoji_lookup").
		Where(squirrel.Eq{"id": emojiID})
	sqlText, args, err := q.ToSql()
	if err != nil {
		return emoji, fmt.Errorf("build get emoji lookup query: %w", err)
	}
	if err = e.c.GetContext(ctx, &emoji, sqlText, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emoji, ErrEmojiNotFound
		}
		return emoji, fmt.Errorf("get emoji lookup: %w", err)
	}
	return emoji, nil
}

func (e *Entity) ListReadyGuildEmojis(ctx context.Context, guildID int64) ([]model.GuildEmoji, error) {
	return e.listReady(ctx, squirrel.Eq{"guild_id": guildID})
}

func (e *Entity) ListReadyGuildEmojisByGuilds(ctx context.Context, guildIDs []int64) ([]model.GuildEmoji, error) {
	if len(guildIDs) == 0 {
		return []model.GuildEmoji{}, nil
	}
	return e.listReady(ctx, squirrel.Eq{"guild_id": guildIDs})
}

func (e *Entity) MarkReady(ctx context.Context, guildID, emojiID int64, animated bool, actualFileSize int64, width, height int64) (_ model.GuildEmoji, err error) {
	tx, err := e.c.BeginTxx(ctx, nil)
	if err != nil {
		return model.GuildEmoji{}, fmt.Errorf("begin mark ready tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	emoji, err := getGuildEmojiForUpdate(ctx, tx, guildID, emojiID)
	if err != nil {
		return model.GuildEmoji{}, err
	}
	if emoji.Done {
		return emoji, nil
	}
	if time.Now().UTC().After(emoji.UploadExpiresAt) {
		return model.GuildEmoji{}, ErrEmojiUploadExpired
	}

	count, err := countReadyByAnimated(ctx, tx, guildID, animated, emojiID)
	if err != nil {
		return model.GuildEmoji{}, err
	}
	limit := emojiutil.MaxStaticPerGuild
	if animated {
		limit = emojiutil.MaxAnimatedPerGuild
	}
	if count >= int64(limit) {
		return model.GuildEmoji{}, ErrEmojiQuotaExceeded
	}

	now := time.Now().UTC()
	update := `
		UPDATE guild_emojis
		SET done = TRUE,
		    animated = $3,
		    actual_file_size = $4,
		    content_type = 'image/webp',
		    width = $5,
		    height = $6,
		    updated_at = $7
		WHERE guild_id = $1 AND id = $2`
	if _, err = tx.ExecContext(ctx, update, guildID, emojiID, animated, actualFileSize, width, height, now); err != nil {
		return model.GuildEmoji{}, fmt.Errorf("update guild emoji ready: %w", err)
	}

	updateLookup := `
		UPDATE emoji_lookup
		SET done = TRUE,
		    animated = $2,
		    width = $3,
		    height = $4,
		    updated_at = $5
		WHERE id = $1`
	if _, err = tx.ExecContext(ctx, updateLookup, emojiID, animated, width, height, now); err != nil {
		return model.GuildEmoji{}, fmt.Errorf("update emoji lookup ready: %w", err)
	}

	emoji.Done = true
	emoji.Animated = animated
	emoji.ActualFileSize = &actualFileSize
	contentType := "image/webp"
	emoji.ContentType = &contentType
	emoji.Width = &width
	emoji.Height = &height
	emoji.UpdatedAt = now
	return emoji, nil
}

func (e *Entity) Rename(ctx context.Context, guildID, emojiID int64, name, normalized string) (_ model.GuildEmoji, err error) {
	tx, err := e.c.BeginTxx(ctx, nil)
	if err != nil {
		return model.GuildEmoji{}, fmt.Errorf("begin rename emoji tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	emoji, err := getGuildEmojiForUpdate(ctx, tx, guildID, emojiID)
	if err != nil {
		return model.GuildEmoji{}, err
	}

	now := time.Now().UTC()
	update := squirrel.Update("guild_emojis").
		PlaceholderFormat(squirrel.Dollar).
		Set("name", name).
		Set("name_normalized", normalized).
		Set("updated_at", now).
		Where(squirrel.Eq{"guild_id": guildID, "id": emojiID})
	sqlText, args, err := update.ToSql()
	if err != nil {
		return model.GuildEmoji{}, fmt.Errorf("build rename guild emoji query: %w", err)
	}
	if _, err = tx.ExecContext(ctx, sqlText, args...); err != nil {
		return model.GuildEmoji{}, mapSQLError("rename guild emoji", err)
	}

	update = squirrel.Update("emoji_lookup").
		PlaceholderFormat(squirrel.Dollar).
		Set("name", name).
		Set("updated_at", now).
		Where(squirrel.Eq{"id": emojiID})
	sqlText, args, err = update.ToSql()
	if err != nil {
		return model.GuildEmoji{}, fmt.Errorf("build rename emoji lookup query: %w", err)
	}
	if _, err = tx.ExecContext(ctx, sqlText, args...); err != nil {
		return model.GuildEmoji{}, fmt.Errorf("rename emoji lookup: %w", err)
	}

	emoji.Name = name
	emoji.NameNormalized = normalized
	emoji.UpdatedAt = now
	return emoji, nil
}

func (e *Entity) Delete(ctx context.Context, guildID, emojiID int64) (model.GuildEmoji, error) {
	emoji, err := e.GetGuildEmoji(ctx, guildID, emojiID)
	if err != nil {
		return model.GuildEmoji{}, err
	}
	if err = e.deleteByRows(ctx, []model.GuildEmoji{emoji}); err != nil {
		return model.GuildEmoji{}, err
	}
	return emoji, nil
}

func (e *Entity) DeleteGuildEmojis(ctx context.Context, guildID int64) ([]model.GuildEmoji, error) {
	rows, err := e.listAllByGuild(ctx, guildID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []model.GuildEmoji{}, nil
	}
	if err = e.deleteByRows(ctx, rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (e *Entity) listReady(ctx context.Context, predicate squirrel.Sqlizer) ([]model.GuildEmoji, error) {
	rows := make([]model.GuildEmoji, 0)
	q := squirrel.Select(
		"guild_id", "id", "name", "name_normalized", "creator_id", "done", "animated",
		"declared_file_size", "actual_file_size", "content_type", "width", "height",
		"upload_expires_at", "created_at", "updated_at",
	).
		PlaceholderFormat(squirrel.Dollar).
		From("guild_emojis").
		Where(predicate).
		Where(squirrel.Eq{"done": true}).
		OrderBy("name ASC")
	sqlText, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list ready guild emojis query: %w", err)
	}
	if err = e.c.SelectContext(ctx, &rows, sqlText, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("list ready guild emojis: %w", err)
	}
	if rows == nil {
		return []model.GuildEmoji{}, nil
	}
	return rows, nil
}

func (e *Entity) listAllByGuild(ctx context.Context, guildID int64) ([]model.GuildEmoji, error) {
	rows := make([]model.GuildEmoji, 0)
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("guild_emojis").
		Where(squirrel.Eq{"guild_id": guildID})
	sqlText, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list all guild emojis query: %w", err)
	}
	if err = e.c.SelectContext(ctx, &rows, sqlText, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("list all guild emojis: %w", err)
	}
	if rows == nil {
		return []model.GuildEmoji{}, nil
	}
	return rows, nil
}

func (e *Entity) listExpired(ctx context.Context, guildID int64) ([]model.GuildEmoji, error) {
	rows := make([]model.GuildEmoji, 0)
	q := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("guild_emojis").
		Where(squirrel.Eq{"guild_id": guildID, "done": false}).
		Where(squirrel.LtOrEq{"upload_expires_at": time.Now().UTC()})
	sqlText, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list expired guild emojis query: %w", err)
	}
	if err = e.c.SelectContext(ctx, &rows, sqlText, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("list expired guild emojis: %w", err)
	}
	if rows == nil {
		return []model.GuildEmoji{}, nil
	}
	return rows, nil
}

func (e *Entity) deleteByRows(ctx context.Context, rows []model.GuildEmoji) (err error) {
	tx, err := e.c.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete emoji tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.Id)
	}

	delGuild := squirrel.Delete("guild_emojis").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"guild_id": rows[0].GuildId, "id": ids})
	sqlText, args, err := delGuild.ToSql()
	if err != nil {
		return fmt.Errorf("build delete guild emojis query: %w", err)
	}
	if _, err = tx.ExecContext(ctx, sqlText, args...); err != nil {
		return fmt.Errorf("delete guild emojis: %w", err)
	}

	delLookup := squirrel.Delete("emoji_lookup").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": ids})
	sqlText, args, err = delLookup.ToSql()
	if err != nil {
		return fmt.Errorf("build delete emoji lookup query: %w", err)
	}
	if _, err = tx.ExecContext(ctx, sqlText, args...); err != nil {
		return fmt.Errorf("delete emoji lookup: %w", err)
	}
	return nil
}

func getGuildEmojiForUpdate(ctx context.Context, tx *sqlx.Tx, guildID, emojiID int64) (model.GuildEmoji, error) {
	var emoji model.GuildEmoji
	query := `
		SELECT *
		FROM guild_emojis
		WHERE guild_id = $1 AND id = $2
		FOR UPDATE`
	if err := tx.GetContext(ctx, &emoji, query, guildID, emojiID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emoji, ErrEmojiNotFound
		}
		return emoji, fmt.Errorf("get guild emoji for update: %w", err)
	}
	return emoji, nil
}

func countReadyByAnimated(ctx context.Context, tx *sqlx.Tx, guildID int64, animated bool, excludeID int64) (int64, error) {
	var count int64
	query := `
		SELECT count(*)
		FROM guild_emojis
		WHERE guild_id = $1
		  AND done = TRUE
		  AND animated = $2
		  AND id <> $3`
	if err := tx.GetContext(ctx, &count, query, guildID, animated, excludeID); err != nil {
		return 0, fmt.Errorf("count ready emojis by animation: %w", err)
	}
	return count, nil
}

func mapSQLError(operation string, err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrEmojiNameTaken
	}
	return fmt.Errorf("%s: %w", operation, err)
}
