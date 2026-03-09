package attachment

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocql/gocql"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	createAttachment        = `INSERT INTO gochat.attachments (id, channel_id, author_id, done, filesize, name) VALUES (?, ?, ?, false, ?, ?) USING TTL ?`
	removeAttachment        = `DELETE FROM gochat.attachments WHERE channel_id = ? AND id = ?`
	getAttachment           = `SELECT id, channel_id, name, filesize, content_type, height, width, url, preview_url, author_id, done FROM gochat.attachments WHERE channel_id = ? AND id = ?`
	getAttachmentsByChannel = `SELECT id, channel_id, name, filesize, content_type, height, width, url, preview_url, author_id, done FROM gochat.attachments WHERE channel_id = ? AND id IN ?`
	listDoneZeroSize        = `SELECT id, channel_id, name, filesize, content_type, height, width, url, preview_url, author_id, done FROM gochat.attachments WHERE done = true ALLOW FILTERING`
	updateFileSize          = `UPDATE gochat.attachments SET filesize = ? WHERE channel_id = ? AND id = ?`
	updateName              = `UPDATE gochat.attachments SET name = ? WHERE channel_id = ? AND id = ?`
	doneAttachment          = `UPDATE gochat.attachments USING TTL 0 SET done = true, content_type = ?, url = ?, preview_url = ?, height = ?, width = ?, filesize = ?, name = ?, author_id = ? WHERE channel_id = ? AND id = ?`
)

func (e *Entity) CreateAttachment(ctx context.Context, id, channelId, authorId, ttlSeconds, fileSize int64, name string) error {
	err := e.c.Session().
		Query(createAttachment).
		WithContext(ctx).
		Bind(id, channelId, authorId, fileSize, name, ttlSeconds).
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
		Bind(channelId, id).
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
		Bind(channelId, id).
		Scan(&a.Id, &a.ChannelId, &a.Name, &a.FileSize, &a.ContentType, &a.Height, &a.Width, &a.URL, &a.PreviewURL, &a.AuthorId, &a.Done)
	if err != nil {
		return a, fmt.Errorf("unable to get attachment: %w", err)
	}
	return a, nil
}

func (e *Entity) DoneAttachment(ctx context.Context, id, channelId int64, contentType, url, previewURL *string, height, width, fileSize *int64, name *string, authorId *int64) error {
	err := e.c.Session().
		Query(doneAttachment).
		WithContext(ctx).
		Bind(contentType, url, previewURL, height, width, fileSize, name, authorId, channelId, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to done attachment: %w", err)
	}
	return nil
}

func (e *Entity) SelectAttachmentsByChannel(ctx context.Context, channelId int64, ids []int64) ([]model.Attachment, error) {
	if len(ids) == 0 {
		return []model.Attachment{}, nil
	}

	var attachments []model.Attachment
	iter := e.c.Session().
		Query(getAttachmentsByChannel).
		WithContext(ctx).
		Bind(channelId, ids).
		Iter()
	var a model.Attachment
	for iter.Scan(&a.Id, &a.ChannelId, &a.Name, &a.FileSize, &a.ContentType, &a.Height, &a.Width, &a.URL, &a.PreviewURL, &a.AuthorId, &a.Done) {
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

func (e *Entity) UpdateFileSize(ctx context.Context, id, channelId int64, fileSize int64) error {
	err := e.c.Session().
		Query(updateFileSize).
		WithContext(ctx).
		Bind(fileSize, channelId, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to update attachment filesize: %w", err)
	}
	return nil
}

func (e *Entity) ListDoneZeroSize(ctx context.Context) ([]model.Attachment, error) {
	var attachments []model.Attachment
	iter := e.c.Session().
		Query(listDoneZeroSize).
		WithContext(ctx).
		Iter()
	var a model.Attachment
	for iter.Scan(&a.Id, &a.ChannelId, &a.Name, &a.FileSize, &a.ContentType, &a.Height, &a.Width, &a.URL, &a.PreviewURL, &a.AuthorId, &a.Done) {
		if a.FileSize == 0 {
			attachments = append(attachments, a)
		}
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("unable to list attachments: %w", err)
	}
	return attachments, nil
}

func (e *Entity) UpdateName(ctx context.Context, id, channelId int64, name string) error {
	err := e.c.Session().
		Query(updateName).
		WithContext(ctx).
		Bind(name, channelId, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to update attachment name: %w", err)
	}
	return nil
}
