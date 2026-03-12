package threadmember

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) AddThreadMember(ctx context.Context, threadID, userID int64) (model.ThreadMember, error) {
	var member model.ThreadMember
	q := squirrel.Insert("thread_members").
		PlaceholderFormat(squirrel.Dollar).
		Columns("thread_id", "user_id").
		Values(threadID, userID).
		Suffix("ON CONFLICT (thread_id, user_id) DO UPDATE SET user_id = EXCLUDED.user_id RETURNING thread_id, user_id, flags, join_at")
	raw, args, err := q.ToSql()
	if err != nil {
		return member, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.GetContext(ctx, &member, raw, args...); err != nil {
		return model.ThreadMember{}, fmt.Errorf("unable to add thread member: %w", err)
	}
	return member, nil
}

func (e *Entity) RemoveThreadMember(ctx context.Context, threadID, userID int64) error {
	q := squirrel.Delete("thread_members").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.And{
			squirrel.Eq{"thread_id": threadID},
			squirrel.Eq{"user_id": userID},
		})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err := e.c.ExecContext(ctx, raw, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("unable to remove thread member: %w", err)
	}
	return nil
}

func (e *Entity) RemoveThreadMembers(ctx context.Context, threadID int64) error {
	q := squirrel.Delete("thread_members").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"thread_id": threadID})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	if _, err := e.c.ExecContext(ctx, raw, args...); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("unable to remove thread members: %w", err)
	}
	return nil
}

func (e *Entity) GetThreadMember(ctx context.Context, threadID, userID int64) (model.ThreadMember, error) {
	var member model.ThreadMember
	q := squirrel.Select("thread_id", "user_id", "flags", "join_at").
		PlaceholderFormat(squirrel.Dollar).
		From("thread_members").
		Where(squirrel.And{
			squirrel.Eq{"thread_id": threadID},
			squirrel.Eq{"user_id": userID},
		}).
		Limit(1)
	raw, args, err := q.ToSql()
	if err != nil {
		return member, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.GetContext(ctx, &member, raw, args...); err != nil {
		return model.ThreadMember{}, fmt.Errorf("unable to get thread member: %w", err)
	}
	return member, nil
}

func (e *Entity) GetThreadMembers(ctx context.Context, threadID int64) ([]model.ThreadMember, error) {
	var members []model.ThreadMember
	q := squirrel.Select("thread_id", "user_id", "flags", "join_at").
		PlaceholderFormat(squirrel.Dollar).
		From("thread_members").
		Where(squirrel.Eq{"thread_id": threadID}).
		OrderBy("join_at ASC", "user_id ASC")
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.SelectContext(ctx, &members, raw, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.ThreadMember{}, nil
		}
		return nil, fmt.Errorf("unable to get thread members: %w", err)
	}
	return members, nil
}

func (e *Entity) GetThreadMembersBulk(ctx context.Context, threadIDs []int64) ([]model.ThreadMember, error) {
	if len(threadIDs) == 0 {
		return []model.ThreadMember{}, nil
	}
	var members []model.ThreadMember
	q := squirrel.Select("thread_id", "user_id", "flags", "join_at").
		PlaceholderFormat(squirrel.Dollar).
		From("thread_members").
		Where(squirrel.Eq{"thread_id": threadIDs}).
		OrderBy("thread_id ASC", "join_at ASC", "user_id ASC")
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.SelectContext(ctx, &members, raw, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.ThreadMember{}, nil
		}
		return nil, fmt.Errorf("unable to get thread members bulk: %w", err)
	}
	return members, nil
}

func (e *Entity) GetThreadMembersByUser(ctx context.Context, userID int64, threadIDs []int64) ([]model.ThreadMember, error) {
	if len(threadIDs) == 0 {
		return []model.ThreadMember{}, nil
	}
	var members []model.ThreadMember
	q := squirrel.Select("thread_id", "user_id", "flags", "join_at").
		PlaceholderFormat(squirrel.Dollar).
		From("thread_members").
		Where(squirrel.And{
			squirrel.Eq{"user_id": userID},
			squirrel.Eq{"thread_id": threadIDs},
		})
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.SelectContext(ctx, &members, raw, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.ThreadMember{}, nil
		}
		return nil, fmt.Errorf("unable to get thread members: %w", err)
	}
	return members, nil
}

func (e *Entity) GetUserThreadMembers(ctx context.Context, userID int64) ([]model.ThreadMember, error) {
	var members []model.ThreadMember
	q := squirrel.Select("thread_id", "user_id", "flags", "join_at").
		PlaceholderFormat(squirrel.Dollar).
		From("thread_members").
		Where(squirrel.Eq{"user_id": userID})
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	if err := e.c.SelectContext(ctx, &members, raw, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.ThreadMember{}, nil
		}
		return nil, fmt.Errorf("unable to get user thread members: %w", err)
	}
	return members, nil
}
