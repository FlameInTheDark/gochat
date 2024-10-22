package helper

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

// RandomToken generates a random string of a fixed length using crypto/rand
func RandomToken(n int) (string, error) {
	// Calculate how many random bytes we need to generate
	byteLength := (n * 3) / 4 // Base64 encoding expands by 4/3, so generate fewer bytes

	// Create a byte slice to hold the random bytes
	randomBytes := make([]byte, byteLength)

	// Read random bytes into the slice
	if _, err := io.ReadFull(rand.Reader, randomBytes); err != nil {
		return "", err
	}

	// Encode to base64 and return the first n characters
	return base64.RawURLEncoding.EncodeToString(randomBytes)[:n], nil
}
