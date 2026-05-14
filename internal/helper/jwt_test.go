package helper

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func parseClaims(t *testing.T, tokenString string, secret string) *Claims {
	t.Helper()

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if !token.Valid {
		t.Fatal("token is invalid")
	}

	return claims
}

func expectExpiryNear(t *testing.T, got *jwt.NumericDate, want time.Duration) {
	t.Helper()

	if got == nil {
		t.Fatal("missing expiry")
	}

	remaining := time.Until(got.Time)
	if remaining < want-time.Minute || remaining > want+time.Minute {
		t.Fatalf("expiry remaining = %s, want near %s", remaining, want)
	}
}

func expectAudience(t *testing.T, got jwt.ClaimStrings, want string) {
	t.Helper()

	for _, audience := range got {
		if audience == want {
			return
		}
	}

	t.Fatalf("audience = %v, want %q", got, want)
}

func TestIssueTokensSetsExpectedLifetimesAndClaims(t *testing.T) {
	const secret = "test-secret"
	const userID int64 = 42
	const sessionVersion int64 = 7

	access, refresh, err := IssueTokens(userID, sessionVersion, secret)
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	accessClaims := parseClaims(t, access, secret)
	refreshClaims := parseClaims(t, refresh, secret)

	if accessClaims.UserID != userID || accessClaims.SessionVersion != sessionVersion {
		t.Fatalf("access identity = (%d, %d), want (%d, %d)", accessClaims.UserID, accessClaims.SessionVersion, userID, sessionVersion)
	}
	if refreshClaims.UserID != userID || refreshClaims.SessionVersion != sessionVersion {
		t.Fatalf("refresh identity = (%d, %d), want (%d, %d)", refreshClaims.UserID, refreshClaims.SessionVersion, userID, sessionVersion)
	}

	if accessClaims.TokenType != "access" {
		t.Fatalf("access token type = %q, want access", accessClaims.TokenType)
	}
	if refreshClaims.TokenType != "refresh" {
		t.Fatalf("refresh token type = %q, want refresh", refreshClaims.TokenType)
	}

	expectAudience(t, accessClaims.Audience, "api")
	expectAudience(t, refreshClaims.Audience, "refresh")

	expectExpiryNear(t, accessClaims.ExpiresAt, 15*time.Minute)
	expectExpiryNear(t, refreshClaims.ExpiresAt, 30*24*time.Hour)

	if refreshClaims.ID == "" {
		t.Fatal("refresh token JTI is empty")
	}
}
