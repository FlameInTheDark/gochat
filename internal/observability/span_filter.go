package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// filteringSpanProcessor wraps a SpanProcessor and drops spans whose
// severity is below minLevel:
//
//	LevelInfo  → all spans pass through
//	LevelWarn  → 4xx, 5xx, and spans with codes.Error
//	LevelError → only spans with codes.Error (5xx / explicit errors)
type filteringSpanProcessor struct {
	next     sdktrace.SpanProcessor
	minLevel slog.Level
}

// newFilteringSpanProcessor returns next unwrapped when minLevel is Info
// (nothing to filter), otherwise wraps it.
func newFilteringSpanProcessor(next sdktrace.SpanProcessor, minLevel slog.Level) sdktrace.SpanProcessor {
	if minLevel <= slog.LevelInfo {
		return next
	}
	return &filteringSpanProcessor{next: next, minLevel: minLevel}
}

func (p *filteringSpanProcessor) OnStart(parent context.Context, s sdktrace.ReadWriteSpan) {
	p.next.OnStart(parent, s)
}

func (p *filteringSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	if p.shouldExport(s) {
		p.next.OnEnd(s)
	}
}

func (p *filteringSpanProcessor) Shutdown(ctx context.Context) error {
	return p.next.Shutdown(ctx)
}

func (p *filteringSpanProcessor) ForceFlush(ctx context.Context) error {
	return p.next.ForceFlush(ctx)
}

func (p *filteringSpanProcessor) shouldExport(s sdktrace.ReadOnlySpan) bool {
	// codes.Error is set for 5xx — always export at any non-info level.
	if s.Status().Code == codes.Error {
		return true
	}
	// warn level also exports 4xx: look for the HTTP status attribute.
	if p.minLevel <= slog.LevelWarn {
		for _, attr := range s.Attributes() {
			if attr.Key == semconv.HTTPResponseStatusCodeKey {
				return attr.Value.AsInt64() >= 400
			}
		}
	}
	return false
}
