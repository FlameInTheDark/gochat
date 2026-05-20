package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"sort"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/pgentities/rolecheck"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/FlameInTheDark/gochat/cmd/ws/auth"
	"github.com/FlameInTheDark/gochat/cmd/ws/subscriber"
	"github.com/FlameInTheDark/gochat/internal/cache/kvs"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guildchannelmessages"
	"github.com/FlameInTheDark/gochat/internal/database/entities/readstates"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/dmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/friend"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/groupdmchannel"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/usersettings"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/gatewaystate"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/presence"
	"github.com/nats-io/nats.go"
)

type helloMessage struct {
	Token              string `json:"token"`
	HeartbeatSessionID string `json:"heartbeat_session_id,omitempty"`
	ClientInstanceID   string `json:"client_instance_id,omitempty"`
	ResumeSessionID    string `json:"resume_session_id,omitempty"`
	ResumeGeneration   int64  `json:"resume_generation,omitempty"`
	LastSeq            int64  `json:"last_seq,omitempty"`
	KnownReadyVersion  int64  `json:"known_ready_version,omitempty"`
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
	dm       dmchannel.DmChannel
	gdm      groupdmchannel.GroupDMChannel
	fr       friend.Friend
	uset     usersettings.UserSettings
	rs       readstates.ReadStates
	gclm     guildchannelmessages.GuildChannelMessages
	u        user.User
	gc       guildchannels.GuildChannels
	perm     rolecheck.RoleCheck
	jwt      *auth.Auth
	sendJSON func(v any) error
	nats     *nats.Conn
	pstore   *presence.Store
	gstate   *gatewaystate.Store
	// IDs this connection is watching for presence updates
	psubs     map[int64]struct{}
	autoPsubs map[int64]struct{}
	// Channel IDs this connection is explicitly subscribed to.
	csubs     map[int64]struct{}
	hTimer    *time.Timer
	initTimer *time.Timer
	closer    func()
	log       *slog.Logger
	cache     *kvs.Cache
	ctx       context.Context
	telemetry *observability.WSTelemetry

	// lastPresenceTouch throttles TouchSessionTTL calls to avoid
	// redundant Redis round-trips on every heartbeat.
	lastPresenceTouch time.Time
	// callback used to expose the authenticated user id to the outer ws connection
	onAuthenticated func(userID int64)
	// session identifier for this ws connection
	sessionID        string
	connectionID     string
	clientInstanceID string
	generation       int64
	lastEventId      int64
	hbTimeout        int64
	// Whether we successfully set presence after hello
	presenceSet bool
}

func New(c *db.CQLCon, pg *pgdb.DB, sub *subscriber.Subscriber, sendJSON func(v any) error, jwt *auth.Auth, hbTimeout int64, closer func(), logger *slog.Logger, nats *nats.Conn, pstore *presence.Store, gstate *gatewaystate.Store, cache *kvs.Cache, onAuthenticated func(userID int64), baseCtx context.Context, telemetry *observability.WSTelemetry) *Handler {
	initTimer := time.AfterFunc(time.Second*5, closer)
	return &Handler{
		sub:       sub,
		g:         guild.New(pg.Conn()),
		m:         member.New(pg.Conn()),
		dm:        dmchannel.New(pg.Conn()),
		gdm:       groupdmchannel.New(pg.Conn()),
		fr:        friend.New(pg.Conn()),
		uset:      usersettings.New(pg.Conn()),
		rs:        readstates.New(c),
		gclm:      guildchannelmessages.New(c),
		u:         user.New(pg.Conn()),
		gc:        guildchannels.New(pg.Conn()),
		perm:      rolecheck.New(pg),
		jwt:       jwt,
		sendJSON:  sendJSON,
		nats:      nats,
		pstore:    pstore,
		gstate:    gstate,
		psubs:     make(map[int64]struct{}),
		autoPsubs: make(map[int64]struct{}),
		csubs:     make(map[int64]struct{}),

		hbTimeout:       hbTimeout,
		initTimer:       initTimer,
		closer:          closer,
		log:             logger,
		cache:           cache,
		onAuthenticated: onAuthenticated,
		ctx:             baseCtx,
		telemetry:       telemetry,
	}
}

func (h *Handler) HandleMessage(e mqmsg.Message) {
	ctx, span := h.startMessageSpan(e)
	defer span.End()
	log := helper.WithContext(h.log, ctx)

	if e.Operation != mqmsg.OPCodeHello && h.user == nil {
		return
	}
	h.telemetry.MessageIn(ctx, int(e.Operation), messageEventType(e.EventType))
	switch e.Operation {
	case mqmsg.OPCodeHello:
		h.hello(&e)
	case mqmsg.OPCodeHeartBeat:
		if len(e.Data) == 0 || string(bytes.TrimSpace(e.Data)) == "null" {
			h.telemetry.Heartbeat(ctx, "empty")
			return
		}
		var m heartbeatMessage
		err := json.Unmarshal(e.Data, &m)
		if err != nil {
			h.telemetry.Heartbeat(ctx, "invalid")
			log.Warn("Error unmarshalling heart beat msg", "error", err)
			return
		}
		if m.LastEventId >= h.lastEventId {
			h.telemetry.Heartbeat(ctx, "accepted")
			// add grace to tolerate network jitter (10s)
			h.hTimer.Reset(time.Millisecond * time.Duration(h.hbTimeout+10000))
			if h.gstate != nil && h.connectionID != "" {
				opCtx, cancel := context.WithTimeout(ctx, time.Second)
				ttl := h.connectionTTLSeconds()
				_ = h.gstate.TouchConnection(opCtx, h.connectionID, m.LastEventId, ttl)
				cancel()
			}
			// Refresh this session TTL: heartbeat_interval * 2
			// Throttled: skip if we touched within the last 10s.
			if h.user != nil && h.pstore != nil && h.sessionID != "" && h.presenceSet &&
				time.Since(h.lastPresenceTouch) > 10*time.Second {
				opCtx, cancel := context.WithTimeout(ctx, time.Second*2)
				// TTL expects seconds
				ttl := h.presenceTTLSeconds()
				if h.connectionID != "" && h.generation > 0 {
					_ = h.pstore.TouchSessionTTLIfOwner(opCtx, h.user.Id, h.sessionID, h.connectionID, h.generation, ttl)
				} else {
					_ = h.pstore.TouchSessionTTL(opCtx, h.user.Id, h.sessionID, ttl)
				}
				cancel()
				h.lastPresenceTouch = time.Now()
			}
			h.lastEventId = m.LastEventId
			_ = h.sendHeartbeatAck(m.LastEventId)
		} else {
			h.telemetry.Heartbeat(ctx, "stale")
		}
	case mqmsg.OPCodeChannelSubscription:
		var m mqmsg.Subscribe
		err := json.Unmarshal(e.Data, &m)
		if err != nil {
			log.Warn("Error unmarshalling channel subscription msg", "error", err)
			return
		}

		if requestedChannels, ok := resolveRequestedChannels(m); ok {
			h.syncChannelSubscriptions(ctx, requestedChannels)
		}

		for _, guildID := range m.Guilds {
			ok, err := h.m.IsGuildMember(ctx, guildID, h.user.Id)
			if err != nil {
				log.Warn("Error checking guild access", "error", err)
			} else if ok {
				err := h.sub.Subscribe(ctx, fmt.Sprintf("guild.%d", guildID), fmt.Sprintf("guild.%d", guildID))
				if err != nil {
					log.Warn("Error subscribing to guild", "error", err)
				}
			} else {
				log.Warn("User does not have permission to view guild", "user_id", h.user.Id, "guild_id", guildID)
			}
		}

	case mqmsg.OPCodePresenceSubscription:
		var m mqmsg.PresenceSubscription
		if err := json.Unmarshal(e.Data, &m); err != nil {
			log.Warn("Error unmarshalling presence subscription msg", "error", err)
			return
		}

		if m.Clear {
			for uid := range h.psubs {
				if h.isAutoPresenceSubscription(uid) {
					continue
				}
				_ = h.sub.Unsubscribe(ctx, fmt.Sprintf("presence.%d", uid))
				delete(h.psubs, uid)
			}
		}
		if len(m.Set) > 0 {
			for uid := range h.psubs {
				if h.isAutoPresenceSubscription(uid) {
					continue
				}
				_ = h.sub.Unsubscribe(ctx, fmt.Sprintf("presence.%d", uid))
				delete(h.psubs, uid)
			}
			for _, uid := range m.Set {
				key := fmt.Sprintf("presence.%d", uid)
				if err := h.sub.Subscribe(ctx, key, fmt.Sprintf("presence.user.%d", uid)); err != nil {
					log.Warn("Error subscribing to presence", "error", err, "user_id", uid)
					continue
				}
				h.psubs[uid] = struct{}{}
				h.sendPresenceSnapshot(ctx, uid)
			}
		}

		for _, uid := range m.Add {
			if _, ok := h.psubs[uid]; ok {
				continue
			}
			key := fmt.Sprintf("presence.%d", uid)
			if err := h.sub.Subscribe(ctx, key, fmt.Sprintf("presence.user.%d", uid)); err != nil {
				log.Warn("Error subscribing to presence", "error", err, "user_id", uid)
				continue
			}
			h.psubs[uid] = struct{}{}
			h.sendPresenceSnapshot(ctx, uid)
		}

		for _, uid := range m.Remove {
			if _, ok := h.psubs[uid]; !ok {
				continue
			}
			if h.isAutoPresenceSubscription(uid) {
				continue
			}
			_ = h.sub.Unsubscribe(ctx, fmt.Sprintf("presence.%d", uid))
			delete(h.psubs, uid)
		}
		h.persistClientState(ctx)

	case mqmsg.OPCodeRTC:
		// Only handle RTCBindingAlive keepalive to refresh per-channel route TTL
		if e.EventType == nil {
			return
		}
		if *e.EventType != mqmsg.EventTypeRTCBindingAlive {
			return
		}
		if h.cache == nil {
			return
		}
		var m struct {
			Channel int64 `json:"channel"`
			DMCall  bool  `json:"dm_call"`
		}
		if err := json.Unmarshal(e.Data, &m); err != nil {
			return
		}
		if m.Channel <= 0 {
			return
		}
		opCtx, cancel := context.WithTimeout(ctx, time.Second)
		if m.DMCall {
			_ = h.cache.SetTTL(opCtx, fmt.Sprintf("dmcall:route:%d", m.Channel), 60)
			cancel()
			return
		}
		_ = h.cache.SetTTL(opCtx, fmt.Sprintf("voice:route:%d", m.Channel), 60)
		// Update this session's voice channel and publish aggregated presence
		if h.pstore != nil && h.sessionID != "" && h.user != nil {
			// Set session voice channel
			ch := m.Channel
			if h.connectionID != "" && h.generation > 0 {
				_ = h.pstore.SetSessionVoiceChannelIfOwner(opCtx, h.user.Id, h.sessionID, h.connectionID, h.generation, &ch, h.presenceTTLSeconds())
			} else {
				_ = h.pstore.SetSessionVoiceChannel(opCtx, h.user.Id, h.sessionID, &ch, h.presenceTTLSeconds())
			}
			agg, _, _ := h.pstore.Aggregate(opCtx, h.user.Id, time.Now().Unix())
			// cache aggregated presence and publish
			_ = h.pstore.SetAggregated(opCtx, agg, h.hbTimeout*2/1000)
			h.publishPresence(agg)
		}
		cancel()
		return
	case mqmsg.OPCodePresenceUpdate:
		if h.user == nil || h.sessionID == "" || h.pstore == nil {
			return
		}

		var m mqmsg.PresenceUpdateRequest
		if err := json.Unmarshal(e.Data, &m); err != nil {
			log.Warn("Error unmarshalling presence update msg", "error", err)
			return
		}
		// Allow offline for manual invisible mode; other valid statuses are online/idle/dnd
		now := time.Now().Unix()
		ttl := h.presenceTTLSeconds()

		opCtx, cancel := context.WithTimeout(ctx, time.Second*2)
		defer cancel()

		if m.Status == presence.StatusOffline {
			// Set global override to appear offline
			if err := h.pstore.SetOverride(opCtx, h.user.Id, presence.StatusOffline, now, m.CustomStatusText); err != nil {
				log.Warn("Error setting offline override", "error", err)
				return
			}
			agg, _, _ := h.pstore.Aggregate(opCtx, h.user.Id, now)
			_ = h.pstore.SetAggregated(opCtx, agg, ttl)
			h.publishPresence(agg)
			return
		}

		// Clear override and upsert session presence
		if err := h.pstore.ClearOverride(opCtx, h.user.Id); err != nil {
			log.Warn("Error clearing presence override", "error", err)
		}
		if h.sessionID == "" {
			h.sessionID = fmt.Sprintf("%d-%d", h.user.Id, now)
		}

		switch m.Status {
		case presence.StatusOnline, presence.StatusIdle, presence.StatusDND:
		default:
			return
		}

		existingSession, ok, err := h.pstore.GetSession(opCtx, h.user.Id, h.sessionID)
		if err != nil {
			log.Warn("Error loading existing session presence", "error", err)
			return
		}
		if !ok {
			existingSession = presence.SessionPresence{}
		}

		sp := mergePresenceUpdate(existingSession, h.sessionID, m, now, ttl)
		sp.ConnectionID = h.connectionID
		sp.Generation = h.generation

		if err := h.pstore.UpsertSession(opCtx, h.user.Id, h.sessionID, sp, ttl); err != nil {
			log.Warn("Error upserting session presence", "error", err)
			return
		}

		agg, _, _ := h.pstore.Aggregate(opCtx, h.user.Id, now)
		_ = h.pstore.SetAggregated(opCtx, agg, ttl)
		h.publishPresence(agg)
		h.presenceSet = true

	default:
		log.Warn("Unknown operation", "operation", e.Operation)
	}
}

func mergePresenceUpdate(existing presence.SessionPresence, sessionID string, update mqmsg.PresenceUpdateRequest, now, ttl int64) presence.SessionPresence {
	sp := existing
	sp.SessionID = sessionID
	sp.Status = update.Status
	sp.Platform = update.Platform
	sp.Since = now
	sp.UpdatedAt = now
	sp.ExpiresAt = now + ttl
	sp.CustomStatusText = update.CustomStatusText

	if update.VoiceChannelID != nil {
		if *update.VoiceChannelID > 0 {
			voiceID := *update.VoiceChannelID
			sp.VoiceChannelID = &voiceID
		} else {
			sp.VoiceChannelID = nil
		}
	}
	if update.Mute != nil {
		sp.Mute = *update.Mute
	}
	if update.Deafen != nil {
		sp.Deafen = *update.Deafen
	}
	if update.SelfVideo != nil {
		sp.SelfVideo = *update.SelfVideo
	}

	return sp
}

func resolveRequestedChannels(m mqmsg.Subscribe) ([]int64, bool) {
	if m.Channels != nil {
		return normalizeChannelIDs(m.Channels), true
	}
	if m.Channel != nil {
		if *m.Channel <= 0 {
			return nil, false
		}
		return []int64{*m.Channel}, true
	}
	return nil, false
}

func normalizeChannelIDs(channelIDs []int64) []int64 {
	if len(channelIDs) == 0 {
		return []int64{}
	}

	set := make(map[int64]struct{}, len(channelIDs))
	result := make([]int64, 0, len(channelIDs))
	for _, channelID := range channelIDs {
		if channelID <= 0 {
			continue
		}
		if _, ok := set[channelID]; ok {
			continue
		}
		set[channelID] = struct{}{}
		result = append(result, channelID)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

func buildChannelSubscriptionDiff(current map[int64]struct{}, requested []int64) (subscribe []int64, unsubscribe []int64) {
	requestedSet := make(map[int64]struct{}, len(requested))
	for _, channelID := range requested {
		if _, ok := requestedSet[channelID]; ok {
			continue
		}
		requestedSet[channelID] = struct{}{}
		if _, ok := current[channelID]; !ok {
			subscribe = append(subscribe, channelID)
		}
	}
	for channelID := range current {
		if _, ok := requestedSet[channelID]; !ok {
			unsubscribe = append(unsubscribe, channelID)
		}
	}
	sort.Slice(subscribe, func(i, j int) bool {
		return subscribe[i] < subscribe[j]
	})
	sort.Slice(unsubscribe, func(i, j int) bool {
		return unsubscribe[i] < unsubscribe[j]
	})
	return subscribe, unsubscribe
}

func channelSubscriptionKey(channelID int64) string {
	return fmt.Sprintf("channel.%d", channelID)
}

func (h *Handler) syncChannelSubscriptions(ctx context.Context, requested []int64) {
	allowed := make([]int64, 0, len(requested))
	for _, channelID := range requested {
		if h.canSubscribeChannel(ctx, channelID) {
			allowed = append(allowed, channelID)
		}
	}

	toSubscribe, toUnsubscribe := buildChannelSubscriptionDiff(h.csubs, allowed)
	for _, channelID := range toSubscribe {
		key := channelSubscriptionKey(channelID)
		if err := h.sub.Subscribe(ctx, key, key); err != nil {
			h.log.Warn("Error subscribing to channel", "error", err, "channel_id", channelID)
			continue
		}
		h.csubs[channelID] = struct{}{}
	}
	for _, channelID := range toUnsubscribe {
		_ = h.sub.Unsubscribe(ctx, channelSubscriptionKey(channelID))
		delete(h.csubs, channelID)
	}
	h.persistClientState(ctx)
}

func (h *Handler) canSubscribeChannel(ctx context.Context, channelID int64) bool {
	log := helper.WithContext(h.log, ctx)
	if gcinfo, err := h.gc.GetGuildByChannel(ctx, channelID); err == nil {
		_, _, _, ok, perr := h.perm.ChannelPerm(ctx, gcinfo.GuildId, gcinfo.ChannelId, h.user.Id, permissions.PermServerViewChannels)
		if perr != nil {
			log.Warn("Error checking channel permissions", "error", perr, "channel_id", channelID)
			return false
		}
		if ok {
			return true
		}
	}

	if ok, err := h.dm.IsDmChannelParticipant(ctx, channelID, h.user.Id); err == nil && ok {
		return true
	} else if err != nil {
		log.Warn("Error checking DM participation", "error", err, "channel_id", channelID)
	}

	if ok, err := h.gdm.IsGroupDmParticipant(ctx, channelID, h.user.Id); err == nil && ok {
		return true
	} else if err != nil {
		log.Warn("Error checking Group DM participation", "error", err, "channel_id", channelID)
	}

	log.Warn("User does not have permission/access to channel", "user_id", h.user.Id, "channel_id", channelID)
	return false
}

func (h *Handler) Close() error {
	h.OnWSClosed()
	h.closer()
	return nil
}

func (h *Handler) OnWSClosed() {
	if h.user != nil && h.gstate != nil && h.connectionID != "" {
		ctx, cancel := context.WithTimeout(h.baseContext(), time.Second)
		_ = h.gstate.DeleteConnection(ctx, h.connectionID, h.user.Id)
		cancel()
	}
	if h.user == nil || h.pstore == nil || h.nats == nil || !h.presenceSet || h.sessionID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(h.baseContext(), time.Second*2)
	defer cancel()
	// Read previous aggregated presence
	prev, _, _ := h.pstore.Get(ctx, h.user.Id)
	// Remove this session
	ttl := h.presenceTTLSeconds()
	removed := true
	if h.connectionID != "" && h.generation > 0 {
		var err error
		removed, err = h.pstore.RemoveSessionIfOwner(ctx, h.user.Id, h.sessionID, h.connectionID, h.generation, ttl)
		if err != nil {
			h.log.Warn("Error removing owned presence session", "error", err, "user_id", h.user.Id, "session_id", h.sessionID)
			return
		}
	} else {
		_ = h.pstore.RemoveSession(ctx, h.user.Id, h.sessionID, ttl)
	}
	if !removed {
		return
	}
	// Re-aggregate
	now := time.Now().Unix()
	agg, _, _ := h.pstore.Aggregate(ctx, h.user.Id, now)
	if presenceChanged(prev, agg) {
		_ = h.pstore.SetAggregated(ctx, agg, ttl)
		h.publishPresence(agg)
	}
}

func (h *Handler) sendPresenceSnapshot(ctx context.Context, userID int64) {
	// Read presence from cache and send to this connection only
	if h.pstore == nil {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	p, ok, _ := h.pstore.Get(ctx, userID)
	status := presence.StatusOffline
	since := time.Now().Unix()
	text := ""
	var voiceID *int64
	if ok {
		status = p.Status
		since = p.Since
		text = p.CustomStatusText
	}
	var mute, deafen, selfVideo bool
	var activeStream = p.ActiveStream
	if ok {
		mute = p.Mute
		deafen = p.Deafen
		selfVideo = p.SelfVideo
	}
	// include voice channel id if present
	if ok && p.VoiceChannelID != nil {
		vid := *p.VoiceChannelID
		voiceID = &vid
	}
	if status == presence.StatusOffline && voiceID == nil {
		clearVoice := int64(0)
		voiceID = &clearVoice
	}
	msg, err := mqmsg.BuildEventMessage(&mqmsg.PresenceUpdate{
		UserID:           userID,
		Status:           status,
		Since:            since,
		CustomStatusText: text,
		ClientStatus:     p.ClientStatus,
		VoiceChannelID:   voiceID,
		Mute:             mute,
		Deafen:           deafen,
		SelfVideo:        selfVideo,
		ActiveStream:     activeStream,
	})
	if err != nil {
		return
	}
	_ = h.sendJSON(msg)
}

func (h *Handler) publishPresence(agg presence.Presence) {
	_ = presence.Publish(h.baseContext(), h.nats, agg)
}

func (h *Handler) presenceTTLSeconds() int64 {
	ttl := h.hbTimeout * 2 / 1000
	if ttl < 1 {
		return 1
	}
	return ttl
}

func (h *Handler) connectionTTLSeconds() int64 {
	ttl := (h.hbTimeout + 10000) * 2 / 1000
	if ttl < 1 {
		return 1
	}
	return ttl
}

func (h *Handler) clientStateTTLSeconds() int64 {
	return int64((24 * time.Hour) / time.Second)
}

func (h *Handler) sendHeartbeatAck(lastEventID int64) error {
	msg, err := mqmsg.BuildEventMessage(&mqmsg.HeartbeatAck{
		LastEventID:  lastEventID,
		ServerTime:   time.Now().UnixMilli(),
		ConnectionID: h.connectionID,
	})
	if err != nil {
		return err
	}
	return h.sendJSON(msg)
}

func (h *Handler) persistClientState(ctx context.Context) {
	if h.gstate == nil || h.user == nil || h.clientInstanceID == "" {
		return
	}
	channels := make([]int64, 0, len(h.csubs))
	for channelID := range h.csubs {
		channels = append(channels, channelID)
	}
	sort.Slice(channels, func(i, j int) bool { return channels[i] < channels[j] })

	presenceSet := make([]int64, 0, len(h.psubs))
	for userID := range h.psubs {
		if h.isAutoPresenceSubscription(userID) {
			continue
		}
		presenceSet = append(presenceSet, userID)
	}
	sort.Slice(presenceSet, func(i, j int) bool { return presenceSet[i] < presenceSet[j] })

	opCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	_ = h.gstate.SetClientState(opCtx, h.user.Id, h.clientInstanceID, gatewaystate.ClientState{
		Channels:    channels,
		PresenceSet: presenceSet,
	}, h.clientStateTTLSeconds())
}

func (h *Handler) isAutoPresenceSubscription(userID int64) bool {
	_, ok := h.autoPsubs[userID]
	return ok
}

func presenceChanged(prev, next presence.Presence) bool {
	return prev.Status != next.Status ||
		prev.CustomStatusText != next.CustomStatusText ||
		!reflect.DeepEqual(prev.ClientStatus, next.ClientStatus) ||
		!sameInt64Ptr(prev.VoiceChannelID, next.VoiceChannelID) ||
		prev.Mute != next.Mute ||
		prev.Deafen != next.Deafen ||
		prev.SelfVideo != next.SelfVideo ||
		!reflect.DeepEqual(prev.ActiveStream, next.ActiveStream)
}

func sameInt64Ptr(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func (h *Handler) baseContext() context.Context {
	ctx := observability.BackgroundFromContext(h.ctx)
	if h.user != nil {
		ctx = helper.ContextWithUserID(ctx, h.user.Id)
	}
	return ctx
}

func (h *Handler) startMessageSpan(e mqmsg.Message) (context.Context, trace.Span) {
	ctx := h.baseContext()
	attrs := []attribute.KeyValue{
		attribute.Int("message.operation", int(e.Operation)),
	}
	if e.EventType != nil {
		attrs = append(attrs, attribute.Int("message.event_type", int(*e.EventType)))
	}
	ctx, span := observability.Tracer("gochat/ws.handler").Start(ctx, "ws.message")
	span.SetAttributes(attrs...)
	span.SetStatus(codes.Ok, "handled")
	return ctx, span
}

func messageEventType(eventType *mqmsg.EventType) *int {
	if eventType == nil {
		return nil
	}
	value := int(*eventType)
	return &value
}
