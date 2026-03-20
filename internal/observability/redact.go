package observability

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const redactedValue = "[REDACTED]"

func RedactSecret(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return redactedValue
}

func RedactEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(email))
	return "sha256:" + hex.EncodeToString(sum[:8])
}
