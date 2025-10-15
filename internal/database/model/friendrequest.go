package model

type FriendRequest struct {
	UserId   int64 `db:"user_id"`
	FriendId int64 `db:"friend_id"`
}
