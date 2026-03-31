package sfu

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	cachepkg "github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/cache/testutil"
)

type fakeRouteCache struct {
	testutil.Noop
	jsonValues map[string][]byte
}

func (f *fakeRouteCache) ensure() {
	if f.jsonValues == nil {
		f.jsonValues = make(map[string][]byte)
	}
}

func (f *fakeRouteCache) Set(ctx context.Context, key, val string) error { return nil }
func (f *fakeRouteCache) Get(ctx context.Context, key string) (string, error) {
	raw, ok := f.jsonValues[key]
	if !ok {
		return "", errors.New("cache miss")
	}
	return string(raw), nil
}
func (f *fakeRouteCache) SetJSON(ctx context.Context, key string, val interface{}) error {
	f.ensure()
	raw, err := json.Marshal(val)
	if err != nil {
		return err
	}
	f.jsonValues[key] = raw
	return nil
}
func (f *fakeRouteCache) SetTimedJSON(ctx context.Context, key string, val interface{}, ttl int64, _ ...cachepkg.TimedOption) error {
	return f.SetJSON(ctx, key, val)
}
func (f *fakeRouteCache) SetTimedJSONNX(ctx context.Context, key string, val interface{}, ttl int64, _ ...cachepkg.TimedOption) (bool, error) {
	if _, ok := f.jsonValues[key]; ok {
		return false, nil
	}
	return true, f.SetJSON(ctx, key, val)
}
func (f *fakeRouteCache) GetJSON(ctx context.Context, key string, v interface{}) error {
	raw, ok := f.jsonValues[key]
	if !ok {
		return errors.New("cache miss")
	}
	return json.Unmarshal(raw, v)
}

func TestRefreshVoiceRouteWritesBinding(t *testing.T) {
	cache := &fakeRouteCache{jsonValues: map[string][]byte{}}

	if err := refreshVoiceRoute(context.Background(), cache, 12, "sfu-1", "wss://one.example/signal", "eu-central"); err != nil {
		t.Fatalf("refreshVoiceRoute returned error: %v", err)
	}

	var binding voiceRouteBinding
	if err := cache.GetJSON(context.Background(), "voice:route:12", &binding); err != nil {
		t.Fatalf("expected stored route binding: %v", err)
	}
	if binding.ID != "sfu-1" || binding.URL != "wss://one.example/signal" || binding.Region != "eu-central" {
		t.Fatalf("unexpected binding: %#v", binding)
	}
}

func TestRefreshVoiceRouteIgnoresOldSFUDuringRebind(t *testing.T) {
	cache := &fakeRouteCache{jsonValues: map[string][]byte{}}
	newBinding := voiceRouteBinding{ID: "sfu-new", URL: "wss://new.example/signal", Region: "eu-central"}
	if err := cache.SetJSON(context.Background(), "voice:rebind:44", newBinding); err != nil {
		t.Fatalf("unable to seed rebind marker: %v", err)
	}
	if err := cache.SetJSON(context.Background(), "voice:route:44", newBinding); err != nil {
		t.Fatalf("unable to seed current route: %v", err)
	}

	if err := refreshVoiceRoute(context.Background(), cache, 44, "sfu-old", "wss://old.example/signal", "us-east"); err != nil {
		t.Fatalf("refreshVoiceRoute returned error: %v", err)
	}

	var binding voiceRouteBinding
	if err := cache.GetJSON(context.Background(), "voice:route:44", &binding); err != nil {
		t.Fatalf("expected stored route binding: %v", err)
	}
	if binding != newBinding {
		t.Fatalf("expected old SFU rewrite to be ignored, got %#v", binding)
	}

	if err := refreshVoiceRoute(context.Background(), cache, 44, "sfu-new", "wss://new.example/signal", "eu-central"); err != nil {
		t.Fatalf("refreshVoiceRoute returned error for new route: %v", err)
	}
	if err := cache.GetJSON(context.Background(), "voice:route:44", &binding); err != nil {
		t.Fatalf("expected stored route binding: %v", err)
	}
	if binding != newBinding {
		t.Fatalf("expected new SFU rewrite to succeed, got %#v", binding)
	}
}
