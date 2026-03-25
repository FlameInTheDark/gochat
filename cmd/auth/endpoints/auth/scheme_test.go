package auth

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestConfirmationRequestBodyParserAcceptsStringAndNumericID(t *testing.T) {
	const expectedID int64 = 2297802081286750208

	tests := []struct {
		name string
		body string
	}{
		{
			name: "string id",
			body: `{"id":"2297802081286750208","token":"4fuoUwnpdrgSssKCZOlN0_m1Ek_d9bGmKWqEgR3v","name":"Menchikoff","discriminator":"menchikoff","password":"c30pl2p6h977"}`,
		},
		{
			name: "numeric id",
			body: `{"id":2297802081286750208,"token":"4fuoUwnpdrgSssKCZOlN0_m1Ek_d9bGmKWqEgR3v","name":"Menchikoff","discriminator":"menchikoff","password":"c30pl2p6h977"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/", func(c *fiber.Ctx) error {
				var req ConfirmationRequest
				if err := c.BodyParser(&req); err != nil {
					return err
				}
				if req.Id != expectedID {
					t.Fatalf("expected id %d, got %d", expectedID, req.Id)
				}
				if req.Discriminator != "menchikoff" {
					t.Fatalf("expected discriminator to be parsed, got %q", req.Discriminator)
				}
				return c.SendStatus(fiber.StatusNoContent)
			})

			req := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != fiber.StatusNoContent {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("expected status %d, got %d: %s", fiber.StatusNoContent, resp.StatusCode, strings.TrimSpace(string(body)))
			}
		})
	}
}

func TestPasswordResetRequestBodyParserAcceptsStringAndNumericID(t *testing.T) {
	const expectedID int64 = 2297802081286750208

	tests := []struct {
		name string
		body string
	}{
		{
			name: "string id",
			body: `{"id":"2297802081286750208","token":"4fuoUwnpdrgSssKCZOlN0_m1Ek_d9bGmKWqEgR3v","password":"c30pl2p6h977"}`,
		},
		{
			name: "numeric id",
			body: `{"id":2297802081286750208,"token":"4fuoUwnpdrgSssKCZOlN0_m1Ek_d9bGmKWqEgR3v","password":"c30pl2p6h977"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/", func(c *fiber.Ctx) error {
				var req PasswordResetRequest
				if err := c.BodyParser(&req); err != nil {
					return err
				}
				if req.Id != expectedID {
					t.Fatalf("expected id %d, got %d", expectedID, req.Id)
				}
				if req.Password != "c30pl2p6h977" {
					t.Fatalf("expected password to be parsed, got %q", req.Password)
				}
				return c.SendStatus(fiber.StatusNoContent)
			})

			req := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != fiber.StatusNoContent {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("expected status %d, got %d: %s", fiber.StatusNoContent, resp.StatusCode, strings.TrimSpace(string(body)))
			}
		})
	}
}
