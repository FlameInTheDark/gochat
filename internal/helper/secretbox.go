package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type EncryptedValue struct {
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}

type SecretBox struct {
	gcm cipher.AEAD
}

func NewSecretBoxFromBase64(raw string) (*SecretBox, error) {
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode MFA encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("decode MFA encryption key: expected 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("init MFA cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("init MFA GCM: %w", err)
	}

	return &SecretBox{gcm: gcm}, nil
}

func (b *SecretBox) Encrypt(plaintext []byte) (EncryptedValue, error) {
	if b == nil || b.gcm == nil {
		return EncryptedValue{}, fmt.Errorf("secret box is not configured")
	}

	nonce := make([]byte, b.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return EncryptedValue{}, fmt.Errorf("generate nonce: %w", err)
	}

	return EncryptedValue{
		Nonce:      nonce,
		Ciphertext: b.gcm.Seal(nil, nonce, plaintext, nil),
	}, nil
}

func (b *SecretBox) Decrypt(value EncryptedValue) ([]byte, error) {
	if b == nil || b.gcm == nil {
		return nil, fmt.Errorf("secret box is not configured")
	}
	if len(value.Nonce) != b.gcm.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size")
	}
	plaintext, err := b.gcm.Open(nil, value.Nonce, value.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret: %w", err)
	}
	return plaintext, nil
}
