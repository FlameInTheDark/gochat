package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FlameInTheDark/gochat/cmd/search/config"
	searchendpoints "github.com/FlameInTheDark/gochat/cmd/search/endpoints/search"
	"github.com/FlameInTheDark/gochat/internal/botsearch"
	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	authenticationrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/authentication"
	botrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/bot"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guilddiscovery"
	userrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/guildsearch"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/msgsearch"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/searchmq"
	"github.com/FlameInTheDark/gochat/internal/server"
	"github.com/FlameInTheDark/gochat/internal/shutter"
	"github.com/nats-io/nats.go"
)

type App struct {
	server *server.Server
	logger *slog.Logger
	addr   string
	nats   *nats.Conn
	subs   []*nats.Subscription

	guildRepo guild.Guild
	botRepo   botrepo.Bot
	userRepo  userrepo.User
	discovery guilddiscovery.GuildDiscovery
	guildOS   *guildsearch.Search
	botOS     *botsearch.Search
}

const startupProbeTimeout = 3 * time.Second
const startupProbeMaxDelay = 5 * time.Second

func NewApp(shut *shutter.Shut, logger *slog.Logger) (*App, error) {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		return nil, err
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
		DriverName:   cfg.PGDriver,
		MaxRetries:   cfg.PGRetries,
		QueryLog:     cfg.PGQueryLog,
		MaxOpenConns: cfg.PGMaxOpenConns,
		MaxIdleConns: cfg.PGMaxIdleConns,
	})
	if err != nil {
		return nil, err
	}
	shut.Up(pg)

	logger.Info("Connecting to KeyDB")
	cache, err := kvs.New(cfg.KeyDB, kvs.Options{
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
	})
	if err != nil {
		return nil, err
	}
	shut.Up(cache)

	logger.Info("Connecting to OpenSearch")
	messageSearch, err := msgsearch.NewSearch(cfg.OSAddresses, cfg.OSInsecureSkipVerify, cfg.OSUsername, cfg.OSPassword)
	if err != nil {
		return nil, err
	}
	guildSearch, err := guildsearch.NewSearch(cfg.OSAddresses, cfg.OSInsecureSkipVerify, cfg.OSUsername, cfg.OSPassword)
	if err != nil {
		return nil, err
	}
	botSearch, err := botsearch.NewSearch(cfg.OSAddresses, cfg.OSInsecureSkipVerify, cfg.OSUsername, cfg.OSPassword)
	if err != nil {
		return nil, err
	}

	logger.Info("Connecting to Search NATS")
	nc, err := nats.Connect(cfg.NATSConnString, nats.Compression(true))
	if err != nil {
		return nil, err
	}
	shut.UpFunc(func() { nc.Close() })

	if err := verifyInfrastructure(logger, cfg.PGRetries, pg, database, cache, nc); err != nil {
		return nil, err
	}

	s := server.NewServer()
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
	s.WithMetrics("gochat-search")
	s.AuthMiddleware(cfg.AuthSecret)
	s.Use(helper.RequireSessionVersion(helper.NewSessionVersionChecker(authenticationrepo.New(pg.Conn()), cache)))
	s.Use(helper.RequireTokenType("access", "api"))
	s.Register("/api/v1", searchendpoints.New(database, pg, messageSearch, guildSearch, botSearch, logger))

	app := &App{
		server:    s,
		logger:    logger,
		addr:      cfg.ServerAddress,
		nats:      nc,
		guildRepo: guild.New(pg.Conn()),
		botRepo:   botrepo.New(pg.Conn()),
		userRepo:  userrepo.New(pg.Conn()),
		discovery: guilddiscovery.New(pg.Conn()),
		guildOS:   guildSearch,
		botOS:     botSearch,
	}
	if err := app.subscribeGuildIndexing(); err != nil {
		return nil, err
	}
	if err := app.subscribeBotIndexing(); err != nil {
		return nil, err
	}
	go app.backfillBotIndex()
	shut.Up(app)
	return app, nil
}

func (a *App) Start() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalCh
		_ = a.Close()
	}()
	if err := a.server.Start(a.addr); err != nil {
		a.logger.Error("search service stopped", slog.String("error", err.Error()))
	}
}

func (a *App) Close() error {
	for _, sub := range a.subs {
		if err := sub.Unsubscribe(); err != nil {
			a.logger.Error("unable to unsubscribe", slog.String("error", err.Error()))
		}
	}
	if a.nats != nil {
		a.nats.Close()
	}
	return a.server.Close()
}

func (a *App) subscribeGuildIndexing() error {
	upsertSub, err := a.nats.Subscribe(searchmq.GuildUpsertSubject, func(msg *nats.Msg) {
		ctx := observability.ExtractNATSContext(context.Background(), msg)
		ctx, finish := observability.StartNATSConsumeSpan(ctx, msg.Subject)
		var processErr error
		defer func() { finish(processErr) }()

		var payload dto.GuildIndexMessage
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			processErr = err
			a.logger.ErrorContext(ctx, "failed to decode guild index message", slog.String("error", err.Error()))
			return
		}
		if err := a.indexGuild(ctx, payload.GuildId); err != nil {
			processErr = err
			a.logger.ErrorContext(ctx, "failed to index guild", slog.Int64("guild_id", payload.GuildId), slog.String("error", err.Error()))
		}
	})
	if err != nil {
		return err
	}
	a.subs = append(a.subs, upsertSub)

	deleteSub, err := a.nats.Subscribe(searchmq.GuildDeleteSubject, func(msg *nats.Msg) {
		ctx := observability.ExtractNATSContext(context.Background(), msg)
		ctx, finish := observability.StartNATSConsumeSpan(ctx, msg.Subject)
		var processErr error
		defer func() { finish(processErr) }()

		var payload dto.GuildIndexDeleteMessage
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			processErr = err
			a.logger.ErrorContext(ctx, "failed to decode guild delete index message", slog.String("error", err.Error()))
			return
		}
		if err := a.guildOS.DeleteGuild(ctx, payload.GuildId); err != nil {
			processErr = err
			a.logger.ErrorContext(ctx, "failed to delete indexed guild", slog.Int64("guild_id", payload.GuildId), slog.String("error", err.Error()))
		}
	})
	if err != nil {
		return err
	}
	a.subs = append(a.subs, deleteSub)
	return nil
}

func (a *App) subscribeBotIndexing() error {
	upsertSub, err := a.nats.Subscribe(searchmq.BotUpsertSubject, func(msg *nats.Msg) {
		ctx := observability.ExtractNATSContext(context.Background(), msg)
		ctx, finish := observability.StartNATSConsumeSpan(ctx, msg.Subject)
		var processErr error
		defer func() { finish(processErr) }()

		var payload dto.BotIndexMessage
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			processErr = err
			a.logger.ErrorContext(ctx, "failed to decode bot index message", slog.String("error", err.Error()))
			return
		}
		if err := a.indexBot(ctx, payload.BotUserId); err != nil {
			processErr = err
			a.logger.ErrorContext(ctx, "failed to index bot", slog.Int64("bot_user_id", payload.BotUserId), slog.String("error", err.Error()))
		}
	})
	if err != nil {
		return err
	}
	a.subs = append(a.subs, upsertSub)

	deleteSub, err := a.nats.Subscribe(searchmq.BotDeleteSubject, func(msg *nats.Msg) {
		ctx := observability.ExtractNATSContext(context.Background(), msg)
		ctx, finish := observability.StartNATSConsumeSpan(ctx, msg.Subject)
		var processErr error
		defer func() { finish(processErr) }()

		var payload dto.BotIndexDeleteMessage
		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			processErr = err
			a.logger.ErrorContext(ctx, "failed to decode bot delete index message", slog.String("error", err.Error()))
			return
		}
		if err := a.botOS.DeleteBot(ctx, payload.BotUserId); err != nil {
			processErr = err
			a.logger.ErrorContext(ctx, "failed to delete indexed bot", slog.Int64("bot_user_id", payload.BotUserId), slog.String("error", err.Error()))
		}
	})
	if err != nil {
		return err
	}
	a.subs = append(a.subs, deleteSub)
	return nil
}

func (a *App) indexGuild(ctx context.Context, guildID int64) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	g, err := a.guildRepo.GetGuildById(ctx, guildID)
	if err != nil {
		return err
	}
	tagsByGuild, err := a.discovery.GetTagsByGuilds(ctx, []int64{guildID})
	if err != nil {
		return err
	}
	statsByGuild, err := a.discovery.GetStatsByGuilds(ctx, []int64{guildID})
	if err != nil {
		return err
	}
	return a.guildOS.IndexGuild(ctx, guildsearch.Guild{
		GuildID:      g.Id,
		Name:         g.Name,
		Description:  g.Description,
		Tags:         tagsByGuild[g.Id],
		Public:       g.Public,
		MembersCount: statsByGuild[g.Id].MembersCount,
	})
}

func (a *App) indexBot(ctx context.Context, botUserID int64) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	b, err := a.botRepo.GetBot(ctx, botUserID)
	if err != nil {
		return err
	}
	if !b.Public || b.Disabled {
		return a.botOS.DeleteBot(ctx, botUserID)
	}
	u, err := a.userRepo.GetUserById(ctx, botUserID)
	if err != nil {
		return err
	}
	tagsByBot, err := a.botRepo.GetTagsByBots(ctx, []int64{botUserID})
	if err != nil {
		return err
	}
	countsByBot, err := a.botRepo.GetInstallCounts(ctx, []int64{botUserID})
	if err != nil {
		return err
	}
	bio := ""
	if u.Bio != nil {
		bio = *u.Bio
	}
	return a.botOS.IndexBot(ctx, botsearch.Bot{
		BotUserID:     b.BotUserId,
		Name:          u.Name,
		Description:   b.Description,
		Bio:           bio,
		Tags:          tagsByBot[b.BotUserId],
		Public:        b.Public,
		Disabled:      b.Disabled,
		InstallsCount: countsByBot[b.BotUserId],
	})
}

func (a *App) backfillBotIndex() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ids, err := a.botRepo.ListPublicEnabledBotIDs(ctx, 10000)
	if err != nil {
		a.logger.ErrorContext(ctx, "failed to list public bots for backfill", slog.String("error", err.Error()))
		return
	}
	for _, id := range ids {
		if err := a.indexBot(ctx, id); err != nil {
			a.logger.ErrorContext(ctx, "failed to backfill bot index", slog.Int64("bot_user_id", id), slog.String("error", err.Error()))
		}
	}
}

func verifyInfrastructure(logger *slog.Logger, retries int, pg *pgdb.DB, database *db.CQLCon, cache *kvs.Cache, nc *nats.Conn) error {
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
	return retryStartupProbe(baseCtx, logger, "nats", retries, nc.FlushWithContext)
}

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
			logger.Warn("startup dependency probe failed; retrying",
				slog.String("dependency", dependency),
				slog.Int("attempt", attempt),
				slog.String("error", err.Error()))
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
