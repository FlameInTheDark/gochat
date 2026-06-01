package applicationcommands

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const interactionTokenPrefix = "gci_"

func GenerateInteractionToken() (plain, prefix, hash string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", "", err
	}
	plain = interactionTokenPrefix + base64.RawURLEncoding.EncodeToString(raw)
	if len(plain) < 16 {
		return "", "", "", fmt.Errorf("generated interaction token too short")
	}
	return plain, plain[:16], HashInteractionToken(plain), nil
}

func HashInteractionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
