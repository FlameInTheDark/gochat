package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel              string        `yaml:"log_level" env:"LOG_LEVEL" env-default:"warn"`
	BotNATSConnString     string        `yaml:"bot_nats_conn_string" env:"BOT_NATS_CONN_STRING" env-default:"nats://bot-nats:4222"`
	BotEventPartitions    int           `yaml:"bot_event_partitions" env:"BOT_EVENT_PARTITIONS" env-default:"1024"`
	BotRouterID           string        `yaml:"bot_router_id" env:"BOT_ROUTER_ID"`
	BotRouterLeaseTTL     time.Duration `yaml:"bot_router_lease_ttl" env:"BOT_ROUTER_LEASE_TTL" env-default:"15s"`
	BotRouterLeaseRenew   time.Duration `yaml:"bot_router_lease_renew" env:"BOT_ROUTER_LEASE_RENEW" env-default:"5s"`
	BotRouterAcquireEvery time.Duration `yaml:"bot_router_acquire_every" env:"BOT_ROUTER_ACQUIRE_EVERY" env-default:"5s"`
	KeyDB                 string        `yaml:"keydb" env:"KEYDB" env-default:"127.0.0.1:6379"`
	PGDriver              string        `yaml:"pg_driver" env:"PG_DRIVER" env-default:"postgres"`
	PGDSN                 string        `yaml:"pg_dsn" env:"PG_DSN" env-default:""`
	PGRetries             int           `yaml:"pg_retries" env:"PG_RETRIES" env-default:"5"`
	PGQueryLog            bool          `yaml:"pg_query_log" env:"PG_QUERY_LOG" env-default:"false"`
	PGMaxOpenConns        int           `yaml:"pg_max_open_conns" env:"PG_MAX_OPEN_CONNS" env-default:"50"`
	PGMaxIdleConns        int           `yaml:"pg_max_idle_conns" env:"PG_MAX_IDLE_CONNS" env-default:"25"`
	RedisPoolSize         int           `yaml:"redis_pool_size" env:"REDIS_POOL_SIZE" env-default:"100"`
	RedisMinIdleConns     int           `yaml:"redis_min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" env-default:"20"`
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
	return &config, nil
}
