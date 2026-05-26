package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	crand "crypto/rand"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/gatewaystate"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/userbootstrap"

	pgmodel "github.com/FlameInTheDark/gochat/internal/database/model"
)

type gatewayReadySession struct {
	SessionID        string `json:"session_id"`
	ConnectionID     string `json:"connection_id"`
	ClientInstanceID string `json:"client_instance_id,omitempty"`
	Generation       int64  `json:"generation"`
	ProtocolVersion  int    `json:"protocol_version"`
}

type gatewayReadyAutoSubscriptions struct {
	Guilds   []int64 `json:"guilds"`
	Friends  []int64 `json:"friends"`
	Presence []int64 `json:"presence"`
}

type gatewayReadyPayload struct {
	User                dto.User                      `json:"user"`
	Settings            *pgmodel.UserSettingsData     `json:"settings,omitempty"`
	SettingsVersion     int64                         `json:"settings_version,omitempty"`
	ReadStates          map[int64]int64               `json:"read_states"`
	GuildsLastMessages  map[int64]map[int64]int64     `json:"guilds_last_messages"`
	ThreadsLastMessages map[int64]int64               `json:"threads_last_messages"`
	JoinedThreads       map[int64]map[int64][]int64   `json:"joined_threads"`
	DMCalls             []mqmsg.DMCallSummary         `json:"dm_calls"`
	Guilds              []dto.Guild                   `json:"guilds"`
	Friends             []int64                       `json:"friends"`
	FriendRequests      []int64                       `json:"friend_requests"`
	ServerTime          int64                         `json:"server_time"`
	Session             gatewayReadySession           `json:"session"`
	AutoSubscriptions   gatewayReadyAutoSubscriptions `json:"auto_subscriptions"`
}

type gatewayDMCallState struct {
	CallID       int64           `json:"call_id"`
	ChannelID    int64           `json:"channel_id"`
	CallerID     int64           `json:"caller_id"`
	RecipientID  int64           `json:"recipient_id"`
	Region       string          `json:"region,omitempty"`
	Participants map[int64]int64 `json:"participants,omitempty"`
	DismissedBy  map[int64]bool  `json:"dismissed_by,omitempty"`
	StartedAt    int64           `json:"started_at"`
	EndedAt      int64           `json:"ended_at,omitempty"`
	SoloSince    int64           `json:"solo_since,omitempty"`
}

func (h *Handler) hello(msg *mqmsg.Message) {
	ctx := h.baseContext()
	log := helper.WithContext(h.log, ctx)
	var m helloMessage
	err := json.Unmarshal(msg.Data, &m)
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		log.Error("Error unmarshalling hello message", "error", err)
		return
	}
	token, err := h.jwt.ParseAccess(ctx, m.Token)
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.telemetry.AuthFailure(ctx, "invalid_token")
		log.Error("Error parsing token", "error", err)
		return
	}

	// --- Parallel DB fetch: user + guilds ---
	ctx, cancel := context.WithTimeout(helper.ContextWithUserID(ctx, token.UserID), time.Second*time.Duration(h.hbTimeout))
	defer cancel()

	type userResult struct {
		user pgmodel.User
		err  error
	}
	type guildsResult struct {
		guilds []pgmodel.UserGuild
		err    error
	}
	userCh := make(chan userResult, 1)
	guildsCh := make(chan guildsResult, 1)

	go func() {
		u, e := h.u.GetUserById(ctx, token.UserID)
		userCh <- userResult{u, e}
	}()
	go func() {
		g, e := h.m.GetUserGuilds(ctx, token.UserID)
		guildsCh <- guildsResult{g, e}
	}()

	ur := <-userCh
	gr := <-guildsCh

	if ur.err != nil {
		h.initTimer.Stop()
		h.closer()
		h.telemetry.AuthFailure(ctx, "user_lookup")
		log.Error("Error getting user", "error", ur.err)
		return
	}
	if gr.err != nil {
		h.initTimer.Stop()
		h.closer()
		log.Error("Error getting user's guilds", "error", gr.err)
		return
	}

	h.initTimer.Stop()

	h.user = &dto.User{
		Id:          ur.user.Id,
		Name:        ur.user.Name,
		Bio:         ur.user.Bio,
		BannerColor: ur.user.BannerColor,
		PanelColor:  ur.user.PanelColor,
	}
	if h.onAuthenticated != nil {
		h.onAuthenticated(token.UserID)
	}

	// Establish or resume the logical session. v2 clients use resume_session_id
	// with an incremented generation; legacy clients can still pass heartbeat_session_id.
	switch {
	case m.ResumeSessionID != "":
		h.sessionID = m.ResumeSessionID
	case m.HeartbeatSessionID != "":
		h.sessionID = m.HeartbeatSessionID
	default:
		h.sessionID = newSessionID()
	}
	h.connectionID = newSessionID()
	h.clientInstanceID = m.ClientInstanceID
	if h.clientInstanceID == "" {
		h.clientInstanceID = h.connectionID
	}
	h.generation = m.ResumeGeneration + 1
	if h.generation <= 0 {
		h.generation = time.Now().UnixNano()
	}

	if h.gstate != nil {
		opCtx, cancel := context.WithTimeout(ctx, time.Second*2)
		_ = h.gstate.UpsertConnection(opCtx, gatewaystate.ConnectionState{
			ConnectionID:     h.connectionID,
			UserID:           token.UserID,
			SessionID:        h.sessionID,
			ClientInstanceID: h.clientInstanceID,
			Generation:       h.generation,
			ProtocolVersion:  2,
			LastSeq:          m.LastSeq,
		}, h.connectionTTLSeconds())
		cancel()
	}

	// Do not auto-set presence here. Presence is set only after client sends PresenceUpdate.
	hellomsg, err := mqmsg.BuildEventMessage(&mqmsg.HeartbeatInterval{
		HeartbeatInterval: h.hbTimeout,
		SessionID:         h.sessionID,
		ConnectionID:      h.connectionID,
		Generation:        h.generation,
		ProtocolVersion:   2,
	})
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		return
	}
	err = h.sendJSON(hellomsg)
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		log.Error("Error sending hello message", "error", err)
		return
	}
	h.hTimer = time.AfterFunc(time.Millisecond*time.Duration(h.hbTimeout+10000), func() {
		timeoutCtx := h.baseContext()
		h.telemetry.HeartbeatTimeout(timeoutCtx)
		helper.WithContext(h.log, timeoutCtx).Warn("Heartbeat timeout; closing WS", "user_id", func() any {
			if h.user != nil {
				return h.user.Id
			}
			return int64(0)
		}())
		err := h.Close()
		if err != nil {
			helper.WithContext(h.log, timeoutCtx).Error("Error closing WS connection after timeout", "error", err)
		}
	})

	// Subscribe to personal user topic
	err = h.sub.Subscribe(ctx, "user", fmt.Sprintf("user.%d", token.UserID))
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		log.Error("Error subscribing to user", "error", err)
		return
	}

	// Subscribe to all guilds (hub registrations are fast in-memory ops)
	for _, g := range gr.guilds {
		if err := h.sub.Subscribe(ctx, fmt.Sprintf("guild.%d", g.GuildId), fmt.Sprintf("guild.%d", g.GuildId)); err != nil {
			log.Warn("Error subscribing to guild", "error", err, "guild_id", g.GuildId)
		}
	}

	friendIDs, friendRequestIDs := h.subscribeFriendPresence(ctx, token.UserID)
	h.restoreClientSubscriptions(ctx)
	if err := h.sendGatewayReady(ctx, ur.user, gr.guilds, friendIDs, friendRequestIDs); err != nil {
		log.Warn("Error sending gateway ready message", "error", err)
	}
}

func (h *Handler) subscribeFriendPresence(ctx context.Context, userID int64) ([]int64, []int64) {
	if h.fr == nil {
		return nil, nil
	}

	friends, err := h.fr.GetFriends(ctx, userID)
	if err != nil {
		helper.WithContext(h.log, ctx).Warn("Error loading friends for auto presence subscriptions", "error", err)
		return nil, nil
	}
	friendIDs := make([]int64, 0, len(friends))
	for _, f := range friends {
		if f.FriendID == 0 {
			continue
		}
		friendIDs = append(friendIDs, f.FriendID)
		key := fmt.Sprintf("presence.%d", f.FriendID)
		if err := h.sub.Subscribe(ctx, key, fmt.Sprintf("presence.user.%d", f.FriendID)); err != nil {
			helper.WithContext(h.log, ctx).Warn("Error subscribing to friend presence", "error", err, "friend_id", f.FriendID)
			continue
		}
		h.psubs[f.FriendID] = struct{}{}
		h.autoPsubs[f.FriendID] = struct{}{}
		h.sendPresenceSnapshot(ctx, f.FriendID)
	}

	requests, err := h.fr.GetFriendRequests(ctx, userID)
	if err != nil {
		helper.WithContext(h.log, ctx).Warn("Error loading friend requests for gateway ready", "error", err)
		return friendIDs, nil
	}
	requestIDs := make([]int64, 0, len(requests))
	for _, r := range requests {
		if r.FriendId != 0 {
			requestIDs = append(requestIDs, r.FriendId)
		}
	}
	return friendIDs, requestIDs
}

func (h *Handler) restoreClientSubscriptions(ctx context.Context) {
	if h.gstate == nil || h.user == nil || h.clientInstanceID == "" {
		return
	}
	state, ok, err := h.gstate.GetClientState(ctx, h.user.Id, h.clientInstanceID)
	if err != nil || !ok {
		return
	}
	if state.Channels != nil {
		h.syncChannelSubscriptions(ctx, state.Channels)
	}
	if len(state.PresenceSet) > 0 {
		for _, uid := range state.PresenceSet {
			if _, exists := h.psubs[uid]; exists {
				continue
			}
			key := fmt.Sprintf("presence.%d", uid)
			if err := h.sub.Subscribe(ctx, key, fmt.Sprintf("presence.user.%d", uid)); err != nil {
				helper.WithContext(h.log, ctx).Warn("Error restoring presence subscription", "error", err, "user_id", uid)
				continue
			}
			h.psubs[uid] = struct{}{}
			h.sendPresenceSnapshot(ctx, uid)
		}
	}
}

func (h *Handler) sendGatewayReady(ctx context.Context, user pgmodel.User, memberships []pgmodel.UserGuild, friendIDs, friendRequestIDs []int64) error {
	guildIDs := make([]int64, 0, len(memberships))
	for _, membership := range memberships {
		if membership.GuildId != 0 {
			guildIDs = append(guildIDs, membership.GuildId)
		}
	}

	guildDTOs := make([]dto.Guild, 0, len(guildIDs))
	if len(guildIDs) > 0 {
		guilds, err := h.g.GetGuildsList(ctx, guildIDs)
		if err != nil {
			return fmt.Errorf("load gateway ready guilds: %w", err)
		}
		guildDTOs = guildModelsToDTOs(guilds)
	}

	var settings *pgmodel.UserSettingsData
	var settingsVersion int64
	if h.uset != nil {
		if stored, err := h.uset.GetUserSettings(ctx, user.Id, 0); err == nil && stored.UserId != 0 {
			if data, err := pgmodel.UnmarshalStoredUserSettingsData(stored.Settings); err == nil {
				settings = &data
				settingsVersion = stored.Version
			}
		}
	}

	readStates, guildsLastMessages, threadsLastMessages, joinedThreads := h.loadGatewayReadyReadState(ctx, user.Id, guildIDs)

	ready := gatewayReadyPayload{
		User:                userModelToDTO(user),
		Settings:            settings,
		SettingsVersion:     settingsVersion,
		ReadStates:          readStates,
		GuildsLastMessages:  guildsLastMessages,
		ThreadsLastMessages: threadsLastMessages,
		JoinedThreads:       joinedThreads,
		DMCalls:             h.activeDMCallSummaries(ctx, user.Id),
		Guilds:              guildDTOs,
		Friends:             friendIDs,
		FriendRequests:      friendRequestIDs,
		ServerTime:          time.Now().UnixMilli(),
		Session: gatewayReadySession{
			SessionID:        h.sessionID,
			ConnectionID:     h.connectionID,
			ClientInstanceID: h.clientInstanceID,
			Generation:       h.generation,
			ProtocolVersion:  2,
		},
		AutoSubscriptions: gatewayReadyAutoSubscriptions{
			Guilds:   guildIDs,
			Friends:  friendIDs,
			Presence: friendIDs,
		},
	}

	body, err := json.Marshal(ready)
	if err != nil {
		return err
	}
	t := mqmsg.EventTypeGatewayReady
	return h.sendJSON(mqmsg.Message{
		Operation: mqmsg.OpCodeDispatch,
		EventType: &t,
		Data:      body,
	})
}

func (h *Handler) activeDMCallSummaries(ctx context.Context, userID int64) []mqmsg.DMCallSummary {
	if h.cache == nil || userID == 0 {
		return []mqmsg.DMCallSummary{}
	}
	index, err := h.cache.HGetAll(ctx, fmt.Sprintf("dmcall:user:%d", userID))
	if err != nil || len(index) == 0 {
		return []mqmsg.DMCallSummary{}
	}

	out := make([]mqmsg.DMCallSummary, 0, len(index))
	for channelID := range index {
		id, err := strconv.ParseInt(channelID, 10, 64)
		if err != nil || id == 0 {
			continue
		}
		var call gatewayDMCallState
		if err := h.cache.GetJSON(ctx, fmt.Sprintf("dmcall:state:%d", id), &call); err != nil {
			continue
		}
		if call.CallID == 0 || call.ChannelID == 0 || call.EndedAt != 0 {
			continue
		}
		out = append(out, mqmsg.DMCallSummary{
			CallID:       call.CallID,
			ChannelID:    call.ChannelID,
			CallerID:     call.CallerID,
			RecipientID:  call.RecipientID,
			Region:       call.Region,
			Participants: call.Participants,
			StartedAt:    call.StartedAt,
			SoloSince:    call.SoloSince,
			Dismissed:    call.DismissedBy != nil && call.DismissedBy[userID],
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].StartedAt == out[j].StartedAt {
			return out[i].ChannelID < out[j].ChannelID
		}
		return out[i].StartedAt > out[j].StartedAt
	})
	return out
}

func userModelToDTO(u pgmodel.User) dto.User {
	return dto.User{
		Id:          u.Id,
		Name:        u.Name,
		Bio:         u.Bio,
		BannerColor: u.BannerColor,
		PanelColor:  u.PanelColor,
		IsBot:       u.IsBot(),
	}
}

func (h *Handler) loadGatewayReadyReadState(ctx context.Context, userID int64, guildIDs []int64) (map[int64]int64, map[int64]map[int64]int64, map[int64]int64, map[int64]map[int64][]int64) {
	log := helper.WithContext(h.log, ctx)

	readStates := map[int64]int64{}
	if h.rs != nil {
		rs, err := h.rs.GetReadStates(ctx, userID)
		if err != nil {
			log.Warn("Error loading read states for gateway ready", "error", err)
		} else if rs != nil {
			readStates = rs
		}
	}

	guildsLastMessages := map[int64]map[int64]int64{}
	if h.gclm == nil || len(guildIDs) == 0 {
		return readStates, guildsLastMessages, map[int64]int64{}, map[int64]map[int64][]int64{}
	}

	rawGuildLastMessages, err := h.gclm.GetChannelsMessagesForGuilds(ctx, guildIDs)
	if err != nil {
		log.Warn("Error loading guild last messages for gateway ready", "error", err)
		return readStates, guildsLastMessages, map[int64]int64{}, map[int64]map[int64][]int64{}
	}
	if len(rawGuildLastMessages) == 0 {
		return readStates, guildsLastMessages, map[int64]int64{}, map[int64]map[int64][]int64{}
	}

	if h.gc == nil || h.ch == nil || h.tm == nil {
		return readStates, userbootstrap.FilterGuildLastMessages(rawGuildLastMessages, nil), map[int64]int64{}, map[int64]map[int64][]int64{}
	}

	channelIDs, err := h.gc.GetGuildsChannelsIDsMany(ctx, guildIDs)
	if err != nil {
		log.Warn("Error loading guild channel IDs for gateway ready", "error", err)
		return readStates, guildsLastMessages, map[int64]int64{}, map[int64]map[int64][]int64{}
	}
	if len(channelIDs) == 0 {
		return readStates, guildsLastMessages, map[int64]int64{}, map[int64]map[int64][]int64{}
	}

	channels, err := h.ch.GetChannelsBulk(ctx, channelIDs)
	if err != nil {
		log.Warn("Error loading channels for gateway ready read-state filter", "error", err)
		return readStates, guildsLastMessages, map[int64]int64{}, map[int64]map[int64][]int64{}
	}
	guildsLastMessages = userbootstrap.FilterGuildLastMessages(rawGuildLastMessages, channels)

	threadMembers, err := h.tm.GetUserThreadMembers(ctx, userID)
	if err != nil {
		log.Warn("Error loading joined threads for gateway ready", "error", err)
		return readStates, guildsLastMessages, map[int64]int64{}, map[int64]map[int64][]int64{}
	}
	joinedThreadSet := userbootstrap.BuildJoinedThreadSet(threadMembers)
	if len(joinedThreadSet) == 0 {
		return readStates, guildsLastMessages, map[int64]int64{}, map[int64]map[int64][]int64{}
	}

	guildChannels, err := h.gc.GetGuildChannelsByChannelIDs(ctx, channelIDs)
	if err != nil {
		log.Warn("Error loading guild channel rows for gateway ready joined threads", "error", err)
		return readStates, guildsLastMessages, map[int64]int64{}, map[int64]map[int64][]int64{}
	}

	threadsLastMessages := userbootstrap.FilterThreadLastMessages(joinedThreadSet, channels, rawGuildLastMessages)
	joinedThreads := userbootstrap.BuildJoinedThreads(joinedThreadSet, channels, guildChannels)
	return readStates, guildsLastMessages, threadsLastMessages, joinedThreads
}

func guildModelsToDTOs(guilds []pgmodel.Guild) []dto.Guild {
	out := make([]dto.Guild, len(guilds))
	for i, g := range guilds {
		out[i] = dto.Guild{
			Id:              g.Id,
			Name:            g.Name,
			Owner:           g.OwnerId,
			Public:          g.Public,
			Permissions:     g.Permissions,
			SystemChannelId: g.SystemMessages,
		}
	}
	return out
}

// newSessionID generates a random UUIDv4-like string without external deps.
func newSessionID() string {
	var b [16]byte
	if n, err := randRead(b[:]); err == nil && n == len(b) {
		b[6] = (b[6] & 0x0f) | 0x40
		b[8] = (b[8] & 0x3f) | 0x80
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]),
			uint16(b[4])<<8|uint16(b[5]),
			uint16(b[6])<<8|uint16(b[7]),
			uint16(b[8])<<8|uint16(b[9]),
			uint64(b[10])<<40|uint64(b[11])<<32|uint64(b[12])<<24|uint64(b[13])<<16|uint64(b[14])<<8|uint64(b[15]),
		)
	}

	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// indirection to avoid importing crypto/rand in multiple places
var randRead = func(p []byte) (int, error) {
	return crand.Read(p)
}
