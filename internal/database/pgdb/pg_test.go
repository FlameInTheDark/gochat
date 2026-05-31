package pgdb

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestRunProbeRecordsStatusTransitionsAndCounters(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	prevProvider := otel.GetMeterProvider()
	otel.SetMeterProvider(provider)
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
		otel.SetMeterProvider(prevProvider)
	})

	results := []error{errors.New("db unavailable"), nil}
	db := &DB{
		pingFn: func(ctx context.Context) error {
			if len(results) == 0 {
				return nil
			}
			err := results[0]
			results = results[1:]
			return err
		},
	}

	if err := db.runProbe(context.Background(), 50*time.Millisecond); err == nil {
		t.Fatal("expected first probe to fail")
	}
	if err := db.runProbe(context.Background(), 50*time.Millisecond); err != nil {
		t.Fatalf("expected second probe to succeed, got %v", err)
	}

	rm := collectResourceMetrics(t, reader)

	if got := sumIntMetricValue(t, rm, "gochat.postgres.probe.success"); got != 1 {
		t.Fatalf("expected 1 success probe, got %d", got)
	}
	if got := sumIntMetricValue(t, rm, "gochat.postgres.probe.failure"); got != 1 {
		t.Fatalf("expected 1 failed probe, got %d", got)
	}
	if got := gaugeIntMetricValue(t, rm, "gochat.postgres.probe.status"); got != 1 {
		t.Fatalf("expected final probe status to be 1, got %d", got)
	}
	if got := gaugeIntMetricValue(t, rm, "gochat.postgres.probe.last_success_unix"); got <= 0 {
		t.Fatalf("expected last success unix timestamp to be recorded, got %d", got)
	}
	if got := histogramCountValue(t, rm, "gochat.postgres.probe.duration"); got != 2 {
		t.Fatalf("expected 2 duration records, got %d", got)
	}
}

func TestStartProbeLoopRunsImmediateProbe(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	prevProvider := otel.GetMeterProvider()
	otel.SetMeterProvider(provider)
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
		otel.SetMeterProvider(prevProvider)
	})

	var calls atomic.Int64
	db := &DB{
		pingFn: func(ctx context.Context) error {
			calls.Add(1)
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go db.StartProbeLoop(ctx, time.Hour)

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if calls.Load() > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if calls.Load() == 0 {
		t.Fatal("expected StartProbeLoop to run an immediate probe")
	}
	cancel()

	rm := collectResourceMetrics(t, reader)
	if got := sumIntMetricValue(t, rm, "gochat.postgres.probe.success"); got != 1 {
		t.Fatalf("expected 1 successful immediate probe, got %d", got)
	}
}

func TestWithPGXExecMode(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want string
	}{
		{
			name: "keyword dsn",
			dsn:  "host=yugabyte port=5433 user=yugabyte password=yugabyte dbname=gochat sslmode=disable",
			want: "host=yugabyte port=5433 user=yugabyte password=yugabyte dbname=gochat sslmode=disable default_query_exec_mode=exec",
		},
		{
			name: "keeps explicit mode",
			dsn:  "host=yugabyte port=5433 default_query_exec_mode=simple_protocol",
			want: "host=yugabyte port=5433 default_query_exec_mode=simple_protocol",
		},
		{
			name: "url dsn",
			dsn:  "postgres://user:pass@yugabyte:5433/gochat?sslmode=disable",
			want: "postgres://user:pass@yugabyte:5433/gochat?default_query_exec_mode=exec&sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := withPGXExecMode(tt.dsn); got != tt.want {
				t.Fatalf("withPGXExecMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func collectResourceMetrics(t *testing.T, reader *sdkmetric.ManualReader) metricdata.ResourceMetrics {
	t.Helper()

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}
	return rm
}

func sumIntMetricValue(t *testing.T, rm metricdata.ResourceMetrics, name string) int64 {
	t.Helper()

	for _, scope := range rm.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name != name {
				continue
			}
			sum, ok := metric.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("metric %s is not an int64 sum", name)
			}
			var total int64
			for _, point := range sum.DataPoints {
				total += point.Value
			}
			return total
		}
	}
	t.Fatalf("metric %s not found", name)
	return 0
}

func gaugeIntMetricValue(t *testing.T, rm metricdata.ResourceMetrics, name string) int64 {
	t.Helper()

	for _, scope := range rm.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name != name {
				continue
			}
			gauge, ok := metric.Data.(metricdata.Gauge[int64])
			if !ok {
				t.Fatalf("metric %s is not an int64 gauge", name)
			}
			if len(gauge.DataPoints) == 0 {
				t.Fatalf("metric %s has no datapoints", name)
			}
			return gauge.DataPoints[0].Value
		}
	}
	t.Fatalf("metric %s not found", name)
	return 0
}

func histogramCountValue(t *testing.T, rm metricdata.ResourceMetrics, name string) uint64 {
	t.Helper()

	for _, scope := range rm.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name != name {
				continue
			}
			histogram, ok := metric.Data.(metricdata.Histogram[float64])
			if !ok {
				t.Fatalf("metric %s is not a float64 histogram", name)
			}
			var count uint64
			for _, point := range histogram.DataPoints {
				count += point.Count
			}
			return count
		}
	}
	t.Fatalf("metric %s not found", name)
	return 0
}
