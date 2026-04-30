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
		"OTEL_METRIC_EXPORT_INTERVAL",
		"OTEL_LOGS_EXPORTER",
	)

	cfg := &Config{
		TelemetryOTLPEndpoint:         "https://telemetry.example.com",
		TelemetryOTLPHeaders:          "Authorization=Bearer example",
		TelemetryOTLPProtocol:         "http/protobuf",
		TelemetryMetricExportInterval: "60000",
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
	if got := getenvOrEmpty("OTEL_METRIC_EXPORT_INTERVAL"); got != "60000" {
		t.Fatalf("OTEL_METRIC_EXPORT_INTERVAL = %q", got)
	}
	if got := getenvOrEmpty("OTEL_LOGS_EXPORTER"); got != "otlp" {
		t.Fatalf("OTEL_LOGS_EXPORTER = %q", got)
	}
}

func TestApplyObservabilityEnvPreservesExplicitEnv(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://override.example.com")
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Bearer override")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "30000")
	t.Setenv("OTEL_LOGS_EXPORTER", "none")

	cfg := &Config{
		TelemetryOTLPEndpoint:         "https://telemetry.example.com",
		TelemetryOTLPHeaders:          "Authorization=Bearer example",
		TelemetryOTLPProtocol:         "http/protobuf",
		TelemetryMetricExportInterval: "60000",
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
	if got := getenvOrEmpty("OTEL_METRIC_EXPORT_INTERVAL"); got != "30000" {
		t.Fatalf("OTEL_METRIC_EXPORT_INTERVAL = %q", got)
	}
	if got := getenvOrEmpty("OTEL_LOGS_EXPORTER"); got != "none" {
		t.Fatalf("OTEL_LOGS_EXPORTER = %q", got)
	}
}

func TestValidateAcceptsDefaultUDPPortRange(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidateRejectsPartialUDPPortRange(t *testing.T) {
	cfg := &Config{UDPPortRangeStart: 40000}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected partial udp port range to fail validation")
	}
}

func TestValidateRejectsReversedUDPPortRange(t *testing.T) {
	cfg := &Config{UDPPortRangeStart: 40100, UDPPortRangeEnd: 40000}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected reversed udp port range to fail validation")
	}
}

func TestValidateAcceptsICEPublicIP(t *testing.T) {
	cfg := &Config{ICEPublicIP: "203.0.113.10"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidateRejectsInvalidICEPublicIP(t *testing.T) {
	cfg := &Config{ICEPublicIP: "not-an-ip"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid ice public ip to fail validation")
	}
}

func TestValidateRejectsPartialInlineDTLSConfig(t *testing.T) {
	cfg := &Config{DTLSCertificatePEM: "-----BEGIN CERTIFICATE-----\n..."}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected partial inline dtls config to fail validation")
	}
}

func TestValidateRejectsPartialDTLSFileConfig(t *testing.T) {
	cfg := &Config{DTLSCertificateFile: "server.crt"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected partial dtls file config to fail validation")
	}
}

func TestValidateRejectsMixedDTLSConfigSources(t *testing.T) {
	cfg := &Config{
		DTLSCertificatePEM:  "-----BEGIN CERTIFICATE-----\n...",
		DTLSPrivateKeyPEM:   "-----BEGIN PRIVATE KEY-----\n...",
		DTLSCertificateFile: "server.crt",
		DTLSPrivateKeyFile:  "server.key",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mixed dtls config sources to fail validation")
	}
}

func TestValidateAcceptsInlineDTLSConfig(t *testing.T) {
	cfg := &Config{
		DTLSCertificatePEM: "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
		DTLSPrivateKeyPEM:  "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidateAcceptsDTLSFileConfig(t *testing.T) {
	cfg := &Config{
		DTLSCertificateFile: "server.crt",
		DTLSPrivateKeyFile:  "server.key",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
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
