package config

import (
	"fmt"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/configutil"
	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel                   string   `yaml:"log_level" env:"LOG_LEVEL" env-default:"warn"`
	ServerAddress              string   `yaml:"server_address" env:"SERVER_ADDRESS" env-default:":3102"`
	AuthSecret                 string   `yaml:"auth_secret" env:"AUTH_SECRET"`
	AuthSecretEnforcement      string   `yaml:"auth_secret_enforcement" env:"AUTH_SECRET_ENFORCEMENT" env-default:"warn"`
	Cluster                    []string `yaml:"cluster" env:"CLUSTER" env-default:""`
	ClusterKeyspace            string   `yaml:"cluster_keyspace" env:"CLUSTER_KEYSPACE" env-default:"gochat"`
	KeyDB                      string   `yaml:"keydb" env:"KEYDB" env-default:"127.0.0.1:6379"`
	UploadLimit                int64    `yaml:"upload_limit" env:"UPLOAD_LIMIT" env-default:"50000000"`
	AttachmentTTLMinutes       int64    `yaml:"attachment_ttl_minutes" env:"ATTACHMENT_TTL_MINUTES" env-default:"10"`
	NATSConnString             string   `yaml:"nats_conn_string" env:"NATS_CONN_STRING" env-default:"nats://nats:4222"`
	BotNATSConnString          string   `yaml:"bot_nats_conn_string" env:"BOT_NATS_CONN_STRING" env-default:"nats://bot-nats:4222"`
	BotEventPartitions         int      `yaml:"bot_event_partitions" env:"BOT_EVENT_PARTITIONS" env-default:"1024"`
	IndexerNATSConnString      string   `yaml:"indexer_nats_conn_string" env:"INDEX_NATS_CONN_STRING" env-default:"nats://indexer-nats:4222"`
	PGDriver                   string   `yaml:"pg_driver" env:"PG_DRIVER" env-default:"postgres"`
	PGDSN                      string   `yaml:"pg_dsn" env:"PG_DSN" env-default:""`
	PGRetries                  int      `yaml:"pg_retries" env:"PG_RETRIES" env-default:"5"`
	PGQueryLog                 bool     `yaml:"pg_query_log" env:"PG_QUERY_LOG" env-default:"false"`
	PGMaxOpenConns             int      `yaml:"pg_max_open_conns" env:"PG_MAX_OPEN_CONNS" env-default:"50"`
	PGMaxIdleConns             int      `yaml:"pg_max_idle_conns" env:"PG_MAX_IDLE_CONNS" env-default:"25"`
	RedisPoolSize              int      `yaml:"redis_pool_size" env:"REDIS_POOL_SIZE" env-default:"100"`
	RedisMinIdleConns          int      `yaml:"redis_min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" env-default:"20"`
	IdempotencyStorageLifetime int64    `yaml:"idempotency_storage_lifetime" env:"IDEMPOTENCY_STORAGE_LIFETIME" env-default:"10"`
	RateLimitTime              int      `yaml:"rate_limit_time" env:"RATE_LIMIT_TIME" env-default:"1"`
	RateLimitRequests          int      `yaml:"rate_limit_requests" env:"RATE_LIMIT_REQUESTS" env-default:"60"`
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
	if err := validator.New().Struct(&config); err != nil {
		return nil, err
	}
	if err := configutil.ValidateAuthSecretWithMode(config.AuthSecret, config.AuthSecretEnforcement); err != nil {
		return nil, err
	}
	configutil.WarnWeakAuthSecret(logger, config.AuthSecret, "botapi")
	return &config, nil
}
