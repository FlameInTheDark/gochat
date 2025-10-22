package config

import (
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ApiLog                     bool     `yaml:"api_log" env:"API_LOG" env-default:"true"`
	ServerAddress              string   `yaml:"server_address" env:"SERVER_ADDRESS" env-default:":3100"`
	IdempotencyStorageLifetime int64    `yaml:"idempotency_storage_lifetime" env:"IDEMPOTENCY_STORAGE_LIFETIME" env-default:"10"`
	RateLimitTime              int      `yaml:"rate_limit_time" env:"RATE_LIMIT_TIME" env-default:"1"`
	RateLimitRequests          int      `yaml:"rate_limit_requests" env:"RATE_LIMIT_REQUESTS" env-default:"20"`
	AppName                    string   `yaml:"app_name" env:"APP_NAME" env-default:"GoChat"`
	BaseUrl                    string   `yaml:"base_url" env:"BASE_URL" env-default:"http://example.com" validation:"http_url"`
	Cluster                    []string `yaml:"cluster" env:"CLUSTER" env-default:""`
	ClusterKeyspace            string   `yaml:"cluster_keyspace" env:"CLUSTER_KEYSPACE" env-default:"gochat"`
	AuthSecret                 string   `yaml:"auth_secret" env:"AUTH_SECRET" env-default:"change_me_before_use_it_in_production"`
	Swagger                    bool     `yaml:"swagger" env:"SWAGGER" env-default:"false"`
	KeyDB                      string   `yaml:"keydb" env:"KEYDB" env-default:"127.0.0.1"`
	UploadLimit                int64    `yaml:"upload_limit" env:"UPLOAD_LIMIT" env-default:"50000000"`
	AttachmentTTLMinutes       int64    `yaml:"attachment_ttl_minutes" env:"ATTACHMENT_TTL_MINUTES" env-default:"10"`
	NatsConnString             string   `yaml:"nats_conn_string" env:"NATS_CONN_STRING" env-default:"nats://nats:4222"`
	IndexerNatsConnString      string   `yaml:"indexer_nats_conn_string" env:"INDEX_NATS_CONN_STRING" env-default:"nats://indexer-nats:4222"`
	PGDSN                      string   `yaml:"pg_dsn" env:"PG_DSN" env-default:""`
	PGRetries                  int      `yaml:"pg_retries" env:"PG_RETRIES" env-default:"5"`
	OSInsecureSkipVerify       bool     `yaml:"os_insecure_skip_verify" env:"OS_INSECURE_SKIP_VERIFY"`
	OSAddresses                []string `yaml:"os_addresses" env:"OS_ADDRESSES"`
	OSUsername                 string   `yaml:"os_username" env:"OS_USERNAME"`
	OSPassword                 string   `yaml:"os_password" env:"OS_PASSWORD"`
	// Voice/Discovery
	VoiceRegions       []VoiceRegion `yaml:"voice_regions"`
	VoiceDefaultRegion string        `yaml:"voice_region" env:"VOICE_REGION" env-default:"global"`
	EtcdEndpoints      []string      `yaml:"etcd_endpoints" env:"ETCD_ENDPOINTS" env-separator:","`
	EtcdPrefix         string        `yaml:"etcd_prefix" env:"ETCD_PREFIX" env-default:"/gochat/sfu"`
	EtcdUsername       string        `yaml:"etcd_username" env:"ETCD_USERNAME"`
	EtcdPassword       string        `yaml:"etcd_password" env:"ETCD_PASSWORD"`
}

type VoiceRegion struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
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
