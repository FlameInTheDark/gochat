package user

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	friendrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/friend"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type friendRepoMock struct {
	isFriendErr            error
	isFriendResult         bool
	createFriendRequestErr error
}

func (m *friendRepoMock) AddFriend(context.Context, int64, int64) error { return nil }
func (m *friendRepoMock) RemoveFriend(context.Context, int64, int64) error {
	return nil
}
func (m *friendRepoMock) GetFriends(context.Context, int64) ([]model.Friend, error) {
	return nil, nil
}
func (m *friendRepoMock) CreateFriendRequest(context.Context, int64, int64) error {
	return m.createFriendRequestErr
}
func (m *friendRepoMock) RemoveFriendRequest(context.Context, int64, int64) error {
	return nil
}
func (m *friendRepoMock) GetFriendRequests(context.Context, int64) ([]model.FriendRequest, error) {
	return nil, nil
}
func (m *friendRepoMock) IsFriend(context.Context, int64, int64) (bool, error) {
	return m.isFriendResult, m.isFriendErr
}

type discriminatorRepoMock struct {
	userByDiscriminator model.Discriminator
	getByDiscriminator  error
}

func (m *discriminatorRepoMock) CreateDiscriminator(context.Context, int64, string) error { return nil }
func (m *discriminatorRepoMock) GetDiscriminatorByUserId(context.Context, int64) (model.Discriminator, error) {
	return model.Discriminator{}, nil
}
func (m *discriminatorRepoMock) GetUserIdByDiscriminator(context.Context, string) (model.Discriminator, error) {
	return m.userByDiscriminator, m.getByDiscriminator
}
func (m *discriminatorRepoMock) GetDiscriminatorsByUserIDs(context.Context, []int64) ([]model.Discriminator, error) {
	return nil, nil
}

func TestCreateFriendRequestReturnsOKWhenRequestAlreadyExists(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", &jwt.Token{Claims: &helper.Claims{UserID: 1}})
		return c.Next()
	})

	e := &entity{
		disc: &discriminatorRepoMock{
			userByDiscriminator: model.Discriminator{UserId: 2, Discriminator: "menchikoff"},
		},
		fr: &friendRepoMock{
			createFriendRequestErr: friendrepo.ErrFriendRequestAlreadyExists,
		},
	}
	app.Post("/friends", e.CreateFriendRequest)

	req := httptest.NewRequest(http.MethodPost, "/friends", strings.NewReader(`{"discriminator":"menchikoff"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected request to complete, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}
}

func TestCreateFriendRequestReturnsConflictWhenUsersAreAlreadyFriends(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", &jwt.Token{Claims: &helper.Claims{UserID: 1}})
		return c.Next()
	})

	e := &entity{
		disc: &discriminatorRepoMock{
			userByDiscriminator: model.Discriminator{UserId: 2, Discriminator: "menchikoff"},
		},
		fr: &friendRepoMock{
			isFriendResult: true,
		},
	}
	app.Post("/friends", e.CreateFriendRequest)

	req := httptest.NewRequest(http.MethodPost, "/friends", strings.NewReader(`{"discriminator":"menchikoff"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("expected request to complete, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected status %d, got %d", fiber.StatusConflict, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("expected response body, got %v", err)
	}
	if strings.TrimSpace(string(body)) != ErrAlreadyFriends {
		t.Fatalf("expected body %q, got %q", ErrAlreadyFriends, strings.TrimSpace(string(body)))
	}
}
