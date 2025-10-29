package config

import (
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Log           bool   `yaml:"api_log" env:"API_LOG" env-default:"true"`
	ServerAddress string `yaml:"server_address" env-default:":3200"`
	// Discovery for SFU heartbeats
	EtcdEndpoints []string `yaml:"etcd_endpoints" env:"ETCD_ENDPOINTS" env-separator:","`
	EtcdPrefix    string   `yaml:"etcd_prefix" env:"ETCD_PREFIX" env-default:"/gochat/sfu"`
	EtcdUsername  string   `yaml:"etcd_username" env:"ETCD_USERNAME"`
	EtcdPassword  string   `yaml:"etcd_password" env:"ETCD_PASSWORD"`
	// Cassandra for attachment finalization
	Cluster         []string `yaml:"cluster" env:"CLUSTER" env-separator:","`
	ClusterKeyspace string   `yaml:"cluster_keyspace" env:"CLUSTER_KEYSPACE" env-default:"gochat"`
	// JWT secret for validating webhook tokens (HS256). Tokens carry service type and id.
	JWTSecret string `yaml:"jwt_secret" env:"WEBHOOK_JWT_SECRET"`
	// Swagger
	Swagger bool `yaml:"swagger" env:"SWAGGER" env-default:"false"`
	// Cache
	KeyDB string `yaml:"keydb" env:"KEYDB" env-default:"127.0.0.1"`
	// NATS
	NatsConnString string `yaml:"nats_conn_string" env:"NATS_CONN_STRING" env-default:"nats://nats:4222"`
}

func LoadConfig(logger *slog.Logger) (*Config, error) {
	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		path = "./config.yaml"
	}
	var cfg Config
	if _, err := os.Stat(path); err == nil {
		if rerr := cleanenv.ReadConfig(path, &cfg); rerr != nil {
			return nil, rerr
		}
		return &cfg, nil
	}
	if rerr := cleanenv.ReadEnv(&cfg); rerr != nil {
		return nil, rerr
	}
	return &cfg, nil
}
