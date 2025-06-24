package config

import (
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ApiLog                bool     `yaml:"api_log" env:"API_LOG" env-default:"true"`
	RateLimitTime         int      `yaml:"rate_limit_time" env:"RATE_LIMIT_TIME" env-default:"1"`
	RateLimitRequests     int      `yaml:"rate_limit_requests" env:"RATE_LIMIT_REQUESTS" env-default:"20"`
	AppName               string   `yaml:"app_name" env:"APP_NAME" env-default:"GoChat"`
	BaseUrl               string   `yaml:"base_url" env:"BASE_URL" env-default:"http://example.com" validation:"http_url"`
	Cluster               []string `yaml:"cluster" env:"CLUSTER" env-default:""`
	ClusterKeyspace       string   `yaml:"cluster_keyspace" env:"CLUSTER_KEYSPACE" env-default:"gochat"`
	EmailSource           string   `yaml:"email_source" env:"EMAIL_SOURCE" env-default:"no-reply@example.com" validation:"email"`
	EmailName             string   `yaml:"email_name" env:"EMAIL_NAME" env-default:"no-reply"`
	EmailTemplate         string   `yaml:"email_template" env:"EMAIL_TEMPLATE" env-default:"./email_notify.tmpl"`
	EmailProvider         string   `yaml:"email_provider" env:"EMAIL_PROVIDER" env-default:"log"`
	SendpulseUserId       string   `yaml:"sendpulse_user_id" env:"SENDPULSE_USER_ID" env-default:""`
	SendpulseSecret       string   `yaml:"sendpulse_secret" env:"SENDPULSE_SECRET" env-default:""`
	AuthSecret            string   `yaml:"auth_secret" env:"AUTH_SECRET" env-default:"change_me_before_use_it_in_production"`
	Swagger               bool     `yaml:"swagger" env:"SWAGGER" env-default:"false"`
	KeyDB                 string   `yaml:"keydb" env:"KEYDB" env-default:"127.0.0.1"`
	S3Endpoint            string   `yaml:"s3_endpoint" env:"S3_ENDPOINT" env-default:""`
	S3AccessKeyID         string   `yaml:"s3_access_key_id" env:"S3_ACCESS_KEY_ID" env-default:""`
	S3SecretAccessKey     string   `yaml:"s3_secret_access_key" env:"S3_SECRET_ACCESS_KEY" env-default:""`
	S3UseSSL              bool     `yaml:"s3_use_ssl" env:"S3_USE_SSL" env-default:"false"`
	UploadLimit           int64    `yaml:"upload_limit" env:"UPLOAD_LIMIT" env-default:"50000000"`
	NatsConnString        string   `yaml:"nats_conn_string" env:"NATS_CONN_STRING" env-default:"nats://nats:4222"`
	IndexerNatsConnString string   `yaml:"indexer_nats_conn_string" env:"INDEX_NATS_CONN_STRING" env-default:"nats://indexer-nats:4222"`
	SolrBaseURL           string   `yaml:"solr_base_url" env:"SOLR_BASE_URL" env-default:"http://solr:8983"`
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
