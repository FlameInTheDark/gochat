package observability

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

func TestNATSHeaderPropagationIncludesRequestID(t *testing.T) {
	ctx := helper.ContextWithRequestID(context.Background(), "req-42")
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    [16]byte{1, 2, 3, 4},
		SpanID:     [8]byte{5, 6, 7, 8},
		TraceFlags: trace.FlagsSampled,
		Remote:     false,
	})
	ctx = trace.ContextWithSpanContext(ctx, spanCtx)

	headers := InjectNATSHeaders(ctx, nil)
	if got := headers.Get(RequestIDHeader); got != "req-42" {
		t.Fatalf("unexpected request id header: %q", got)
	}
	if headers.Get("traceparent") == "" {
		t.Fatalf("expected traceparent header")
	}

	msg := &nats.Msg{Header: headers}
	out := ExtractNATSContext(context.Background(), msg)
	if got, ok := helper.RequestIDFromContext(out); !ok || got == "" {
		t.Fatalf("expected request id in extracted context")
	}
}
