package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Cluster               []string      `yaml:"cluster" env:"CLUSTER" env-default:""`
	ClusterKeyspace       string        `yaml:"cluster_keyspace" env:"CLUSTER_KEYSPACE" env-default:"gochat"`
	NatsConnString        string        `yaml:"nats_conn_string" env:"NATS_CONN_STRING" env-default:"nats://nats:4222"`
	KeyDB                 string        `yaml:"keydb" env:"KEYDB" env-default:"127.0.0.1:6379"`
	CacheTTL              time.Duration `yaml:"cache_ttl" env:"CACHE_TTL" env-default:"6h"`
	NegativeCacheTTL      time.Duration `yaml:"negative_cache_ttl" env:"NEGATIVE_CACHE_TTL" env-default:"30m"`
	ExcludedURLPatterns   []string      `yaml:"excluded_url_patterns" env:"EXCLUDED_URL_PATTERNS" env-separator:","`
	FetchTimeout          time.Duration `yaml:"fetch_timeout" env:"FETCH_TIMEOUT" env-default:"10s"`
	MaxBodyBytes          int64         `yaml:"max_body_bytes" env:"MAX_BODY_BYTES" env-default:"2097152"`
	AllowPrivateHosts     bool          `yaml:"allow_private_hosts" env:"ALLOW_PRIVATE_HOSTS" env-default:"false"`
	YouTubeOEmbedEndpoint string        `yaml:"youtube_oembed_endpoint" env:"YOUTUBE_OEMBED_ENDPOINT" env-default:"https://www.youtube.com/oembed"`
	YouTubeEmbedBaseURL   string        `yaml:"youtube_embed_base_url" env:"YOUTUBE_EMBED_BASE_URL" env-default:"https://www.youtube.com/embed"`
}

func LoadConfig(logger *slog.Logger) (*Config, error) {
	var config Config
	err := cleanenv.ReadConfig("./config.yaml", &config)
	if err != nil {
		logger.Warn("unable to read config", slog.String("error", err.Error()))
		err = cleanenv.ReadEnv(&config)
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}
	return &config, validator.New().Struct(&config)
}
