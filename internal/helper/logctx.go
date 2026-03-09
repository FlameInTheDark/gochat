package helper

import (
	"context"
	"log/slog"
)

// ctxKey is an unexported type to avoid collisions with keys from other packages.
type ctxKey string

// Well-known context keys for logging correlation.
const (
	// ContextKeyRequestID is the key under which a request ID may be stored in context.
	// Use ContextWithRequestID to set it.
	ContextKeyRequestID ctxKey = "request_id"

	// ContextKeyUserID is an optional user identifier key. Not required, but commonly useful.
	ContextKeyUserID ctxKey = "user_id"
)

// ContextWithRequestID returns a new context with the provided request ID attached.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	if ctx == nil || id == "" {
		return ctx
	}
	return context.WithValue(ctx, ContextKeyRequestID, id)
}

// RequestIDFromContext extracts a request ID from context if present.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	if v := ctx.Value(ContextKeyRequestID); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s, true
		}
	}
	return "", false
}

// AttrsFromContext collects well-known attributes from context to attach to logs.
func AttrsFromContext(ctx context.Context) []slog.Attr {
	if ctx == nil {
		return nil
	}
	attrs := make([]slog.Attr, 0, 2)
	if rid, ok := RequestIDFromContext(ctx); ok {
		attrs = append(attrs, slog.String("request_id", rid))
	}
	if v := ctx.Value(ContextKeyUserID); v != nil {
		switch t := v.(type) {
		case int64:
			attrs = append(attrs, slog.Int64("user_id", t))
		case int:
			attrs = append(attrs, slog.Int("user_id", t))
		case string:
			if t != "" {
				attrs = append(attrs, slog.String("user_id", t))
			}
		}
	}
	return attrs
}

// WithContext returns a logger decorated with attributes extracted from context
// (e.g., request_id, user_id). If no attributes found, the original logger is returned.
func WithContext(log *slog.Logger, ctx context.Context) *slog.Logger {
	if log == nil || ctx == nil {
		return log
	}
	attrs := AttrsFromContext(ctx)
	if len(attrs) == 0 {
		return log
	}
	// Convert []slog.Attr to []any to satisfy slog.Logger.With signature
	args := make([]any, 0, len(attrs))
	for _, a := range attrs {
		args = append(args, a)
	}
	return log.With(args...)
}
