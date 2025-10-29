package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServerAddress string   `yaml:"server_address" env-default:":3300"`
	AuthSecret    string   `yaml:"auth_secret" env:"AUTH_SECRET" env-required:"true"`
	STUNServers   []string `yaml:"stun_servers" env:"STUN_SERVERS" env-separator:"," env-default:"stun:stun.l.google.com:19302"`
	Region        string   `yaml:"region" env:"SFU_REGION" env-default:"global"`
	PublicBaseURL string   `yaml:"public_base_url" env:"SFU_PUBLIC_BASE_URL" env-required:"true"`
	// Discovery
	WebhookURL   string `yaml:"webhook_url" env:"WEBHOOK_URL" env-required:"true"`
	WebhookToken string `yaml:"webhook_token" env:"WEBHOOK_TOKEN" env-required:"true"`
	ServiceID    string `yaml:"service_id" env:"SFU_SERVICE_ID" env-required:"true"`

	// Media limits
	// MaxAudioBitrateKbps, when > 0, injects SDP constraints to cap OPUS
	// encoder average bitrate on clients and advertises bandwidth limits per
	// RFC7587 (maxaveragebitrate) and RFC3890 (b=AS/TIAS).
	// Example values: 64, 96, 128. 0 disables limiting.
	MaxAudioBitrateKbps int  `yaml:"max_audio_bitrate_kbps" env:"SFU_MAX_AUDIO_BITRATE_KBPS" env-default:"0"`
	EnforceAudioBitrate bool `yaml:"enforce_audio_bitrate" env:"SFU_ENFORCE_AUDIO_BITRATE" env-default:"false"`

	// AudioBitrateMarginPercent adds tolerance to measured inbound RTP bitrate
	// during enforcement to account for headers/Jitter/overhead. E.g. 15 means
	// 15% over the configured cap is tolerated before disconnect. Range 0..100.
	AudioBitrateMarginPercent int `yaml:"audio_bitrate_margin_percent" env:"SFU_AUDIO_BITRATE_MARGIN_PERCENT" env-default:"15"`
}

func LoadConfig() (*Config, error) {
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
