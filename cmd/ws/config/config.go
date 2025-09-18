package config

import (
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Host              string   `yaml:"host" env:"HOST" envDefault:":3100"`
	AuthSecret        string   `yaml:"auth_secret" env:"AUTH_SECRET" env-default:"change_me_before_use_it_in_production"`
	Cluster           []string `yaml:"cluster" env:"CLUSTER" env-default:""`
	ClusterKeyspace   string   `yaml:"cluster_keyspace" env:"CLUSTER_KEYSPACE" env-default:"gochat"`
	HearthBeatTimeout int64    `yaml:"hearth_beat_timeout" env:"HEARTH_BEAT_TIME" env-default:"35000"`
	RabbitMQHost      string   `yaml:"rabbitmq_host" env:"RABBITMQ_HOST" env-default:"rabbitmq"`
	RabbitMQPort      int      `yaml:"rabbitmq_port" env:"RABBITMQ_PORT" env-default:"5672"`
	RabbitMQUsername  string   `yaml:"rabbitmq_username" env:"RABBITMQ_USERNAME" env-default:"guest"`
	RabbitMQPassword  string   `yaml:"rabbitmq_password" env:"RABBITMQ_PASSWORD" env-default:"guest"`
	NatsConnString    string   `yaml:"nats_conn_string" env:"NATS_CONN_STRING" env-default:"nats://nats:4222"`
	PGDSN             string   `yaml:"pg_dsn" env:"PG_DSN"`
	PGRetries         int      `yaml:"pg_retries" env:"PG_RETRIES" env-default:"5"`
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
