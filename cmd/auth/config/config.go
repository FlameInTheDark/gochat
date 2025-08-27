package config

import (
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ApiLog                     bool   `yaml:"api_log" env:"API_LOG" env-default:"true"`
	ServerAddress              string `yaml:"server_address" env:"SERVER_ADDRESS" env-default:":3100"`
	IdempotencyStorageLifetime int64  `yaml:"idempotency_storage_lifetime" env:"IDEMPOTENCY_STORAGE_LIFETIME" env-default:"10"`
	RateLimitTime              int    `yaml:"rate_limit_time" env:"RATE_LIMIT_TIME" env-default:"1"`
	RateLimitRequests          int    `yaml:"rate_limit_requests" env:"RATE_LIMIT_REQUESTS" env-default:"20"`
	AppName                    string `yaml:"app_name" env:"APP_NAME" env-default:"GoChat"`
	BaseUrl                    string `yaml:"base_url" env:"BASE_URL" env-default:"http://example.com" validation:"http_url"`
	EmailSource                string `yaml:"email_source" env:"EMAIL_SOURCE" env-default:"no-reply@example.com" validation:"email"`
	EmailName                  string `yaml:"email_name" env:"EMAIL_NAME" env-default:"no-reply"`
	EmailTemplate              string `yaml:"email_template" env:"EMAIL_TEMPLATE" env-default:"./email_notify.tmpl"`
	PasswordResetTemplate      string `yaml:"password_reset_template" env:"PASSWORD_RESET_TEMPLATE" env-default:"./password_reset.tmpl"`
	EmailProvider              string `yaml:"email_provider" env:"EMAIL_PROVIDER" env-default:"log"`
	SendpulseUserId            string `yaml:"sendpulse_user_id" env:"SENDPULSE_USER_ID" env-default:""`
	SendpulseSecret            string `yaml:"sendpulse_secret" env:"SENDPULSE_SECRET" env-default:""`
	AuthSecret                 string `yaml:"auth_secret" env:"AUTH_SECRET" env-default:"change_me_before_use_it_in_production"`
	Swagger                    bool   `yaml:"swagger" env:"SWAGGER" env-default:"false"`
	KeyDB                      string `yaml:"keydb" env:"KEYDB" env-default:"127.0.0.1"`
	PGDSN                      string `yaml:"pg_dsn" env:"PG_DSN" env-default:""`
	PGRetries                  int    `yaml:"pg_retries" env:"PG_RETRIES" env-default:"5"`
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
