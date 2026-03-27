package config

import (
	"os"
	"testing"
)

func TestApplyObservabilityEnvSetsOTLPVarsFromConfig(t *testing.T) {
	unsetEnvForTest(t,
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_HEADERS",
		"OTEL_EXPORTER_OTLP_PROTOCOL",
		"OTEL_LOGS_EXPORTER",
	)

	cfg := &Config{
		TelemetryOTLPEndpoint: "https://telemetry.example.com",
		TelemetryOTLPHeaders:  "Authorization=Bearer example",
		TelemetryOTLPProtocol: "http/protobuf",
	}
	if err := cfg.ApplyObservabilityEnv(); err != nil {
		t.Fatalf("ApplyObservabilityEnv: %v", err)
	}

	if got := getenvOrEmpty("OTEL_EXPORTER_OTLP_ENDPOINT"); got != "https://telemetry.example.com" {
		t.Fatalf("OTEL_EXPORTER_OTLP_ENDPOINT = %q", got)
	}
	if got := getenvOrEmpty("OTEL_EXPORTER_OTLP_HEADERS"); got != "Authorization=Bearer example" {
		t.Fatalf("OTEL_EXPORTER_OTLP_HEADERS = %q", got)
	}
	if got := getenvOrEmpty("OTEL_EXPORTER_OTLP_PROTOCOL"); got != "http/protobuf" {
		t.Fatalf("OTEL_EXPORTER_OTLP_PROTOCOL = %q", got)
	}
	if got := getenvOrEmpty("OTEL_LOGS_EXPORTER"); got != "otlp" {
		t.Fatalf("OTEL_LOGS_EXPORTER = %q", got)
	}
}

func TestApplyObservabilityEnvPreservesExplicitEnv(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://override.example.com")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Bearer override")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_LOGS_EXPORTER", "none")

	cfg := &Config{
		TelemetryOTLPEndpoint: "https://telemetry.example.com",
		TelemetryOTLPHeaders:  "Authorization=Bearer example",
		TelemetryOTLPProtocol: "http/protobuf",
	}
	if err := cfg.ApplyObservabilityEnv(); err != nil {
		t.Fatalf("ApplyObservabilityEnv: %v", err)
	}

	if got := getenvOrEmpty("OTEL_EXPORTER_OTLP_ENDPOINT"); got != "https://override.example.com" {
		t.Fatalf("OTEL_EXPORTER_OTLP_ENDPOINT = %q", got)
	}
	if got := getenvOrEmpty("OTEL_EXPORTER_OTLP_HEADERS"); got != "Authorization=Bearer override" {
		t.Fatalf("OTEL_EXPORTER_OTLP_HEADERS = %q", got)
	}
	if got := getenvOrEmpty("OTEL_LOGS_EXPORTER"); got != "none" {
		t.Fatalf("OTEL_LOGS_EXPORTER = %q", got)
	}
}

func getenvOrEmpty(key string) string {
	return os.Getenv(key)
}

func unsetEnvForTest(t *testing.T, keys ...string) {
	t.Helper()

	snapshots := make(map[string]struct {
		value string
		ok    bool
	}, len(keys))
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		snapshots[key] = struct {
			value string
			ok    bool
		}{value: value, ok: ok}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%q): %v", key, err)
		}
	}

	t.Cleanup(func() {
		for key, snapshot := range snapshots {
			if snapshot.ok {
				_ = os.Setenv(key, snapshot.value)
				continue
			}
			_ = os.Unsetenv(key)
		}
	})
}
