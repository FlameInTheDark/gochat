package mqmsg

import (
	"encoding/json"
	"github.com/FlameInTheDark/gochat/internal/dto"
)

// UserBrief is a lightweight user payload for events
type UserBrief struct {
	Id            int64           `json:"id"`
	Name          string          `json:"name"`
	Discriminator string          `json:"discriminator"`
	Avatar        *int64          `json:"avatar,omitempty"`      // legacy numeric ID
	AvatarData    *dto.AvatarData `json:"avatar_data,omitempty"` // full avatar metadata
}

// IncomingFriendRequest notifies recipient about a new friend request
type IncomingFriendRequest struct {
	From UserBrief `json:"from"`
}

func (m *IncomingFriendRequest) EventType() *EventType {
	e := EventTypeUserFriendRequest
	return &e
}

func (m *IncomingFriendRequest) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *IncomingFriendRequest) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// FriendAdded notifies users when a friendship was established
type FriendAdded struct {
	Friend UserBrief `json:"friend"`
}

func (m *FriendAdded) EventType() *EventType {
	e := EventTypeUserFriendAdded
	return &e
}

func (m *FriendAdded) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *FriendAdded) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// FriendRemoved notifies users when a friendship was removed
type FriendRemoved struct {
	Friend UserBrief `json:"friend"`
}

func (m *FriendRemoved) EventType() *EventType {
	e := EventTypeUserFriendRemoved
	return &e
}

func (m *FriendRemoved) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *FriendRemoved) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
