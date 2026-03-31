package guild

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestReceiveInviteReturnsNotFoundWhenInviteIsMissing(t *testing.T) {
	e := &entity{
		inv: &fakeInviteRepo{err: sql.ErrNoRows},
	}
	app := newGuildTestApp(t, 10, "/guild/invites/receive/:invite_code", e.ReceiveInvite)

	req := httptest.NewRequest("GET", "/guild/invites/receive/ABCDEFGH", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestAcceptInviteReturnsNotFoundWhenInviteIsMissing(t *testing.T) {
	e := &entity{
		inv: &fakeInviteRepo{err: sql.ErrNoRows},
	}
	app := newGuildTestApp(t, 10, "/guild/invites/accept/:invite_code", e.AcceptInvite)

	req := httptest.NewRequest("POST", "/guild/invites/accept/ABCDEFGH", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
