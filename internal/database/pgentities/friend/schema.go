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
	tx, err := e.c.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	qu := squirrel.Insert("friends").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "friend_id").
		Values(userID, friendID)
	raw, args, err := qu.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = tx.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to add friend: %w", err)
	}

	qf := squirrel.Insert("friends").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "friend_id").
		Values(friendID, userID)
	rawf, argsf, err := qf.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = tx.ExecContext(ctx, rawf, argsf...)
	if err != nil {
		return fmt.Errorf("unable to add friend: %w", err)
	}

	frr := squirrel.Delete("friend_requests").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"user_id": userID, "friend_id": friendID})
	rawfrr, argsfrr, err := frr.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = tx.ExecContext(ctx, rawfrr, argsfrr...)
	if err != nil {
		return fmt.Errorf("unable to add friend: %w", err)
	}

	return nil
}

func (e *Entity) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	tx, err := e.c.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	q := squirrel.Delete("friends").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"user_id": userID, "friend_id": friendID})
	raw, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = tx.ExecContext(ctx, raw, args...)
	if err != nil {
		return fmt.Errorf("unable to remove friend: %w", err)
	}

	qf := squirrel.Delete("friends").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"user_id": friendID, "friend_id": userID})
	rawf, argsf, err := qf.ToSql()
	if err != nil {
		return fmt.Errorf("unable to create SQL query: %w", err)
	}
	_, err = tx.ExecContext(ctx, rawf, argsf...)
	if err != nil {
		return fmt.Errorf("unable to remove friend: %w", err)
	}
	return nil
}

func (e *Entity) GetFriends(ctx context.Context, userID int64) ([]model.Friend, error) {
	var f []model.Friend
	q := squirrel.Select("user_id", "friend_id").
		PlaceholderFormat(squirrel.Dollar).
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

func (e *Entity) CreateFriendRequest(ctx context.Context, userId, friendId int64) error {
	q := squirrel.Insert("friend_requests").
		PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "friend_id").
		Values(friendId, userId)
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

func (e *Entity) RemoveFriendRequest(ctx context.Context, userId, friendId int64) error {
	q := squirrel.Delete("friend_requests").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"user_id": userId, "friend_id": friendId})
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

func (e *Entity) GetFriendRequests(ctx context.Context, userId int64) ([]model.FriendRequest, error) {
	var reqs []model.FriendRequest
	q := squirrel.Select("user_id", "friend_id").
		PlaceholderFormat(squirrel.Dollar).
		From("friend_requests").
		Where(squirrel.Eq{"user_id": userId})
	raw, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("unable to create SQL query: %w", err)
	}
	err = e.c.SelectContext(ctx, &reqs, raw, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get friends: %w", err)
	}
	return reqs, nil
}

func (e *Entity) IsFriend(ctx context.Context, userId, friendId int64) (bool, error) {
	subq := squirrel.
		Select("1").
		PlaceholderFormat(squirrel.Dollar).
		From("friends").
		Where(
			squirrel.Or{
				squirrel.And{
					squirrel.Eq{"user_id": userId},
					squirrel.Eq{"friend_id": friendId},
				},
				squirrel.And{
					squirrel.Eq{"user_id": friendId},
					squirrel.Eq{"friend_id": userId},
				},
			},
		)

	sqlStr, args, err := subq.ToSql()
	if err != nil {
		return false, err
	}

	var areFriends bool
	if err := e.c.QueryRowContext(ctx, fmt.Sprintf("EXISTS (%s) AS are_friends", sqlStr), args...).
		Scan(&areFriends); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return areFriends, nil
}
