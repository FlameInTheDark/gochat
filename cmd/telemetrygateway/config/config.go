package config

import (
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Log           bool   `yaml:"api_log" env:"API_LOG" env-default:"true"`
	LogLevel      string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
	ServerAddress string `yaml:"server_address" env-default:":4318"`
	JWTSecret     string `yaml:"jwt_secret" env:"WEBHOOK_JWT_SECRET"`

	UpstreamEndpoint   string  `yaml:"upstream_endpoint" env:"OTLP_UPSTREAM_ENDPOINT" env-default:"http://otel-collector:4318"`
	BodyLimitBytes     int     `yaml:"body_limit_bytes" env:"TELEMETRY_GATEWAY_BODY_LIMIT_BYTES" env-default:"10485760"`
	RequestTimeoutMS   int     `yaml:"request_timeout_ms" env:"TELEMETRY_GATEWAY_REQUEST_TIMEOUT_MS" env-default:"5000"`
	RateLimitPerSecond float64 `yaml:"rate_limit_per_second" env:"TELEMETRY_GATEWAY_RATE_LIMIT_PER_SECOND" env-default:"20"`
	RateLimitBurst     int     `yaml:"rate_limit_burst" env:"TELEMETRY_GATEWAY_RATE_LIMIT_BURST" env-default:"40"`
}

func LoadConfig(logger *slog.Logger) (*Config, error) {
	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		path = "./config.yaml"
	}

	var cfg Config
	if _, err := os.Stat(path); err == nil {
		if readErr := cleanenv.ReadConfig(path, &cfg); readErr != nil {
			return nil, readErr
		}
		return &cfg, nil
	}
	if readErr := cleanenv.ReadEnv(&cfg); readErr != nil {
		return nil, readErr
	}
	return &cfg, nil
}
