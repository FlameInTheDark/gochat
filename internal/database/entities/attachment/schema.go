package attachment

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocql/gocql"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	createAttachment   = `INSERT INTO gochat.attachments (id, channel_id, done, filesize, name, height, width, url, content_type) VALUES (?, ?, true, ?, ?, ?, ?, ?, ?)`
	removeAttachment   = `DELETE FROM gochat.attachments WHERE id = ? AND channel_id = ?`
	getAttachment      = `SELECT id, channel_id, name, filesize, content_type, height, width, url, done FROM gochat.attachments WHERE id = ? AND channel_id = ?`
	getAttachmentsByID = `SELECT id, channel_id, name, filesize, content_type, height, width, url, done FROM gochat.attachments WHERE id IN ?`
	doneAttachment     = `UPDATE gochat.attachments USING TTL 0 SET done = true, content_type = ?, url = ? WHERE id = ? AND channel_id = ?`
)

func (e *Entity) CreateAttachment(ctx context.Context, id, channelId, fileSize int64, height, width int64, name, url, contentType string) error {
	err := e.c.Session().
		Query(createAttachment).
		WithContext(ctx).
		Bind(id, channelId, fileSize, name, height, width, url, contentType).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create attachment: %w", err)
	}
	return nil
}

func (e *Entity) RemoveAttachment(ctx context.Context, id, channelId int64) error {
	err := e.c.Session().
		Query(removeAttachment).
		WithContext(ctx).
		Bind(id, channelId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove attachment: %w", err)
	}
	return nil
}

func (e *Entity) GetAttachment(ctx context.Context, id, channelId int64) (model.Attachment, error) {
	var a model.Attachment
	err := e.c.Session().
		Query(getAttachment).
		WithContext(ctx).
		Bind(id, channelId).
		Scan(&a.Id, &a.ChannelId, &a.Name, &a.FileSize, &a.ContentType, &a.Height, &a.Width, &a.URL, &a.Done)
	if err != nil {
		return a, fmt.Errorf("unable to get attachment: %w", err)
	}
	return a, nil
}

func (e *Entity) DoneAttachment(ctx context.Context, id, channelId int64, contentType, url *string) error {
	err := e.c.Session().
		Query(doneAttachment).
		WithContext(ctx).
		Bind(contentType, url, id, channelId).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to done attachment: %w", err)
	}
	return nil
}

func (e *Entity) SelectAttachmentByIDs(ctx context.Context, ids []int64) ([]model.Attachment, error) {
	var attachments []model.Attachment
	iter := e.c.Session().
		Query(getAttachmentsByID).
		WithContext(ctx).
		Bind(ids).
		Iter()
	var a model.Attachment
	for iter.Scan(&a.Id, &a.ChannelId, &a.Name, &a.FileSize, &a.ContentType, &a.Height, &a.Width, &a.URL, &a.Done) {
		attachments = append(attachments, a)
	}
	err := iter.Close()
	if errors.Is(err, gocql.ErrNotFound) {
		return []model.Attachment{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get attachments: %w", err)
	}
	return attachments, nil
}
