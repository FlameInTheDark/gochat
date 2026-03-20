package observability

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

func TestBackgroundFromContextPreservesCorrelationFields(t *testing.T) {
	ctx := helper.ContextWithRequestID(context.Background(), "req-42")
	ctx = helper.ContextWithUserID(ctx, 77)

	member, err := baggage.NewMember("region", "eu-west")
	if err != nil {
		t.Fatalf("unexpected baggage member error: %v", err)
	}
	bag, err := baggage.New(member)
	if err != nil {
		t.Fatalf("unexpected baggage creation error: %v", err)
	}
	ctx = baggage.ContextWithBaggage(ctx, bag)

	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    [16]byte{1, 2, 3, 4},
		SpanID:     [8]byte{5, 6, 7, 8},
		TraceFlags: trace.FlagsSampled,
	})
	ctx = trace.ContextWithSpanContext(ctx, spanCtx)

	out := BackgroundFromContext(ctx)

	if got, ok := helper.RequestIDFromContext(out); !ok || got != "req-42" {
		t.Fatalf("unexpected request id after background copy: %q ok=%v", got, ok)
	}
	if got, ok := helper.UserIDFromContext(out); !ok || got != 77 {
		t.Fatalf("unexpected user id after background copy: %d ok=%v", got, ok)
	}
	if got := baggage.FromContext(out).Member("region").Value(); got != "eu-west" {
		t.Fatalf("unexpected baggage value after background copy: %q", got)
	}
	if got := trace.SpanContextFromContext(out); got.TraceID() != spanCtx.TraceID() || got.SpanID() != spanCtx.SpanID() {
		t.Fatalf("unexpected span context after background copy: %+v", got)
	}
}
