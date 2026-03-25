package friend

import (
	"errors"
	"strings"
	"testing"

	"github.com/lib/pq"
)

func TestMapFriendSQLErrorDuplicateFriendRequest(t *testing.T) {
	err := mapFriendSQLError("create friend request", &pq.Error{
		Code:       "23505",
		Constraint: "friend_requests_pkey_102651",
	})
	if !errors.Is(err, ErrFriendRequestAlreadyExists) {
		t.Fatalf("expected ErrFriendRequestAlreadyExists, got %v", err)
	}
}

func TestMapFriendSQLErrorDuplicateFriendship(t *testing.T) {
	err := mapFriendSQLError("add friend", &pq.Error{
		Code:       "23505",
		Constraint: "idx_unique_friend_102651",
	})
	if !errors.Is(err, ErrAlreadyFriends) {
		t.Fatalf("expected ErrAlreadyFriends, got %v", err)
	}
}

func TestMapFriendSQLErrorWrapsUnknownErrors(t *testing.T) {
	root := errors.New("boom")
	err := mapFriendSQLError("create friend request", root)
	if err == nil {
		t.Fatal("expected wrapped error")
	}
	if !strings.Contains(err.Error(), "create friend request: boom") {
		t.Fatalf("expected wrapped operation context, got %q", err.Error())
	}
}
