package friend

import (
	"context"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
)

const (
	addFriend = `
		INSERT INTO gochat.friends (user_id, friend_id, created_at) VALUES (?, ?, toTimestamp(now()));
		INSERT INTO gochat.friends (user_id, friend_id, created_at) VALUES (?, ?, toTimestamp(now()));`
	removeFriend = `
		DELETE FROM gochat.friends WHERE user_id = ? AND friend_id = ?;
		DELETE FROM gochat.friends WHERE user_id = ? AND friend_id = ?;`
	getFriends = `SELECT user_id, friend_id, created_at FROM gochat.friends WHERE user_id = ?;`
)

func (e *Entity) AddFriend(ctx context.Context, userID, friendID int64) error {
	err := e.c.Session().
		Query(addFriend).
		WithContext(ctx).
		Bind(userID, friendID, friendID, userID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to add friend: %w", err)
	}
	return nil
}

func (e *Entity) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	err := e.c.Session().
		Query(removeFriend).
		WithContext(ctx).
		Bind(userID, friendID, friendID, userID).
		Exec()
	if err != nil {
		return fmt.Errorf("unable to remove friend: %w", err)
	}
	return nil
}

func (e *Entity) GetFriends(ctx context.Context, userID int64) ([]model.Friend, error) {
	var f []model.Friend
	iter := e.c.Session().
		Query(getFriends).
		WithContext(ctx).
		Bind(userID).
		Iter()
	var friend model.Friend
	for iter.Scan(&friend.UserID, &friend.FriendID, &friend.CreatedAt) {
		f = append(f, friend)
	}
	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to get friends: %w", err)
	}
	return f, nil
}
