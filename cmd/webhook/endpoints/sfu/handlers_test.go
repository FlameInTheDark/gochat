package sfu

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	cachepkg "github.com/FlameInTheDark/gochat/internal/cache"
)

type fakeRouteCache struct {
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
func (f *fakeRouteCache) GetWithTTL(ctx context.Context, key string) (string, cachepkg.LookupMeta, error) {
	raw, ok := f.jsonValues[key]
	if !ok {
		return "", cachepkg.LookupMeta{}, nil
	}
	return string(raw), cachepkg.LookupMeta{Hit: true, TTL: time.Hour, HasTTL: true}, nil
}
func (f *fakeRouteCache) Delete(ctx context.Context, key string) error { return nil }
func (f *fakeRouteCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	raw, ok := f.jsonValues[key]
	if !ok {
		return nil, errors.New("cache miss")
	}
	return append([]byte(nil), raw...), nil
}
func (f *fakeRouteCache) SetTimed(ctx context.Context, key, val string, ttl int64) error { return nil }
func (f *fakeRouteCache) SetTimedInt64(ctx context.Context, key string, val int64, ttl int64) error {
	return nil
}
func (f *fakeRouteCache) SetInt64(ctx context.Context, key string, val int64) error { return nil }
func (f *fakeRouteCache) SetTTL(ctx context.Context, key string, ttl int64) error   { return nil }
func (f *fakeRouteCache) Incr(ctx context.Context, key string) (int64, error)       { return 0, nil }
func (f *fakeRouteCache) GetInt64(ctx context.Context, key string) (int64, error)   { return 0, nil }
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
func (f *fakeRouteCache) GetJSONWithTTL(ctx context.Context, key string, v interface{}) (cachepkg.LookupMeta, error) {
	raw, ok := f.jsonValues[key]
	if !ok {
		return cachepkg.LookupMeta{}, nil
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return cachepkg.LookupMeta{}, err
	}
	return cachepkg.LookupMeta{Hit: true, TTL: time.Hour, HasTTL: true}, nil
}
func (f *fakeRouteCache) TryAcquireRefreshLock(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	return true, nil
}
func (f *fakeRouteCache) ReleaseRefreshLock(ctx context.Context, key, token string) error { return nil }
func (f *fakeRouteCache) HGet(ctx context.Context, key, field string) (string, error)     { return "", nil }
func (f *fakeRouteCache) HSet(ctx context.Context, key, field, value string) error        { return nil }
func (f *fakeRouteCache) HDel(ctx context.Context, key, field string) error               { return nil }
func (f *fakeRouteCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return nil, nil
}
func (f *fakeRouteCache) HGetAllMulti(_ context.Context, keys []string) ([]map[string]string, error) {
	return make([]map[string]string, len(keys)), nil
}
func (f *fakeRouteCache) MGetBytes(_ context.Context, keys ...string) ([][]byte, error) {
	return make([][]byte, len(keys)), nil
}
func (f *fakeRouteCache) HIncrBy(ctx context.Context, key, field string, delta int64) (int64, error) {
	return 0, nil
}
func (f *fakeRouteCache) SetTimedJSONBatch(_ context.Context, keys []string, _ []interface{}, _ int64, _ ...cachepkg.TimedOption) error {
	return nil
}
func (f *fakeRouteCache) ZAddBatch(_ context.Context, _ string, _ []cachepkg.ZBatchMember) error {
	return nil
}
func (f *fakeRouteCache) ZAdd(ctx context.Context, key string, score float64, member string) error {
	return nil
}
func (f *fakeRouteCache) ZRem(ctx context.Context, key string, members ...string) error { return nil }
func (f *fakeRouteCache) ZRevRangeByScore(ctx context.Context, key, max, min string, offset, count int64) ([]string, error) {
	return nil, nil
}
func (f *fakeRouteCache) XAdd(ctx context.Context, stream string, maxLen int64, approx bool, values map[string]interface{}) error {
	return nil
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
