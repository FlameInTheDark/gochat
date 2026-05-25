package gatewaystate

import (
	"context"
	"testing"

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

func TestConnectionAndClientStateRoundTrip(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	conn := ConnectionState{
		ConnectionID:     "conn-1",
		UserID:           42,
		SessionID:        "session-1",
		ClientInstanceID: "client-1",
		Generation:       3,
		ProtocolVersion:  2,
		LastSeq:          9,
	}
	if err := store.UpsertConnection(ctx, conn, 60); err != nil {
		t.Fatalf("upsert connection: %v", err)
	}
	if err := store.TouchConnection(ctx, "conn-1", 10, 60); err != nil {
		t.Fatalf("touch connection: %v", err)
	}

	state := ClientState{Channels: []int64{1, 2}, PresenceSet: []int64{7, 8}}
	if err := store.SetClientState(ctx, 42, "client-1", state, 60); err != nil {
		t.Fatalf("set client state: %v", err)
	}
	got, ok, err := store.GetClientState(ctx, 42, "client-1")
	if err != nil {
		t.Fatalf("get client state: %v", err)
	}
	if !ok {
		t.Fatalf("expected client state")
	}
	if len(got.Channels) != 2 || got.Channels[0] != 1 || got.Channels[1] != 2 {
		t.Fatalf("unexpected channels: %#v", got.Channels)
	}
	if len(got.PresenceSet) != 2 || got.PresenceSet[0] != 7 || got.PresenceSet[1] != 8 {
		t.Fatalf("unexpected presence set: %#v", got.PresenceSet)
	}

	if err := store.DeleteConnection(ctx, "conn-1", 42); err != nil {
		t.Fatalf("delete connection: %v", err)
	}
}
