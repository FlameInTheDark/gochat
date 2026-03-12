package helper

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"unicode/utf8"
)

const MaxMessageNonceLength = 25

var (
	ErrMessageNonceType      = errors.New("nonce must be a string or integer")
	ErrMessageNonceTooLong   = errors.New("nonce must be 25 characters or fewer")
	messageNonceIntegerRegex = regexp.MustCompile(`^-?[0-9]+$`)
)

// MessageNonce preserves the client's original JSON nonce representation so it
// can be echoed back exactly as sent.
type MessageNonce []byte

func (n *MessageNonce) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*n = nil
		return nil
	}

	if b[0] == '"' {
		var value string
		if err := json.Unmarshal(b, &value); err != nil {
			return ErrMessageNonceType
		}
		if utf8.RuneCountInString(value) > MaxMessageNonceLength {
			return ErrMessageNonceTooLong
		}
		cp := make([]byte, len(b))
		copy(cp, b)
		*n = MessageNonce(cp)
		return nil
	}

	if !messageNonceIntegerRegex.Match(b) {
		return ErrMessageNonceType
	}
	if len(b) > MaxMessageNonceLength {
		return ErrMessageNonceTooLong
	}

	cp := make([]byte, len(b))
	copy(cp, b)
	*n = MessageNonce(cp)
	return nil
}

func (n MessageNonce) MarshalJSON() ([]byte, error) {
	if len(n) == 0 {
		return []byte("null"), nil
	}

	cp := make([]byte, len(n))
	copy(cp, n)
	return cp, nil
}

func (n MessageNonce) Clone() *MessageNonce {
	if len(n) == 0 {
		return nil
	}

	cp := make(MessageNonce, len(n))
	copy(cp, n)
	return &cp
}

func (n MessageNonce) IsZero() bool {
	return len(n) == 0
}

func (n MessageNonce) CacheKeyPart() string {
	if len(n) == 0 {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(n)
}
