package icon

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	createIcon      = `INSERT INTO gochat.icons (id, guild_id, object) VALUES (?, ?, ?)`
	removeIcon      = `DELETE FROM gochat.icons WHERE id = ?`
	getIcon         = `SELECT id, guild_id, object FROM gochat.icons WHERE id = ?`
	getIconByUserId = `SELECT id, guild_id, object FROM gochat.icons WHERE guild_id = ?`
)

func (e *Entity) CreateIcon(ctx context.Context, id, guildId int64, object string) error {
	err := e.c.Session().
		Query(createIcon).
		WithContext(ctx).
		Bind(id, guildId, object).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create icon: %w", err)
	}
	return nil
}

func (e *Entity) RemoveIcon(ctx context.Context, id int64) error {
	err := e.c.Session().
		Query(removeIcon).
		WithContext(ctx).
		Bind(id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove icon: %w", err)
	}
	return nil
}

func (e *Entity) GetIcon(ctx context.Context, id int64) (model.Icon, error) {
	var i model.Icon
	err := e.c.Session().
		Query(getIcon).
		WithContext(ctx).
		Bind(id).
		Exec()
	if err != nil {
		return i, fmt.Errorf("unable to get icon: %w", err)
	}
	return i, nil
}

func (e *Entity) GetIconsByUserId(ctx context.Context, userId int64) ([]model.Icon, error) {
	var icons []model.Icon
	iter := e.c.Session().
		Query(getIconByUserId).
		WithContext(ctx).
		Bind(userId).
		Iter()
	var i model.Icon
	for iter.Scan(&i.Id, &i.GuildId, &i.Object) {
		icons = append(icons, i)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get icon: %w", err)
	}
	return icons, nil
}
