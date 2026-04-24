package guild

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
)

func decodeStreamToken(t *testing.T, secret, token string) streammeta.Claims {
	t.Helper()

	var claims streammeta.Claims
	_, err := jwt.ParseWithClaims(token, &claims, func(tok *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("unable to parse stream token: %v", err)
	}
	return claims
}

func seedVoiceMembership(t *testing.T, cache *fakeCache, channelID, userID int64) {
	t.Helper()
	if err := cache.HSet(context.Background(), sessionHashKey(channelID), fmtInt64(userID), "true"); err != nil {
		t.Fatalf("unable to seed voice membership: %v", err)
	}
}

func TestStartStreamUsesStrictVoiceRegion(t *testing.T) {
	cache := &fakeCache{jsonValues: map[string][]byte{}}
	seedVoiceMembership(t, cache, 42, 10)

	streamDisco := &fakeDiscoveryManager{
		lists: map[string][]discovery.Instance{
			"us-east":    nil,
			"eu-central": {{ID: "stream-eu-1", Region: "eu-central", URL: "wss://stream-eu.example/signal", Load: 1}},
		},
	}
	e := &entity{
		cache:              cache,
		ch:                 &fakeCreateChannelRepo{},
		defaultVoiceRegion: "us-east",
		streamDisco:        streamDisco,
		allowedRegions:     map[string]struct{}{"us-east": {}, "eu-central": {}},
		allowedRegionIDs:   []string{"eu-central", "us-east"},
		streamSelector:     newVoiceSelector(newVoiceTestLogger()),
		streamAuthSecret:   "stream-secret",
		perm: &fakePermissionChecker{
			channel:      &model.Channel{Id: 42, Type: model.ChannelTypeGuildVoice},
			channelOK:    true,
			channelPerms: int64(permissions.PermVoiceConnect | permissions.PermVoiceVideo),
		},
	}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/voice/:channel_id/streams", e.StartStream)

	req := httptest.NewRequest("POST", "/guild/1/voice/42/streams", strings.NewReader(`{"source_type":"screen","audio_mode":"desktop"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}
	if got := streamDisco.listCalls["us-east"]; got != 1 {
		t.Fatalf("expected one us-east lookup, got %d", got)
	}
	if got := streamDisco.listCalls["eu-central"]; got != 0 {
		t.Fatalf("expected no fallback lookup, got %d", got)
	}
}

func TestStartStreamIsIdempotentPerUserAndChannel(t *testing.T) {
	cache := &fakeCache{jsonValues: map[string][]byte{}}
	seedVoiceMembership(t, cache, 42, 10)

	streamDisco := &fakeDiscoveryManager{
		lists: map[string][]discovery.Instance{
			"us-east": {{ID: "stream-us-1", Region: "us-east", URL: "wss://stream-us.example/signal", Load: 1}},
		},
	}
	e := &entity{
		cache:              cache,
		ch:                 &fakeCreateChannelRepo{},
		defaultVoiceRegion: "us-east",
		streamDisco:        streamDisco,
		allowedRegions:     map[string]struct{}{"us-east": {}},
		allowedRegionIDs:   []string{"us-east"},
		streamSelector:     newVoiceSelector(newVoiceTestLogger()),
		streamAuthSecret:   "stream-secret",
		perm: &fakePermissionChecker{
			channel:      &model.Channel{Id: 42, Type: model.ChannelTypeGuildVoice},
			channelOK:    true,
			channelPerms: int64(permissions.PermVoiceConnect | permissions.PermVoiceVideo),
		},
	}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/voice/:channel_id/streams", e.StartStream)

	body := `{"source_type":"screen","audio_mode":"desktop"}`
	req1 := httptest.NewRequest("POST", "/guild/1/voice/42/streams", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	resp1, err := app.Test(req1, -1)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if resp1.StatusCode != fiber.StatusOK {
		t.Fatalf("expected first status 200, got %d", resp1.StatusCode)
	}
	var first CreateVoiceStreamResponse
	if err := json.NewDecoder(resp1.Body).Decode(&first); err != nil {
		t.Fatalf("unable to decode first response: %v", err)
	}

	req2 := httptest.NewRequest("POST", "/guild/1/voice/42/streams", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app.Test(req2, -1)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	if resp2.StatusCode != fiber.StatusOK {
		t.Fatalf("expected second status 200, got %d", resp2.StatusCode)
	}
	var second CreateVoiceStreamResponse
	if err := json.NewDecoder(resp2.Body).Decode(&second); err != nil {
		t.Fatalf("unable to decode second response: %v", err)
	}

	if first.StreamID == 0 || first.StreamID != second.StreamID {
		t.Fatalf("expected idempotent stream id, got first=%d second=%d", first.StreamID, second.StreamID)
	}
	if first.Stream.OwnerUserID != 10 || second.Stream.OwnerUserID != 10 {
		t.Fatalf("expected owner_user_id to round-trip in stream summaries, got first=%d second=%d", first.Stream.OwnerUserID, second.Stream.OwnerUserID)
	}
	if got := streamDisco.listCalls["us-east"]; got != 1 {
		t.Fatalf("expected a single discovery lookup, got %d", got)
	}
}

func TestJoinStreamIssuesViewerOnlyToken(t *testing.T) {
	cache := &fakeCache{jsonValues: map[string][]byte{}}
	seedVoiceMembership(t, cache, 42, 10)

	meta := streammeta.Metadata{
		ActiveStream: streammeta.ActiveStream{
			ID:         777,
			ChannelID:  42,
			SourceType: streammeta.SourceTypeScreen,
			AudioMode:  streammeta.AudioModeDesktop,
			StartedAt:  12345,
		},
		GuildID:     1,
		OwnerUserID: 55,
		Region:      "us-east",
		RouteID:     "stream-us-1",
		RouteURL:    "wss://stream-us.example/signal",
	}
	if err := cache.SetJSON(context.Background(), streammeta.MetaKey(meta.ID), meta); err != nil {
		t.Fatalf("unable to seed stream metadata: %v", err)
	}
	if err := cache.SetJSON(context.Background(), streammeta.RouteKey(meta.ID), streammeta.RouteBinding{ID: meta.RouteID, URL: meta.RouteURL, Region: meta.Region}); err != nil {
		t.Fatalf("unable to seed stream route: %v", err)
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("unable to marshal metadata: %v", err)
	}
	if err := cache.HSet(context.Background(), streammeta.ChannelKey(meta.ChannelID), fmtInt64(meta.ID), string(raw)); err != nil {
		t.Fatalf("unable to seed channel hash: %v", err)
	}

	e := &entity{
		cache:              cache,
		ch:                 &fakeCreateChannelRepo{},
		defaultVoiceRegion: "us-east",
		streamSelector:     newVoiceSelector(newVoiceTestLogger()),
		streamAuthSecret:   "stream-secret",
		perm: &fakePermissionChecker{
			channel:      &model.Channel{Id: 42, Type: model.ChannelTypeGuildVoice},
			channelOK:    true,
			channelPerms: int64(permissions.PermVoiceConnect),
		},
	}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/voice/:channel_id/streams/:stream_id/join", e.JoinStream)

	req := httptest.NewRequest("POST", "/guild/1/voice/42/streams/777/join", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body JoinVoiceStreamResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}
	claims := decodeStreamToken(t, e.streamAuthSecret, body.StreamToken)
	if claims.Role != streammeta.RoleViewer {
		t.Fatalf("expected viewer role, got %q", claims.Role)
	}
	if claims.StreamID != meta.ID || claims.ChannelID != meta.ChannelID || claims.GuildID != meta.GuildID {
		t.Fatalf("unexpected stream token claims: %#v", claims)
	}
}

func TestStopStreamIsIdempotentWhenStreamIsAlreadyGone(t *testing.T) {
	cache := &fakeCache{jsonValues: map[string][]byte{}}
	e := &entity{cache: cache}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/voice/:channel_id/streams/:stream_id", e.StopStream)

	req := httptest.NewRequest("DELETE", "/guild/1/voice/42/streams/777", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}
