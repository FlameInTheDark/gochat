package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"

	"github.com/FlameInTheDark/gochat/cmd/ws/hub"
	"github.com/FlameInTheDark/gochat/cmd/ws/subscriber"
	"github.com/FlameInTheDark/gochat/internal/botauth"
	"github.com/FlameInTheDark/gochat/internal/botgateway"
	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	botrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/bot"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/dmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/groupdmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/presence"
	natsio "github.com/nats-io/nats.go"
)

const botShardLeaseTTL = 75 * time.Second
const botPresencePlatform = "bot"
const maxBotCustomStatusTextLength = 255

type presencePayload struct {
	Status           string `json:"status"`
	CustomStatusText string `json:"custom_status_text,omitempty"`
}

type identifyPayload struct {
	ShardID    int              `json:"shard_id"`
	ShardCount int              `json:"shard_count"`
	SessionID  string           `json:"session_id,omitempty"`
	Presence   *presencePayload `json:"presence,omitempty"`
}

type readyPayload struct {
	Bot                dto.User        `json:"bot"`
	SessionID          string          `json:"session_id"`
	ShardID            int             `json:"shard_id"`
	ShardCount         int             `json:"shard_count"`
	GuildIDs           []int64         `json:"guild_ids"`
	DMChannelIDs       []int64         `json:"dm_channel_ids"`
	GroupDMChannelIDs  []int64         `json:"group_dm_channel_ids"`
	ReceivesDMEvents   bool            `json:"receives_dm_events"`
	GrantedPermissions map[int64]int64 `json:"granted_permissions"`
}

type Gateway struct {
	hub              *hub.Hub
	cache            *kvs.Cache
	pstore           *presence.Store
	presenceNATS     *natsio.Conn
	registry         *botgateway.Registry
	instanceID       string
	bot              botrepo.Bot
	dm               dmchannel.DmChannel
	gdm              groupdmchannel.GroupDMChannel
	perm             rolecheck.RoleCheck
	log              *slog.Logger
	heartbeatTimeout int64
}

func New(h *hub.Hub, cache *kvs.Cache, pstore *presence.Store, presenceNATS *natsio.Conn, instanceID string, bot botrepo.Bot, dm dmchannel.DmChannel, gdm groupdmchannel.GroupDMChannel, perm rolecheck.RoleCheck, heartbeatTimeout int64, logger *slog.Logger) *Gateway {
	if logger == nil {
		logger = observability.Logger()
	}
	return &Gateway{
		hub:              h,
		cache:            cache,
		pstore:           pstore,
		presenceNATS:     presenceNATS,
		registry:         botgateway.NewRegistry(cache, botShardLeaseTTL),
		instanceID:       instanceID,
		bot:              bot,
		dm:               dm,
		gdm:              gdm,
		perm:             perm,
		log:              logger,
		heartbeatTimeout: heartbeatTimeout,
	}
}

type botGatewayConn struct {
	gateway   *Gateway
	principal *botauth.Principal
	sessionID atomic.Value
	out       chan []byte
	closed    atomic.Bool
}

func (c *botGatewayConn) Send(delivery hub.Delivery) {
	if c.closed.Load() {
		return
	}
	sessionID, _ := c.sessionID.Load().(string)
	if sessionID == "" {
		return
	}
	var msg botgateway.DeliveryMessage
	if err := json.Unmarshal(delivery.Data, &msg); err != nil {
		return
	}
	if !containsSession(msg.SessionIDs, sessionID) {
		return
	}
	select {
	case c.out <- msg.Event:
	default:
	}
}

func containsSession(ids []string, sessionID string) bool {
	for _, id := range ids {
		if id == sessionID {
			return true
		}
	}
	return false
}

func (g *Gateway) WSHandler(c *websocket.Conn) {
	raw := c.Locals("bot_principal")
	principal, ok := raw.(*botauth.Principal)
	if !ok || principal == nil {
		_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "missing bot principal"), time.Now().Add(time.Second))
		_ = c.Close()
		return
	}
	connCtx := observability.BackgroundFromContext(context.Background())
	bg := &botGatewayConn{gateway: g, principal: principal, out: make(chan []byte, 256)}
	subs := subscriber.New(g.hub, bg, nil, func() context.Context { return connCtx })
	defer close(bg.out)
	defer func() { _ = subs.Close() }()
	defer bg.closed.Store(true)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for msg := range bg.out {
			_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	if err := subs.Subscribe(connCtx, "bot-deliver-instance", botgateway.InstanceDeliverySubject(g.instanceID)); err != nil {
		_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "unable to subscribe"), time.Now().Add(time.Second))
		return
	}

	sessionID, leaseKey, leaseValue := "", "", ""
	var registeredSession *botgateway.Session
	presenceSet := false
	releaseLease := func() {
		if leaseKey == "" {
			return
		}
		val, err := g.cache.Client().Get(context.Background(), leaseKey).Result()
		if err == nil && val == leaseValue {
			_ = g.cache.Client().Del(context.Background(), leaseKey).Err()
		}
	}
	defer releaseLease()
	defer func() {
		if registeredSession != nil {
			_ = g.registry.Unregister(context.Background(), *registeredSession)
		}
		if presenceSet && sessionID != "" {
			if err := g.clearBotPresence(context.Background(), principal.BotUserID, sessionID); err != nil {
				g.log.Warn("unable to clear bot presence", "error", err, "bot_user_id", principal.BotUserID, "session_id", sessionID)
			}
		}
	}()

	_ = c.SetReadDeadline(time.Now().Add(time.Duration(g.heartbeatTimeout+15000) * time.Millisecond))
	for {
		_, rawMsg, err := c.ReadMessage()
		if err != nil {
			if !isExpectedBotWSReadError(err) {
				g.log.Warn("bot ws read error", "error", err)
			}
			return
		}
		var msg mqmsg.Message
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			continue
		}
		switch msg.Operation {
		case mqmsg.OPCodeHello:
			if sessionID != "" {
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "already identified"), time.Now().Add(time.Second))
				return
			}
			var identify identifyPayload
			if err := json.Unmarshal(msg.Data, &identify); err != nil {
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "invalid identify"), time.Now().Add(time.Second))
				return
			}
			if identify.ShardCount <= 0 || identify.ShardID < 0 || identify.ShardID >= identify.ShardCount {
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "invalid shard"), time.Now().Add(time.Second))
				return
			}
			if existing, found, err := g.registry.ExistingShardCount(connCtx, principal.BotUserID); err != nil {
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "unable to identify"), time.Now().Add(time.Second))
				return
			} else if found && existing != identify.ShardCount {
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "shard count mismatch"), time.Now().Add(time.Second))
				return
			}
			sessionID = identify.SessionID
			if sessionID == "" {
				sessionID = uuid.NewString()
			}
			if identify.ShardCount > 1 {
				leaseKey = fmt.Sprintf("botgw:lease:%d:%d:%d", principal.BotUserID, identify.ShardCount, identify.ShardID)
				leaseValue = sessionID + ":" + uuid.NewString()
				ok, err := g.cache.Client().SetNX(connCtx, leaseKey, leaseValue, botShardLeaseTTL).Result()
				if err != nil || !ok {
					_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "duplicate shard"), time.Now().Add(time.Second))
					return
				}
			}
			ready, err := g.readyForShard(connCtx, principal, identify, sessionID)
			if err != nil {
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "unable to identify"), time.Now().Add(time.Second))
				return
			}
			session := botgateway.Session{
				SessionID:        sessionID,
				BotUserID:        principal.BotUserID,
				InstanceID:       g.instanceID,
				ShardID:          identify.ShardID,
				ShardCount:       identify.ShardCount,
				ReceivesDMEvents: identify.ShardID == 0,
			}
			status, text, err := normalizeBotPresencePayload(identify.Presence)
			if err != nil {
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, err.Error()), time.Now().Add(time.Second))
				return
			}
			if err := g.registry.Register(connCtx, session); err != nil {
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "unable to identify"), time.Now().Add(time.Second))
				return
			}
			if err := g.setBotPresence(connCtx, principal.BotUserID, sessionID, status, text); err != nil {
				g.log.Warn("unable to set bot presence", "error", err, "bot_user_id", principal.BotUserID, "session_id", sessionID)
			} else {
				presenceSet = true
			}
			registeredSession = &session
			bg.sessionID.Store(sessionID)
			_ = sendBotJSON(c, mqmsg.Message{Operation: mqmsg.OPCodeHello, Data: mustRaw(mqmsg.HeartbeatInterval{HeartbeatInterval: g.heartbeatTimeout, SessionID: sessionID})})
			t := mqmsg.EventTypeGatewayReady
			_ = sendBotJSON(c, mqmsg.Message{Operation: mqmsg.OpCodeDispatch, EventType: &t, Data: mustRaw(ready)})
		case mqmsg.OPCodeHeartBeat:
			_ = c.SetReadDeadline(time.Now().Add(time.Duration(g.heartbeatTimeout+15000) * time.Millisecond))
			if leaseKey != "" {
				_ = g.cache.Client().Expire(connCtx, leaseKey, botShardLeaseTTL).Err()
			}
			if registeredSession != nil {
				_ = g.registry.Touch(connCtx, *registeredSession)
			}
			if presenceSet && sessionID != "" {
				_ = g.touchBotPresence(connCtx, principal.BotUserID, sessionID)
			}
			_ = sendBotJSON(c, mqmsg.Message{Operation: mqmsg.OPCodeHeartbeatAck, Data: mustRaw(map[string]int64{"server_time": time.Now().UnixMilli()})})
		case mqmsg.OPCodePresenceUpdate:
			if sessionID == "" {
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "identify required"), time.Now().Add(time.Second))
				return
			}
			var update mqmsg.PresenceUpdateRequest
			if err := json.Unmarshal(msg.Data, &update); err != nil {
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "invalid presence update"), time.Now().Add(time.Second))
				return
			}
			status, text, err := normalizeBotPresenceRequest(update)
			if err != nil {
				_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, err.Error()), time.Now().Add(time.Second))
				return
			}
			if err := g.setBotPresence(connCtx, principal.BotUserID, sessionID, status, text); err != nil {
				g.log.Warn("unable to update bot presence", "error", err, "bot_user_id", principal.BotUserID, "session_id", sessionID)
				continue
			}
			presenceSet = true
		default:
		}
	}
}

func (g *Gateway) readyForShard(ctx context.Context, principal *botauth.Principal, identify identifyPayload, sessionID string) (readyPayload, error) {
	installs, err := g.bot.ListBotGuilds(ctx, principal.BotUserID)
	if err != nil {
		return readyPayload{}, err
	}
	ready := readyPayload{
		Bot: dto.User{
			Id:          principal.User.Id,
			Name:        principal.User.Name,
			Bio:         principal.User.Bio,
			BannerColor: principal.User.BannerColor,
			PanelColor:  principal.User.PanelColor,
			IsBot:       principal.User.IsBot(),
		},
		SessionID:          sessionID,
		ShardID:            identify.ShardID,
		ShardCount:         identify.ShardCount,
		GuildIDs:           []int64{},
		DMChannelIDs:       []int64{},
		GroupDMChannelIDs:  []int64{},
		GrantedPermissions: map[int64]int64{},
	}
	ready.ReceivesDMEvents = identify.ShardID == 0
	for _, install := range installs {
		if int(install.GuildId%int64(identify.ShardCount)) != identify.ShardID {
			continue
		}
		ready.GuildIDs = append(ready.GuildIDs, install.GuildId)
		ready.GrantedPermissions[install.GuildId] = install.GrantedPermissions
	}
	dms, err := g.dm.GetUserDmChannels(ctx, principal.BotUserID)
	if err != nil {
		return ready, err
	}
	for _, item := range dms {
		if int(item.ChannelId%int64(identify.ShardCount)) != identify.ShardID {
			continue
		}
		ready.DMChannelIDs = append(ready.DMChannelIDs, item.ChannelId)
	}
	groupDMs, err := g.gdm.GetUserGroupDmChannels(ctx, principal.BotUserID)
	if err != nil {
		return ready, err
	}
	for _, item := range groupDMs {
		if int(item.ChannelId%int64(identify.ShardCount)) != identify.ShardID {
			continue
		}
		ready.GroupDMChannelIDs = append(ready.GroupDMChannelIDs, item.ChannelId)
	}
	return ready, nil
}

func (g *Gateway) setBotPresence(ctx context.Context, botUserID int64, sessionID, status, customText string) error {
	if g.pstore == nil || botUserID == 0 || sessionID == "" {
		return nil
	}
	now := time.Now().Unix()
	ttl := g.presenceTTLSeconds()
	existing, ok, err := g.pstore.GetSession(ctx, botUserID, sessionID)
	if err != nil {
		return err
	}
	since := now
	if ok && existing.Since > 0 && existing.Status == status {
		since = existing.Since
	}
	sp := presence.SessionPresence{
		SessionID:        sessionID,
		Status:           status,
		Platform:         botPresencePlatform,
		Since:            since,
		UpdatedAt:        now,
		ExpiresAt:        now + ttl,
		CustomStatusText: customText,
	}
	_ = g.pstore.ClearOverride(ctx, botUserID)
	if err := g.pstore.UpsertSession(ctx, botUserID, sessionID, sp, ttl); err != nil {
		return err
	}
	_, err = presence.Refresh(ctx, g.pstore, g.presenceNATS, botUserID, ttl)
	return err
}

func (g *Gateway) touchBotPresence(ctx context.Context, botUserID int64, sessionID string) error {
	if g.pstore == nil || botUserID == 0 || sessionID == "" {
		return nil
	}
	return g.pstore.TouchSessionTTL(ctx, botUserID, sessionID, g.presenceTTLSeconds())
}

func (g *Gateway) clearBotPresence(ctx context.Context, botUserID int64, sessionID string) error {
	if g.pstore == nil || botUserID == 0 || sessionID == "" {
		return nil
	}
	ttl := g.presenceTTLSeconds()
	if err := g.pstore.RemoveSession(ctx, botUserID, sessionID, ttl); err != nil {
		return err
	}
	_, err := presence.Refresh(ctx, g.pstore, g.presenceNATS, botUserID, ttl)
	return err
}

func (g *Gateway) presenceTTLSeconds() int64 {
	ttl := g.heartbeatTimeout * 2 / 1000
	if ttl < 1 {
		return 1
	}
	return ttl
}

func normalizeBotPresencePayload(input *presencePayload) (string, string, error) {
	if input == nil {
		return presence.StatusOnline, "", nil
	}
	return normalizeBotPresence(input.Status, input.CustomStatusText)
}

func normalizeBotPresenceRequest(input mqmsg.PresenceUpdateRequest) (string, string, error) {
	return normalizeBotPresence(input.Status, input.CustomStatusText)
}

func normalizeBotPresence(status, customText string) (string, string, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		status = presence.StatusOnline
	}
	switch status {
	case presence.StatusOnline, presence.StatusIdle, presence.StatusDND, presence.StatusOffline:
	default:
		return "", "", fmt.Errorf("invalid presence status")
	}
	customText = strings.TrimSpace(customText)
	if len([]rune(customText)) > maxBotCustomStatusTextLength {
		return "", "", fmt.Errorf("custom status text is too long")
	}
	return status, customText, nil
}

func sendBotJSON(c *websocket.Conn, v any) error {
	_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.WriteJSON(v)
}

func mustRaw(v any) json.RawMessage {
	raw, _ := json.Marshal(v)
	return raw
}

func isExpectedBotWSReadError(err error) bool {
	if err == nil {
		return false
	}
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseProtocolError, websocket.CloseNoStatusReceived, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		return true
	}
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, net.ErrClosed)
}
