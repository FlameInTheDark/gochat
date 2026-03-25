package auth

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestParseAndValidateReturnsDetailedParseErrorAndTelemetry(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	exporter := &spanCaptureExporter{}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)))

	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	defer otel.SetTracerProvider(previous)

	e := &entity{log: logger}
	app := fiber.New()
	app.Post("/", func(c *fiber.Ctx) error {
		ctx, span := otel.Tracer("auth-test").Start(c.UserContext(), "request")
		defer span.End()
		c.SetUserContext(ctx)

		var req ConfirmationRequest
		return e.parseAndValidate(c, "confirmation", &req)
	})

	body := `{"id":"not-a-number","token":"4fuoUwnpdrgSssKCZOlN0_m1Ek_d9bGmKWqEgR3v","name":"Menchikoff","discriminator":"menchikoff","password":"c30pl2p6h977"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	respBody, _ := io.ReadAll(resp.Body)
	gotBody := strings.TrimSpace(string(respBody))
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	if !strings.Contains(gotBody, "unable to parse body:") || !strings.Contains(gotBody, "invalid syntax") {
		t.Fatalf("expected detailed parse error response, got %q", gotBody)
	}

	logOutput := logs.String()
	if !strings.Contains(logOutput, "auth request rejected") {
		t.Fatalf("expected rejection log, got %q", logOutput)
	}
	if !strings.Contains(logOutput, "auth_operation=confirmation") || !strings.Contains(logOutput, "stage=parse") {
		t.Fatalf("expected structured operation/stage in logs, got %q", logOutput)
	}

	spans := exporter.spans()
	if len(spans) != 1 {
		t.Fatalf("expected one ended span, got %d", len(spans))
	}
	if !hasSpanEvent(spans[0], "auth.request.rejected") {
		t.Fatalf("expected auth.request.rejected span event, got %#v", spans[0].Events())
	}
}

func TestParseAndValidateReturnsDetailedValidationErrorAndTelemetry(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	exporter := &spanCaptureExporter{}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)))

	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	defer otel.SetTracerProvider(previous)

	e := &entity{log: logger}
	app := fiber.New()
	app.Post("/", func(c *fiber.Ctx) error {
		ctx, span := otel.Tracer("auth-test").Start(c.UserContext(), "request")
		defer span.End()
		c.SetUserContext(ctx)

		var req ConfirmationRequest
		return e.parseAndValidate(c, "confirmation", &req)
	})

	body := `{"id":"2297802081286750208","token":"4fuoUwnpdrgSssKCZOlN0_m1Ek_d9bGmKWqEgR3v","name":"Menchikoff","discriminator":"menchikoff","password":"short"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	respBody, _ := io.ReadAll(resp.Body)
	gotBody := strings.TrimSpace(string(respBody))
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	if !strings.Contains(gotBody, ErrPasswordIsTooShort) {
		t.Fatalf("expected validation message to contain %q, got %q", ErrPasswordIsTooShort, gotBody)
	}

	logOutput := logs.String()
	if !strings.Contains(logOutput, "auth request rejected") {
		t.Fatalf("expected rejection log, got %q", logOutput)
	}
	if !strings.Contains(logOutput, "stage=validation") {
		t.Fatalf("expected validation stage in logs, got %q", logOutput)
	}

	spans := exporter.spans()
	if len(spans) != 1 {
		t.Fatalf("expected one ended span, got %d", len(spans))
	}
	if !hasSpanEvent(spans[0], "auth.request.rejected") {
		t.Fatalf("expected auth.request.rejected span event, got %#v", spans[0].Events())
	}
}

type spanCaptureExporter struct {
	mu    sync.Mutex
	items []sdktrace.ReadOnlySpan
}

func (e *spanCaptureExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.items = append(e.items, spans...)
	return nil
}

func (e *spanCaptureExporter) Shutdown(context.Context) error {
	return nil
}

func (e *spanCaptureExporter) spans() []sdktrace.ReadOnlySpan {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]sdktrace.ReadOnlySpan, len(e.items))
	copy(out, e.items)
	return out
}

func hasSpanEvent(span sdktrace.ReadOnlySpan, name string) bool {
	for _, event := range span.Events() {
		if event.Name == name {
			return true
		}
	}
	return false
}
