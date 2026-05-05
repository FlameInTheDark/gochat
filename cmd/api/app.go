package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/FlameInTheDark/gochat/cmd/api/config"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/emoji"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/guild"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/message"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/user"
	"github.com/FlameInTheDark/gochat/cmd/api/endpoints/voice"
	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	reactionrepo "github.com/FlameInTheDark/gochat/internal/database/entities/reaction"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	authenticationrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/authentication"
	channelrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/channel"
	"github.com/FlameInTheDark/gochat/internal/embedmq"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/indexmq"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/nats"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/presence"
	reactionutil "github.com/FlameInTheDark/gochat/internal/reaction"
	"github.com/FlameInTheDark/gochat/internal/s3"
	"github.com/FlameInTheDark/gochat/internal/searchmq"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/shutter"
	"github.com/FlameInTheDark/gochat/internal/threadcount"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
	"github.com/gofiber/fiber/v2"
	natsio "github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

type App struct {
	server *server.Server
	db     *db.CQLCon
	logger *slog.Logger

	addr string
}

const threadMessageCountFlushInterval = 5 * time.Second
const reactionFlushReadBlock = 2 * time.Second
const startupProbeTimeout = 3 * time.Second
const startupProbeMaxDelay = 5 * time.Second

func retryStartupProbe(ctx context.Context, logger *slog.Logger, dependency string, maxRetries int, probe func(context.Context) error) error {
	if ctx == nil {
		return fmt.Errorf("probe %s: nil context", dependency)
	}

	delay := time.Second
	for attempt := 1; ; attempt++ {
		probeCtx, cancel := context.WithTimeout(ctx, startupProbeTimeout)
		err := probe(probeCtx)
		cancel()
		if err == nil {
			return nil
		}

		if logger != nil {
			logger.Warn(
				"startup dependency probe failed; retrying",
				slog.String("dependency", dependency),
				slog.Int("attempt", attempt),
				slog.String("error", err.Error()),
			)
		}
		if maxRetries > 0 && attempt >= maxRetries {
			return fmt.Errorf("probe %s after %d attempts: %w", dependency, attempt, err)
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("probe %s canceled: %w", dependency, ctx.Err())
		case <-time.After(delay):
		}

		delay *= 2
		if delay > startupProbeMaxDelay {
			delay = startupProbeMaxDelay
		}
	}
}

func verifyAPIInfrastructure(logger *slog.Logger, retries int, pg *pgdb.DB, database *db.CQLCon, cache *kvs.Cache, natsQueue *nats.NatsQueue, indexMQ *indexmq.IndexMQ, embedMQ *embedmq.Queue) error {
	baseCtx := context.Background()
	if err := retryStartupProbe(baseCtx, logger, "postgres", retries, func(ctx context.Context) error {
		return pg.Conn().PingContext(ctx)
	}); err != nil {
		return err
	}
	if err := retryStartupProbe(baseCtx, logger, "scylla", retries, database.Ping); err != nil {
		return err
	}
	if err := retryStartupProbe(baseCtx, logger, "redis", retries, cache.Ping); err != nil {
		return err
	}
	if err := retryStartupProbe(baseCtx, logger, "nats", retries, natsQueue.Ping); err != nil {
		return err
	}
	if err := retryStartupProbe(baseCtx, logger, "indexer_nats", retries, indexMQ.Ping); err != nil {
		return err
	}
	return retryStartupProbe(baseCtx, logger, "embedder_nats", retries, embedMQ.Ping)
}

func startThreadMessageCountFlusher(ctx context.Context, cache *kvs.Cache, repo channelrepo.Channel, logger *slog.Logger) {
	if cache == nil || repo == nil {
		return
	}
	ticker := time.NewTicker(threadMessageCountFlushInterval)
	defer ticker.Stop()

	flush := func() {
		var cursor uint64
		for {
			keys, next, err := cache.Client().Scan(ctx, cursor, threadcount.DeltaKeyPattern, 128).Result()
			if err != nil {
				if logger != nil {
					logger.Error("failed to scan thread message count deltas", slog.String("error", err.Error()))
				}
				return
			}
			for _, key := range keys {
				delta, err := cache.Client().GetSet(ctx, key, "0").Int64()
				if errors.Is(err, redis.Nil) || delta <= 0 {
					continue
				}
				if err != nil {
					if logger != nil {
						logger.Error("failed to swap thread message count delta", slog.String("key", key), slog.String("error", err.Error()))
					}
					continue
				}

				threadID, err := threadcount.ParseDeltaKey(key)
				if err != nil {
					if logger != nil {
						logger.Error("failed to parse thread message count key", slog.String("key", key), slog.String("error", err.Error()))
					}
					_, _ = cache.Client().IncrBy(ctx, key, delta).Result()
					continue
				}
				if err := repo.AdjustMessageCount(ctx, threadID, delta); err != nil {
					if logger != nil {
						logger.Error("failed to flush thread message count delta",
							slog.Int64("thread_id", threadID),
							slog.Int64("delta", delta),
							slog.String("error", err.Error()))
					}
					_, _ = cache.Client().IncrBy(ctx, key, delta).Result()
					continue
				}

				_ = cache.Client().Expire(ctx, key, time.Duration(threadcount.DeltaTTLSeconds)*time.Second).Err()
			}
			cursor = next
			if cursor == 0 {
				return
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case <-ticker.C:
			flush()
		}
	}
}

func startReactionFlusher(ctx context.Context, cache *kvs.Cache, repo reactionrepo.Reaction, logger *slog.Logger) {
	if cache == nil || repo == nil {
		return
	}

	lastID := "0-0"
	for {
		if ctx.Err() != nil {
			return
		}

		streams, err := cache.Client().XRead(ctx, &redis.XReadArgs{
			Streams: []string{reactionutil.FlushStreamKey, lastID},
			Count:   128,
			Block:   reactionFlushReadBlock,
		}).Result()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if errors.Is(err, redis.Nil) {
				continue
			}
			if logger != nil {
				logger.Error("failed to read reaction flush stream", slog.String("error", err.Error()))
			}
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				lastID = message.ID
				raw, ok := message.Values["event"]
				if !ok {
					continue
				}
				payload, ok := raw.(string)
				if !ok || payload == "" {
					continue
				}

				var event reactionutil.FlushEvent
				if err := json.Unmarshal([]byte(payload), &event); err != nil {
					if logger != nil {
						logger.Error("failed to decode reaction flush event", slog.String("error", err.Error()))
					}
					continue
				}

				switch event.Op {
				case "add":
					if err := repo.UpsertReaction(ctx, event.Reaction); err != nil && logger != nil {
						logger.Error("failed to flush reaction add", slog.String("error", err.Error()))
					}
				case "remove":
					if err := repo.DeleteReaction(ctx, event.Reaction); err != nil && logger != nil {
						logger.Error("failed to flush reaction remove", slog.String("error", err.Error()))
					}
				default:
					continue
				}

				summary := model.ReactionSummary{
					MessageId: event.Reaction.MessageId,
					BucketKey: event.Reaction.BucketKey,
					Custom:    event.Reaction.Custom,
					EmojiId:   event.Reaction.EmojiId,
					EmojiName: event.Reaction.EmojiName,
					Count:     event.Count,
				}
				if err := repo.SetSummary(ctx, summary); err != nil && logger != nil {
					logger.Error("failed to flush reaction summary", slog.String("error", err.Error()))
				}
			}
		}
	}
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) (*App, error) {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		return nil, err
	}

	// In Prefork mode Fiber spawns child processes via os.StartProcess; each child
	// re-executes main(). Only children should open DB/cache connections.
	if cfg.Prefork && !fiber.IsChild() {
		logger.Info("Prefork master process — skipping DB connections")
		s := server.NewServer(true)
		shut.Up(s)
		return &App{server: s, logger: logger, addr: cfg.ServerAddress}, nil
	}

	logger.Info("Connecting to ScyllaDB")
	database, err := db.NewCQLCon(cfg.ClusterKeyspace, db.NewDBLogger(logger), cfg.Cluster...)
	if err != nil {
		return nil, err
	}
	shut.Up(database)

	logger.Info("Connecting to PostgreSQL")
	pg := pgdb.NewDB(logger)
	err = pg.Connect(cfg.PGDSN, pgdb.ConnectOptions{
		MaxRetries:   cfg.PGRetries,
		QueryLog:     cfg.PGQueryLog,
		MaxOpenConns: cfg.PGMaxOpenConns,
		MaxIdleConns: cfg.PGMaxIdleConns,
	})
	if err != nil {
		return nil, err
	}
	shut.Up(pg)
	postgresProbeCtx, cancelPostgresProbe := context.WithCancel(context.Background())
	shut.UpFunc(cancelPostgresProbe)
	go pg.StartProbeLoop(postgresProbeCtx, 30*time.Second)

	var storage *s3.Client
	if cfg.S3Endpoint != "" {
		logger.Info("Connecting to S3")
		storage, err = s3.NewClient(cfg.S3Endpoint, cfg.S3AccessKeyID, cfg.S3SecretAccessKey, cfg.S3Region, cfg.S3Bucket, cfg.S3UseSSL)
		if err != nil {
			return nil, err
		}
	}

	logger.Info("Connecting to NATS")
	var qt mq.SendTransporter
	nt, err := nats.New(cfg.NATSConnString)
	if err != nil {
		return nil, err
	}
	shut.Up(nt)
	qt = nt

	logger.Info("Connecting to Indexer NATS")
	imq, err := indexmq.NewIndexMQ(cfg.IndexerNATSConnString)
	if err != nil {
		return nil, err
	}
	shut.Up(imq)

	logger.Info("Connecting to Embedder NATS")
	emq, err := embedmq.New(cfg.NATSConnString)
	if err != nil {
		return nil, err
	}
	shut.Up(emq)

	logger.Info("Connecting to KeyDB")
	cache, err := kvs.New(cfg.KeyDB, kvs.Options{
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
	})
	if err != nil {
		return nil, err
	}
	shut.Up(cache)

	if err := verifyAPIInfrastructure(logger, cfg.PGRetries, pg, database, cache, nt, imq, emq); err != nil {
		return nil, err
	}
	pstore := presence.NewStore(cache)

	threadCountCtx, cancelThreadCount := context.WithCancel(context.Background())
	shut.UpFunc(cancelThreadCount)
	go startThreadMessageCountFlusher(threadCountCtx, cache, channelrepo.New(pg.Conn()), logger)

	reactionFlushCtx, cancelReactionFlush := context.WithCancel(context.Background())
	shut.UpFunc(cancelReactionFlush)
	go startReactionFlusher(reactionFlushCtx, cache, reactionrepo.New(database), logger)

	logger.Info("Connecting to NATS for SFU occupancy updates")
	if cfg.NATSConnString != "" {
		if occNc, err := natsio.Connect(cfg.NATSConnString, natsio.Compression(true)); err == nil {
			shut.UpFunc(func() { _ = occNc.Drain() })
			_, _ = occNc.Subscribe("voice.occ", func(m *natsio.Msg) {
				type occ struct {
					Channel int64 `json:"channel"`
					Delta   int   `json:"delta"`
				}
				var o occ
				if err := json.Unmarshal(m.Data, &o); err != nil || o.Channel <= 0 || (o.Delta != 1 && o.Delta != -1) {
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				key := fmt.Sprintf("voice:occ:%d", o.Channel)
				if o.Delta == 1 {
					_, _ = cache.Incr(ctx, key)
				} else {
					if n, err := cache.GetInt64(ctx, key); err == nil {
						if n > 0 {
							n--
						}
						_ = cache.SetInt64(ctx, key, n)
						if n <= 0 {
							_ = cache.Delete(ctx, fmt.Sprintf("voice:route:%d", o.Channel))
						}
					}
				}
			})
		}
	}

	logger.Info("Connecting to Etcd")
	disco, err := discovery.NewManager(cfg.EtcdEndpoints, cfg.EtcdPrefix, cfg.EtcdUsername, cfg.EtcdPassword)
	if err != nil {
		return nil, err
	}
	streamDisco, err := discovery.NewManager(cfg.EtcdEndpoints, cfg.StreamEtcdPrefix, cfg.EtcdUsername, cfg.EtcdPassword)
	if err != nil {
		return nil, err
	}

	idgen.New(0)

	logger.Info("Registering HTTP server")
	searchQueue, err := searchmq.New(cfg.IndexerNATSConnString)
	if err != nil {
		return nil, err
	}
	shut.Up(searchQueue)

	s := server.NewServer(cfg.Prefork)
	shut.Up(s)

	s.WithCache(cache)
	if cfg.Swagger {
		s.WithSwagger("api")
	}
	if cfg.ApiLog {
		s.WithLoggerLevel(logger, observability.ParseLogLevel(cfg.LogLevel))
	}
	s.WithCORS()
	s.WithCompression()
	s.WithMetrics("gochat-api")
	s.WithIdempotency(cache.Client(), cfg.IdempotencyStorageLifetime)
	s.AuthMiddleware(cfg.AuthSecret)
	s.Use(helper.RequireSessionVersion(helper.NewSessionVersionChecker(authenticationrepo.New(pg.Conn()), cache)))
	s.RateLimitPipedMiddleware(cfg.RateLimitRequests, cfg.RateLimitTime)
	s.Use(helper.RequireTokenType("access", "api"))
	s.Use(func(c *fiber.Ctx) error {
		c.Locals("base_url", cfg.BaseURL)
		return c.Next()
	})

	contentHosts, err := buildContentHosts(cfg.ContentHosts)
	if err != nil {
		return nil, err
	}

	s.Register(
		"/api/v1",
		emoji.New(database, pg, cache, logger),
		user.New(database, pg, qt, searchQueue, cache, cfg.AttachmentTTLMinutes*60, contentHosts, logger),
		message.New(database, pg, qt, imq, emq, cfg.UploadLimit, cfg.AttachmentTTLMinutes*60, cache, logger),
		guild.New(database, pg, qt, imq, searchQueue, cache, storage, cfg.AttachmentTTLMinutes*60, cfg.AuthSecret, pstore, nt.Conn(), cfg.VoiceDefaultRegion, disco, streamDisco, extractRegionIDs(cfg.VoiceRegions), logger),
		voice.New(convertRegions(cfg.VoiceRegions), logger),
	)

	return &App{server: s, db: database, logger: logger, addr: cfg.ServerAddress}, nil
}

func extractRegionIDs(v []config.VoiceRegion) []string {
	if len(v) == 0 {
		return nil
	}
	out := make([]string, 0, len(v))
	for _, r := range v {
		if r.ID != "" {
			out = append(out, r.ID)
		}
	}
	return out
}

func buildContentHosts(rawHosts []string) ([]string, error) {
	seen := make(map[string]struct{}, len(rawHosts))
	hosts := make([]string, 0, len(rawHosts))

	add := func(raw string) error {
		normalized, err := normalizeContentHost(raw)
		if err != nil {
			return err
		}
		if normalized == "" {
			return nil
		}
		if _, ok := seen[normalized]; ok {
			return nil
		}
		seen[normalized] = struct{}{}
		hosts = append(hosts, normalized)
		return nil
	}

	for _, raw := range rawHosts {
		if err := add(raw); err != nil {
			return nil, fmt.Errorf("normalize content host %q: %w", raw, err)
		}
	}

	return hosts, nil
}

func normalizeContentHost(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("content host must include scheme and host")
	}

	return u.Scheme + "://" + u.Host, nil
}

func convertRegions(v []config.VoiceRegion) []voice.Region {
	if len(v) == 0 {
		return nil
	}
	out := make([]voice.Region, 0, len(v))
	for _, r := range v {
		if r.ID == "" {
			continue
		}
		out = append(out, voice.Region{ID: r.ID, Name: r.Name})
	}
	return out
}

func (app *App) Start() {
	app.logger.Info("Starting")
	go func() {
		err := app.server.Start(app.addr)
		if err != nil {
			app.logger.Error("Error starting server", "error", err)
			os.Exit(1)
		}
	}()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
}
