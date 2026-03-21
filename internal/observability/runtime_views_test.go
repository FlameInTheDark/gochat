package observability

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestMetricViewsApplyShortLatencyBucketsAndDropHTTPStatusCode(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(metricViews()...),
	)

	histogram, err := provider.Meter("test").Float64Histogram("gochat.http.server.duration")
	if err != nil {
		t.Fatalf("Float64Histogram returned error: %v", err)
	}
	histogram.Record(context.Background(), 0.25, metric.WithAttributes(
		attribute.String("service", "gochat-auth"),
		attribute.String("http.method", "GET"),
		attribute.String("http.route", "/api/v1/auth/refresh"),
		attribute.Int("http.status_code", 200),
	))

	collected := collectMetricData(t, reader)
	hist := findHistogramMetric(t, collected, "gochat.http.server.duration")
	if len(hist.DataPoints) != 1 {
		t.Fatalf("expected one duration datapoint, got %d", len(hist.DataPoints))
	}
	if got := hist.DataPoints[0].Bounds; !floatSlicesEqual(got, shortLatencyHistogramBuckets) {
		t.Fatalf("unexpected short latency bounds: %#v", got)
	}

	attrs := attrSetToMap(hist.DataPoints[0].Attributes)
	if _, ok := attrs["http.status_code"]; ok {
		t.Fatalf("expected http.status_code to be filtered from duration attrs, got %#v", attrs)
	}
	if attrs["http.route"] != "/api/v1/auth/refresh" {
		t.Fatalf("expected http.route to be preserved, got %#v", attrs)
	}
}

func TestMetricViewsApplyLongLivedBuckets(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(metricViews()...),
	)

	histogram, err := provider.Meter("test").Float64Histogram("gochat.ws.connection.duration")
	if err != nil {
		t.Fatalf("Float64Histogram returned error: %v", err)
	}
	histogram.Record(context.Background(), 15, metric.WithAttributes(
		attribute.Bool("ws.compression", true),
		attribute.String("ignored", "value"),
	))

	collected := collectMetricData(t, reader)
	hist := findHistogramMetric(t, collected, "gochat.ws.connection.duration")
	if len(hist.DataPoints) != 1 {
		t.Fatalf("expected one ws connection datapoint, got %d", len(hist.DataPoints))
	}
	if got := hist.DataPoints[0].Bounds; !floatSlicesEqual(got, longLivedHistogramBuckets) {
		t.Fatalf("unexpected long-lived bounds: %#v", got)
	}

	attrs := attrSetToMap(hist.DataPoints[0].Attributes)
	if _, ok := attrs["ignored"]; ok {
		t.Fatalf("expected ignored attribute to be filtered, got %#v", attrs)
	}
	if attrs["ws.compression"] != true {
		t.Fatalf("expected ws.compression attr to be preserved, got %#v", attrs)
	}
}

func collectMetricData(t *testing.T, reader *sdkmetric.ManualReader) metricdata.ResourceMetrics {
	t.Helper()
	var collected metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &collected); err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	return collected
}

func findHistogramMetric(t *testing.T, collected metricdata.ResourceMetrics, name string) metricdata.Histogram[float64] {
	t.Helper()
	for _, scopeMetrics := range collected.ScopeMetrics {
		for _, metric := range scopeMetrics.Metrics {
			if metric.Name != name {
				continue
			}
			hist, ok := metric.Data.(metricdata.Histogram[float64])
			if !ok {
				t.Fatalf("metric %q was not a float64 histogram: %#v", name, metric.Data)
			}
			return hist
		}
	}
	t.Fatalf("metric %q was not collected", name)
	return metricdata.Histogram[float64]{}
}

func attrSetToMap(set attribute.Set) map[string]any {
	attrs := set.ToSlice()
	out := make(map[string]any, len(attrs))
	for _, attr := range attrs {
		out[string(attr.Key)] = attr.Value.AsInterface()
	}
	return out
}

func floatSlicesEqual(left, right []float64) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
