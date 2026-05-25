package usernote

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) GetNote(ctx context.Context, ownerUserId, targetUserId int64) (model.UserNote, error) {
	var note model.UserNote
	q := squirrel.Select("owner_user_id", "target_user_id", "note", "updated_at").
		PlaceholderFormat(squirrel.Dollar).
		From("user_notes").
		Where(squirrel.Eq{"owner_user_id": ownerUserId, "target_user_id": targetUserId}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return note, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.GetContext(ctx, &note, raw, args...); err != nil {
		return note, fmt.Errorf("unable to get user note: %w", err)
	}
	return note, nil
}

func (e *Entity) UpsertNote(ctx context.Context, ownerUserId, targetUserId int64, note string) error {
	q := squirrel.Insert("user_notes").
		PlaceholderFormat(squirrel.Dollar).
		Columns("owner_user_id", "target_user_id", "note").
		Values(ownerUserId, targetUserId, note).
		Suffix("ON CONFLICT (owner_user_id, target_user_id) DO UPDATE SET note = EXCLUDED.note, updated_at = NOW()")
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err := e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to upsert user note: %w", err)
	}
	return nil
}

func (e *Entity) DeleteNote(ctx context.Context, ownerUserId, targetUserId int64) error {
	q := squirrel.Delete("user_notes").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"owner_user_id": ownerUserId, "target_user_id": targetUserId})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err := e.c.ExecContext(ctx, raw, args...); err != nil {
		return fmt.Errorf("unable to delete user note: %w", err)
	}
	return nil
}
