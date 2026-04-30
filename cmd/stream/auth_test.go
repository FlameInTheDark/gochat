package main

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/cmd/stream/config"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
)

func testStreamApp() *App {
	return &App{cfg: &config.Config{AuthSecret: "auth-secret", ServiceID: "stream-a"}}
}

func signedStreamToken(t *testing.T, secret string, mutate func(*streammeta.Claims)) string {
	t.Helper()

	now := time.Now()
	claims := streammeta.Claims{
		Claims: helper.Claims{
			UserID:    10,
			TokenType: "stream",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "gochat",
				Audience:  []string{"stream"},
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
			},
		},
		StreamID:    777,
		ChannelID:   42,
		GuildID:     1,
		OwnerUserID: 10,
		RouteID:     "stream-a",
		Role:        streammeta.RolePublisher,
		SourceType:  streammeta.SourceTypeScreen,
		AudioMode:   streammeta.AudioModeDesktop,
	}
	if mutate != nil {
		mutate(&claims)
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("unable to sign token: %v", err)
	}
	return token
}

func TestValidateJoinTokenAcceptsRouteBoundPublisherToken(t *testing.T) {
	app := testStreamApp()
	token := signedStreamToken(t, app.cfg.AuthSecret, nil)

	userID, streamID, voiceChannelID, guildID, perms, role, sourceType, audioMode, err := app.validateJoinToken(token)
	if err != nil {
		t.Fatalf("validateJoinToken returned error: %v", err)
	}
	if userID != 10 || streamID != 777 || voiceChannelID != 42 || guildID == nil || *guildID != 1 {
		t.Fatalf("unexpected identity claims user=%d stream=%d voice=%d guild=%v", userID, streamID, voiceChannelID, guildID)
	}
	expectedPerms := int64(permissions.PermVoiceSpeak | permissions.PermVoiceVideo)
	if perms != expectedPerms || role != streammeta.RolePublisher || sourceType != streammeta.SourceTypeScreen || audioMode != streammeta.AudioModeDesktop {
		t.Fatalf("unexpected media claims perms=%d role=%q source=%q audio=%q", perms, role, sourceType, audioMode)
	}
}

func TestValidateJoinTokenRejectsWrongRoute(t *testing.T) {
	app := testStreamApp()
	token := signedStreamToken(t, app.cfg.AuthSecret, func(claims *streammeta.Claims) {
		claims.RouteID = "stream-b"
	})

	if _, _, _, _, _, _, _, _, err := app.validateJoinToken(token); err == nil {
		t.Fatal("expected wrong route token to be rejected")
	}
}

func TestAuthorizeJoinFieldsRejectsWrongStreamID(t *testing.T) {
	app := testStreamApp()
	token := signedStreamToken(t, app.cfg.AuthSecret, nil)

	if _, _, _, _, _, _, _, _, err := app.authorizeJoinFields(778, token); err == nil {
		t.Fatal("expected wrong stream id token to be rejected")
	}
}

func TestValidateJoinTokenRejectsExpiredToken(t *testing.T) {
	app := testStreamApp()
	token := signedStreamToken(t, app.cfg.AuthSecret, func(claims *streammeta.Claims) {
		now := time.Now()
		claims.IssuedAt = jwt.NewNumericDate(now.Add(-2 * time.Minute))
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Minute))
	})

	if _, _, _, _, _, _, _, _, err := app.validateJoinToken(token); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

func TestValidateJoinTokenRejectsInvalidRole(t *testing.T) {
	app := testStreamApp()
	token := signedStreamToken(t, app.cfg.AuthSecret, func(claims *streammeta.Claims) {
		claims.Role = "admin"
	})

	if _, _, _, _, _, _, _, _, err := app.validateJoinToken(token); err == nil {
		t.Fatal("expected invalid role token to be rejected")
	}
}

func TestValidateJoinTokenRejectsPublisherOwnerMismatch(t *testing.T) {
	app := testStreamApp()
	token := signedStreamToken(t, app.cfg.AuthSecret, func(claims *streammeta.Claims) {
		claims.UserID = claims.OwnerUserID + 1
	})

	if _, _, _, _, _, _, _, _, err := app.validateJoinToken(token); err == nil {
		t.Fatal("expected publisher token for a different owner to be rejected")
	}
}

func TestValidateJoinTokenViewerTokenIsReceiveOnly(t *testing.T) {
	app := testStreamApp()
	token := signedStreamToken(t, app.cfg.AuthSecret, func(claims *streammeta.Claims) {
		claims.Role = streammeta.RoleViewer
	})

	_, _, _, _, perms, role, _, _, err := app.validateJoinToken(token)
	if err != nil {
		t.Fatalf("validateJoinToken returned error: %v", err)
	}
	if role != streammeta.RoleViewer {
		t.Fatalf("expected viewer role, got %q", role)
	}
	if perms != 0 {
		t.Fatalf("expected viewer token to grant no publish permissions, got %d", perms)
	}
}
