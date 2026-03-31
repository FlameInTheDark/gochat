package configutil

import (
	"fmt"
	"log/slog"
	"strings"
)

const (
	recommendedAuthSecretLength = 32
)

func ValidateAuthSecret(secret string) error {
	if strings.TrimSpace(secret) == "" {
		return fmt.Errorf("auth secret is required")
	}

	return nil
}

func WarnWeakAuthSecret(logger *slog.Logger, secret string, service string) {
	if logger == nil {
		return
	}
	if len(strings.TrimSpace(secret)) >= recommendedAuthSecretLength {
		return
	}

	logger.Warn(
		"auth secret is shorter than the recommended minimum",
		slog.String("service", service),
		slog.Int("recommended_length", recommendedAuthSecretLength),
	)
}
