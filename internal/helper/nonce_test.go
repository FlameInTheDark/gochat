package helper

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestMessageNonceUnmarshalAcceptsStringAndInteger(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "string", input: `"draft-1"`, want: `"draft-1"`},
		{name: "integer", input: `12345`, want: `12345`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nonce MessageNonce
			if err := json.Unmarshal([]byte(tt.input), &nonce); err != nil {
				t.Fatalf("Unmarshal returned error: %v", err)
			}
			if got := string(nonce); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestMessageNonceUnmarshalRejectsUnsupportedTypes(t *testing.T) {
	tests := []string{
		`true`,
		`12.5`,
		`{"x":1}`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			var nonce MessageNonce
			err := json.Unmarshal([]byte(input), &nonce)
			if !errors.Is(err, ErrMessageNonceType) {
				t.Fatalf("expected ErrMessageNonceType, got %v", err)
			}
		})
	}
}

func TestMessageNonceUnmarshalRejectsTooLongValues(t *testing.T) {
	tests := []string{
		`"` + "abcdefghijklmnopqrstuvwxyz" + `"`,
		`12345678901234567890123456`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			var nonce MessageNonce
			err := json.Unmarshal([]byte(input), &nonce)
			if !errors.Is(err, ErrMessageNonceTooLong) {
				t.Fatalf("expected ErrMessageNonceTooLong, got %v", err)
			}
		})
	}
}

func TestMessageNonceCloneProducesIndependentCopy(t *testing.T) {
	var nonce MessageNonce
	if err := json.Unmarshal([]byte(`"draft-1"`), &nonce); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	cloned := nonce.Clone()
	if cloned == nil {
		t.Fatal("expected clone")
	}

	(*cloned)[0] = 'x'
	if string(nonce) != `"draft-1"` {
		t.Fatalf("expected original nonce to stay unchanged, got %q", string(nonce))
	}
}
