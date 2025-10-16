package config

import (
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ApiLog            bool     `yaml:"api_log" env:"API_LOG" env-default:"true"`
	ServerAddress     string   `yaml:"server_address" env:"SERVER_ADDRESS" env-default:":3200"`
	AuthSecret        string   `yaml:"auth_secret" env:"AUTH_SECRET" env-default:"change_me_before_use_it_in_production"`
	Cluster           []string `yaml:"cluster" env:"CLUSTER" env-default:""`
	ClusterKeyspace   string   `yaml:"cluster_keyspace" env:"CLUSTER_KEYSPACE" env-default:"gochat"`
	S3Endpoint        string   `yaml:"s3_endpoint" env:"S3_ENDPOINT" env-default:""`
	S3AccessKeyID     string   `yaml:"s3_access_key_id" env:"S3_ACCESS_KEY_ID" env-default:""`
	S3SecretAccessKey string   `yaml:"s3_secret_access_key" env:"S3_SECRET_ACCESS_KEY" env-default:""`
	S3UseSSL          bool     `yaml:"s3_use_ssl" env:"S3_USE_SSL" env-default:"false"`
	S3Bucket          string   `yaml:"s3_bucket" env:"S3_BUCKET" env-default:"gochat"`
	S3Region          string   `yaml:"s3_region" env:"S3_REGION"`
	S3ExternalURL     string   `yaml:"s3_external_url" env:"S3_EXTERNAL_URL"`
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
