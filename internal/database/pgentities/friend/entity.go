package friend

import (
	"context"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/jmoiron/sqlx"
)

type Friend interface {
	AddFriend(ctx context.Context, userID, friendID int64) error
	RemoveFriend(ctx context.Context, userID, friendID int64) error
	GetFriends(ctx context.Context, userID int64) ([]model.Friend, error)
	CreateFriendRequest(ctx context.Context, userId, friendId int64) error
	RemoveFriendRequest(ctx context.Context, userId, friendId int64) error
	GetFriendRequests(ctx context.Context, userId int64) ([]model.FriendRequest, error)
	IsFriend(ctx context.Context, userId, friendId int64) (bool, error)
}

type Entity struct {
	c *sqlx.DB
}

func New(c *sqlx.DB) *Entity {
	return &Entity{c: c}
}
