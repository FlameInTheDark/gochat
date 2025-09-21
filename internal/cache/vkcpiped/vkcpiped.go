package vkcpiped

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// VKStorage implements fiber.Storage
type VKStorage struct {
	write *redis.Client
	read  *redis.Client

	pipeSize      int
	flushInterval time.Duration
	workers       int

	enqueueTimeout     time.Duration
	waitAckTimeout     time.Duration
	execTimeout        time.Duration
	directWriteTimeout time.Duration
	queueCapacity      int

	prefix string

	resetUseFlushDB bool
	resetScanCount  int

	jobs       chan setJob
	wg         sync.WaitGroup
	closed     atomic.Bool
	closeWrite bool
	closeRead  bool
}

type VKOptions struct {
	WriteClient *redis.Client
	ReadClient  *redis.Client
	Addr        string
	Password    string
	DB          int
	PoolSize    int

	PipeSize      int
	FlushInterval time.Duration
	Workers       int
	QueueCapacity int

	EnqueueTimeout     time.Duration
	WaitAckTimeout     time.Duration
	ExecTimeout        time.Duration
	DirectWriteTimeout time.Duration

	Prefix string

	ResetUseFlushDB bool
	ResetScanCount  int
}

type setJob struct {
	ctx context.Context
	key string
	val []byte
	exp time.Duration
	res chan error
}

func NewVKStorage(opts VKOptions) (*VKStorage, error) {
	wc := opts.WriteClient
	rc := opts.ReadClient

	newIfNil := func() *redis.Client {
		addr := opts.Addr
		if addr == "" {
			addr = "localhost:6379"
		}
		pool := opts.PoolSize
		if pool == 0 {
			pool = 2048
		}
		return redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: opts.Password,
			DB:       opts.DB,
			PoolSize: pool,
		})
	}

	st := &VKStorage{
		pipeSize:           max(1, nz(opts.PipeSize, 16)),
		flushInterval:      nzd(opts.FlushInterval, time.Millisecond),
		workers:            max(1, nz(opts.Workers, 256)),
		queueCapacity:      nz(opts.QueueCapacity, 100_000),
		enqueueTimeout:     nzd(opts.EnqueueTimeout, 2*time.Millisecond),
		waitAckTimeout:     opts.WaitAckTimeout,
		execTimeout:        nzd(opts.ExecTimeout, 250*time.Millisecond),
		directWriteTimeout: nzd(opts.DirectWriteTimeout, 50*time.Millisecond),
		prefix:             opts.Prefix,
		resetScanCount:     nz(opts.ResetScanCount, 1000),
	}
	st.jobs = make(chan setJob, st.queueCapacity)

	if wc == nil {
		wc = newIfNil()
		st.closeWrite = true
	}
	if rc == nil {
		rc = newIfNil()
		st.closeRead = true
	}
	st.write, st.read = wc, rc

	if st.prefix == "" {
		st.resetUseFlushDB = true
	}
	if opts.ResetUseFlushDB {
		st.resetUseFlushDB = true
	}

	if err := st.write.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis write ping: %w", err)
	}
	if st.read != st.write {
		if err := st.read.Ping(context.Background()).Err(); err != nil {
			return nil, fmt.Errorf("redis read ping: %w", err)
		}
	}

	for i := 0; i < st.workers; i++ {
		st.wg.Add(1)
		go st.worker()
	}

	return st, nil
}

func NewVKStorageFromClient(c *redis.Client, tweak func(*VKOptions)) (*VKStorage, error) {
	opts := VKOptions{
		WriteClient: c,
		ReadClient:  c,
	}
	if tweak != nil {
		tweak(&opts)
	}
	return NewVKStorage(opts)
}

func (s *VKStorage) GetWithContext(ctx context.Context, key string) ([]byte, error) {
	val, err := s.read.Get(ctx, s.k(key)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (s *VKStorage) SetWithContext(ctx context.Context, key string, val []byte, exp time.Duration) error {
	if s.closed.Load() {
		return errors.New("storage closed")
	}
	res := make(chan error, 1)
	j := setJob{ctx: ctx, key: s.k(key), val: val, exp: exp, res: res}

	select {
	case s.jobs <- j:
		if s.waitAckTimeout <= 0 {
			return nil
		}
		select {
		case err := <-res:
			return err
		case <-time.After(s.waitAckTimeout):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}

	case <-time.After(s.enqueueTimeout):
		wctx, cancel := context.WithTimeout(ctx, s.directWriteTimeout)
		defer cancel()
		if exp > 0 {
			return s.write.SetArgs(wctx, j.key, j.val, redis.SetArgs{TTL: exp}).Err()
		}
		return s.write.Set(wctx, j.key, j.val, 0).Err()

	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *VKStorage) DeleteWithContext(ctx context.Context, key string) error {
	return s.write.Del(ctx, s.k(key)).Err()
}

func (s *VKStorage) ResetWithContext(ctx context.Context) error {
	if s.prefix == "" && s.resetUseFlushDB {
		return s.write.FlushDB(ctx).Err()
	}

	var cursor uint64
	pat := s.prefix + "*"
	count := int64(max(1, s.resetScanCount))
	for {
		keys, next, err := s.write.Scan(ctx, cursor, pat, count).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := s.write.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}

func (s *VKStorage) Get(key string) ([]byte, error) {
	return s.GetWithContext(context.Background(), key)
}

func (s *VKStorage) Set(key string, val []byte, exp time.Duration) error {
	return s.SetWithContext(context.Background(), key, val, exp)
}

func (s *VKStorage) Delete(key string) error {
	return s.DeleteWithContext(context.Background(), key)
}

func (s *VKStorage) Reset() error {
	return s.ResetWithContext(context.Background())
}

func (s *VKStorage) Close() error {
	if s.closed.Swap(true) {
		return nil
	}
	close(s.jobs)
	s.wg.Wait()
	var err error
	if s.closeWrite {
		if e := s.write.Close(); e != nil {
			err = e
		}
	}
	if s.closeRead && s.read != s.write {
		if e := s.read.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}

func (s *VKStorage) worker() {
	defer s.wg.Done()

	type item struct {
		key string
		val []byte
		exp time.Duration
		ctx context.Context
		res chan error
	}

	buf := make([]item, 0, s.pipeSize)
	timer := time.NewTimer(s.flushInterval)
	defer timer.Stop()

	flush := func() {
		if len(buf) == 0 {
			return
		}
		pipe := s.write.Pipeline()
		cmds := make([]*redis.StatusCmd, 0, len(buf))
		for _, it := range buf {
			if it.exp > 0 {
				cmds = append(cmds, pipe.SetArgs(it.ctx, it.key, it.val, redis.SetArgs{TTL: it.exp}))
			} else {
				cmds = append(cmds, pipe.Set(it.ctx, it.key, it.val, 0))
			}
		}

		execCtx, cancel := context.WithTimeout(context.Background(), s.execTimeout)
		_, execErr := pipe.Exec(execCtx)
		cancel()

		for i, c := range cmds {
			err := c.Err()
			if execErr != nil && err == nil {
				err = execErr
			}
			select {
			case buf[i].res <- err:
			default:
			}
		}
		buf = buf[:0]

		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(s.flushInterval)
	}

	for {
		select {
		case j, ok := <-s.jobs:
			if !ok {
				flush()
				return
			}
			if s.closed.Load() {
				select {
				case j.res <- errors.New("storage closed"):
				default:
				}
				continue
			}
			buf = append(buf, item{key: j.key, val: j.val, exp: j.exp, ctx: j.ctx, res: j.res})
			if len(buf) >= s.pipeSize {
				flush()
			}
		case <-timer.C:
			flush()
		}
	}
}

func (s *VKStorage) k(key string) string {
	if s.prefix == "" {
		return key
	}

	if strings.HasPrefix(key, s.prefix) {
		return key
	}
	return s.prefix + key
}

func nz(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
func nzd(v, def time.Duration) time.Duration {
	if v == 0 {
		return def
	}
	return v
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
