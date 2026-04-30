package guild

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
)

func parseStreamToken(secret, token string) (streammeta.Claims, error) {
	var claims streammeta.Claims
	_, err := jwt.ParseWithClaims(token, &claims, func(tok *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	return claims, err
}

func decodeStreamToken(t *testing.T, secret, token string) streammeta.Claims {
	t.Helper()

	claims, err := parseStreamToken(secret, token)
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
		authSecret:         "auth-secret",
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
		authSecret:         "auth-secret",
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
	claims := decodeStreamToken(t, e.authSecret, first.StreamToken)
	if claims.Role != streammeta.RolePublisher || claims.RouteID != "stream-us-1" || claims.OwnerUserID != 10 {
		t.Fatalf("unexpected publisher token claims: %#v", claims)
	}
	if _, err := parseStreamToken("stream-secret", first.StreamToken); err == nil {
		t.Fatalf("expected stream token to reject legacy stream secret")
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil || time.Duration(claims.ExpiresAt.Unix()-claims.IssuedAt.Unix())*time.Second != streamTokenTTL {
		t.Fatalf("expected %s stream token TTL, got iat=%v exp=%v", streamTokenTTL, claims.IssuedAt, claims.ExpiresAt)
	}
	if got := streamDisco.listCalls["us-east"]; got != 1 {
		t.Fatalf("expected a single discovery lookup, got %d", got)
	}
}

func TestStartStreamsInSameVoiceChannelCanUseDifferentServices(t *testing.T) {
	cache := &fakeCache{jsonValues: map[string][]byte{}}
	seedVoiceMembership(t, cache, 42, 10)
	seedVoiceMembership(t, cache, 42, 11)

	streamDisco := &fakeDiscoveryManager{
		lists: map[string][]discovery.Instance{
			"us-east": {
				{ID: "stream-us-1", Region: "us-east", URL: "wss://stream-us-1.example/signal", Load: 0},
				{ID: "stream-us-2", Region: "us-east", URL: "wss://stream-us-2.example/signal", Load: 0},
			},
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
		authSecret:         "auth-secret",
		perm: &fakePermissionChecker{
			channel:      &model.Channel{Id: 42, Type: model.ChannelTypeGuildVoice},
			channelOK:    true,
			channelPerms: int64(permissions.PermVoiceConnect | permissions.PermVoiceVideo),
		},
	}

	body := `{"source_type":"screen","audio_mode":"desktop"}`
	app10 := newGuildTestApp(t, 10, "/guild/:guild_id/voice/:channel_id/streams", e.StartStream)
	req10 := httptest.NewRequest("POST", "/guild/1/voice/42/streams", strings.NewReader(body))
	req10.Header.Set("Content-Type", "application/json")
	resp10, err := app10.Test(req10, -1)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if resp10.StatusCode != fiber.StatusOK {
		t.Fatalf("expected first status 200, got %d", resp10.StatusCode)
	}
	var first CreateVoiceStreamResponse
	if err := json.NewDecoder(resp10.Body).Decode(&first); err != nil {
		t.Fatalf("unable to decode first response: %v", err)
	}

	app11 := newGuildTestApp(t, 11, "/guild/:guild_id/voice/:channel_id/streams", e.StartStream)
	req11 := httptest.NewRequest("POST", "/guild/1/voice/42/streams", strings.NewReader(body))
	req11.Header.Set("Content-Type", "application/json")
	resp11, err := app11.Test(req11, -1)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	if resp11.StatusCode != fiber.StatusOK {
		t.Fatalf("expected second status 200, got %d", resp11.StatusCode)
	}
	var second CreateVoiceStreamResponse
	if err := json.NewDecoder(resp11.Body).Decode(&second); err != nil {
		t.Fatalf("unable to decode second response: %v", err)
	}

	firstClaims := decodeStreamToken(t, e.authSecret, first.StreamToken)
	secondClaims := decodeStreamToken(t, e.authSecret, second.StreamToken)
	if firstClaims.RouteID == "" || secondClaims.RouteID == "" {
		t.Fatalf("expected route-bound tokens, got first=%#v second=%#v", firstClaims, secondClaims)
	}
	if firstClaims.RouteID == secondClaims.RouteID {
		t.Fatalf("expected reservations to spread same-channel streams across services, got %q twice", firstClaims.RouteID)
	}
	if first.StreamURL == second.StreamURL {
		t.Fatalf("expected different stream service URLs, got %q twice", first.StreamURL)
	}
	if got := streamDisco.listCalls["us-east"]; got != 1 {
		t.Fatalf("expected one discovery lookup shared by snapshot, got %d", got)
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
		authSecret:         "auth-secret",
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
	claims := decodeStreamToken(t, e.authSecret, body.StreamToken)
	if claims.Role != streammeta.RoleViewer {
		t.Fatalf("expected viewer role, got %q", claims.Role)
	}
	if claims.StreamID != meta.ID || claims.ChannelID != meta.ChannelID || claims.GuildID != meta.GuildID {
		t.Fatalf("unexpected stream token claims: %#v", claims)
	}
	if claims.OwnerUserID != meta.OwnerUserID || claims.RouteID != meta.RouteID {
		t.Fatalf("expected token bound to streamer and route, got %#v", claims)
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil || time.Duration(claims.ExpiresAt.Unix()-claims.IssuedAt.Unix())*time.Second != time.Minute {
		t.Fatalf("expected 1 minute viewer token TTL, got iat=%v exp=%v", claims.IssuedAt, claims.ExpiresAt)
	}
}

func TestJoinStreamTokenUsesCurrentRouteBinding(t *testing.T) {
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
		RouteID:     "stream-us-old",
		RouteURL:    "wss://stream-old.example/signal",
	}
	if err := cache.SetJSON(context.Background(), streammeta.MetaKey(meta.ID), meta); err != nil {
		t.Fatalf("unable to seed stream metadata: %v", err)
	}
	currentRoute := streammeta.RouteBinding{ID: "stream-us-new", URL: "wss://stream-new.example/signal", Region: "us-east"}
	if err := cache.SetJSON(context.Background(), streammeta.RouteKey(meta.ID), currentRoute); err != nil {
		t.Fatalf("unable to seed current stream route: %v", err)
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
		authSecret:         "auth-secret",
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
	if body.StreamURL != currentRoute.URL {
		t.Fatalf("expected response to use current route %q, got %q", currentRoute.URL, body.StreamURL)
	}
	claims := decodeStreamToken(t, e.authSecret, body.StreamToken)
	if claims.RouteID != currentRoute.ID {
		t.Fatalf("expected token route %q, got %#v", currentRoute.ID, claims)
	}
}

func TestListStreamsDoesNotRequireVoiceMembership(t *testing.T) {
	cache := &fakeCache{jsonValues: map[string][]byte{}}
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
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("unable to marshal metadata: %v", err)
	}
	if err := cache.HSet(context.Background(), streammeta.ChannelKey(meta.ChannelID), fmtInt64(meta.ID), string(raw)); err != nil {
		t.Fatalf("unable to seed channel hash: %v", err)
	}

	e := &entity{
		cache: cache,
		perm: &fakePermissionChecker{
			channel:      &model.Channel{Id: 42, Type: model.ChannelTypeGuildVoice},
			channelOK:    true,
			channelPerms: int64(permissions.PermVoiceConnect),
		},
	}
	app := newGuildTestApp(t, 10, "/guild/:guild_id/voice/:channel_id/streams", e.ListStreams)

	req := httptest.NewRequest("GET", "/guild/1/voice/42/streams", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var streams []VoiceStreamSummary
	if err := json.NewDecoder(resp.Body).Decode(&streams); err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}
	if len(streams) != 1 || streams[0].ID != meta.ID {
		t.Fatalf("expected seeded stream summary, got %#v", streams)
	}
}

func TestJoinStreamStillRequiresVoiceMembership(t *testing.T) {
	cache := &fakeCache{jsonValues: map[string][]byte{}}
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
	raw, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("unable to marshal metadata: %v", err)
	}
	if err := cache.HSet(context.Background(), streammeta.ChannelKey(meta.ChannelID), fmtInt64(meta.ID), string(raw)); err != nil {
		t.Fatalf("unable to seed channel hash: %v", err)
	}

	e := &entity{
		cache: cache,
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
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
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
