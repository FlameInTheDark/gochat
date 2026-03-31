package observability

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

type dependencyInstruments struct {
	calls    metric.Int64Counter
	errors   metric.Int64Counter
	duration metric.Float64Histogram
}

var (
	dependencyOnce sync.Once
	dependencyInst dependencyInstruments
)

type dependencyStateKey struct{}

type dependencyState struct {
	result string
	mu     sync.RWMutex
}

func dependencyMetrics() dependencyInstruments {
	dependencyOnce.Do(func() {
		meter := otel.Meter("gochat/dependencies")
		dependencyInst.calls, _ = meter.Int64Counter("gochat.dependency.calls")
		dependencyInst.errors, _ = meter.Int64Counter("gochat.dependency.errors")
		dependencyInst.duration, _ = meter.Float64Histogram("gochat.dependency.duration")
	})
	return dependencyInst
}

func StartDependencySpan(ctx context.Context, system, operation, target string, attrs ...attribute.KeyValue) (context.Context, func(error)) {
	if ctx == nil {
		ctx = context.Background()
	}
	inst := dependencyMetrics()
	metricAttrs := []attribute.KeyValue{
		attribute.String("dependency.system", system),
		attribute.String("dependency.operation", operation),
	}
	spanAttrs := make([]attribute.KeyValue, 0, len(attrs)+4)
	spanAttrs = append(spanAttrs, metricAttrs...)
	if target != "" {
		spanAttrs = append(spanAttrs, attribute.String("dependency.target", target))
	}
	spanAttrs = append(spanAttrs, attrs...)

	state := &dependencyState{}
	ctx = context.WithValue(ctx, dependencyStateKey{}, state)
	ctx, span := otel.Tracer("gochat/dependencies").Start(ctx, fmt.Sprintf("%s %s", system, operation))
	span.SetAttributes(spanAttrs...)
	start := time.Now()

	return ctx, func(err error) {
		result := state.resultValue(err)
		finalMetricAttrs := append(metricAttrs[:len(metricAttrs):len(metricAttrs)], attribute.String("dependency.result", result))
		duration := time.Since(start).Seconds()
		inst.calls.Add(ctx, 1, metric.WithAttributes(finalMetricAttrs...))
		inst.duration.Record(ctx, duration, metric.WithAttributes(finalMetricAttrs...))
		span.SetAttributes(attribute.String("dependency.result", result))
		if err != nil && result == "error" {
			inst.errors.Add(ctx, 1, metric.WithAttributes(finalMetricAttrs...))
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, result)
		}
		span.End()
	}
}

func SetDependencyResult(ctx context.Context, result string) {
	if ctx == nil {
		return
	}
	state, _ := ctx.Value(dependencyStateKey{}).(*dependencyState)
	if state == nil {
		return
	}
	state.setResult(result)
}

func (s *dependencyState) setResult(result string) {
	if s == nil {
		return
	}
	result = normalizeDependencyResult(result)
	if result == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.result = result
}

func (s *dependencyState) resultValue(err error) string {
	if s == nil {
		if err != nil {
			return "error"
		}
		return "ok"
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.result != "" {
		return s.result
	}
	if err != nil {
		return "error"
	}
	return "ok"
}

func normalizeDependencyResult(result string) string {
	result = strings.TrimSpace(strings.ToLower(result))
	switch result {
	case "ok", "miss", "error":
		return result
	default:
		return ""
	}
}
