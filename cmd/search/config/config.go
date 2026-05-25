package config

import (
	"fmt"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/configutil"
	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ApiLog                bool     `yaml:"api_log" env:"API_LOG" env-default:"true"`
	LogLevel              string   `yaml:"log_level" env:"LOG_LEVEL" env-default:"warn"`
	ServerAddress         string   `yaml:"server_address" env:"SERVER_ADDRESS" env-default:":3100"`
	AuthSecret            string   `yaml:"auth_secret" env:"AUTH_SECRET"`
	AuthSecretEnforcement string   `yaml:"auth_secret_enforcement" env:"AUTH_SECRET_ENFORCEMENT" env-default:"warn"`
	Swagger               bool     `yaml:"swagger" env:"SWAGGER" env-default:"false"`
	Cluster               []string `yaml:"cluster" env:"CLUSTER" env-default:""`
	ClusterKeyspace       string   `yaml:"cluster_keyspace" env:"CLUSTER_KEYSPACE" env-default:"gochat"`
	KeyDB                 string   `yaml:"keydb" env:"KEYDB" env-default:"127.0.0.1:6379"`
	NATSConnString        string   `yaml:"nats_conn_string" env:"NATS_CONN_STRING" env-default:"nats://indexer-nats:4222"`
	PGDriver              string   `yaml:"pg_driver" env:"PG_DRIVER" env-default:"postgres"`
	PGDSN                 string   `yaml:"pg_dsn" env:"PG_DSN" env-default:""`
	PGRetries             int      `yaml:"pg_retries" env:"PG_RETRIES" env-default:"5"`
	PGQueryLog            bool     `yaml:"pg_query_log" env:"PG_QUERY_LOG" env-default:"false"`
	PGMaxOpenConns        int      `yaml:"pg_max_open_conns" env:"PG_MAX_OPEN_CONNS" env-default:"50"`
	PGMaxIdleConns        int      `yaml:"pg_max_idle_conns" env:"PG_MAX_IDLE_CONNS" env-default:"25"`
	RedisPoolSize         int      `yaml:"redis_pool_size" env:"REDIS_POOL_SIZE" env-default:"100"`
	RedisMinIdleConns     int      `yaml:"redis_min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" env-default:"20"`
	OSInsecureSkipVerify  bool     `yaml:"os_insecure_skip_verify" env:"OS_INSECURE_SKIP_VERIFY"`
	OSAddresses           []string `yaml:"os_addresses" env:"OS_ADDRESSES"`
	OSUsername            string   `yaml:"os_username" env:"OS_USERNAME"`
	OSPassword            string   `yaml:"os_password" env:"OS_PASSWORD"`
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
	configutil.WarnWeakAuthSecret(logger, config.AuthSecret, "search")
	return &config, nil
}
