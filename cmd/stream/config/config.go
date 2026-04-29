package config

import (
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/configutil"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServerAddress         string   `yaml:"server_address" env-default:":3300"`
	AuthSecret            string   `yaml:"auth_secret" env:"AUTH_SECRET"`
	AuthSecretEnforcement string   `yaml:"auth_secret_enforcement" env:"AUTH_SECRET_ENFORCEMENT" env-default:"warn"`
	STUNServers           []string `yaml:"stun_servers" env:"STUN_SERVERS" env-separator:"," env-default:"stun:stun.l.google.com:19302"`
	Region                string   `yaml:"region" env:"STREAM_REGION" env-default:"global"`
	PublicBaseURL         string   `yaml:"public_base_url" env:"STREAM_PUBLIC_BASE_URL" env-required:"true"`
	ICEPublicIP           string   `yaml:"ice_public_ip" env:"STREAM_ICE_PUBLIC_IP"`
	// Optional DTLS certificate/key pair for WebRTC peer connections. Provide
	// either the PEM values directly or file paths for both halves of the pair.
	// When left empty, the SFU generates a self-signed certificate at startup.
	DTLSCertificatePEM  string `yaml:"dtls_certificate_pem" env:"STREAM_DTLS_CERTIFICATE_PEM"`
	DTLSPrivateKeyPEM   string `yaml:"dtls_private_key_pem" env:"STREAM_DTLS_PRIVATE_KEY_PEM"`
	DTLSCertificateFile string `yaml:"dtls_certificate_file" env:"STREAM_DTLS_CERTIFICATE_FILE"`
	DTLSPrivateKeyFile  string `yaml:"dtls_private_key_file" env:"STREAM_DTLS_PRIVATE_KEY_FILE"`
	// Discovery
	WebhookURL   string `yaml:"webhook_url" env:"WEBHOOK_URL" env-required:"true"`
	WebhookToken string `yaml:"webhook_token" env:"WEBHOOK_TOKEN" env-required:"true"`
	ServiceID    string `yaml:"service_id" env:"STREAM_SERVICE_ID" env-required:"true"`
	// Telemetry configuration for the external OTLP gateway. These values are
	// projected back into the standard OTEL env vars before observability init.
	TelemetryOTLPEndpoint         string `yaml:"telemetry_otlp_endpoint" env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	TelemetryOTLPHeaders          string `yaml:"telemetry_otlp_headers" env:"OTEL_EXPORTER_OTLP_HEADERS"`
	TelemetryOTLPProtocol         string `yaml:"telemetry_otlp_protocol" env:"OTEL_EXPORTER_OTLP_PROTOCOL" env-default:"http/protobuf"`
	TelemetryMetricExportInterval string `yaml:"telemetry_metric_export_interval" env:"OTEL_METRIC_EXPORT_INTERVAL"`

	// SignalHeartbeatIntervalMS is used by the v2 `/signal?v=2` protocol hello.
	SignalHeartbeatIntervalMS int64 `yaml:"signal_heartbeat_interval_ms" env:"STREAM_SIGNAL_HEARTBEAT_INTERVAL_MS" env-default:"15000"`

	// DAVE / E2EE controls for the v2 voice gateway.
	DAVEEnabled             bool  `yaml:"dave_enabled" env:"STREAM_DAVE_ENABLED" env-default:"true"`
	DAVERequiredDefault     bool  `yaml:"dave_required_default" env:"STREAM_DAVE_REQUIRED_DEFAULT" env-default:"false"`
	DAVETransitionTimeoutMS int64 `yaml:"dave_transition_timeout_ms" env:"STREAM_DAVE_TRANSITION_TIMEOUT_MS" env-default:"2000"`
	DAVEOldRatchetWindowMS  int64 `yaml:"dave_old_ratchet_window_ms" env:"STREAM_DAVE_OLD_RATCHET_WINDOW_MS" env-default:"10000"`
	DAVEAllowAV1            bool  `yaml:"dave_allow_av1" env:"STREAM_DAVE_ALLOW_AV1" env-default:"false"`

	// Media limits
	// MaxAudioBitrateKbps, when > 0, injects SDP constraints to cap OPUS
	// encoder average bitrate on clients and advertises bandwidth limits per
	// RFC7587 (maxaveragebitrate) and RFC3890 (b=AS/TIAS).
	// Example values: 64, 96, 128. 0 disables limiting.
	MaxAudioBitrateKbps int  `yaml:"max_audio_bitrate_kbps" env:"STREAM_MAX_AUDIO_BITRATE_KBPS" env-default:"0"`
	EnforceAudioBitrate bool `yaml:"enforce_audio_bitrate" env:"STREAM_ENFORCE_AUDIO_BITRATE" env-default:"false"`

	// MaxVideoBitrateKbps, when > 0, advertises SDP bandwidth limits for video.
	// It is intentionally a high ceiling, not an enforcement path, so browsers
	// can start at high quality and WebRTC congestion control can downshift.
	// Example values: 30000 for 1080p60, 60000 for 1440p60, 100000 for 4K60.
	MaxVideoBitrateKbps int `yaml:"max_video_bitrate_kbps" env:"STREAM_MAX_VIDEO_BITRATE_KBPS" env-default:"0"`

	// AudioBitrateMarginPercent adds tolerance to measured inbound RTP bitrate
	// during enforcement to account for headers/Jitter/overhead. E.g. 15 means
	// 15% over the configured cap is tolerated before disconnect. Range 0..100.
	AudioBitrateMarginPercent int `yaml:"audio_bitrate_margin_percent" env:"STREAM_AUDIO_BITRATE_MARGIN_PERCENT" env-default:"15"`

	// Optional UDP port range for WebRTC transports. Leave both as 0 to keep the
	// current OS-managed ephemeral port behavior unchanged.
	UDPPortRangeStart int `yaml:"udp_port_range_start" env:"STREAM_UDP_PORT_RANGE_START" env-default:"0"`
	UDPPortRangeEnd   int `yaml:"udp_port_range_end" env:"STREAM_UDP_PORT_RANGE_END" env-default:"0"`
}

func LoadConfig() (*Config, error) {
	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		path = "./config.yaml"
	}
	var cfg Config
	if _, err := os.Stat(path); err == nil {
		if rerr := cleanenv.ReadConfig(path, &cfg); rerr != nil {
			return nil, rerr
		}
		cfg.applyAuthSecretMigrationFallback()
		if verr := cfg.Validate(); verr != nil {
			return nil, verr
		}
		if err := configutil.ValidateAuthSecretWithMode(cfg.AuthSecret, cfg.AuthSecretEnforcement); err != nil {
			return nil, err
		}
		configutil.WarnWeakAuthSecret(slog.Default(), cfg.AuthSecret, "stream")
		return &cfg, nil
	}
	if rerr := cleanenv.ReadEnv(&cfg); rerr != nil {
		return nil, rerr
	}
	cfg.applyAuthSecretMigrationFallback()
	if verr := cfg.Validate(); verr != nil {
		return nil, verr
	}
	if err := configutil.ValidateAuthSecretWithMode(cfg.AuthSecret, cfg.AuthSecretEnforcement); err != nil {
		return nil, err
	}
	configutil.WarnWeakAuthSecret(slog.Default(), cfg.AuthSecret, "stream")
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c == nil {
		return nil
	}
	if ip := strings.TrimSpace(c.ICEPublicIP); ip != "" {
		if _, err := netip.ParseAddr(ip); err != nil {
			return fmt.Errorf("ice_public_ip must be a valid IP address")
		}
	}
	if err := c.validateDTLSConfig(); err != nil {
		return err
	}
	if c.MaxAudioBitrateKbps < 0 {
		return fmt.Errorf("max_audio_bitrate_kbps must be greater than or equal to 0")
	}
	if c.MaxVideoBitrateKbps < 0 {
		return fmt.Errorf("max_video_bitrate_kbps must be greater than or equal to 0")
	}
	start, end := c.UDPPortRangeStart, c.UDPPortRangeEnd
	if start == 0 && end == 0 {
		return nil
	}
	if start == 0 || end == 0 {
		return fmt.Errorf("udp_port_range_start and udp_port_range_end must both be set")
	}
	if start < 1 || start > 65535 || end < 1 || end > 65535 {
		return fmt.Errorf("udp port range must be within 1..65535")
	}
	if start > end {
		return fmt.Errorf("udp_port_range_start must be less than or equal to udp_port_range_end")
	}
	return nil
}

func (c *Config) applyAuthSecretMigrationFallback() {
	if c == nil || strings.TrimSpace(c.AuthSecret) != "" {
		return
	}
	c.AuthSecret = strings.TrimSpace(os.Getenv("STREAM_AUTH_SECRET"))
}

func (c *Config) validateDTLSConfig() error {
	inlineCert := strings.TrimSpace(c.DTLSCertificatePEM)
	inlineKey := strings.TrimSpace(c.DTLSPrivateKeyPEM)
	fileCert := strings.TrimSpace(c.DTLSCertificateFile)
	fileKey := strings.TrimSpace(c.DTLSPrivateKeyFile)

	hasInline := inlineCert != "" || inlineKey != ""
	hasFiles := fileCert != "" || fileKey != ""

	if hasInline && hasFiles {
		return fmt.Errorf("configure dtls certificate using either pem values or file paths, not both")
	}
	if hasInline && (inlineCert == "" || inlineKey == "") {
		return fmt.Errorf("dtls_certificate_pem and dtls_private_key_pem must both be set")
	}
	if hasFiles && (fileCert == "" || fileKey == "") {
		return fmt.Errorf("dtls_certificate_file and dtls_private_key_file must both be set")
	}

	return nil
}

func (c *Config) ApplyObservabilityEnv() error {
	if c == nil {
		return nil
	}

	if err := setEnvIfMissing("OTEL_EXPORTER_OTLP_ENDPOINT", c.TelemetryOTLPEndpoint); err != nil {
		return err
	}
	if err := setEnvIfMissing("OTEL_EXPORTER_OTLP_HEADERS", c.TelemetryOTLPHeaders); err != nil {
		return err
	}
	if err := setEnvIfMissing("OTEL_EXPORTER_OTLP_PROTOCOL", c.TelemetryOTLPProtocol); err != nil {
		return err
	}
	if err := setEnvIfMissing("OTEL_METRIC_EXPORT_INTERVAL", c.TelemetryMetricExportInterval); err != nil {
		return err
	}
	if strings.TrimSpace(c.TelemetryOTLPEndpoint) != "" {
		if err := setEnvIfMissing("OTEL_LOGS_EXPORTER", "otlp"); err != nil {
			return err
		}
	}
	return nil
}

func setEnvIfMissing(key, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if _, exists := os.LookupEnv(key); exists {
		return nil
	}
	return os.Setenv(key, value)
}
