package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/FlameInTheDark/gochat/cmd/ws/auth"
	"github.com/FlameInTheDark/gochat/cmd/ws/subscriber"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/presence"
	"github.com/nats-io/nats.go"
)

type helloMessage struct {
	Token              string `json:"token"`
	HeartbeatSessionID string `json:"heartbeat_session_id,omitempty"`
}

type heartbeatMessage struct {
	// Seconds since connection opened
	LastEventId int64 `json:"e"`
}

type Handler struct {
	user     *dto.User
	sub      *subscriber.Subscriber
	g        guild.Guild
	m        member.Member
	u        user.User
	gc       guildchannels.GuildChannels
	perm     rolecheck.RoleCheck
	jwt      *auth.Auth
	sendJSON func(v any) error
	nats     *nats.Conn
	pstore   *presence.Store
	// IDs this connection is watching for presence updates
	psubs map[int64]struct{}
	// Whether we successfully set presence after hello
	presenceSet bool
	// session identifier for this ws connection
	sessionID string

	lastEventId int64
	hbTimeout   int64
	hTimer      *time.Timer
	initTimer   *time.Timer
	closer      func()
	log         *slog.Logger
}

func New(c *db.CQLCon, pg *pgdb.DB, sub *subscriber.Subscriber, sendJSON func(v any) error, jwt *auth.Auth, hbTimeout int64, closer func(), logger *slog.Logger, nats *nats.Conn, pstore *presence.Store) *Handler {
	initTimer := time.AfterFunc(time.Second*5, closer)
	return &Handler{
		sub:      sub,
		g:        guild.New(pg.Conn()),
		m:        member.New(pg.Conn()),
		u:        user.New(pg.Conn()),
		gc:       guildchannels.New(pg.Conn()),
		perm:     rolecheck.New(c, pg),
		jwt:      jwt,
		sendJSON: sendJSON,
		nats:     nats,
		pstore:   pstore,
		psubs:    make(map[int64]struct{}),

		hbTimeout: hbTimeout,
		initTimer: initTimer,
		closer:    closer,
		log:       logger,
	}
}

func (h *Handler) HandleMessage(e mqmsg.Message) {
	if e.Operation != mqmsg.OPCodeHello && h.user == nil {
		return
	}
	switch e.Operation {
	case mqmsg.OPCodeHello:
		h.hello(&e)
	case mqmsg.OPCodeHeartBeat:
		if len(e.Data) == 0 || string(bytes.TrimSpace(e.Data)) == "null" {
			return
		}
		var m heartbeatMessage
		err := json.Unmarshal(e.Data, &m)
		if err != nil {
			h.log.Warn("Error unmarshalling heart beat msg", "error", err)
			return
		}
		if m.LastEventId >= h.lastEventId {
			prev := h.lastEventId
			// add grace to tolerate network jitter (10s)
			h.hTimer.Reset(time.Millisecond * time.Duration(h.hbTimeout+10000))
			// Refresh this session TTL: heartbeat_interval * 2
			if h.user != nil && h.pstore != nil && h.sessionID != "" && h.presenceSet {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				// TTL expects seconds
				ttl := h.hbTimeout * 2 / 1000
				if ttl < 1 {
					ttl = 1
				}
				_ = h.pstore.TouchSessionTTL(ctx, h.user.Id, h.sessionID, ttl)
				cancel()
			}
			h.lastEventId = m.LastEventId
			h.log.Info("Heartbeat OK; timer reset", "user_id", func() any {
				if h.user != nil {
					return h.user.Id
				}
				return int64(0)
			}(), "prev_e", prev, "new_e", m.LastEventId)
		}
	case mqmsg.OPCodeChannelSubscription:
		var m mqmsg.Subscribe
		err := json.Unmarshal(e.Data, &m)
		if err != nil {
			h.log.Warn("Error unmarshalling channel subscription msg", "error", err)
			return
		}

		if m.Channel != nil {
			gc, err := h.gc.GetGuildByChannel(context.Background(), *m.Channel)
			if err != nil {
				h.log.Warn("Error getting guild channel", "error", err, "channel_id", *m.Channel)
			} else {
				_, _, _, ok, err := h.perm.ChannelPerm(context.Background(), gc.GuildId, gc.ChannelId, h.user.Id, permissions.PermServerViewChannels)
				if err != nil {
					h.log.Warn("Error checking channel permissions", "error", err)
				} else if ok {
					err := h.sub.Subscribe("channel", fmt.Sprintf("channel.%d", *m.Channel))
					if err != nil {
						h.log.Warn("Error subscribing to channel", "error", err)
					}
				} else {
					h.log.Warn("User does not have permission to view channel", "user_id", h.user.Id, "channel_id", *m.Channel)
				}
			}
		}

		for _, guildID := range m.Guilds {
			ok, err := h.m.IsGuildMember(context.Background(), guildID, h.user.Id)
			if err != nil {
				h.log.Warn("Error checking guild access", "error", err)
			} else if ok {
				err := h.sub.Subscribe(fmt.Sprintf("guild.%d", guildID), fmt.Sprintf("guild.%d", guildID))
				if err != nil {
					h.log.Warn("Error subscribing to guild", "error", err)
				}
			} else {
				h.log.Warn("User does not have permission to view guild", "user_id", h.user.Id, "guild_id", guildID)
			}
		}

	case mqmsg.OPCodePresenceSubscription:
		var m mqmsg.PresenceSubscription
		if err := json.Unmarshal(e.Data, &m); err != nil {
			h.log.Warn("Error unmarshalling presence subscription msg", "error", err)
			return
		}

		if m.Clear {
			for uid := range h.psubs {
				_ = h.sub.Unsubscribe(fmt.Sprintf("presence.%d", uid))
				delete(h.psubs, uid)
			}
		}
		if len(m.Set) > 0 {
			for uid := range h.psubs {
				_ = h.sub.Unsubscribe(fmt.Sprintf("presence.%d", uid))
				delete(h.psubs, uid)
			}
			for _, uid := range m.Set {
				key := fmt.Sprintf("presence.%d", uid)
				if err := h.sub.Subscribe(key, fmt.Sprintf("presence.user.%d", uid)); err != nil {
					h.log.Warn("Error subscribing to presence", "error", err, "user_id", uid)
					continue
				}
				h.psubs[uid] = struct{}{}
				h.sendPresenceSnapshot(uid)
			}
		}

		for _, uid := range m.Add {
			if _, ok := h.psubs[uid]; ok {
				continue
			}
			key := fmt.Sprintf("presence.%d", uid)
			if err := h.sub.Subscribe(key, fmt.Sprintf("presence.user.%d", uid)); err != nil {
				h.log.Warn("Error subscribing to presence", "error", err, "user_id", uid)
				continue
			}
			h.psubs[uid] = struct{}{}
			h.sendPresenceSnapshot(uid)
		}

		for _, uid := range m.Remove {
			if _, ok := h.psubs[uid]; !ok {
				continue
			}
			_ = h.sub.Unsubscribe(fmt.Sprintf("presence.%d", uid))
			delete(h.psubs, uid)
		}

	case mqmsg.OPCodePresenceUpdate:
		if h.user == nil || h.sessionID == "" || h.pstore == nil {
			return
		}
		var m struct {
			Status           string `json:"status"`
			Platform         string `json:"platform,omitempty"`
			CustomStatusText string `json:"custom_status_text,omitempty"`
		}
		if err := json.Unmarshal(e.Data, &m); err != nil {
			h.log.Warn("Error unmarshalling presence update msg", "error", err)
			return
		}
		// Allow offline for manual invisible mode; other valid statuses are online/idle/dnd
		now := time.Now().Unix()
		ttl := h.hbTimeout * 2 / 1000
		if ttl < 1 {
			ttl = 1
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		if m.Status == presence.StatusOffline {
			// Set global override to appear offline
			if err := h.pstore.SetOverride(ctx, h.user.Id, presence.StatusOffline, now, m.CustomStatusText); err != nil {
				h.log.Warn("Error setting offline override", "error", err)
				return
			}
			agg, _, _ := h.pstore.Aggregate(ctx, h.user.Id, now)
			_ = h.pstore.SetAggregated(ctx, agg, ttl)
			h.publishPresence(agg)
			return
		}
		// Clear override and upsert session presence
		if err := h.pstore.ClearOverride(ctx, h.user.Id); err != nil {
			h.log.Warn("Error clearing presence override", "error", err)
		}
		if h.sessionID == "" {
			h.sessionID = fmt.Sprintf("%d-%d", h.user.Id, now)
		}
		switch m.Status {
		case presence.StatusOnline, presence.StatusIdle, presence.StatusDND:
		default:
			return
		}
		sp := presence.SessionPresence{SessionID: h.sessionID, Status: m.Status, Platform: m.Platform, Since: now, UpdatedAt: now, ExpiresAt: now + ttl, CustomStatusText: m.CustomStatusText}
		if err := h.pstore.UpsertSession(ctx, h.user.Id, h.sessionID, sp, ttl); err != nil {
			h.log.Warn("Error upserting session presence", "error", err)
			return
		}
		agg, _, _ := h.pstore.Aggregate(ctx, h.user.Id, now)
		_ = h.pstore.SetAggregated(ctx, agg, ttl)
		h.publishPresence(agg)
		h.presenceSet = true

	default:
		h.log.Warn("Unknown operation", "operation", e.Operation)
	}
}

func (h *Handler) Close() error {
	h.OnWSClosed()
	h.closer()
	return nil
}

func (h *Handler) OnWSClosed() {
	if h.user == nil || h.pstore == nil || h.nats == nil || !h.presenceSet || h.sessionID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	// Read previous aggregated presence
	prev, _, _ := h.pstore.Get(ctx, h.user.Id)
	// Remove this session
	ttl := h.hbTimeout * 2 / 1000
	if ttl < 1 {
		ttl = 1
	}
	_ = h.pstore.RemoveSession(ctx, h.user.Id, h.sessionID, ttl)
	// Re-aggregate
	now := time.Now().Unix()
	agg, _, _ := h.pstore.Aggregate(ctx, h.user.Id, now)
	// If changed, store and publish (status or text)
	if agg.Status != prev.Status || agg.CustomStatusText != prev.CustomStatusText {
		_ = h.pstore.SetAggregated(ctx, agg, ttl)
		h.publishPresence(agg)
	}
}

func (h *Handler) sendPresenceSnapshot(userID int64) {
	// Read presence from cache and send to this connection only
	if h.pstore == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	p, ok, _ := h.pstore.Get(ctx, userID)
	status := presence.StatusOffline
	since := time.Now().Unix()
	text := ""
	if ok {
		status = p.Status
		since = p.Since
		text = p.CustomStatusText
	}
	msg, err := mqmsg.BuildEventMessage(&mqmsg.PresenceUpdate{UserID: userID, Status: status, Since: since, CustomStatusText: text})
	if err != nil {
		return
	}
	_ = h.sendJSON(msg)
}

func (h *Handler) publishPresence(agg presence.Presence) {
	if h.nats == nil {
		return
	}
	msg, err := mqmsg.BuildEventMessage(&mqmsg.PresenceUpdate{UserID: agg.UserID, Status: agg.Status, Since: agg.Since, CustomStatusText: agg.CustomStatusText})
	if err != nil {
		return
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	_ = h.nats.Publish(fmt.Sprintf("presence.user.%d", agg.UserID), b)
}
