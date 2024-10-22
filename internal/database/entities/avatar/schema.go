package avatar

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	createAvatar       = `INSERT INTO gochat.avatars (id, user_id, object) VALUES (?, ?, ?)`
	removeAvatar       = `DELETE FROM gochat.avatars WHERE id = ?`
	getAvatar          = `SELECT id, user_id, object FROM gochat.avatars WHERE id = ?`
	getAvatarsByUserId = `SELECT id, user_id, object FROM gochat.avatars WHERE user_id = ?`
)

func (e *Entity) CreateAvatar(ctx context.Context, id, userId int64, object string) error {
	err := e.c.Session().
		Query(createAvatar).
		WithContext(ctx).
		Bind(id, userId, object).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create avatar: %w", err)
	}
	return nil
}

func (e *Entity) RemoveAvatar(ctx context.Context, id int64) error {
	err := e.c.Session().
		Query(removeAvatar).
		WithContext(ctx).
		Bind(id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove avatar: %w", err)
	}
	return nil
}

func (e *Entity) GetAvatar(ctx context.Context, id int64) (model.Avatar, error) {
	var a model.Avatar
	err := e.c.Session().
		Query(getAvatar).
		WithContext(ctx).
		Bind(id).
		Exec()
	if err != nil {
		return a, fmt.Errorf("unable to get avatar: %w", err)
	}
	return a, nil
}

func (e *Entity) GetAvatarsByUserId(ctx context.Context, userId int64) ([]model.Avatar, error) {
	var avatars []model.Avatar
	iter := e.c.Session().
		Query(getAvatarsByUserId).
		WithContext(ctx).
		Bind(userId).
		Iter()
	var a model.Avatar
	for iter.Scan(&a.Id, &a.UserId, &a.Object) {
		avatars = append(avatars, a)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get avatars: %w", err)
	}
	return avatars, nil
}
