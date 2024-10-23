package channel

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	getChannel            = `SELECT id, name, type, parent_it, permissions, created_at FROM gochat.channels WHERE id = ?`
	getChannelThreads     = `SELECT id, name, type, parent_it, permissions, created_at FROM gochat.channels WHERE type = ? AND parent_it = ?`
	createChannel         = `INSERT INTO gochat.channels (id, name, type, parent_id, permissions, created_at) VALUES (?, ?, ?, ?, ?, toTimestamp(now()))`
	deleteChannel         = `DELETE FROM gochat.channels WHERE id = ?`
	renameChannel         = `UPDATE gochat.channels SET name = ? WHERE id = ?`
	setChannelPermissions = `UPDATE gochat.channels SET permissions = ? WHERE id = ?`
)

func (e *Entity) GetChannel(ctx context.Context, id int64) (model.Channel, error) {
	var ch model.Channel
	err := e.c.Session().
		Query(getChannel).
		WithContext(ctx).
		Bind(id).
		Scan(&ch.Id, &ch.Name, &ch.Type, &ch.ParentID, &ch.Permissions, &ch.CreatedAt)
	if err != nil {
		return model.Channel{}, fmt.Errorf("unable to get channel: %w", err)
	}
	return ch, nil
}

func (e *Entity) GetChannelThreads(ctx context.Context, channelId int64) ([]model.Channel, error) {
	var channels []model.Channel
	iter := e.c.Session().
		Query(getChannelThreads).
		WithContext(ctx).
		Bind(model.ChannelTypeThread, channelId).
		Iter()
	var ch model.Channel
	for iter.Scan(&ch.Id, &ch.Name, &ch.Type, &ch.ParentID, &ch.Permissions, &ch.CreatedAt) {
		channels = append(channels, ch)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get channel threads: %w", err)
	}
	return channels, nil
}

func (e *Entity) CreateChannel(ctx context.Context, id int64, name, channelType string, parent int64, permissions int) error {
	err := e.c.Session().
		Query(createChannel).
		WithContext(ctx).
		Bind(id, name, channelType, parent, permissions).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create channel: %w", err)
	}
	return nil
}

func (e *Entity) DeleteChannel(ctx context.Context, id int64) error {
	err := e.c.Session().
		Query(deleteChannel).
		WithContext(ctx).
		Bind(id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to delete channel: %w", err)
	}
	return nil
}

func (e *Entity) RenameChannel(ctx context.Context, id int64, newName string) error {
	err := e.c.Session().
		Query(renameChannel).
		WithContext(ctx).
		Bind(id, newName).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to rename channel: %w", err)
	}
	return nil
}

func (e *Entity) SetChannelPermissions(ctx context.Context, id int64, permissions int) error {
	err := e.c.Session().
		Query(setChannelPermissions).
		WithContext(ctx).
		Bind(id, permissions).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set channel permissions: %w", err)
	}
	return nil
}
