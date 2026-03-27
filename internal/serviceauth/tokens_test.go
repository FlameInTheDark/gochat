package serviceauth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenManagerParseAndValidate(t *testing.T) {
	manager := NewTokenManager("supersecret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"typ": "sfu",
		"id":  "node-1",
	})
	signed, err := token.SignedString([]byte("supersecret"))
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}

	claims, err := manager.Parse(signed)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.ServiceType != "sfu" || claims.ServiceID != "node-1" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	if !manager.Validate("sfu", "node-1", signed) {
		t.Fatal("expected token to validate")
	}
	if manager.Validate("sfu", "node-2", signed) {
		t.Fatal("expected token id mismatch to fail")
	}
}

func TestExtractBearerToken(t *testing.T) {
	if got := ExtractBearerToken("Bearer abc123"); got != "abc123" {
		t.Fatalf("ExtractBearerToken = %q, want %q", got, "abc123")
	}
	if got := ExtractBearerToken("bearer abc123"); got != "abc123" {
		t.Fatalf("ExtractBearerToken should be case-insensitive, got %q", got)
	}
	if got := ExtractBearerToken("Basic abc123"); got != "" {
		t.Fatalf("ExtractBearerToken = %q, want empty", got)
	}
}
