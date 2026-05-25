package config

import (
	"fmt"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/configutil"
	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Host                  string   `yaml:"host" env:"HOST" envDefault:":3100"`
	AuthSecret            string   `yaml:"auth_secret" env:"AUTH_SECRET"`
	AuthSecretEnforcement string   `yaml:"auth_secret_enforcement" env:"AUTH_SECRET_ENFORCEMENT" env-default:"warn"`
	Cluster               []string `yaml:"cluster" env:"CLUSTER" env-default:""`
	ClusterKeyspace       string   `yaml:"cluster_keyspace" env:"CLUSTER_KEYSPACE" env-default:"gochat"`
	HearthBeatTimeout     int64    `yaml:"hearth_beat_timeout" env:"HEARTH_BEAT_TIME" env-default:"35000"`
	NATSConnString        string   `yaml:"nats_conn_string" env:"NATS_CONN_STRING" env-default:"nats://nats:4222"`
	PGDriver              string   `yaml:"pg_driver" env:"PG_DRIVER" env-default:"postgres"`
	PGDSN                 string   `yaml:"pg_dsn" env:"PG_DSN"`
	PGRetries             int      `yaml:"pg_retries" env:"PG_RETRIES" env-default:"5"`
	CacheAddr             string   `yaml:"cache_addr" env:"CACHE_ADDR" env-default:"keydb:6379"`
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
	configutil.WarnWeakAuthSecret(logger, config.AuthSecret, "ws")
	return &config, nil
}
