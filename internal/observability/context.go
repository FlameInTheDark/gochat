package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

func BackgroundFromContext(ctx context.Context) context.Context {
	out := context.Background()
	if ctx == nil {
		return out
	}

	if rid, ok := helper.RequestIDFromContext(ctx); ok {
		out = helper.ContextWithRequestID(out, rid)
	}
	if userID, ok := helper.UserIDFromContext(ctx); ok {
		out = helper.ContextWithUserID(out, userID)
	}
	if bag := baggage.FromContext(ctx); bag.Len() > 0 {
		out = baggage.ContextWithBaggage(out, bag)
	}
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		out = trace.ContextWithSpanContext(out, spanCtx)
	}
	return out
}

func RequestIDFromContext(ctx context.Context) string {
	if rid, ok := helper.RequestIDFromContext(ctx); ok {
		return rid
	}
	return ""
}

func generateRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}
