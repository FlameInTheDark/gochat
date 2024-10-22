package attachment

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	createAttachment = `INSERT INTO gochat.attachments (id, object) VALUES (?, ?)`
	removeAttachment = `DELETE FROM gochat.attachments WHERE id = ?`
	getAttachment    = `SELECT id, object FROM gochat.attachments WHERE id = ?`
)

func (e *Entity) CreateAttachment(ctx context.Context, id, userId int64, object string) error {
	err := e.c.Session().
		Query(createAttachment).
		WithContext(ctx).
		Bind(id, object).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to create attachment: %w", err)
	}
	return nil
}

func (e *Entity) RemoveAttachment(ctx context.Context, id int64) error {
	err := e.c.Session().
		Query(removeAttachment).
		WithContext(ctx).
		Bind(id).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove attachment: %w", err)
	}
	return nil
}

func (e *Entity) GetAttachment(ctx context.Context, id int64) (model.Attachment, error) {
	var a model.Attachment
	err := e.c.Session().
		Query(getAttachment).
		WithContext(ctx).
		Bind(id).
		Exec()
	if err != nil {
		return a, fmt.Errorf("unable to get attachment: %w", err)
	}
	return a, nil
}
