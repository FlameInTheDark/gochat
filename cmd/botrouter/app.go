package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/FlameInTheDark/gochat/cmd/botrouter/config"
	"github.com/FlameInTheDark/gochat/internal/botgateway"
	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	botrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/bot"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/shutter"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const sessionTTL = 75 * time.Second

const renewPartitionLeaseScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("EXPIRE", KEYS[1], ARGV[2])
end
return 0
`

type App struct {
	cfg        *config.Config
	log        *slog.Logger
	nc         *nats.Conn
	pg         *pgdb.DB
	cache      *kvs.Cache
	bot        botrepo.Bot
	perm       rolecheck.RoleCheck
	registry   *botgateway.Registry
	routerID   string
	partitions int

	mu    sync.Mutex
	owned map[int]*nats.Subscription
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) (*App, error) {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		return nil, err
	}
	routerID := cfg.BotRouterID
	if routerID == "" {
		routerID = uuid.NewString()
	}

	nc, err := nats.Connect(cfg.BotNATSConnString, nats.Compression(true))
	if err != nil {
		return nil, err
	}
	shut.UpFunc(nc.Close)

	pg := pgdb.NewDB(logger)
	if err := pg.Connect(cfg.PGDSN, pgdb.ConnectOptions{
		DriverName:   cfg.PGDriver,
		MaxRetries:   cfg.PGRetries,
		QueryLog:     cfg.PGQueryLog,
		MaxOpenConns: cfg.PGMaxOpenConns,
		MaxIdleConns: cfg.PGMaxIdleConns,
	}); err != nil {
		return nil, err
	}
	shut.Up(pg)
	postgresProbeCtx, cancelPostgresProbe := context.WithCancel(context.Background())
	shut.UpFunc(cancelPostgresProbe)
	go pg.StartProbeLoop(postgresProbeCtx, 30*time.Second)

	cache, err := kvs.New(cfg.KeyDB, kvs.Options{
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
	})
	if err != nil {
		return nil, err
	}
	shut.Up(cache)

	app := &App{
		cfg:        cfg,
		log:        logger,
		nc:         nc,
		pg:         pg,
		cache:      cache,
		bot:        botrepo.New(pg.Conn()),
		perm:       rolecheck.New(pg),
		registry:   botgateway.NewRegistry(cache, sessionTTL),
		routerID:   routerID,
		partitions: botgateway.NormalizePartitions(cfg.BotEventPartitions),
		owned:      map[int]*nats.Subscription{},
	}
	return app, nil
}

func (a *App) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.log.Info("Bot router starting", slog.String("router_id", a.routerID), slog.Int("partitions", a.partitions))
	go a.acquireLoop(ctx)
	go a.renewLoop(ctx)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
}

func (a *App) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	var errs []error
	for partition, sub := range a.owned {
		if err := sub.Unsubscribe(); err != nil {
			errs = append(errs, err)
		}
		delete(a.owned, partition)
	}
	return errors.Join(errs...)
}

func (a *App) acquireLoop(ctx context.Context) {
	a.acquirePartitions(ctx)
	ticker := time.NewTicker(a.cfg.BotRouterAcquireEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.acquirePartitions(ctx)
		}
	}
}

func (a *App) renewLoop(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.BotRouterLeaseRenew)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.renewPartitions(ctx)
		}
	}
}

func (a *App) acquirePartitions(ctx context.Context) {
	for partition := 0; partition < a.partitions; partition++ {
		if a.hasPartition(partition) {
			continue
		}
		ok, err := a.cache.Client().SetNX(ctx, partitionLeaseKey(partition), a.routerID, a.cfg.BotRouterLeaseTTL).Result()
		if err != nil {
			a.log.Warn("bot router partition lease acquire failed", slog.Int("partition", partition), slog.String("error", err.Error()))
			continue
		}
		if !ok {
			continue
		}
		if err := a.subscribePartition(ctx, partition); err != nil {
			a.log.Error("bot router partition subscribe failed", slog.Int("partition", partition), slog.String("error", err.Error()))
			_ = a.cache.Client().Del(ctx, partitionLeaseKey(partition)).Err()
		}
	}
}

func (a *App) hasPartition(partition int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, ok := a.owned[partition]
	return ok
}

func (a *App) subscribePartition(ctx context.Context, partition int) error {
	subject := botgateway.PartitionWildcardSubject(partition)
	sub, err := a.nc.Subscribe(subject, func(msg *nats.Msg) {
		a.handleEvent(ctx, msg)
	})
	if err != nil {
		return err
	}
	a.mu.Lock()
	a.owned[partition] = sub
	a.mu.Unlock()
	a.log.Info("bot router partition acquired", slog.Int("partition", partition), slog.String("subject", subject))
	return nil
}

func (a *App) renewPartitions(ctx context.Context) {
	a.mu.Lock()
	partitions := make([]int, 0, len(a.owned))
	for partition := range a.owned {
		partitions = append(partitions, partition)
	}
	a.mu.Unlock()
	sort.Ints(partitions)

	for _, partition := range partitions {
		ok, err := a.cache.Client().Eval(ctx, renewPartitionLeaseScript, []string{partitionLeaseKey(partition)}, a.routerID, int(a.cfg.BotRouterLeaseTTL.Seconds())).Int()
		if err != nil || ok != 1 {
			a.releasePartition(partition)
		}
	}
}

func (a *App) releasePartition(partition int) {
	a.mu.Lock()
	sub := a.owned[partition]
	delete(a.owned, partition)
	a.mu.Unlock()
	if sub != nil {
		_ = sub.Unsubscribe()
	}
	a.log.Warn("bot router partition released", slog.Int("partition", partition))
}

func (a *App) handleEvent(ctx context.Context, msg *nats.Msg) {
	subject, ok := botgateway.ParseEventSubject(msg.Subject)
	if !ok {
		return
	}
	ctx = observability.ExtractNATSContext(ctx, msg)
	switch subject.Kind {
	case botgateway.EventKindGuild, botgateway.EventKindGuildChannel:
		a.routeGuildEvent(ctx, subject, msg.Data)
	case botgateway.EventKindUserDM:
		a.routeDMEvent(ctx, subject, msg.Data)
	}
}

func (a *App) routeGuildEvent(ctx context.Context, subject botgateway.EventSubject, raw []byte) {
	installs, err := a.bot.ListGuildBots(ctx, subject.GuildID)
	if err != nil {
		a.log.Warn("bot router list guild bots failed", slog.Int64("guild_id", subject.GuildID), slog.String("error", err.Error()))
		return
	}
	channelScoped := subject.Kind == botgateway.EventKindGuildChannel
	required := requiredPermissionsForEvent(raw, channelScoped)
	for _, install := range installs {
		if !a.botCanReceive(ctx, install.BotUserId, subject.GuildID, subject.ChannelID, channelScoped, required) {
			continue
		}
		sessions, err := a.registry.BotSessions(ctx, install.BotUserId)
		if err != nil {
			a.log.Warn("bot router list sessions failed", slog.Int64("bot_user_id", install.BotUserId), slog.String("error", err.Error()))
			continue
		}
		a.publishToSessions(ctx, botgateway.TargetGuildSessions(sessions, subject.GuildID), raw)
	}
}

func (a *App) routeDMEvent(ctx context.Context, subject botgateway.EventSubject, raw []byte) {
	if !allowedDMEvent(raw) {
		return
	}
	sessions, err := a.registry.BotSessions(ctx, subject.UserID)
	if err != nil {
		a.log.Warn("bot router list dm sessions failed", slog.Int64("bot_user_id", subject.UserID), slog.String("error", err.Error()))
		return
	}
	a.publishToSessions(ctx, botgateway.TargetDMSessions(sessions), raw)
}

func (a *App) botCanReceive(ctx context.Context, botUserID, guildID, channelID int64, channelScoped bool, required []permissions.RolePermission) bool {
	if channelScoped {
		_, _, _, ok, err := a.perm.ChannelPerm(ctx, guildID, channelID, botUserID, required...)
		return err == nil && ok
	}
	_, ok, err := a.perm.GuildPerm(ctx, guildID, botUserID, required...)
	return err == nil && ok
}

func (a *App) publishToSessions(ctx context.Context, sessions []botgateway.Session, raw []byte) {
	grouped := botgateway.GroupByInstance(sessions)
	for instanceID, sessionIDs := range grouped {
		body, err := json.Marshal(botgateway.DeliveryMessage{SessionIDs: sessionIDs, Event: raw})
		if err != nil {
			continue
		}
		subject := botgateway.InstanceDeliverySubject(instanceID)
		headers := observability.InjectNATSHeaders(ctx, nil)
		if err := a.nc.PublishMsg(&nats.Msg{Subject: subject, Header: headers, Data: body}); err != nil {
			a.log.Warn("bot router delivery publish failed", slog.String("instance_id", instanceID), slog.String("error", err.Error()))
		}
	}
}

func partitionLeaseKey(partition int) string {
	return fmt.Sprintf("botrouter:partition:%d", partition)
}

func requiredPermissionsForEvent(raw []byte, channelScoped bool) []permissions.RolePermission {
	if !channelScoped {
		return []permissions.RolePermission{permissions.PermServerViewChannels}
	}
	var envelope mqmsg.Message
	if err := json.Unmarshal(raw, &envelope); err != nil || envelope.EventType == nil {
		return []permissions.RolePermission{permissions.PermServerViewChannels}
	}
	switch *envelope.EventType {
	case mqmsg.EventTypeMessageCreate, mqmsg.EventTypeMessageUpdate, mqmsg.EventTypeMessageDelete,
		mqmsg.EventTypeMessageReactionAdd, mqmsg.EventTypeMessageReactionRemove:
		return []permissions.RolePermission{permissions.PermServerViewChannels, permissions.PermTextReadMessageHistory}
	default:
		return []permissions.RolePermission{permissions.PermServerViewChannels}
	}
}

func allowedDMEvent(raw []byte) bool {
	var envelope mqmsg.Message
	if err := json.Unmarshal(raw, &envelope); err != nil || envelope.EventType == nil {
		return false
	}
	return *envelope.EventType == mqmsg.EventTypeUserDMMessage
}
