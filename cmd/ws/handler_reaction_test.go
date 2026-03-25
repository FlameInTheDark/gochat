package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	reactionutil "github.com/FlameInTheDark/gochat/internal/reaction"
)

type fakeReactionCache struct {
	values map[string]map[string]string
}

func (f *fakeReactionCache) Set(ctx context.Context, key, val string) error      { return nil }
func (f *fakeReactionCache) Get(ctx context.Context, key string) (string, error) { return "", nil }
func (f *fakeReactionCache) Delete(ctx context.Context, key string) error        { return nil }
func (f *fakeReactionCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}
func (f *fakeReactionCache) SetTimed(ctx context.Context, key, val string, ttl int64) error {
	return nil
}
func (f *fakeReactionCache) SetTimedInt64(ctx context.Context, key string, val int64, ttl int64) error {
	return nil
}
func (f *fakeReactionCache) SetInt64(ctx context.Context, key string, val int64) error { return nil }
func (f *fakeReactionCache) SetTTL(ctx context.Context, key string, ttl int64) error   { return nil }
func (f *fakeReactionCache) Incr(ctx context.Context, key string) (int64, error)       { return 0, nil }
func (f *fakeReactionCache) GetInt64(ctx context.Context, key string) (int64, error)   { return 0, nil }
func (f *fakeReactionCache) SetJSON(ctx context.Context, key string, val interface{}) error {
	return nil
}
func (f *fakeReactionCache) SetTimedJSON(ctx context.Context, key string, val interface{}, ttl int64) error {
	return nil
}
func (f *fakeReactionCache) SetTimedJSONNX(ctx context.Context, key string, val interface{}, ttl int64) (bool, error) {
	return true, nil
}
func (f *fakeReactionCache) GetJSON(ctx context.Context, key string, v interface{}) error { return nil }
func (f *fakeReactionCache) HGet(ctx context.Context, key, field string) (string, error) {
	if f.values[key] == nil {
		return "", nil
	}
	return f.values[key][field], nil
}
func (f *fakeReactionCache) HSet(ctx context.Context, key, field, value string) error { return nil }
func (f *fakeReactionCache) HDel(ctx context.Context, key, field string) error        { return nil }
func (f *fakeReactionCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (f *fakeReactionCache) HIncrBy(ctx context.Context, key, field string, delta int64) (int64, error) {
	return 0, nil
}
func (f *fakeReactionCache) ZAdd(ctx context.Context, key string, score float64, member string) error {
	return nil
}
func (f *fakeReactionCache) ZRem(ctx context.Context, key string, members ...string) error {
	return nil
}
func (f *fakeReactionCache) ZRevRangeByScore(ctx context.Context, key, max, min string, offset, count int64) ([]string, error) {
	return nil, nil
}
func (f *fakeReactionCache) XAdd(ctx context.Context, stream string, maxLen int64, approx bool, values map[string]interface{}) error {
	return nil
}
func (f *fakeReactionCache) HGetAllMulti(_ context.Context, keys []string) ([]map[string]string, error) {
	return make([]map[string]string, len(keys)), nil
}
func (f *fakeReactionCache) MGetBytes(_ context.Context, keys ...string) ([][]byte, error) {
	return make([][]byte, len(keys)), nil
}
func (f *fakeReactionCache) SetTimedJSONBatch(_ context.Context, _ []string, _ []interface{}, _ int64) error {
	return nil
}
func (f *fakeReactionCache) ZAddBatch(_ context.Context, _ string, _ []cache.ZBatchMember) error {
	return nil
}

func decodeReactionAdd(t *testing.T, payload []byte) mqmsg.MessageReactionAdd {
	t.Helper()
	var envelope mqmsg.Message
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	var message mqmsg.MessageReactionAdd
	if err := json.Unmarshal(envelope.Data, &message); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	return message
}

func TestPersonalizeReactionMessageUsesCacheForMe(t *testing.T) {
	event, err := mqmsg.BuildEventMessage(&mqmsg.MessageReactionAdd{
		ChannelId: 7,
		MessageId: 9,
		Reaction: dto.MessageReaction{
			Count: 3,
			Emoji: dto.MessageReactionEmoji{Name: "❤️"},
		},
	})
	if err != nil {
		t.Fatalf("BuildEventMessage returned error: %v", err)
	}
	wire, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	cache := &fakeReactionCache{
		values: map[string]map[string]string{
			reactionutil.UserKey(9, 55): {
				reactionutil.SystemBucketKey("❤️"): "123",
			},
		},
	}

	personalized := personalizeMessageForRecipientWithCache(cache, "channel.7", 55, wire)
	got := decodeReactionAdd(t, personalized)
	if !got.Reaction.Me {
		t.Fatalf("expected personalized reaction me=true, got %#v", got.Reaction)
	}
}
