package observability

import (
	"context"
	"errors"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestDependencyMetricsRecordMissResultWithoutErrorCounter(t *testing.T) {
	resetDependencyMetricsForTest()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(metricViews()...),
	)
	previous := otel.GetMeterProvider()
	otel.SetMeterProvider(provider)
	defer otel.SetMeterProvider(previous)

	ctx, end := StartDependencySpan(context.Background(), "redis", "get", "user")
	SetDependencyResult(ctx, "miss")
	end(nil)

	collected := collectMetricData(t, reader)
	calls := findSumMetric(t, collected, "gochat.dependency.calls")
	if len(calls.DataPoints) != 1 {
		t.Fatalf("expected one dependency call datapoint, got %d", len(calls.DataPoints))
	}
	attrs := attrSetToMap(calls.DataPoints[0].Attributes)
	if attrs["dependency.result"] != "miss" {
		t.Fatalf("expected dependency.result=miss, got %#v", attrs)
	}

	if metricExists(collected, "gochat.dependency.errors") {
		errorsMetric := findSumMetric(t, collected, "gochat.dependency.errors")
		if len(errorsMetric.DataPoints) != 0 {
			t.Fatalf("expected no dependency error datapoints for miss, got %#v", errorsMetric.DataPoints)
		}
	}
}

func TestDependencyMetricsRecordErrorResultForFailures(t *testing.T) {
	resetDependencyMetricsForTest()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(metricViews()...),
	)
	previous := otel.GetMeterProvider()
	otel.SetMeterProvider(provider)
	defer otel.SetMeterProvider(previous)

	_, end := StartDependencySpan(context.Background(), "redis", "get", "user")
	end(errors.New("dial tcp timeout"))

	collected := collectMetricData(t, reader)
	errorsMetric := findSumMetric(t, collected, "gochat.dependency.errors")
	if len(errorsMetric.DataPoints) != 1 {
		t.Fatalf("expected one dependency error datapoint, got %d", len(errorsMetric.DataPoints))
	}
	attrs := attrSetToMap(errorsMetric.DataPoints[0].Attributes)
	if attrs["dependency.result"] != "error" {
		t.Fatalf("expected dependency.result=error, got %#v", attrs)
	}
}

func resetDependencyMetricsForTest() {
	dependencyOnce = sync.Once{}
	dependencyInst = dependencyInstruments{}
}

func findSumMetric(t *testing.T, collected metricdata.ResourceMetrics, name string) metricdata.Sum[int64] {
	t.Helper()
	for _, scopeMetrics := range collected.ScopeMetrics {
		for _, metric := range scopeMetrics.Metrics {
			if metric.Name != name {
				continue
			}
			sum, ok := metric.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("metric %q was not an int64 sum: %#v", name, metric.Data)
			}
			return sum
		}
	}
	t.Fatalf("metric %q was not collected", name)
	return metricdata.Sum[int64]{}
}

func metricExists(collected metricdata.ResourceMetrics, name string) bool {
	for _, scopeMetrics := range collected.ScopeMetrics {
		for _, metric := range scopeMetrics.Metrics {
			if metric.Name == name {
				return true
			}
		}
	}
	return false
}
