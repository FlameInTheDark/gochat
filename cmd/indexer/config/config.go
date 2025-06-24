package config

import (
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	NatsConnString       string   `yaml:"nats_conn_string" env:"NATS_CONN_STRING" env-default:"nats://indexer-nats:4222"`
	OSInsecureSkipVerify bool     `yaml:"os_insecure_skip_verify" env:"OS_INSECURE_SKIP_VERIFY"`
	OSAddresses          []string `yaml:"os_addresses" env:"OS_ADDRESSES"`
	OSUsername           string   `yaml:"os_username" env:"OS_USERNAME"`
	OSPassword           string   `yaml:"os_password" env:"OS_PASSWORD"`
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
