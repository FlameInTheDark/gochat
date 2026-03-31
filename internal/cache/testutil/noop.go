package testutil

import (
	"context"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
)

type Noop struct{}

func (Noop) Set(context.Context, string, string) error { return nil }

func (Noop) Get(context.Context, string) (string, error) { return "", nil }

func (Noop) GetWithTTL(context.Context, string) (string, cache.LookupMeta, error) {
	return "", cache.LookupMeta{Hit: false, TTL: 0, HasTTL: false}, nil
}

func (Noop) Delete(context.Context, string) error { return nil }

func (Noop) GetBytes(context.Context, string) ([]byte, error) { return nil, nil }

func (Noop) SetTimed(context.Context, string, string, int64) error { return nil }

func (Noop) SetTimedInt64(context.Context, string, int64, int64) error { return nil }

func (Noop) SetInt64(context.Context, string, int64) error { return nil }

func (Noop) SetTTL(context.Context, string, int64) error { return nil }

func (Noop) Incr(context.Context, string) (int64, error) { return 0, nil }

func (Noop) GetInt64(context.Context, string) (int64, error) { return 0, nil }

func (Noop) SetJSON(context.Context, string, interface{}) error { return nil }

func (Noop) SetTimedJSON(context.Context, string, interface{}, int64, ...cache.TimedOption) error {
	return nil
}

func (Noop) SetTimedJSONNX(context.Context, string, interface{}, int64, ...cache.TimedOption) (bool, error) {
	return false, nil
}

func (Noop) GetJSON(context.Context, string, interface{}) error { return nil }

func (Noop) GetJSONWithTTL(context.Context, string, interface{}) (cache.LookupMeta, error) {
	return cache.LookupMeta{Hit: false, TTL: 0, HasTTL: false}, nil
}

func (Noop) TryAcquireRefreshLock(context.Context, string, string, time.Duration) (bool, error) {
	return false, nil
}

func (Noop) ReleaseRefreshLock(context.Context, string, string) error { return nil }

func (Noop) HGet(context.Context, string, string) (string, error) { return "", nil }

func (Noop) HSet(context.Context, string, string, string) error { return nil }

func (Noop) HDel(context.Context, string, string) error { return nil }

func (Noop) HGetAll(context.Context, string) (map[string]string, error) { return nil, nil }

func (Noop) HGetAllMulti(_ context.Context, keys []string) ([]map[string]string, error) {
	return make([]map[string]string, len(keys)), nil
}

func (Noop) MGetBytes(_ context.Context, keys ...string) ([][]byte, error) {
	return make([][]byte, len(keys)), nil
}

func (Noop) HIncrBy(context.Context, string, string, int64) (int64, error) { return 0, nil }

func (Noop) SetTimedJSONBatch(context.Context, []string, []interface{}, int64, ...cache.TimedOption) error {
	return nil
}

func (Noop) ZAddBatch(context.Context, string, []cache.ZBatchMember) error { return nil }

func (Noop) ZAdd(context.Context, string, float64, string) error { return nil }

func (Noop) ZRem(context.Context, string, ...string) error { return nil }

func (Noop) ZRevRangeByScore(context.Context, string, string, string, int64, int64) ([]string, error) {
	return nil, nil
}

func (Noop) XAdd(context.Context, string, int64, bool, map[string]interface{}) error { return nil }
