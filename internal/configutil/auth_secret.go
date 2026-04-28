package configutil

import (
	"fmt"
	"log/slog"
	"strings"
)

const (
	recommendedAuthSecretLength = 32
	AuthSecretEnforcementWarn   = "warn"
	AuthSecretEnforcementStrict = "strict"
)

func ValidateAuthSecret(secret string) error {
	return ValidateAuthSecretWithMode(secret, AuthSecretEnforcementWarn)
}

func ValidateAuthSecretWithMode(secret, enforcement string) error {
	if strings.TrimSpace(secret) == "" {
		return fmt.Errorf("auth secret is required")
	}
	switch normalizeAuthSecretEnforcement(enforcement) {
	case AuthSecretEnforcementWarn:
		return nil
	case AuthSecretEnforcementStrict:
		return validateStrictAuthSecret(secret)
	default:
		return fmt.Errorf("auth_secret_enforcement must be %q or %q", AuthSecretEnforcementWarn, AuthSecretEnforcementStrict)
	}
}

func normalizeAuthSecretEnforcement(enforcement string) string {
	enforcement = strings.ToLower(strings.TrimSpace(enforcement))
	if enforcement == "" {
		return AuthSecretEnforcementWarn
	}
	return enforcement
}

func validateStrictAuthSecret(secret string) error {
	trimmed := strings.TrimSpace(secret)
	if len(trimmed) < recommendedAuthSecretLength {
		return fmt.Errorf("auth secret must be at least %d characters when auth_secret_enforcement=strict", recommendedAuthSecretLength)
	}

	lower := strings.ToLower(trimmed)
	placeholderSecrets := map[string]struct{}{
		"change_me_before_use_it_in_production": {},
		"changeme":                              {},
		"change-me":                             {},
		"change_me":                             {},
		"replace-me":                            {},
		"replace_me":                            {},
	}
	if _, ok := placeholderSecrets[lower]; ok {
		return fmt.Errorf("auth secret placeholder values are not allowed when auth_secret_enforcement=strict")
	}
	return nil
}

func WarnWeakAuthSecret(logger *slog.Logger, secret string, service string) {
	if logger == nil {
		return
	}

	trimmed := strings.TrimSpace(secret)
	lower := strings.ToLower(trimmed)
	_, placeholder := map[string]struct{}{
		"change_me_before_use_it_in_production": {},
		"changeme":                              {},
		"change-me":                             {},
		"change_me":                             {},
		"replace-me":                            {},
		"replace_me":                            {},
	}[lower]
	if len(trimmed) >= recommendedAuthSecretLength && !placeholder {
		return
	}

	logger.Warn(
		"auth secret is weak; rotate the shared secret together across api, auth, ws, sfu, and attachments before enabling strict enforcement",
		slog.String("service", service),
		slog.Int("recommended_length", recommendedAuthSecretLength),
	)
}
