package guild

import (
	"context"
	"fmt"
	"github.com/scylladb/gocqlx/qb"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	getGuild            = `SELECT id, name, owner_id, icon, public, permissions, created_at FROM gochat.guilds WHERE id = ?;`
	createGuild         = `INSERT INTO gochat.guilds (id, name, owner_id, permissions, created_at) VALUES (?, ?, ?, ?, toTimestamp(now()));`
	deleteGuild         = `DELETE FROM gochat.guilds WHERE id = ?;`
	setIcon             = `UPDATE gochat.guilds SET icon = ? WHERE id = ?;`
	setPublic           = `UPDATE gochat.guilds SET public = ? WHERE id = ?;`
	changeOwner         = `UPDATE gochat.guilds SET owner_id = ? WHERE id = ?;`
	getGuildsList       = `SELECT id, name, owner_id, icon, public, created_at FROM gochat.guilds WHERE id IN ?;`
	setGuildPermissions = `UPDATE gochat.guilds SET permissions = ? WHERE id = ?;`
)

func (e *Entity) GetGuildById(ctx context.Context, id int64) (model.Guild, error) {
	var g model.Guild
	err := e.c.Session().
		Query(getGuild).
		WithContext(ctx).
		Bind(id).
		Scan(&g.Id, &g.Name, &g.OwnerId, &g.Icon, &g.Public, &g.Permissions, &g.CreatedAt)
	if err != nil {
		return g, fmt.Errorf("unable to get guild: %w", err)
	}
	return g, nil
}

func (e *Entity) CreateGuild(ctx context.Context, id int64, name string, ownerId, permissions int64) error {
	err := e.c.Session().
		Query(createGuild).
		WithContext(ctx).
		Bind(id, name, ownerId, permissions).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create guild: %w", err)
	}
	return nil
}

func (e *Entity) DeleteGuild(ctx context.Context, id int64) error {
	err := e.c.Session().
		Query(deleteGuild).
		WithContext(ctx).
		Bind(id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to delete guild: %w", err)
	}
	return nil
}

func (e *Entity) SetGuildIcon(ctx context.Context, id, icon int64) error {
	err := e.c.Session().
		Query(setIcon).
		WithContext(ctx).
		Bind(id, icon).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set icon: %w", err)
	}
	return nil
}

func (e *Entity) SetGuildPublic(ctx context.Context, id int64, public bool) error {
	err := e.c.Session().
		Query(setPublic).
		WithContext(ctx).
		Bind(id, public).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set public: %w", err)
	}
	return nil
}

func (e *Entity) ChangeGuildOwner(ctx context.Context, id, ownerId int64) error {
	err := e.c.Session().
		Query(changeOwner).
		WithContext(ctx).
		Bind(id, ownerId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to change owner id: %w", err)
	}
	return nil
}

func (e *Entity) GetGuildsList(ctx context.Context, ids []int64) ([]model.Guild, error) {
	var gs []model.Guild
	iter := e.c.Session().
		Query(getGuildsList).
		WithContext(ctx).
		Bind(ids).
		Iter()
	var g model.Guild
	for iter.Scan(&g.Id, &g.Name, &g.OwnerId, &g.Icon, &g.Public, &g.CreatedAt) {
		gs = append(gs, g)
	}
	if err := iter.Close(); err != nil {
		return gs, fmt.Errorf("unable to get guild list: %w", err)
	}
	return gs, nil
}

func (e *Entity) SetGuildPermissions(ctx context.Context, id int64, permissions int64) error {
	err := e.c.Session().
		Query(setGuildPermissions).
		WithContext(ctx).
		Bind(permissions, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set guild permissions: %w", err)
	}
	return nil
}

func (e *Entity) UpdateGuild(ctx context.Context, id int64, name *string, icon *int64, public *bool) error {
	var args []interface{}
	b := qb.UpdateBuilder{}
	b.Where(qb.Eq("id"))
	b.Table("gochat.guilds")
	if name != nil {
		b.Set("name")
		args = append(args, *name)
	}
	if icon != nil {
		b.Set("icon")
		args = append(args, *icon)
	}
	if public != nil {
		b.Set("public")
		args = append(args, *public)
	}
	args = append(args, id)
	cql, _ := b.ToCql()

	err := e.c.Session().
		Query(cql).
		WithContext(ctx).
		Bind(args...).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to update guild: %w", err)
	}
	return nil
}
