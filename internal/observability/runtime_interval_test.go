package observability

import (
	"testing"
	"time"
)

func TestMetricExportIntervalDefaultsToSixtySeconds(t *testing.T) {
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "")

	if got := metricExportInterval(); got != 60*time.Second {
		t.Fatalf("metricExportInterval() = %v, want %v", got, 60*time.Second)
	}
}

func TestMetricExportIntervalAcceptsMilliseconds(t *testing.T) {
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "30000")

	if got := metricExportInterval(); got != 30*time.Second {
		t.Fatalf("metricExportInterval() = %v, want %v", got, 30*time.Second)
	}
}

func TestMetricExportIntervalAcceptsDurationString(t *testing.T) {
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "45s")

	if got := metricExportInterval(); got != 45*time.Second {
		t.Fatalf("metricExportInterval() = %v, want %v", got, 45*time.Second)
	}
}

func TestMetricExportIntervalFallsBackOnInvalidValue(t *testing.T) {
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "nope")

	if got := metricExportInterval(); got != 60*time.Second {
		t.Fatalf("metricExportInterval() = %v, want %v", got, 60*time.Second)
	}
}
