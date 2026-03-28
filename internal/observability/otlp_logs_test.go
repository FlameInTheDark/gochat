package observability

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FlameInTheDark/gochat/internal/helper"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	collectorpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	"google.golang.org/protobuf/proto"
)

func TestOTLPLogExporterBatchesStructuredLogs(t *testing.T) {
	type requestPayload struct {
		path string
		auth string
		body []byte
	}

	requests := make(chan requestPayload, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		requests <- requestPayload{
			path: r.URL.Path,
			auth: r.Header.Get("Authorization"),
			body: body,
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	exporter, err := newOTLPLogExporter("gochat-sfu", otlpLogExporterConfig{
		endpoint: server.URL + "/v1/logs",
		headers: map[string]string{
			"Authorization": "Bearer example",
		},
		resourceAttrs: []attribute.KeyValue{
			attribute.String("service.name", "gochat-sfu"),
			attribute.String("voice.region", "eu-central"),
		},
		scopeName:      "gochat/logs",
		queueSize:      8,
		batchSize:      2,
		flushInterval:  20 * time.Millisecond,
		requestTimeout: time.Second,
		retryBackoff:   10 * time.Millisecond,
		maxAttempts:    1,
	})
	if err != nil {
		t.Fatalf("newOTLPLogExporter: %v", err)
	}
	defer exporter.Close()

	logger := slog.New(&contextualHandler{
		next:               newOTLPLogHandler(exporter),
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
		if req.path != "/v1/logs" {
			t.Fatalf("unexpected request path: %s", req.path)
		}
		if req.auth != "Bearer example" {
			t.Fatalf("unexpected auth header: %q", req.auth)
		}

		var payload collectorpb.ExportLogsServiceRequest
		if err := proto.Unmarshal(req.body, &payload); err != nil {
			t.Fatalf("proto.Unmarshal: %v", err)
		}
		if len(payload.ResourceLogs) != 1 {
			t.Fatalf("expected one resource logs batch, got %d", len(payload.ResourceLogs))
		}
		resourceLogs := payload.ResourceLogs[0]
		if len(resourceLogs.ScopeLogs) != 1 {
			t.Fatalf("expected one scope logs batch, got %d", len(resourceLogs.ScopeLogs))
		}
		logRecords := resourceLogs.ScopeLogs[0].LogRecords
		if len(logRecords) != 2 {
			t.Fatalf("expected two log records, got %d", len(logRecords))
		}
		if got := logRecords[0].Body.GetStringValue(); got != "peer joined" {
			t.Fatalf("unexpected first log body: %q", got)
		}
		if got := logRecords[0].SeverityText; got != "INFO" {
			t.Fatalf("unexpected first severity: %q", got)
		}
		if got := logRecords[1].SeverityText; got != "WARN" {
			t.Fatalf("unexpected second severity: %q", got)
		}
		attrs := keyValuesToMap(logRecords[0].Attributes)
		if attrs["request_id"] != "req-123" {
			t.Fatalf("missing request_id attr: %#v", attrs)
		}
		if attrs["service.instance.id"] != "sfu-eu-1" {
			t.Fatalf("missing service.instance.id attr: %#v", attrs)
		}
		if len(logRecords[0].TraceId) != 16 || len(logRecords[0].SpanId) != 8 {
			t.Fatalf("expected trace/span ids to be populated, got trace=%x span=%x", logRecords[0].TraceId, logRecords[0].SpanId)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for OTLP log request")
	}
}

func keyValuesToMap(values []*commonpb.KeyValue) map[string]any {
	out := make(map[string]any, len(values))
	for _, kv := range values {
		if kv == nil || kv.Value == nil {
			continue
		}
		switch value := kv.Value.Value.(type) {
		case *commonpb.AnyValue_StringValue:
			out[kv.Key] = value.StringValue
		case *commonpb.AnyValue_IntValue:
			out[kv.Key] = value.IntValue
		case *commonpb.AnyValue_DoubleValue:
			out[kv.Key] = value.DoubleValue
		case *commonpb.AnyValue_BoolValue:
			out[kv.Key] = value.BoolValue
		default:
			out[kv.Key] = kv.Value.String()
		}
	}
	return out
}
