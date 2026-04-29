package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
)

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

// adminClaims defines the admin JWT contents used for API → SFU control calls.
type adminClaims struct {
	helper.Claims
	StreamID int64 `json:"stream_id"`
}

// validateAdminToken parses and validates an admin JWT (typ=admin, aud=stream).
// Returns the stream ID the token is scoped to.
func (a *App) validateAdminToken(token string) (int64, error) {
	var claims adminClaims
	tok := stripBearerPrefix(token)
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
		return 0, err
	}
	if claims.TokenType != "admin" || !containsString(claims.Audience, "stream") {
		return 0, fmt.Errorf("aud/typ mismatch")
	}
	return claims.StreamID, nil
}

// validateJoinToken parses and validates the stream join token.
// Returns user identity, internal stream id, voice channel id, guild id, synthetic permissions, role, source type, and audio mode.
func (a *App) validateJoinToken(token string) (int64, int64, int64, *int64, int64, string, string, string, error) {
	var claims streammeta.Claims

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
		return 0, 0, 0, nil, 0, "", "", "", err
	}

	// Ensure the token type and audience are correct.
	if claims.TokenType != "stream" || !containsString(claims.Audience, "stream") {
		return 0, 0, 0, nil, 0, "", "", "", fmt.Errorf("aud/typ mismatch")
	}
	if claims.RouteID == "" || claims.RouteID != a.cfg.ServiceID {
		return 0, 0, 0, nil, 0, "", "", "", fmt.Errorf("route mismatch")
	}
	if claims.OwnerUserID == 0 {
		return 0, 0, 0, nil, 0, "", "", "", fmt.Errorf("owner missing")
	}
	if claims.Role != streammeta.RolePublisher && claims.Role != streammeta.RoleViewer {
		return 0, 0, 0, nil, 0, "", "", "", fmt.Errorf("role mismatch")
	}
	if claims.Role == streammeta.RolePublisher && claims.UserID != claims.OwnerUserID {
		return 0, 0, 0, nil, 0, "", "", "", fmt.Errorf("owner mismatch")
	}

	var perms int64
	if claims.Role == streammeta.RolePublisher {
		perms = int64(permissions.PermVoiceSpeak | permissions.PermVoiceVideo)
	}

	var guildID *int64
	if claims.GuildID != 0 {
		gid := claims.GuildID
		guildID = &gid
	}

	return claims.UserID, claims.StreamID, claims.ChannelID, guildID, perms, claims.Role, claims.SourceType, claims.AudioMode, nil
}
