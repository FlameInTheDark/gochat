package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	AccessSecret []byte
	Issuer       string // e.g. "gochat"
	Audience     string // e.g. "api" (or "ws" if you prefer)
	Leeway       time.Duration
}

func New(accessSecret, issuer, audience string) *Auth {
	return &Auth{
		AccessSecret: []byte(accessSecret),
		Issuer:       issuer,
		Audience:     audience,
		Leeway:       2 * time.Second,
	}
}

func (a *Auth) ParseAccess(token string) (*helper.Claims, error) {
	tok := strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(tok), "bearer ") {
		tok = strings.TrimSpace(tok[7:])
	}

	claims := &helper.Claims{}
	t, err := jwt.ParseWithClaims(tok, claims, func(t *jwt.Token) (any, error) {
		// lock the algorithm
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected alg: %s", t.Method.Alg())
		}
		return a.AccessSecret, nil
	},
		jwt.WithIssuer(a.Issuer),
		jwt.WithAudience(a.Audience),
		jwt.WithLeeway(a.Leeway),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err // you'll get nice errors: malformed, invalid signature, expired, wrong aud/iss, etc.
	}
	if !t.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.TokenType != "access" { // or compare claims.TokenType if you named it that way
		return nil, errors.New("wrong token type for websocket")
	}
	return claims, nil
}
