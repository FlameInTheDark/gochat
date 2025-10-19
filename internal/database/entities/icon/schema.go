package icon

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	createIcon      = `INSERT INTO gochat.icons (id, guild_id, done, filesize) VALUES (?, ?, false, ?) USING TTL ?`
	doneIcon        = `UPDATE gochat.icons USING TTL 0 SET done = true, content_type = ?, url = ?, height = ?, width = ?, filesize = ? WHERE guild_id = ? AND id = ?`
	removeIcon      = `DELETE FROM gochat.icons WHERE guild_id = ? AND id = ?`
	getIcon         = `SELECT id, guild_id, url, content_type, width, height, filesize, done FROM gochat.icons WHERE guild_id = ? AND id = ?`
	getIconsByGuild = `SELECT id, guild_id, url, content_type, width, height, filesize, done FROM gochat.icons WHERE guild_id = ?`
)

func (e *Entity) CreateIcon(ctx context.Context, id, guildId, ttlSeconds, fileSize int64) error {
	err := e.c.Session().
		Query(createIcon).
		WithContext(ctx).
		Bind(id, guildId, fileSize, ttlSeconds).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create icon: %w", err)
	}
	return nil
}

func (e *Entity) DoneIcon(ctx context.Context, id, guildId int64, contentType, url *string, height, width, fileSize *int64) error {
	err := e.c.Session().
		Query(doneIcon).
		WithContext(ctx).
		Bind(contentType, url, height, width, fileSize, guildId, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to finalize icon: %w", err)
	}
	return nil
}

func (e *Entity) RemoveIcon(ctx context.Context, id, guildId int64) error {
	err := e.c.Session().
		Query(removeIcon).
		WithContext(ctx).
		Bind(guildId, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove icon: %w", err)
	}
	return nil
}

func (e *Entity) GetIcon(ctx context.Context, id, guildId int64) (model.Icon, error) {
	var i model.Icon
	err := e.c.Session().
		Query(getIcon).
		WithContext(ctx).
		Bind(guildId, id).
		Scan(&i.Id, &i.GuildId, &i.URL, &i.ContentType, &i.Width, &i.Height, &i.FileSize, &i.Done)
	if err != nil {
		return i, fmt.Errorf("unable to get icon: %w", err)
	}
	return i, nil
}

func (e *Entity) GetIconsByGuildId(ctx context.Context, guildId int64) ([]model.Icon, error) {
	var icons []model.Icon
	iter := e.c.Session().
		Query(getIconsByGuild).
		WithContext(ctx).
		Bind(guildId).
		Iter()
	var i model.Icon
	for iter.Scan(&i.Id, &i.GuildId, &i.URL, &i.ContentType, &i.Width, &i.Height, &i.FileSize, &i.Done) {
		icons = append(icons, i)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get icons: %w", err)
	}
	return icons, nil
}
