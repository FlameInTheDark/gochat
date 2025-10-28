package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct{ secret []byte }

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{secret: []byte(secret)}
}

type ServiceClaims struct {
	ServiceType string `json:"typ"`
	ServiceID   string `json:"id"`
	jwt.RegisteredClaims
}

// Validate verifies the JWT signature and claims: typ must match, id must match when expectedID is non-empty.
// No expiration is enforced; tokens may omit exp based on current requirements.
func (t *TokenManager) Validate(expectedType, expectedID, token string) bool {
	if len(t.secret) == 0 || token == "" {
		return false
	}
	var claims ServiceClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(tok *jwt.Token) (any, error) {
		if tok.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrTokenUnverifiable
		}
		return t.secret, nil
	})
	if err != nil {
		return false
	}
	if claims.ServiceType != expectedType {
		return false
	}
	if expectedID != "" && claims.ServiceID != expectedID {
		return false
	}
	return true
}
