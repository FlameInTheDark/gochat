package config

import (
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Host              string `yaml:"host" env:"HOST" env-default:":3101"`
	HearthBeatTimeout int64  `yaml:"hearth_beat_timeout" env:"HEARTH_BEAT_TIME" env-default:"35000"`
	NATSConnString    string `yaml:"nats_conn_string" env:"NATS_CONN_STRING" env-default:"nats://nats:4222"`
	BotNATSConnString string `yaml:"bot_nats_conn_string" env:"BOT_NATS_CONN_STRING" env-default:"nats://bot-nats:4222"`
	PGDriver          string `yaml:"pg_driver" env:"PG_DRIVER" env-default:"postgres"`
	PGDSN             string `yaml:"pg_dsn" env:"PG_DSN"`
	PGRetries         int    `yaml:"pg_retries" env:"PG_RETRIES" env-default:"5"`
	CacheAddr         string `yaml:"cache_addr" env:"CACHE_ADDR" env-default:"keydb:6379"`
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
