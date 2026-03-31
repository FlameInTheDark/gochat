package kvs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Cache struct {
	c                *redis.Client
	pendingRefreshes sync.Map
}

const releaseRefreshLockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`

const setTimedJSONNXScript = `
if redis.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2], "NX") then
	redis.call("SET", KEYS[2], ARGV[3], "EX", ARGV[2])
	return 1
end
return 0
`

const (
	refreshAheadPercent = 5
	refreshWindowFloor  = time.Second
)

type timedJSONMetadata struct {
	OriginalTTLSeconds int64 `json:"original_ttl_seconds"`
	Proactive          bool  `json:"proactive"`
}

type jsonLookupState struct {
	Meta     cache.LookupMeta
	Raw      []byte
	Timed    timedJSONMetadata
	HasTimed bool
}

type refreshReservationKey struct {
	ContextPtr uintptr
	Key        string
}

type refreshReservation struct {
	LockKey string
	Token   string
}

// Options configures the Redis connection pool. Zero values use defaults.
type Options struct {
	PoolSize     int           // default 256
	MinIdleConns int           // default 32
	DialTimeout  time.Duration // default 200ms
	ReadTimeout  time.Duration // default 200ms
	WriteTimeout time.Duration // default 200ms
}

func New(addr string, opts ...Options) (*Cache, error) {
	addr = normalizeAddr(addr)

	opt := Options{
		PoolSize:     256,
		MinIdleConns: 32,
		DialTimeout:  200 * time.Millisecond,
		ReadTimeout:  200 * time.Millisecond,
		WriteTimeout: 200 * time.Millisecond,
	}
	if len(opts) > 0 {
		o := opts[0]
		if o.PoolSize > 0 {
			opt.PoolSize = o.PoolSize
		}
		if o.MinIdleConns > 0 {
			opt.MinIdleConns = o.MinIdleConns
		}
		if o.DialTimeout > 0 {
			opt.DialTimeout = o.DialTimeout
		}
		if o.ReadTimeout > 0 {
			opt.ReadTimeout = o.ReadTimeout
		}
		if o.WriteTimeout > 0 {
			opt.WriteTimeout = o.WriteTimeout
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     opt.PoolSize,
		MinIdleConns: opt.MinIdleConns,
		DialTimeout:  opt.DialTimeout,
		ReadTimeout:  opt.ReadTimeout,
		WriteTimeout: opt.WriteTimeout,
	})

	ctx, end := observability.StartDependencySpan(context.Background(), "redis", "ping", addr)
	err := client.Ping(ctx).Err()
	end(err)
	return &Cache{c: client}, err
}

func normalizeAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "localhost:6379"
	}
	if strings.Contains(addr, "://") {
		return addr
	}
	if _, _, err := net.SplitHostPort(addr); err == nil {
		return addr
	}
	if strings.HasPrefix(addr, "[") && strings.HasSuffix(addr, "]") {
		return net.JoinHostPort(strings.TrimSuffix(strings.TrimPrefix(addr, "["), "]"), "6379")
	}
	if strings.Count(addr, ":") >= 2 || !strings.Contains(addr, ":") {
		return net.JoinHostPort(addr, "6379")
	}

	var addrErr *net.AddrError
	if _, _, err := net.SplitHostPort(addr); errors.As(err, &addrErr) && strings.Contains(addrErr.Err, "missing port in address") {
		return net.JoinHostPort(addr, "6379")
	}

	return addr
}

func (c *Cache) Client() *redis.Client {
	return c.c
}

func (c *Cache) Ping(ctx context.Context) error {
	if ctx == nil {
		return errors.New("nil context")
	}
	return c.c.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	return c.c.Close()
}

// Set string value
func (c *Cache) Set(ctx context.Context, key, val string) error {
	ctx, end := c.operation(ctx, "set", key)
	err := c.c.Set(ctx, key, val, 0).Err()
	end(err)
	return err
}

// Get string value
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	ctx, end := c.operation(ctx, "get", key)
	res := c.c.Get(ctx, key)
	err := res.Err()
	finishCacheOperation(ctx, end, err)
	return res.String(), err
}

func (c *Cache) GetWithTTL(ctx context.Context, key string) (string, cache.LookupMeta, error) {
	ctx, end := c.operation(ctx, "get_ttl", key)
	meta, val, err := c.getValueWithTTL(ctx, key)
	if errors.Is(err, redis.Nil) {
		finishCacheOperation(ctx, end, err)
		return "", cache.LookupMeta{}, nil
	}
	if err != nil {
		finishCacheOperation(ctx, end, err)
		return "", cache.LookupMeta{}, err
	}
	finishCacheOperation(ctx, end, nil)
	return string(val), meta, nil
}

// Delete key
func (c *Cache) Delete(ctx context.Context, key string) error {
	ctx, end := c.operation(ctx, "delete", key)
	err := c.c.Del(ctx, key, timedJSONMetadataKey(key), timedJSONRefreshLockKey(key)).Err()
	c.clearPendingRefreshReservation(ctx, key)
	end(err)
	return err
}

func (c *Cache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	ctx, end := c.operation(ctx, "get_bytes", key)
	res := c.c.Get(ctx, key)
	val, err := res.Bytes()
	finishCacheOperation(ctx, end, err)
	return val, err
}

// SetTimed set string value with expiration time in seconds
func (c *Cache) SetTimed(ctx context.Context, key, val string, ttl int64) error {
	ctx, end := c.operation(ctx, "set_timed", key)
	err := c.c.Set(ctx, key, val, time.Duration(ttl)*time.Second).Err()
	end(err)
	return err
}

// SetTimedInt64 set int64 value with expiration time in seconds
func (c *Cache) SetTimedInt64(ctx context.Context, key string, val int64, ttl int64) error {
	ctx, end := c.operation(ctx, "set_timed_int64", key)
	err := c.c.Set(ctx, key, val, time.Duration(ttl)*time.Second).Err()
	end(err)
	return err
}

// SetInt64 set int64 value
func (c *Cache) SetInt64(ctx context.Context, key string, val int64) error {
	ctx, end := c.operation(ctx, "set_int64", key)
	err := c.c.Set(ctx, key, val, 0).Err()
	end(err)
	return err
}

// SetTTL set expiration time for key
func (c *Cache) SetTTL(ctx context.Context, key string, ttl int64) error {
	ctx, end := c.operation(ctx, "expire", key)
	err := c.c.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
	end(err)
	return err
}

// Incr increment numerical value
func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	ctx, end := c.operation(ctx, "incr", key)
	res := c.c.Incr(ctx, key)
	err := res.Err()
	end(err)
	return res.Val(), err
}

// GetInt64 return int64 value of key
func (c *Cache) GetInt64(ctx context.Context, key string) (int64, error) {
	ctx, end := c.operation(ctx, "get_int64", key)
	res := c.c.Get(ctx, key)
	val, err := res.Int64()
	finishCacheOperation(ctx, end, err)
	return val, err
}

// SetJSON marshal set marshaled json of val
func (c *Cache) SetJSON(ctx context.Context, key string, val interface{}) error {
	msg, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ctx, end := c.operation(ctx, "set_json", key)
	pipe := c.c.Pipeline()
	pipe.Set(ctx, key, string(msg), 0)
	pipe.Del(ctx, timedJSONMetadataKey(key), timedJSONRefreshLockKey(key))
	_, err = pipe.Exec(ctx)
	c.clearPendingRefreshReservation(ctx, key)
	end(err)
	return err
}

func (c *Cache) SetTimedJSON(ctx context.Context, key string, val interface{}, ttl int64, opts ...cache.TimedOption) error {
	msg, err := json.Marshal(val)
	if err != nil {
		return err
	}
	meta, err := marshalTimedJSONMetadata(ttl, cache.ResolveTimedOptions(opts...))
	if err != nil {
		return err
	}
	ctx, end := c.operation(ctx, "set_timed_json", key)
	dur := time.Duration(ttl) * time.Second
	pipe := c.c.Pipeline()
	pipe.Set(ctx, key, string(msg), dur)
	pipe.Set(ctx, timedJSONMetadataKey(key), meta, dur)
	_, err = pipe.Exec(ctx)
	if err == nil {
		c.releasePendingRefreshReservation(ctx, key)
	}
	end(err)
	return err
}

func (c *Cache) SetTimedJSONNX(ctx context.Context, key string, val interface{}, ttl int64, opts ...cache.TimedOption) (bool, error) {
	msg, err := json.Marshal(val)
	if err != nil {
		return false, err
	}
	meta, err := marshalTimedJSONMetadata(ttl, cache.ResolveTimedOptions(opts...))
	if err != nil {
		return false, err
	}
	ctx, end := c.operation(ctx, "set_timed_json_nx", key)
	res := c.c.Eval(ctx, setTimedJSONNXScript, []string{key, timedJSONMetadataKey(key)}, string(msg), ttl, meta)
	if err := res.Err(); err != nil {
		end(err)
		return false, err
	}
	end(nil)
	if res.Val() == int64(1) {
		c.releasePendingRefreshReservation(ctx, key)
		return true, nil
	}
	return false, nil
}

// GetJSON unmarshal json into v
func (c *Cache) GetJSON(ctx context.Context, key string, v interface{}) error {
	ctx, end := c.operation(ctx, "get_json", key)
	state, err := c.getJSONLookupState(ctx, key)
	if err != nil {
		finishCacheOperation(ctx, end, err)
		return err
	}
	err = json.Unmarshal(state.Raw, v)
	if err == nil && c.shouldProactivelyRefresh(state) {
		if acquired, lockErr := c.acquirePendingRefresh(ctx, key, state.Meta.TTL); lockErr == nil && acquired {
			err = redis.Nil
		}
	}
	finishCacheOperation(ctx, end, err)
	return err
}

func (c *Cache) GetJSONWithTTL(ctx context.Context, key string, v interface{}) (cache.LookupMeta, error) {
	ctx, end := c.operation(ctx, "get_json_ttl", key)
	meta, raw, err := c.getValueWithTTL(ctx, key)
	if errors.Is(err, redis.Nil) {
		finishCacheOperation(ctx, end, err)
		return cache.LookupMeta{}, nil
	}
	if err != nil {
		finishCacheOperation(ctx, end, err)
		return cache.LookupMeta{}, err
	}
	err = json.Unmarshal(raw, v)
	finishCacheOperation(ctx, end, err)
	return meta, err
}

func (c *Cache) TryAcquireRefreshLock(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return false, nil
	}
	ctx, end := c.operation(ctx, "refresh_lock_acquire", key)
	res := c.c.SetArgs(ctx, key, token, redis.SetArgs{
		TTL:  ttl,
		Mode: "NX",
	})
	if err := res.Err(); err != nil && err != redis.Nil {
		end(err)
		return false, err
	}
	end(nil)
	return res.Val() == "OK", nil
}

func (c *Cache) ReleaseRefreshLock(ctx context.Context, key, token string) error {
	ctx, end := c.operation(ctx, "refresh_lock_release", key)
	err := c.c.Eval(ctx, releaseRefreshLockScript, []string{key}, token).Err()
	end(err)
	return err
}

func (c *Cache) HGet(ctx context.Context, key, field string) (string, error) {
	ctx, end := c.operation(ctx, "hget", key)
	h := c.c.HGet(ctx, key, field)
	if h.Err() != nil {
		finishCacheOperation(ctx, end, h.Err())
		return "", nil
	}
	end(nil)
	return h.Val(), nil
}

func (c *Cache) HSet(ctx context.Context, key, field, value string) error {
	ctx, end := c.operation(ctx, "hset", key)
	h := c.c.HSet(ctx, key, field, value)
	if h.Err() != nil {
		err := h.Err()
		end(err)
		return err
	}
	end(nil)
	return nil
}

func (c *Cache) HDel(ctx context.Context, key, field string) error {
	ctx, end := c.operation(ctx, "hdel", key)
	h := c.c.HDel(ctx, key, field)
	err := h.Err()
	end(err)
	return err
}

func (c *Cache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	ctx, end := c.operation(ctx, "hgetall", key)
	h := c.c.HGetAll(ctx, key)
	if h.Err() != nil {
		err := h.Err()
		end(err)
		return nil, err
	}
	end(nil)
	return h.Val(), nil
}

// HGetAllMulti pipelines len(keys) HGETALL commands in a single round-trip.
func (c *Cache) HGetAllMulti(ctx context.Context, keys []string) ([]map[string]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	ctx, end := c.operation(ctx, "hgetall_multi", keys[0])
	pipe := c.c.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.HGetAll(ctx, key)
	}
	_, err := pipe.Exec(ctx)
	end(err)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	results := make([]map[string]string, len(keys))
	for i, cmd := range cmds {
		if cmd.Err() == nil {
			results[i] = cmd.Val()
		}
	}
	return results, nil
}

// MGetBytes fetches multiple keys in a single MGET round-trip.
func (c *Cache) MGetBytes(ctx context.Context, keys ...string) ([][]byte, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	ctx, end := c.operation(ctx, "mget", keys[0])
	res := c.c.MGet(ctx, keys...)
	if err := res.Err(); err != nil {
		end(err)
		return nil, err
	}
	end(nil)
	vals := res.Val()
	out := make([][]byte, len(vals))
	for i, v := range vals {
		if v != nil {
			out[i] = []byte(v.(string))
		}
	}
	return out, nil
}

// SetTimedJSONBatch pipelines N SETEX commands in a single round-trip.
func (c *Cache) SetTimedJSONBatch(ctx context.Context, keys []string, vals []interface{}, ttl int64, opts ...cache.TimedOption) error {
	if len(keys) == 0 {
		return nil
	}
	ctx, end := c.operation(ctx, "setex_batch", keys[0])
	pipe := c.c.Pipeline()
	dur := time.Duration(ttl) * time.Second
	meta, err := marshalTimedJSONMetadata(ttl, cache.ResolveTimedOptions(opts...))
	if err != nil {
		end(err)
		return err
	}
	for i, key := range keys {
		b, err := json.Marshal(vals[i])
		if err != nil {
			end(err)
			return err
		}
		pipe.Set(ctx, key, string(b), dur)
		pipe.Set(ctx, timedJSONMetadataKey(key), meta, dur)
	}
	_, err = pipe.Exec(ctx)
	if err == nil {
		for _, key := range keys {
			c.releasePendingRefreshReservation(ctx, key)
		}
	}
	end(err)
	return err
}

// ZAddBatch adds multiple members to a sorted set in one ZADD command.
func (c *Cache) ZAddBatch(ctx context.Context, key string, members []cache.ZBatchMember) error {
	if len(members) == 0 {
		return nil
	}
	ctx, end := c.operation(ctx, "zadd_batch", key)
	zs := make([]redis.Z, len(members))
	for i, m := range members {
		zs[i] = redis.Z{Score: m.Score, Member: m.Member}
	}
	err := c.c.ZAdd(ctx, key, zs...).Err()
	end(err)
	return err
}

func (c *Cache) HIncrBy(ctx context.Context, key, field string, delta int64) (int64, error) {
	ctx, end := c.operation(ctx, "hincrby", key)
	h := c.c.HIncrBy(ctx, key, field, delta)
	if h.Err() != nil {
		err := h.Err()
		end(err)
		return 0, err
	}
	end(nil)
	return h.Val(), nil
}

func (c *Cache) ZAdd(ctx context.Context, key string, score float64, member string) error {
	ctx, end := c.operation(ctx, "zadd", key)
	err := c.c.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
	end(err)
	return err
}

func (c *Cache) ZRem(ctx context.Context, key string, members ...string) error {
	ctx, end := c.operation(ctx, "zrem", key)
	args := make([]interface{}, 0, len(members))
	for _, member := range members {
		args = append(args, member)
	}
	err := c.c.ZRem(ctx, key, args...).Err()
	end(err)
	return err
}

func (c *Cache) ZRevRangeByScore(ctx context.Context, key, max, min string, offset, count int64) ([]string, error) {
	ctx, end := c.operation(ctx, "zrevrangebyscore", key)
	res := c.c.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Max:    max,
		Min:    min,
		Offset: offset,
		Count:  count,
	})
	if res.Err() != nil {
		err := res.Err()
		finishCacheOperation(ctx, end, err)
		return nil, err
	}
	end(nil)
	return res.Val(), nil
}

func (c *Cache) XAdd(ctx context.Context, stream string, maxLen int64, approx bool, values map[string]interface{}) error {
	ctx, end := c.operation(ctx, "xadd", stream)
	h := c.c.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: maxLen,
		Approx: approx,
		Values: values,
	})
	if h.Err() != nil {
		err := h.Err()
		end(err)
		return err
	}
	end(nil)
	return nil
}

func (c *Cache) operation(ctx context.Context, operation, key string) (context.Context, func(error)) {
	return observability.StartDependencySpan(ctx, "redis", operation, cacheTarget(key))
}

func (c *Cache) getValueWithTTL(ctx context.Context, key string) (cache.LookupMeta, []byte, error) {
	pipe := c.c.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.PTTL(ctx, key)
	_, execErr := pipe.Exec(ctx)
	if execErr != nil && !errors.Is(execErr, redis.Nil) {
		return cache.LookupMeta{}, nil, execErr
	}
	if err := getCmd.Err(); err != nil {
		return cache.LookupMeta{}, nil, err
	}
	raw, err := getCmd.Bytes()
	if err != nil {
		return cache.LookupMeta{}, nil, err
	}
	meta := cache.LookupMeta{Hit: true}
	ttl := ttlCmd.Val()
	switch {
	case ttl > 0:
		meta.TTL = ttl
		meta.HasTTL = true
	case ttl == -1:
		meta.TTL = 0
		meta.HasTTL = false
	default:
		meta.TTL = 0
		meta.HasTTL = false
	}
	return meta, raw, nil
}

func (c *Cache) getJSONLookupState(ctx context.Context, key string) (jsonLookupState, error) {
	pipe := c.c.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.PTTL(ctx, key)
	metaCmd := pipe.Get(ctx, timedJSONMetadataKey(key))
	_, execErr := pipe.Exec(ctx)
	if execErr != nil && !errors.Is(execErr, redis.Nil) {
		return jsonLookupState{}, execErr
	}
	if err := getCmd.Err(); err != nil {
		return jsonLookupState{}, err
	}
	raw, err := getCmd.Bytes()
	if err != nil {
		return jsonLookupState{}, err
	}

	state := jsonLookupState{
		Meta: cache.LookupMeta{Hit: true},
		Raw:  raw,
	}
	ttl := ttlCmd.Val()
	switch {
	case ttl > 0:
		state.Meta.TTL = ttl
		state.Meta.HasTTL = true
	case ttl == -1:
		state.Meta.TTL = 0
		state.Meta.HasTTL = false
	default:
		state.Meta.TTL = 0
		state.Meta.HasTTL = false
	}

	if metaErr := metaCmd.Err(); metaErr == nil {
		if err := json.Unmarshal([]byte(metaCmd.Val()), &state.Timed); err == nil {
			state.HasTimed = true
		}
	}

	return state, nil
}

func (c *Cache) shouldProactivelyRefresh(state jsonLookupState) bool {
	if !state.Meta.Hit || !state.Meta.HasTTL || state.Meta.TTL <= 0 {
		return false
	}
	if !state.HasTimed || !state.Timed.Proactive || state.Timed.OriginalTTLSeconds <= 0 {
		return false
	}
	window := time.Duration(state.Timed.OriginalTTLSeconds) * time.Second * refreshAheadPercent / 100
	if window < refreshWindowFloor {
		window = refreshWindowFloor
	}
	return state.Meta.TTL <= window
}

func (c *Cache) acquirePendingRefresh(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return false, nil
	}
	token, err := randomRefreshToken()
	if err != nil {
		return false, err
	}
	lockKey := timedJSONRefreshLockKey(key)
	acquired, err := c.TryAcquireRefreshLock(ctx, lockKey, token, ttl)
	if err != nil || !acquired {
		return acquired, err
	}
	if reservationKey, ok := makeRefreshReservationKey(ctx, key); ok {
		c.pendingRefreshes.Store(reservationKey, refreshReservation{LockKey: lockKey, Token: token})
		time.AfterFunc(ttl+time.Second, func() {
			c.pendingRefreshes.Delete(reservationKey)
		})
	}
	return true, nil
}

func (c *Cache) releasePendingRefreshReservation(ctx context.Context, key string) {
	reservationKey, ok := makeRefreshReservationKey(ctx, key)
	if !ok {
		return
	}
	value, found := c.pendingRefreshes.LoadAndDelete(reservationKey)
	if !found {
		return
	}
	reservation, ok := value.(refreshReservation)
	if !ok {
		return
	}
	_ = c.ReleaseRefreshLock(ctx, reservation.LockKey, reservation.Token)
}

func (c *Cache) clearPendingRefreshReservation(ctx context.Context, key string) {
	reservationKey, ok := makeRefreshReservationKey(ctx, key)
	if ok {
		c.pendingRefreshes.Delete(reservationKey)
	}
}

func timedJSONMetadataKey(key string) string {
	return "cache:timed-json:meta:" + key
}

func timedJSONRefreshLockKey(key string) string {
	return "cache:timed-json:refresh-lock:" + key
}

func marshalTimedJSONMetadata(ttl int64, opts cache.TimedOptions) (string, error) {
	meta := timedJSONMetadata{
		OriginalTTLSeconds: ttl,
		Proactive:          opts.Proactive,
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func makeRefreshReservationKey(ctx context.Context, key string) (refreshReservationKey, bool) {
	if ctx == nil || key == "" {
		return refreshReservationKey{}, false
	}
	value := reflect.ValueOf(ctx)
	if !value.IsValid() {
		return refreshReservationKey{}, false
	}
	switch value.Kind() {
	case reflect.Pointer, reflect.UnsafePointer:
		if value.IsNil() {
			return refreshReservationKey{}, false
		}
		return refreshReservationKey{ContextPtr: value.Pointer(), Key: key}, true
	default:
		return refreshReservationKey{}, false
	}
}

func randomRefreshToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func cacheTarget(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "unknown"
	}
	if idx := strings.Index(key, ":"); idx > 0 {
		return key[:idx]
	}
	return key
}

func finishCacheOperation(ctx context.Context, end func(error), err error) {
	if end == nil {
		return
	}
	if errors.Is(err, redis.Nil) {
		observability.SetDependencyResult(ctx, "miss")
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(
			attribute.Bool("cache.miss", true),
			attribute.String("dependency.result", "miss"),
		)
		span.AddEvent("cache.miss")
		end(nil)
		return
	}
	end(err)
}
