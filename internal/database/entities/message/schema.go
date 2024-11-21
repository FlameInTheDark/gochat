package message

import (
	"context"
	"errors"
	"fmt"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/gocql/gocql"
)

const (
	createMessage         = `INSERT INTO gochat.messages (channel_id, bucket, id, user_id, content, attachments) VALUES (?, ?, ?, ?, ?, ?)`
	updateMessage         = `UPDATE gochat.messages SET content = ?, edited_at = toTimestamp(now()) WHERE channel_id = ? AND id = ? AND bucket = ?`
	deleteMessage         = `DELETE FROM gochat.messages WHERE id = ?`
	deleteChannelMessages = `DELETE FROM gochat.messages WHERE channel_id = ?`
	getMessage            = `SELECT id, channel_id, user_id, content, attachments, edited_at FROM gochat.messages WHERE id = ? AND channel_id = ? AND bucket = ?`
	getMessagesBefore     = `SELECT id, channel_id, user_id, content, attachments, edited_at FROM gochat.messages WHERE channel_id = ? AND id <= ? AND bucket = ? ORDER BY id DESC LIMIT ?`
	getMessagesAfter      = `SELECT id, channel_id, user_id, content, attachments, edited_at FROM gochat.messages WHERE channel_id = ? AND id >= ? AND bucket = ? ORDER BY id DESC LIMIT ?`
	getMessagesList       = `SELECT id, channel_id, user_id, content, attachments, edited_at FROM gochat.messages WHERE id IN ?`
)

func (e *Entity) CreateMessage(ctx context.Context, id, channel_id, user_id int64, content string, attachments []int64) error {
	err := e.c.Session().
		Query(createMessage).
		WithContext(ctx).
		Bind(channel_id, idgen.GetBucket(id), id, user_id, content, attachments).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create message: %w", err)
	}
	return nil
}

func (e *Entity) UpdateMessage(ctx context.Context, id, channel_id int64, content string) error {
	err := e.c.Session().
		Query(updateMessage).
		WithContext(ctx).
		Bind(content, channel_id, id, idgen.GetBucket(id)).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to update message: %w", err)
	}
	return nil
}

func (e *Entity) DeleteMessage(ctx context.Context, id int64) error {
	err := e.c.Session().
		Query(deleteMessage).
		WithContext(ctx).
		Bind(id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to delete message: %w", err)
	}
	return nil
}

func (e *Entity) DeleteChannelMessages(ctx context.Context, channel_id int64) error {
	err := e.c.Session().
		Query(deleteChannelMessages).
		WithContext(ctx).
		Bind(channel_id).
		Exec()
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return fmt.Errorf("unable to delete messages: %w", err)
	}
	return nil
}

func (e *Entity) GetMessage(ctx context.Context, id, channelId int64) (model.Message, error) {
	var m model.Message
	err := e.c.Session().
		Query(getMessage).
		WithContext(ctx).
		Bind(id, channelId, idgen.GetBucket(id)).
		Scan(&m.Id, &m.ChannelId, &m.UserId, &m.Content, &m.Attachments, &m.EditedAt)
	if err != nil {
		return m, fmt.Errorf("unable to get message: %w", err)
	}
	return m, nil
}

func (e *Entity) GetMessagesBefore(ctx context.Context, channelId, msgId int64, limit int) ([]model.Message, error) {
	var msgs []model.Message
	if msgId <= channelId {
		return msgs, nil
	}
	lastBucket := idgen.GetBucket(msgId)
	endBucket := idgen.GetBucket(channelId)
	for {
		iter := e.c.Session().
			Query(getMessagesBefore).
			WithContext(ctx).
			Bind(channelId, msgId, lastBucket, limit-len(msgs)).
			Iter()
		var m model.Message
		for iter.Scan(m.Id, &m.ChannelId, &m.UserId, &m.Content, &m.Attachments, &m.EditedAt) {
			msgs = append(msgs, m)
		}
		err := iter.Close()
		if err != nil {
			return nil, fmt.Errorf("unable to get messages before: %w", err)
		}
		if len(msgs) == limit || lastBucket <= endBucket {
			break
		} else {
			lastBucket = lastBucket - 1

		}
	}
	return msgs, nil
}

func (e *Entity) GetMessagesAfter(ctx context.Context, channelId, msgId int64) ([]model.Message, error) {
	var msgs []model.Message
	iter := e.c.Session().
		Query(getMessagesAfter).
		WithContext(ctx).
		Bind(channelId, msgId).
		Iter()
	var m model.Message
	for iter.Scan(&m.ChannelId, &m.UserId, &m.Content, &m.Attachments, &m.EditedAt) {
		msgs = append(msgs, m)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get messages after: %w", err)
	}
	return msgs, nil
}

func (e *Entity) GetMessagesList(ctx context.Context, msgIds []int64) ([]model.Message, error) {
	var msgs []model.Message
	iter := e.c.Session().
		Query(getMessagesList).
		WithContext(ctx).
		Bind(msgIds).
		Iter()
	var m model.Message
	for iter.Scan(&m.ChannelId, &m.UserId, &m.Content, &m.Attachments, &m.EditedAt) {
		msgs = append(msgs, m)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get messages list: %w", err)
	}
	return msgs, nil
}
