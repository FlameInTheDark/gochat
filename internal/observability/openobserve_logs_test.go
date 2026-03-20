package observability

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

func TestOpenObserveLogExporterBatchesAndPreservesCorrelationFields(t *testing.T) {
	type requestPayload struct {
		path string
		auth string
		body []byte
	}

	requests := make(chan requestPayload, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		requests <- requestPayload{
			path: r.URL.Path,
			auth: r.Header.Get("Authorization"),
			body: body,
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	exporter, err := newOpenObserveLogExporter("gochat-sfu", openObserveLogExporterConfig{
		endpoint:       server.URL + "/api/default",
		authHeader:     "Basic example",
		stream:         "gochat_logs",
		queueSize:      8,
		batchSize:      2,
		flushInterval:  20 * time.Millisecond,
		requestTimeout: time.Second,
		retryBackoff:   10 * time.Millisecond,
		maxAttempts:    1,
	})
	if err != nil {
		t.Fatalf("newOpenObserveLogExporter: %v", err)
	}
	defer exporter.Close()

	logger := slog.New(&contextualHandler{
		next:               slog.NewJSONHandler(exporter.Writer(), nil),
		serviceName:        "gochat-sfu",
		deploymentEnvValue: "test",
		fixedAttrs: buildLogAttrs([]attribute.KeyValue{
			attribute.String("voice.region", "eu-central"),
			attribute.String("service.instance.id", "sfu-eu-1"),
		}),
	})

	ctx := helper.ContextWithRequestID(context.Background(), "req-123")
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    [16]byte{1, 2, 3, 4},
		SpanID:     [8]byte{5, 6, 7, 8},
		TraceFlags: trace.FlagsSampled,
	})
	ctx = trace.ContextWithSpanContext(ctx, spanCtx)

	logger.InfoContext(ctx, "peer joined", slog.String("event", "join"))
	logger.WarnContext(ctx, "peer left", slog.String("event", "leave"))

	select {
	case req := <-requests:
		if req.path != "/api/default/gochat_logs/_json" {
			t.Fatalf("unexpected request path: %s", req.path)
		}
		if req.auth != "Basic example" {
			t.Fatalf("unexpected auth header: %q", req.auth)
		}

		var payload []map[string]any
		if err := json.Unmarshal(req.body, &payload); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if len(payload) != 2 {
			t.Fatalf("expected 2 log records, got %d", len(payload))
		}
		if payload[0]["service.name"] != "gochat-sfu" {
			t.Fatalf("missing service.name in payload: %#v", payload[0])
		}
		if payload[0]["voice.region"] != "eu-central" {
			t.Fatalf("missing voice.region in payload: %#v", payload[0])
		}
		if payload[0]["service.instance.id"] != "sfu-eu-1" {
			t.Fatalf("missing service.instance.id in payload: %#v", payload[0])
		}
		if payload[0]["request_id"] != "req-123" {
			t.Fatalf("missing request_id in payload: %#v", payload[0])
		}
		if payload[0]["trace_id"] == "" || payload[0]["span_id"] == "" {
			t.Fatalf("missing trace correlation in payload: %#v", payload[0])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for batched OpenObserve request")
	}
}

func TestOpenObserveLogExporterRetriesTransientFailures(t *testing.T) {
	var attempts atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := attempts.Add(1)
		if current < 3 {
			http.Error(w, "retry", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	exporter, err := newOpenObserveLogExporter("gochat-sfu", openObserveLogExporterConfig{
		endpoint:       server.URL + "/api/default",
		authHeader:     "Basic example",
		stream:         "gochat_logs",
		queueSize:      4,
		batchSize:      1,
		flushInterval:  10 * time.Millisecond,
		requestTimeout: time.Second,
		retryBackoff:   10 * time.Millisecond,
		maxAttempts:    3,
	})
	if err != nil {
		t.Fatalf("newOpenObserveLogExporter: %v", err)
	}
	defer exporter.Close()

	exporter.enqueue([]byte(`{"msg":"retry me"}`))

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if attempts.Load() >= 3 && exporter.successTotal.Load() == 1 {
			if exporter.failureTotal.Load() != 0 {
				t.Fatalf("expected no permanent failures, got %d", exporter.failureTotal.Load())
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("expected three attempts and one successful send, got attempts=%d success=%d failure=%d",
		attempts.Load(), exporter.successTotal.Load(), exporter.failureTotal.Load())
}

func TestOpenObserveLogExporterDropsWhenQueueIsFull(t *testing.T) {
	exporter := &openObserveLogExporter{
		queue: make(chan []byte, 1),
	}
	exporter.queue <- []byte(`{"msg":"existing"}`)

	exporter.enqueue([]byte(`{"msg":"overflow"}`))

	if got := exporter.droppedTotal.Load(); got != 1 {
		t.Fatalf("expected droppedTotal=1, got %d", got)
	}
	if got := exporter.enqueuedTotal.Load(); got != 0 {
		t.Fatalf("expected no enqueued increments on drop, got %d", got)
	}
}

func TestOpenObserveLogExporterPreservesRedaction(t *testing.T) {
	bodyCh := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		bodyCh <- string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	exporter, err := newOpenObserveLogExporter("gochat-sfu", openObserveLogExporterConfig{
		endpoint:       server.URL + "/api/default",
		authHeader:     "Basic example",
		stream:         "gochat_logs",
		queueSize:      4,
		batchSize:      1,
		flushInterval:  10 * time.Millisecond,
		requestTimeout: time.Second,
		retryBackoff:   10 * time.Millisecond,
		maxAttempts:    1,
	})
	if err != nil {
		t.Fatalf("newOpenObserveLogExporter: %v", err)
	}
	defer exporter.Close()

	logger := slog.New(&contextualHandler{
		next:               slog.NewJSONHandler(exporter.Writer(), nil),
		serviceName:        "gochat-sfu",
		deploymentEnvValue: "test",
	})
	logger.Info("redacted payload", slog.String("token", RedactSecret("super-secret-token")))

	select {
	case body := <-bodyCh:
		if strings.Contains(body, "super-secret-token") {
			t.Fatalf("expected secret to stay redacted, got %s", body)
		}
		if !strings.Contains(body, redactedValue) {
			t.Fatalf("expected redacted marker in payload, got %s", body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for redacted payload")
	}
}

func TestBuildResourceAttributesIncludesExtraAttrs(t *testing.T) {
	attrs := buildResourceAttributes("gochat-sfu", []attribute.KeyValue{
		attribute.String("voice.region", "eu-central"),
		attribute.String("service.instance.id", "sfu-eu-1"),
	})

	got := make(map[string]any, len(attrs))
	for _, attr := range attrs {
		got[string(attr.Key)] = attr.Value.AsInterface()
	}

	if got["service.name"] != "gochat-sfu" {
		t.Fatalf("missing service.name: %#v", got)
	}
	if got["deployment.environment"] == "" {
		t.Fatalf("missing deployment.environment: %#v", got)
	}
	if got["voice.region"] != "eu-central" {
		t.Fatalf("missing voice.region: %#v", got)
	}
	if got["service.instance.id"] != "sfu-eu-1" {
		t.Fatalf("missing service.instance.id: %#v", got)
	}
}

func TestSFUTelemetryUsesBaseAttrs(t *testing.T) {
	telemetry := NewSFUTelemetry("gochat-sfu",
		attribute.String("voice.region", "eu-central"),
		attribute.String("service.instance.id", "sfu-eu-1"),
	)

	attrs := telemetry.withBaseAttrs(attribute.String("status", "ok"))
	got := make(map[string]any, len(attrs))
	for _, attr := range attrs {
		got[string(attr.Key)] = attr.Value.AsInterface()
	}

	if got["voice.region"] != "eu-central" {
		t.Fatalf("missing voice.region: %#v", got)
	}
	if got["service.instance.id"] != "sfu-eu-1" {
		t.Fatalf("missing service.instance.id: %#v", got)
	}
	if got["status"] != "ok" {
		t.Fatalf("missing status attr: %#v", got)
	}
}
