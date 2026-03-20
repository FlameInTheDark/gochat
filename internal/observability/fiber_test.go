package observability

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestContextMiddlewareGeneratesRequestID(t *testing.T) {
	app := fiber.New()
	app.Use(RequestContextMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		if RequestIDFromContext(c.UserContext()) == "" {
			t.Fatalf("expected request id in context")
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if got := resp.Header.Get(RequestIDHeader); got == "" {
		t.Fatalf("expected request id response header")
	}
}

func TestRequestContextMiddlewareHonorsRequestID(t *testing.T) {
	app := fiber.New()
	app.Use(RequestContextMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		if got := RequestIDFromContext(c.UserContext()); got != "req-123" {
			t.Fatalf("unexpected request id: %q", got)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, "req-123")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if got := resp.Header.Get(RequestIDHeader); got != "req-123" {
		t.Fatalf("unexpected response request id: %q", got)
	}
}

func TestRequestLoggerSkipsHealthz(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))

	app := fiber.New()
	app.Use(RequestContextMiddleware())
	app.Use(RequestLogger(logger))
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/healthz", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if got := output.String(); got != "" {
		t.Fatalf("expected no log output for /healthz, got %q", got)
	}
}

func TestRequestLoggerLogsApplicationRoutes(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))

	app := fiber.New()
	app.Use(RequestContextMiddleware())
	app.Use(RequestLogger(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if payload["msg"] != "request completed" {
		t.Fatalf("unexpected log message: %#v", payload)
	}
	if payload["route"] != "/test" {
		t.Fatalf("unexpected route field: %#v", payload)
	}
}

func TestLoggerFromFiberIncludesRequestID(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))

	app := fiber.New()
	app.Use(RequestContextMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		LoggerFromFiber(c, logger).Error("scoped route log")
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, "req-123")
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if payload["request_id"] != "req-123" {
		t.Fatalf("unexpected request id field: %#v", payload)
	}
}
