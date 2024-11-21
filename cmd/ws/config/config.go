package config

import (
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AuthSecret       string `yaml:"auth_secret" env:"AUTH_SECRET" env-default:"change_me_before_use_it_in_production"`
	RabbitMQHost     string `yaml:"rabbitmq_host" env:"RABBITMQ_HOST" env-default:"rabbitmq"`
	RabbitMQPort     int    `yaml:"rabbitmq_port" env:"RABBITMQ_PORT" env-default:"5672"`
	RabbitMQUsername string `yaml:"rabbitmq_username" env:"RABBITMQ_USERNAME" env-default:"guest"`
	RabbitMQPassword string `yaml:"rabbitmq_password" env:"RABBITMQ_PASSWORD" env-default:"guest"`
	NatsConnString   string `yaml:"nats_conn_string" env:"NATS_CONN_STRING" env-default:"nats://nats:4222"`
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
