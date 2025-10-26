package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

// sfuClaims defines the SFU JWT contents.
type sfuClaims struct {
	helper.Claims
	ChannelID int64  `json:"channel_id"`
	GuildID   *int64 `json:"guild_id,omitempty"`
	Perms     int64  `json:"perms"`
	// Moved indicates the user was force-moved to this channel by an admin
	// or a user with PermVoiceMoveMembers; bypasses channel-level blocks.
	Moved bool `json:"moved,omitempty"`
}

// stripBearerPrefix removes an optional "Bearer " prefix.
func stripBearerPrefix(token string) string {
	t := strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(t), "bearer ") {
		return strings.TrimSpace(t[7:])
	}
	return t
}

// containsString reports whether v is in xs.
func containsString(xs []string, v string) bool {
	for _, s := range xs {
		if s == v {
			return true
		}
	}
	return false
}

// validateJoinToken parses and validates the SFU join token.
// Returns (userID, channelID, perms, error).
func (a *App) validateJoinToken(token string) (int64, int64, *int64, int64, bool, error) {
	var claims sfuClaims

	tok := stripBearerPrefix(token)

	// Parse and validate the token with expected issuer and algorithm.
	_, err := jwt.ParseWithClaims(
		tok,
		&claims,
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected alg: %s", t.Method.Alg())
			}
			return []byte(a.cfg.AuthSecret), nil
		},
		jwt.WithIssuer("gochat"),
		jwt.WithLeeway(2*time.Second),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return 0, 0, nil, 0, false, err
	}

	// Ensure the token type and audience are correct.
	if claims.TokenType != "sfu" || !containsString(claims.Audience, "sfu") {
		return 0, 0, nil, 0, false, fmt.Errorf("aud/typ mismatch")
	}

	return claims.UserID, claims.ChannelID, claims.GuildID, claims.Perms, claims.Moved, nil
}
