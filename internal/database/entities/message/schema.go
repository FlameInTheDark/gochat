package message

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/idgen"
)

const (
	createMessage     = `INSERT INTO gochat.messages (channel_id, bucket, id, user_id, content, attachments) VALUES (?, ?, ?, ?, ?, ?)`
	updateMessage     = `UPDATE gochat.messages SET content = ? AND edited_at = toTimestamp(now()) WHERE id = ?`
	deleteMessage     = `DELETE FROM gochat.messages WHERE id = ?`
	getMessage        = `SELECT id, channel_id, user_id, content, attachments, edited_at FROM gochat.messages WHERE id = ?`
	getLatestMessages = `SELECT id, channel_id, user_id, content, attachments, edited_at FROM gochat.messages WHERE channel_id = ? ORDER BY id DESC LIMIT 10`
	getMessagesBefore = `SELECT id, channel_id, user_id, content, attachments, edited_at FROM gochat.messages WHERE channel_id = ? AND id < ? ORDER BY id DESC LIMIT 10`
	getMessagesAfter  = `SELECT id, channel_id, user_id, content, attachments, edited_at FROM gochat.messages WHERE channel_id = ? AND id > ? ORDER BY id DESC LIMIT 10`
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

func (e *Entity) UpdateMessage(ctx context.Context, id int64, content string) error {
	err := e.c.Session().
		Query(updateMessage).
		WithContext(ctx).
		Bind(id, content).
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

func (e *Entity) GetMessage(ctx context.Context, id int64) (model.Message, error) {
	var m model.Message
	err := e.c.Session().
		Query(getMessage).
		WithContext(ctx).
		Bind(id).
		Scan(&m.Id, &m.ChannelId, &m.UserId, &m.Content, &m.Attachments, &m.EditedAt)
	if err != nil {
		return m, fmt.Errorf("unable to get message: %w", err)
	}
	return m, nil
}

func (e *Entity) GetLatestMessages(ctx context.Context, channelId int64) ([]model.Message, error) {
	var msgs []model.Message
	iter := e.c.Session().
		Query(getLatestMessages).
		WithContext(ctx).
		Bind(channelId).
		Iter()
	var m model.Message
	for iter.Scan(&m.Id, &m.ChannelId, &m.UserId, &m.Content, &m.Attachments, &m.EditedAt) {
		msgs = append(msgs, m)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get messages: %w", err)
	}
	return msgs, nil
}

func (e *Entity) GetMessagesBefore(ctx context.Context, channelId, msgId int64) ([]model.Message, error) {
	var msgs []model.Message
	iter := e.c.Session().
		Query(getMessagesBefore).
		WithContext(ctx).
		Bind(channelId, msgId).
		Iter()
	var m model.Message
	for iter.Scan(m.Id, &m.ChannelId, &m.UserId, &m.Content, &m.Attachments, &m.EditedAt) {
		msgs = append(msgs, m)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get messages before: %w", err)
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
