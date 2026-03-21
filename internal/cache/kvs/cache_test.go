package kvs

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestFinishCacheOperationTreatsRedisNilAsCacheMiss(t *testing.T) {
	exporter := &spanCaptureExporter{}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)))

	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	defer otel.SetTracerProvider(previous)

	ctx, end := observability.StartDependencySpan(context.Background(), "redis", "get", "user")
	finishCacheOperation(ctx, end, redis.Nil)

	spans := exporter.spans()
	if len(spans) != 1 {
		t.Fatalf("expected one ended span, got %d", len(spans))
	}

	span := spans[0]
	if span.Status().Code == codes.Error {
		t.Fatalf("expected cache miss to avoid error status, got %#v", span.Status())
	}

	var hasMissAttr bool
	for _, attr := range span.Attributes() {
		if string(attr.Key) == "cache.miss" && attr.Value.AsBool() {
			hasMissAttr = true
		}
	}
	if !hasMissAttr {
		t.Fatalf("expected cache.miss attribute on span, got %#v", span.Attributes())
	}

	if len(span.Events()) == 0 || span.Events()[0].Name != "cache.miss" {
		t.Fatalf("expected cache.miss event, got %#v", span.Events())
	}
}

func TestFinishCacheOperationKeepsRealErrors(t *testing.T) {
	exporter := &spanCaptureExporter{}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter)))

	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	defer otel.SetTracerProvider(previous)

	ctx, end := observability.StartDependencySpan(context.Background(), "redis", "get", "user")
	realErr := errors.New("dial tcp timeout")
	finishCacheOperation(ctx, end, realErr)

	spans := exporter.spans()
	if len(spans) != 1 {
		t.Fatalf("expected one ended span, got %d", len(spans))
	}
	if spans[0].Status().Code != codes.Error {
		t.Fatalf("expected real redis error to keep error status, got %#v", spans[0].Status())
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
