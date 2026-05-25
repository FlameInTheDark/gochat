package presence

import (
	"context"
	"testing"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/alicebob/miniredis/v2"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	redis := miniredis.RunT(t)
	cache, err := kvs.New(redis.Addr())
	if err != nil {
		t.Fatalf("new cache: %v", err)
	}
	t.Cleanup(func() { _ = cache.Close() })
	return NewStore(cache)
}

func TestRemoveSessionIfOwnerIgnoresStaleGeneration(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().Unix()
	const userID = int64(10)
	const ttl = int64(60)

	oldSession := SessionPresence{
		SessionID:    "session-1",
		ConnectionID: "conn-old",
		Generation:   1,
		Status:       StatusOnline,
		Platform:     "web",
		Since:        now,
		UpdatedAt:    now,
		ExpiresAt:    now + ttl,
	}
	newSession := oldSession
	newSession.ConnectionID = "conn-new"
	newSession.Generation = 2
	newSession.UpdatedAt = now + 1

	if err := store.UpsertSession(ctx, userID, oldSession.SessionID, oldSession, ttl); err != nil {
		t.Fatalf("upsert old session: %v", err)
	}
	if err := store.UpsertSession(ctx, userID, newSession.SessionID, newSession, ttl); err != nil {
		t.Fatalf("upsert new session: %v", err)
	}

	removed, err := store.RemoveSessionIfOwner(ctx, userID, "session-1", "conn-old", 1, ttl)
	if err != nil {
		t.Fatalf("remove stale owner: %v", err)
	}
	if !removed {
		t.Fatalf("expected old generation to be removed")
	}

	agg, ok, err := store.Aggregate(ctx, userID, now+2)
	if err != nil {
		t.Fatalf("aggregate: %v", err)
	}
	if !ok || agg.Status != StatusOnline {
		t.Fatalf("expected new generation to keep user online, got ok=%v presence=%#v", ok, agg)
	}
	if got := agg.ClientStatus["web"]; got != StatusOnline {
		t.Fatalf("expected web client status online, got %q", got)
	}
}

func TestAggregateReturnsOfflineWithoutStaleVoice(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().Unix()
	const userID = int64(11)
	const ttl = int64(1)
	voiceID := int64(44)

	err := store.UpsertSession(ctx, userID, "session-1", SessionPresence{
		SessionID:      "session-1",
		ConnectionID:   "conn-1",
		Generation:     1,
		Status:         StatusOnline,
		Platform:       "web",
		Since:          now - 10,
		UpdatedAt:      now - 10,
		ExpiresAt:      now - 1,
		VoiceChannelID: &voiceID,
	}, ttl)
	if err != nil {
		t.Fatalf("upsert expired session: %v", err)
	}

	agg, ok, err := store.Aggregate(ctx, userID, now)
	if err != nil {
		t.Fatalf("aggregate: %v", err)
	}
	if ok || agg.Status != StatusOffline {
		t.Fatalf("expected offline aggregate, got ok=%v presence=%#v", ok, agg)
	}
	if agg.VoiceChannelID != nil {
		t.Fatalf("expected stale voice to be cleared, got %v", *agg.VoiceChannelID)
	}
}

func TestReconcileTouchedPublishesCorrectionForExpiredLease(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()
	now := time.Now().Unix()
	const userID = int64(12)
	const ttl = int64(60)
	voiceID := int64(55)

	err := store.UpsertSession(ctx, userID, "session-1", SessionPresence{
		SessionID:      "session-1",
		ConnectionID:   "conn-1",
		Generation:     1,
		Status:         StatusOnline,
		Platform:       "web",
		Since:          now - 60,
		UpdatedAt:      now - 60,
		ExpiresAt:      now - 1,
		VoiceChannelID: &voiceID,
	}, ttl)
	if err != nil {
		t.Fatalf("upsert expired session: %v", err)
	}
	if err := store.SetAggregated(ctx, Presence{UserID: userID, Status: StatusOnline, VoiceChannelID: &voiceID}, ttl); err != nil {
		t.Fatalf("set stale aggregate: %v", err)
	}

	published, err := ReconcileTouched(ctx, store, nil, ttl, 100)
	if err != nil {
		t.Fatalf("reconcile touched: %v", err)
	}
	if published != 1 {
		t.Fatalf("expected one correction publish, got %d", published)
	}

	agg, ok, err := store.CachedAggregate(ctx, userID)
	if err != nil {
		t.Fatalf("cached aggregate: %v", err)
	}
	if !ok || agg.Status != StatusOffline {
		t.Fatalf("expected cached offline aggregate, ok=%v presence=%#v", ok, agg)
	}
	if agg.VoiceChannelID != nil {
		t.Fatalf("expected reconciler to clear stale voice, got %v", *agg.VoiceChannelID)
	}
}
