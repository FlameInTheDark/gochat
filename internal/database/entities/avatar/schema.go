package avatar

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	// Placeholder row with TTL and done=false
	createAvatar     = `INSERT INTO gochat.avatars (id, user_id, done, filesize) VALUES (?, ?, false, ?) USING TTL ?`
	getAvatar        = `SELECT id, user_id, url, content_type, width, height, filesize, done FROM gochat.avatars WHERE user_id = ? AND id = ?`
	doneAvatar       = `UPDATE gochat.avatars USING TTL 0 SET done = true, content_type = ?, url = ?, height = ?, width = ?, filesize = ? WHERE user_id = ? AND id = ?`
	removeAvatar     = `DELETE FROM gochat.avatars WHERE user_id = ? AND id = ?`
	getAvatarsByUser = `SELECT id, user_id, url, content_type, width, height, filesize, done FROM gochat.avatars WHERE user_id = ?`
)

func (e *Entity) CreateAvatar(ctx context.Context, id, userId, ttlSeconds, fileSize int64) error {
	err := e.c.Session().
		Query(createAvatar).
		WithContext(ctx).
		Bind(id, userId, fileSize, ttlSeconds).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create avatar: %w", err)
	}
	return nil
}

func (e *Entity) GetAvatar(ctx context.Context, id, userId int64) (model.Avatar, error) {
	var a model.Avatar
	err := e.c.Session().
		Query(getAvatar).
		WithContext(ctx).
		Bind(userId, id).
		Scan(&a.Id, &a.UserId, &a.URL, &a.ContentType, &a.Width, &a.Height, &a.FileSize, &a.Done)
	if err != nil {
		return a, fmt.Errorf("unable to get avatar: %w", err)
	}
	return a, nil
}

func (e *Entity) DoneAvatar(ctx context.Context, id, userId int64, contentType, url *string, height, width, fileSize *int64) error {
	err := e.c.Session().
		Query(doneAvatar).
		WithContext(ctx).
		Bind(contentType, url, height, width, fileSize, userId, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to done avatar: %w", err)
	}
	return nil
}

func (e *Entity) RemoveAvatar(ctx context.Context, id, userId int64) error {
	err := e.c.Session().
		Query(removeAvatar).
		WithContext(ctx).
		Bind(userId, id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove avatar: %w", err)
	}
	return nil
}

// GetAvatarsByUserId returns list of avatars for a user
func (e *Entity) GetAvatarsByUserId(ctx context.Context, userId int64) ([]model.Avatar, error) {
	var avatars []model.Avatar
	iter := e.c.Session().
		Query(getAvatarsByUser).
		WithContext(ctx).
		Bind(userId).
		Iter()
	var a model.Avatar
	for iter.Scan(&a.Id, &a.UserId, &a.URL, &a.ContentType, &a.Width, &a.Height, &a.FileSize, &a.Done) {
		avatars = append(avatars, a)
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("unable to get avatars: %w", err)
	}
	return avatars, nil
}
