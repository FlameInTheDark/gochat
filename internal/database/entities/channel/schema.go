package channel

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocql/gocql"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	getChannel        = `SELECT id, name, type, parent_id, permissions, topic, private, created_at FROM gochat.channels WHERE id = ?`
	getChannelsBulk   = `SELECT id, name, type, parent_id, permissions, topic, private, created_at FROM gochat.channels WHERE id IN ?`
	getChannelThreads = `SELECT id, name, type, parent_id, permissions, topic, private, created_at FROM gochat.channels WHERE type = ? AND parent_id = ?`
	createChannel     = `INSERT INTO gochat.channels (id, name, type, parent_id, permissions, private, created_at) VALUES (?, ?, ?, ?, ?, ?, toTimestamp(now()))`
	deleteChannel     = `DELETE FROM gochat.channels WHERE id = ?`
	renameChannel     = `UPDATE gochat.channels SET name = ? WHERE id = ?`
	setPermissions    = `UPDATE gochat.channels SET permissions = ? WHERE id = ?`
	setPrivate        = `UPDATE gochat.channels SET private = ? WHERE id = ?`
	setTopic          = `UPDATE gochat.channels SET topic = ? WHERE id = ?`
	setParent         = `UPDATE gochat.channels SET parent_id = ? WHERE id = ?`
	setParentBulk     = `UPDATE gochat.channels SET parent_id = ? WHERE id IN ?`
)

func (e *Entity) GetChannel(ctx context.Context, id int64) (model.Channel, error) {
	var ch model.Channel
	err := e.c.Session().
		Query(getChannel).
		WithContext(ctx).
		Bind(id).
		Scan(&ch.Id, &ch.Name, &ch.Type, &ch.ParentID, &ch.Permissions, &ch.Topic, &ch.Private, &ch.CreatedAt)
	if err != nil {
		return model.Channel{}, fmt.Errorf("unable to get channel: %w", err)
	}
	return ch, nil
}

func (e *Entity) GetChannelsBulk(ctx context.Context, ids []int64) ([]model.Channel, error) {
	var chs []model.Channel
	iter := e.c.Session().
		Query(getChannelsBulk).
		WithContext(ctx).
		Bind(ids).
		Iter()
	var ch model.Channel
	for iter.Scan(&ch.Id, &ch.Name, &ch.Type, &ch.ParentID, &ch.Permissions, &ch.Topic, &ch.Private, &ch.CreatedAt) {
		chs = append(chs, ch)
	}
	err := iter.Close()
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, fmt.Errorf("unable to get channels: %w", err)
	}
	return chs, nil
}

func (e *Entity) GetChannelThreads(ctx context.Context, channelId int64) ([]model.Channel, error) {
	var channels []model.Channel
	iter := e.c.Session().
		Query(getChannelThreads).
		WithContext(ctx).
		Bind(model.ChannelTypeThread, channelId).
		Iter()
	var ch model.Channel
	for iter.Scan(&ch.Id, &ch.Name, &ch.Type, &ch.ParentID, &ch.Permissions, &ch.Topic, &ch.Private, &ch.CreatedAt) {
		channels = append(channels, ch)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get channel threads: %w", err)
	}
	return channels, nil
}

func (e *Entity) CreateChannel(ctx context.Context, id int64, name string, channelType model.ChannelType, parent *int64, permissions *int64, private bool) error {
	err := e.c.Session().
		Query(createChannel).
		WithContext(ctx).
		Bind(id, name, channelType, parent, permissions, private).
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
		Bind(newName, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to rename channel: %w", err)
	}
	return nil
}

func (e *Entity) SetChannelPermissions(ctx context.Context, id int64, permissions int) error {
	err := e.c.Session().
		Query(setPermissions).
		WithContext(ctx).
		Bind(permissions, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set channel permissions: %w", err)
	}
	return nil
}

func (e *Entity) SetChannelPrivate(ctx context.Context, id int64, private bool) error {
	err := e.c.Session().
		Query(setPrivate).
		WithContext(ctx).
		Bind(private, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set channel private: %w", err)
	}
	return nil
}

func (e *Entity) SetChannelTopic(ctx context.Context, id int64, topic *string) error {
	err := e.c.Session().
		Query(setTopic).
		WithContext(ctx).
		Bind(topic, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set channel topic: %w", err)
	}
	return nil
}

func (e *Entity) SetChannelParent(ctx context.Context, id int64, parent *int64) error {
	err := e.c.Session().
		Query(setParent).
		WithContext(ctx).
		Bind(parent, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set channel parent: %w", err)
	}
	return nil
}

func (e *Entity) SetChannelParentBulk(ctx context.Context, id []int64, parent *int64) error {
	err := e.c.Session().
		Query(setParentBulk).
		WithContext(ctx).
		Bind(parent, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to set channel parent bulk: %w", err)
	}
	return nil
}
