package friend

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/Masterminds/squirrel"
)

func (e *Entity) AddFriend(ctx context.Context, userID, friendID int64) error {
	q := squirrel.Insert("friends").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "friend_id").
		Values(userID, friendID)
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to add friend: %w", err)
	}
	return nil
}

func (e *Entity) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	q := squirrel.Delete("friends").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"user_id": userID, "friend_id": friendID})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = e.c.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to remove friend: %w", err)
	}
	return nil
}

func (e *Entity) GetFriends(ctx context.Context, userID int64) ([]model.Friend, error) {
	var f []model.Friend
	q := squirrel.Select("user_id", "friend_id").
		From("friends").
		Where(squirrel.Eq{"user_id": userID})
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &f, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get friends: %w", err)
	}
	return f, nil
}
