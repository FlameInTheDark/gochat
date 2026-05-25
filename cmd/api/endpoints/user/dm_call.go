package user

import (
	"context"
	"hash/fnv"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	cachepkg "github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
)

const (
	dmCallTTLSeconds          = int64(60 * 60 * 6)
	dmCallRouteTTLSeconds     = int64(180)
	dmCallSoloGraceSeconds    = int64(180)
	dmCallStreamStateTTL      = int64(180)
	dmCallStreamRouteTTL      = int64(180)
	dmCallStreamTokenTTL      = time.Minute
	dmCallUserIndexTTLSeconds = int64(60 * 60 * 6)
	dmCallPreferredRegionAuto = "auto"
)

type dmVoiceRouteBinding struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Region string `json:"region,omitempty"`
}

type dmCallState struct {
	CallID       int64           `json:"call_id"`
	ChannelID    int64           `json:"channel_id"`
	CallerID     int64           `json:"caller_id"`
	RecipientID  int64           `json:"recipient_id"`
	Region       string          `json:"region,omitempty"`
	RouteID      string          `json:"route_id,omitempty"`
	RouteURL     string          `json:"route_url,omitempty"`
	Participants map[int64]int64 `json:"participants,omitempty"`
	DismissedBy  map[int64]bool  `json:"dismissed_by,omitempty"`
	StartedAt    int64           `json:"started_at"`
	UpdatedAt    int64           `json:"updated_at"`
	SoloSince    int64           `json:"solo_since,omitempty"`
	EndedAt      int64           `json:"ended_at,omitempty"`
}

type dmVoiceSelector struct {
	mu           sync.Mutex
	reservations map[string][]time.Time
}

func newDMVoiceSelector() *dmVoiceSelector {
	return &dmVoiceSelector{reservations: make(map[string][]time.Time)}
}

func dmCallStateKey(channelID int64) string {
	return "dmcall:state:" + strconv.FormatInt(channelID, 10)
}
func dmCallRouteKey(channelID int64) string {
	return "dmcall:route:" + strconv.FormatInt(channelID, 10)
}
func dmCallUserIndexKey(userID int64) string { return "dmcall:user:" + strconv.FormatInt(userID, 10) }

func voiceClientsKey(channelID int64) string {
	return "voice:clients:" + strconv.FormatInt(channelID, 10)
}

func voiceRouteKey(channelID int64) string {
	return "voice:route:" + strconv.FormatInt(channelID, 10)
}

func voiceRebindKey(channelID int64) string {
	return "voice:rebind:" + strconv.FormatInt(channelID, 10)
}

func (e *entity) parseDMCallChannelID(c *fiber.Ctx) (int64, error) {
	channelID, err := strconv.ParseInt(c.Params("channel_id"), 10, 64)
	if err != nil || channelID <= 0 {
		return 0, fiber.NewError(http.StatusBadRequest, ErrChannelIdInvalid)
	}
	return channelID, nil
}

func (e *entity) parseDMCallStreamID(c *fiber.Ctx) (int64, error) {
	streamID, err := strconv.ParseInt(c.Params("stream_id"), 10, 64)
	if err != nil || streamID <= 0 {
		return 0, fiber.NewError(http.StatusBadRequest, "invalid stream id")
	}
	return streamID, nil
}

func (e *entity) directDMParticipants(ctx context.Context, channelID, userID int64) (int64, int64, error) {
	rows, err := e.dm.GetDmChannelByChannelId(ctx, channelID)
	if err != nil {
		return 0, 0, err
	}
	if len(rows) != 2 {
		return 0, 0, fiber.NewError(fiber.StatusBadRequest, "dm calls are only available in direct messages")
	}
	var peerID int64
	seen := false
	for _, row := range rows {
		if row.UserId == userID {
			seen = true
			peerID = row.ParticipantId
		}
	}
	if !seen || peerID == 0 {
		return 0, 0, fiber.NewError(fiber.StatusForbidden, "not a dm participant")
	}
	return userID, peerID, nil
}

func (e *entity) userPreferredVoiceRegion(ctx context.Context, userID int64) string {
	if e == nil || e.uset == nil {
		return ""
	}
	settings, err := e.loadStoredUserSettings(ctx, userID)
	if err != nil {
		return ""
	}
	region := strings.TrimSpace(settings.Voice.PreferredRegion)
	if region == "" || region == dmCallPreferredRegionAuto {
		return ""
	}
	if len(e.allowedRegions) > 0 {
		if _, ok := e.allowedRegions[region]; !ok {
			return ""
		}
	}
	return region
}

func (e *entity) dmSelectionRegions(ctx context.Context, preferredRegion string, manager discovery.Manager, allowFallback bool) ([]string, error) {
	seen := make(map[string]struct{}, len(e.allowedRegionIDs)+1)
	regions := make([]string, 0, len(e.allowedRegionIDs)+1)
	add := func(region string) {
		if region == "" {
			return
		}
		if _, ok := seen[region]; ok {
			return
		}
		seen[region] = struct{}{}
		regions = append(regions, region)
	}
	add(preferredRegion)
	if !allowFallback {
		return regions, nil
	}
	for _, region := range e.allowedRegionIDs {
		add(region)
	}
	if len(regions) > 0 {
		return regions, nil
	}
	if manager == nil {
		add(e.defaultVoiceRegion)
		return regions, nil
	}
	discovered, err := manager.Regions(ctx)
	if err != nil {
		if len(regions) > 0 {
			return regions, nil
		}
		return nil, err
	}
	sort.Strings(discovered)
	for _, region := range discovered {
		add(region)
	}
	return regions, nil
}

func (s *dmVoiceSelector) pick(roomID int64, region string, instances []discovery.Instance) (dmVoiceRouteBinding, bool) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	var chosen discovery.Instance
	var chosenSet bool
	var bestScore int64
	var bestHash uint64
	for _, inst := range instances {
		if inst.ID == "" || inst.URL == "" {
			continue
		}
		pending := s.pendingLocked(inst.ID, now)
		score := inst.Load + int64(pending)*8
		hash := dmRendezvousScore(roomID, inst.ID)
		if !chosenSet || score < bestScore || (score == bestScore && hash > bestHash) {
			chosen = inst
			chosenSet = true
			bestScore = score
			bestHash = hash
		}
	}
	if !chosenSet {
		return dmVoiceRouteBinding{}, false
	}
	s.reservations[chosen.ID] = append(s.reservations[chosen.ID], now.Add(10*time.Second))
	return dmVoiceRouteBinding{ID: chosen.ID, URL: chosen.URL, Region: region}, true
}

func (s *dmVoiceSelector) pendingLocked(instanceID string, now time.Time) int {
	expiries := s.reservations[instanceID]
	if len(expiries) == 0 {
		return 0
	}
	alive := expiries[:0]
	for _, expiry := range expiries {
		if expiry.After(now) {
			alive = append(alive, expiry)
		}
	}
	if len(alive) == 0 {
		delete(s.reservations, instanceID)
		return 0
	}
	s.reservations[instanceID] = alive
	return len(alive)
}

func dmRendezvousScore(roomID int64, instanceID string) uint64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strconv.FormatInt(roomID, 10)))
	_, _ = hasher.Write([]byte(instanceID))
	return hasher.Sum64()
}

func (e *entity) selectDMBinding(ctx context.Context, roomID int64, preferredRegion string, manager discovery.Manager, selector *dmVoiceSelector, allowFallback bool) (dmVoiceRouteBinding, error) {
	if manager == nil || selector == nil {
		return dmVoiceRouteBinding{}, nil
	}
	regions, err := e.dmSelectionRegions(ctx, preferredRegion, manager, allowFallback)
	if err != nil {
		return dmVoiceRouteBinding{}, err
	}
	var firstErr error
	for _, region := range regions {
		instances, err := manager.List(ctx, region)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if binding, ok := selector.pick(roomID, region, instances); ok {
			return binding, nil
		}
	}
	return dmVoiceRouteBinding{}, firstErr
}

func (e *entity) cachedDMCall(ctx context.Context, channelID int64) (dmCallState, bool) {
	var call dmCallState
	if e.cache == nil {
		return call, false
	}
	if err := e.cache.GetJSON(ctx, dmCallStateKey(channelID), &call); err != nil || call.CallID == 0 || call.EndedAt != 0 {
		return dmCallState{}, false
	}
	return call, true
}

func (e *entity) saveDMCall(ctx context.Context, call dmCallState) error {
	if e.cache == nil {
		return nil
	}
	call.UpdatedAt = time.Now().Unix()
	if call.Participants == nil {
		call.Participants = map[int64]int64{}
	}
	if call.DismissedBy == nil {
		call.DismissedBy = map[int64]bool{}
	}
	if err := e.cache.SetTimedJSON(ctx, dmCallStateKey(call.ChannelID), call, dmCallTTLSeconds, cachepkg.NoneProactive()); err != nil {
		return err
	}
	for _, userID := range []int64{call.CallerID, call.RecipientID} {
		_ = e.cache.HSet(ctx, dmCallUserIndexKey(userID), strconv.FormatInt(call.ChannelID, 10), "1")
		_ = e.cache.SetTTL(ctx, dmCallUserIndexKey(userID), dmCallUserIndexTTLSeconds)
	}
	return nil
}

func (e *entity) dmCallSummaryForUser(call dmCallState, userID int64) mqmsg.DMCallSummary {
	return mqmsg.DMCallSummary{
		CallID:       call.CallID,
		ChannelID:    call.ChannelID,
		CallerID:     call.CallerID,
		RecipientID:  call.RecipientID,
		Region:       call.Region,
		Participants: call.Participants,
		StartedAt:    call.StartedAt,
		SoloSince:    call.SoloSince,
		Dismissed:    call.DismissedBy != nil && call.DismissedBy[userID],
	}
}

func (e *entity) publishDMCall(ctx context.Context, call dmCallState, build func(int64) mqmsg.EventDataMessage) {
	for _, userID := range []int64{call.CallerID, call.RecipientID} {
		_ = mq.SendUserUpdate(ctx, e.mqt, userID, build(userID))
	}
}

func dmCallMediaPermissions() int64 {
	return int64(permissions.CreatePermissions(
		permissions.PermVoiceConnect,
		permissions.PermVoiceSpeak,
		permissions.PermVoiceVideo,
	))
}

func (e *entity) issueDMVoiceToken(userID int64, call dmCallState) (string, error) {
	now := time.Now()
	claims := struct {
		helper.Claims
		ChannelID int64  `json:"channel_id"`
		DMCallID  int64  `json:"dm_call_id"`
		GuildID   *int64 `json:"guild_id,omitempty"`
		Perms     int64  `json:"perms"`
	}{
		Claims: helper.Claims{
			UserID:    userID,
			TokenType: "sfu",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "gochat",
				Audience:  []string{"sfu"},
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Minute)),
			},
		},
		ChannelID: call.ChannelID,
		DMCallID:  call.CallID,
		Perms:     dmCallMediaPermissions(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(e.authSecret))
}

func (e *entity) buildDMCallJoinResponse(userID int64, call dmCallState) (DMCallJoinResponse, error) {
	token, err := e.issueDMVoiceToken(userID, call)
	if err != nil {
		return DMCallJoinResponse{}, err
	}
	return DMCallJoinResponse{
		Call:     e.dmCallSummaryForUser(call, userID),
		SFUURL:   call.RouteURL,
		SFUToken: token,
		Region:   call.Region,
	}, nil
}

func (e *entity) ensureDMCallRoute(ctx context.Context, channelID, userID int64) (dmVoiceRouteBinding, error) {
	if e.cache != nil {
		var cached dmVoiceRouteBinding
		if err := e.cache.GetJSON(ctx, dmCallRouteKey(channelID), &cached); err == nil && cached.URL != "" {
			return cached, nil
		}
	}
	preferredRegion := e.userPreferredVoiceRegion(ctx, userID)
	if preferredRegion == "" {
		preferredRegion = e.defaultVoiceRegion
	}
	binding, err := e.selectDMBinding(ctx, channelID, preferredRegion, e.disco, e.voiceSelector, true)
	if err != nil {
		return dmVoiceRouteBinding{}, err
	}
	if binding.URL == "" {
		return binding, nil
	}
	if e.cache != nil {
		set, err := e.cache.SetTimedJSONNX(ctx, dmCallRouteKey(channelID), binding, dmCallRouteTTLSeconds, cachepkg.NoneProactive())
		if err == nil && !set {
			var winner dmVoiceRouteBinding
			if err := e.cache.GetJSON(ctx, dmCallRouteKey(channelID), &winner); err == nil && winner.URL != "" {
				return winner, nil
			}
		}
	}
	return binding, nil
}

func (e *entity) startOrJoinDMCall(c *fiber.Ctx, joinOnly bool) error {
	channelID, err := e.parseDMCallChannelID(c)
	if err != nil {
		return err
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	callerID, peerID, err := e.directDMParticipants(c.UserContext(), channelID, user.Id)
	if err != nil {
		return err
	}

	call, exists := e.cachedDMCall(c.UserContext(), channelID)
	if !exists {
		if joinOnly {
			return fiber.NewError(fiber.StatusNotFound, "call not found")
		}
		binding, err := e.ensureDMCallRoute(c.UserContext(), channelID, user.Id)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, "voice discovery unavailable")
		}
		if binding.URL == "" {
			return fiber.NewError(fiber.StatusServiceUnavailable, "no sfu available in region")
		}
		now := time.Now().Unix()
		call = dmCallState{
			CallID:       idgen.Next(),
			ChannelID:    channelID,
			CallerID:     callerID,
			RecipientID:  peerID,
			Region:       binding.Region,
			RouteID:      binding.ID,
			RouteURL:     binding.URL,
			Participants: map[int64]int64{user.Id: now},
			DismissedBy:  map[int64]bool{},
			StartedAt:    now,
			UpdatedAt:    now,
		}
		if err := e.saveDMCall(c.UserContext(), call); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "unable to save call")
		}
		e.publishDMCall(c.UserContext(), call, func(uid int64) mqmsg.EventDataMessage {
			return mqmsg.NewDMCallStarted(e.dmCallSummaryForUser(call, uid))
		})
	} else {
		if call.Participants == nil {
			call.Participants = map[int64]int64{}
		}
		call.Participants[user.Id] = time.Now().Unix()
		call.SoloSince = 0
		if call.DismissedBy != nil {
			delete(call.DismissedBy, user.Id)
		}
		if err := e.saveDMCall(c.UserContext(), call); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "unable to save call")
		}
		e.publishDMCall(c.UserContext(), call, func(uid int64) mqmsg.EventDataMessage {
			return mqmsg.NewDMCallJoined(e.dmCallSummaryForUser(call, uid), user.Id)
		})
	}

	resp, err := e.buildDMCallJoinResponse(user.Id, call)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to issue voice token")
	}
	return c.JSON(resp)
}

// StartDMCall
//
//	@Summary	Start or join a direct-message voice call
//	@Produce	json
//	@Tags		User
//	@Param		channel_id	path		int64	true	"Direct DM channel ID"
//	@Success	200			{object}	DMCallJoinResponse
//	@failure	400			{string}	string	"Bad request"
//	@failure	403			{string}	string	"Not a DM participant"
//	@failure	503			{string}	string	"No SFU available"
//	@failure	502			{string}	string	"Voice discovery unavailable"
//	@failure	500			{string}	string	"Internal server error"
//	@Router		/user/me/channels/{channel_id}/call [post]
func (e *entity) StartDMCall(c *fiber.Ctx) error { return e.startOrJoinDMCall(c, false) }

// JoinDMCall
//
//	@Summary	Join an active direct-message voice call
//	@Produce	json
//	@Tags		User
//	@Param		channel_id	path		int64	true	"Direct DM channel ID"
//	@Success	200			{object}	DMCallJoinResponse
//	@failure	400			{string}	string	"Bad request"
//	@failure	403			{string}	string	"Not a DM participant"
//	@failure	404			{string}	string	"Call not found"
//	@failure	500			{string}	string	"Internal server error"
//	@Router		/user/me/channels/{channel_id}/call/join [post]
func (e *entity) JoinDMCall(c *fiber.Ctx) error { return e.startOrJoinDMCall(c, true) }

// DeclineDMCall
//
//	@Summary	Decline or dismiss an incoming direct-message voice call
//	@Produce	json
//	@Tags		User
//	@Param		channel_id	path	int64	true	"Direct DM channel ID"
//	@Success	204
//	@failure	400	{string}	string	"Bad request"
//	@failure	403	{string}	string	"Not a DM participant"
//	@failure	500	{string}	string	"Internal server error"
//	@Router		/user/me/channels/{channel_id}/call/decline [post]
func (e *entity) DeclineDMCall(c *fiber.Ctx) error {
	channelID, err := e.parseDMCallChannelID(c)
	if err != nil {
		return err
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	if _, _, err := e.directDMParticipants(c.UserContext(), channelID, user.Id); err != nil {
		return err
	}
	call, ok := e.cachedDMCall(c.UserContext(), channelID)
	if !ok {
		return c.SendStatus(fiber.StatusNoContent)
	}
	if call.DismissedBy == nil {
		call.DismissedBy = map[int64]bool{}
	}
	call.DismissedBy[user.Id] = true
	if err := e.saveDMCall(c.UserContext(), call); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to save call")
	}
	e.publishDMCall(c.UserContext(), call, func(uid int64) mqmsg.EventDataMessage {
		return mqmsg.NewDMCallDeclined(e.dmCallSummaryForUser(call, uid), user.Id)
	})
	return c.SendStatus(fiber.StatusNoContent)
}

// LeaveDMCall
//
//	@Summary	Leave the current direct-message voice call
//	@Produce	json
//	@Tags		User
//	@Param		channel_id	path	int64	true	"Direct DM channel ID"
//	@Success	204
//	@failure	400	{string}	string	"Bad request"
//	@failure	403	{string}	string	"Not a DM participant"
//	@failure	500	{string}	string	"Internal server error"
//	@Router		/user/me/channels/{channel_id}/call [delete]
func (e *entity) LeaveDMCall(c *fiber.Ctx) error {
	channelID, err := e.parseDMCallChannelID(c)
	if err != nil {
		return err
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	if _, _, err := e.directDMParticipants(c.UserContext(), channelID, user.Id); err != nil {
		return err
	}
	call, ok := e.cachedDMCall(c.UserContext(), channelID)
	if !ok {
		return c.SendStatus(fiber.StatusNoContent)
	}
	delete(call.Participants, user.Id)
	now := time.Now().Unix()
	if len(call.Participants) == 0 {
		call.EndedAt = now
		e.clearDMCall(c.UserContext(), call, "empty")
		return c.SendStatus(fiber.StatusNoContent)
	}
	if len(call.Participants) == 1 && call.SoloSince == 0 {
		call.SoloSince = now
	}
	if err := e.saveDMCall(c.UserContext(), call); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to save call")
	}
	e.publishDMCall(c.UserContext(), call, func(uid int64) mqmsg.EventDataMessage {
		return mqmsg.NewDMCallLeft(e.dmCallSummaryForUser(call, uid), user.Id)
	})
	e.scheduleDMCallSoloCleanup(call.ChannelID, call.SoloSince)
	return c.SendStatus(fiber.StatusNoContent)
}

func (e *entity) clearDMCall(ctx context.Context, call dmCallState, reason string) {
	if e.cache != nil {
		_ = e.cache.Delete(ctx, dmCallStateKey(call.ChannelID))
		_ = e.cache.Delete(ctx, dmCallRouteKey(call.ChannelID))
		_ = e.cache.Delete(ctx, voiceRouteKey(call.ChannelID))
		_ = e.cache.Delete(ctx, voiceRebindKey(call.ChannelID))
		_ = e.cache.Delete(ctx, voiceClientsKey(call.ChannelID))
		_ = e.cache.HDel(ctx, dmCallUserIndexKey(call.CallerID), strconv.FormatInt(call.ChannelID, 10))
		_ = e.cache.HDel(ctx, dmCallUserIndexKey(call.RecipientID), strconv.FormatInt(call.ChannelID, 10))
	}
	e.publishDMCall(ctx, call, func(uid int64) mqmsg.EventDataMessage {
		return mqmsg.NewDMCallEnded(e.dmCallSummaryForUser(call, uid), reason)
	})
}

func (e *entity) scheduleDMCallSoloCleanup(channelID, soloSince int64) {
	if soloSince == 0 {
		return
	}
	go func() {
		time.Sleep(time.Duration(dmCallSoloGraceSeconds) * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		call, ok := e.cachedDMCall(ctx, channelID)
		if !ok || call.SoloSince != soloSince || len(call.Participants) != 1 {
			return
		}
		call.EndedAt = time.Now().Unix()
		e.clearDMCall(ctx, call, "solo_timeout")
	}()
}

func (e *entity) activeDMCallSummaries(ctx context.Context, userID int64) []mqmsg.DMCallSummary {
	if e.cache == nil {
		return []mqmsg.DMCallSummary{}
	}
	index, err := e.cache.HGetAll(ctx, dmCallUserIndexKey(userID))
	if err != nil {
		return []mqmsg.DMCallSummary{}
	}
	out := make([]mqmsg.DMCallSummary, 0, len(index))
	for rawChannelID := range index {
		channelID, err := strconv.ParseInt(rawChannelID, 10, 64)
		if err != nil {
			continue
		}
		call, ok := e.cachedDMCall(ctx, channelID)
		if !ok {
			_ = e.cache.HDel(ctx, dmCallUserIndexKey(userID), rawChannelID)
			continue
		}
		if call.CallerID != userID && call.RecipientID != userID {
			continue
		}
		call, ok = e.reconcileDMCallVoicePresence(ctx, call)
		if !ok {
			continue
		}
		out = append(out, e.dmCallSummaryForUser(call, userID))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt > out[j].StartedAt })
	return out
}

func (e *entity) reconcileDMCallVoicePresence(ctx context.Context, call dmCallState) (dmCallState, bool) {
	if e.cache == nil {
		return call, true
	}
	clients, err := e.cache.HGetAll(ctx, voiceClientsKey(call.ChannelID))
	if err != nil {
		return call, true
	}
	now := time.Now().Unix()
	active := map[int64]int64{}
	for rawUserID := range clients {
		userID, err := strconv.ParseInt(rawUserID, 10, 64)
		if err != nil {
			continue
		}
		if userID == call.CallerID || userID == call.RecipientID {
			active[userID] = now
		}
	}
	if len(active) == 0 {
		call.EndedAt = now
		e.clearDMCall(ctx, call, "empty")
		return dmCallState{}, false
	}
	if len(active) != len(call.Participants) {
		call.Participants = active
		if len(active) == 1 {
			call.SoloSince = now
		} else {
			call.SoloSince = 0
		}
		_ = e.saveDMCall(ctx, call)
	}
	return call, true
}

func (e *entity) validateActiveDMCallParticipant(ctx context.Context, channelID, userID int64) (dmCallState, error) {
	if _, _, err := e.directDMParticipants(ctx, channelID, userID); err != nil {
		return dmCallState{}, err
	}
	call, ok := e.cachedDMCall(ctx, channelID)
	if !ok {
		return dmCallState{}, fiber.NewError(fiber.StatusNotFound, "call not found")
	}
	if call.Participants == nil || call.Participants[userID] == 0 {
		return dmCallState{}, fiber.NewError(fiber.StatusForbidden, "join the call first")
	}
	return call, nil
}

func (e *entity) dmStreamBinding(ctx context.Context, streamID int64, region string) (streammeta.RouteBinding, error) {
	var cached streammeta.RouteBinding
	if e.cache != nil {
		_ = e.cache.GetJSON(ctx, streammeta.RouteKey(streamID), &cached)
		if cached.URL != "" {
			return cached, nil
		}
	}
	binding, err := e.selectDMBinding(ctx, streamID, region, e.streamDisco, e.streamSelector, false)
	if err != nil {
		return streammeta.RouteBinding{}, err
	}
	return streammeta.RouteBinding{ID: binding.ID, URL: binding.URL, Region: binding.Region}, nil
}

func (e *entity) issueDMStreamToken(userID int64, meta streammeta.Metadata, role string, binding streammeta.RouteBinding) (string, error) {
	now := time.Now()
	claims := streammeta.Claims{
		Claims: helper.Claims{
			UserID:    userID,
			TokenType: "stream",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "gochat",
				Audience:  []string{"stream"},
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(dmCallStreamTokenTTL)),
			},
		},
		StreamID:    meta.ID,
		ChannelID:   meta.ChannelID,
		GuildID:     0,
		OwnerUserID: meta.OwnerUserID,
		RouteID:     binding.ID,
		Role:        role,
		SourceType:  meta.SourceType,
		AudioMode:   meta.AudioMode,
		Perms:       dmCallMediaPermissions(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(e.authSecret))
}

// ListDMCallStreams
//
//	@Summary	List active streams in a direct-message voice call
//	@Produce	json
//	@Tags		User
//	@Param		channel_id	path		int64	true	"Direct DM channel ID"
//	@Success	200			{array}		VoiceStreamSummary
//	@failure	400			{string}	string	"Bad request"
//	@failure	403			{string}	string	"Join the call first"
//	@failure	404			{string}	string	"Call not found"
//	@failure	500			{string}	string	"Internal server error"
//	@Router		/user/me/channels/{channel_id}/call/streams [get]
func (e *entity) ListDMCallStreams(c *fiber.Ctx) error {
	channelID, err := e.parseDMCallChannelID(c)
	if err != nil {
		return err
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	if _, err := e.validateActiveDMCallParticipant(c.UserContext(), channelID, user.Id); err != nil {
		return err
	}
	streams, err := e.dmChannelStreams(c.UserContext(), channelID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get streams")
	}
	return c.JSON(streams)
}

func (e *entity) dmChannelStreams(ctx context.Context, channelID int64) ([]VoiceStreamSummary, error) {
	if e.cache == nil {
		return []VoiceStreamSummary{}, nil
	}
	raw, err := e.cache.HGetAll(ctx, streammeta.ChannelKey(channelID))
	if err != nil {
		return nil, err
	}
	out := make([]VoiceStreamSummary, 0, len(raw))
	for streamIDRaw := range raw {
		streamID, err := strconv.ParseInt(streamIDRaw, 10, 64)
		if err != nil {
			continue
		}
		var meta streammeta.Metadata
		if err := e.cache.GetJSON(ctx, streammeta.MetaKey(streamID), &meta); err != nil || meta.ID == 0 || meta.ChannelID != channelID {
			continue
		}
		out = append(out, VoiceStreamSummary{
			ID:          meta.ID,
			OwnerUserID: meta.OwnerUserID,
			ChannelID:   meta.ChannelID,
			SourceType:  meta.SourceType,
			AudioMode:   meta.AudioMode,
			StartedAt:   meta.StartedAt,
		})
	}
	return out, nil
}

// StartDMCallStream
//
//	@Summary	Start or resume screen sharing in a direct-message voice call
//	@Accept		json
//	@Produce	json
//	@Tags		User
//	@Param		channel_id	path		int64						true	"Direct DM channel ID"
//	@Param		request		body		CreateDMCallStreamRequest	true	"Stream options"
//	@Success	200			{object}	CreateDMCallStreamResponse
//	@failure	400			{string}	string	"Bad request"
//	@failure	403			{string}	string	"Join the call first"
//	@failure	404			{string}	string	"Call not found"
//	@failure	409			{string}	string	"Already streaming in another call"
//	@failure	502			{string}	string	"Stream discovery unavailable"
//	@failure	503			{string}	string	"No stream service available"
//	@failure	500			{string}	string	"Internal server error"
//	@Router		/user/me/channels/{channel_id}/call/streams [post]
func (e *entity) StartDMCallStream(c *fiber.Ctx) error {
	channelID, err := e.parseDMCallChannelID(c)
	if err != nil {
		return err
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	call, err := e.validateActiveDMCallParticipant(c.UserContext(), channelID, user.Id)
	if err != nil {
		return err
	}
	var req CreateDMCallStreamRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if existing, ok, _ := e.loadDMUserStream(c.UserContext(), user.Id); ok && existing.ID != 0 {
		if existing.ChannelID != channelID {
			return fiber.NewError(fiber.StatusConflict, "user is already streaming in another call")
		}
		binding, err := e.dmStreamBinding(c.UserContext(), existing.ID, existing.Region)
		if err != nil || binding.URL == "" {
			return fiber.NewError(fiber.StatusServiceUnavailable, "no stream service available")
		}
		token, err := e.issueDMStreamToken(user.Id, existing, streammeta.RolePublisher, binding)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "unable to issue stream token")
		}
		return c.JSON(CreateDMCallStreamResponse{StreamID: existing.ID, StreamURL: binding.URL, StreamToken: token, Stream: voiceStreamSummaryFromMeta(existing)})
	}
	streamID := idgen.Next()
	binding, err := e.dmStreamBinding(c.UserContext(), streamID, call.Region)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "stream discovery unavailable")
	}
	if binding.URL == "" {
		return fiber.NewError(fiber.StatusServiceUnavailable, "no stream service available")
	}
	meta := streammeta.Metadata{
		ActiveStream: streammeta.ActiveStream{
			ID:         streamID,
			ChannelID:  channelID,
			SourceType: req.SourceType,
			AudioMode:  req.AudioMode,
			StartedAt:  time.Now().Unix(),
		},
		GuildID:     0,
		OwnerUserID: user.Id,
		Region:      binding.Region,
		RouteID:     binding.ID,
		RouteURL:    binding.URL,
	}
	if e.cache != nil {
		_ = e.cache.SetTimedJSON(c.UserContext(), streammeta.MetaKey(streamID), meta, dmCallStreamStateTTL)
		_ = e.cache.SetTimedJSON(c.UserContext(), streammeta.RouteKey(streamID), binding, dmCallStreamRouteTTL)
		_ = e.cache.SetTimedJSON(c.UserContext(), dmCallStreamUserKey(user.Id), meta, dmCallStreamStateTTL)
		_ = e.cache.HSet(c.UserContext(), streammeta.ChannelKey(channelID), strconv.FormatInt(streamID, 10), "1")
		_ = e.cache.SetTTL(c.UserContext(), streammeta.ChannelKey(channelID), dmCallStreamStateTTL)
	}
	token, err := e.issueDMStreamToken(user.Id, meta, streammeta.RolePublisher, binding)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to issue stream token")
	}
	e.publishDMCall(c.UserContext(), call, func(uid int64) mqmsg.EventDataMessage {
		return mqmsg.NewDMCallStreamStarted(e.dmCallSummaryForUser(call, uid), user.Id, meta.ActiveStream)
	})
	return c.JSON(CreateDMCallStreamResponse{StreamID: streamID, StreamURL: binding.URL, StreamToken: token, Stream: voiceStreamSummaryFromMeta(meta)})
}

func dmCallStreamUserKey(userID int64) string {
	return "dmcall:stream:user:" + strconv.FormatInt(userID, 10)
}

func (e *entity) loadDMUserStream(ctx context.Context, userID int64) (streammeta.Metadata, bool, error) {
	var meta streammeta.Metadata
	if e.cache == nil {
		return meta, false, nil
	}
	if err := e.cache.GetJSON(ctx, dmCallStreamUserKey(userID), &meta); err != nil || meta.ID == 0 {
		return streammeta.Metadata{}, false, nil
	}
	return meta, true, nil
}

func voiceStreamSummaryFromMeta(meta streammeta.Metadata) VoiceStreamSummary {
	return VoiceStreamSummary{
		ID:          meta.ID,
		OwnerUserID: meta.OwnerUserID,
		ChannelID:   meta.ChannelID,
		SourceType:  meta.SourceType,
		AudioMode:   meta.AudioMode,
		StartedAt:   meta.StartedAt,
	}
}

// JoinDMCallStream
//
//	@Summary	Join a screen share in a direct-message voice call
//	@Produce	json
//	@Tags		User
//	@Param		channel_id	path		int64	true	"Direct DM channel ID"
//	@Param		stream_id	path		int64	true	"Stream ID"
//	@Success	200			{object}	JoinDMCallStreamResponse
//	@failure	400			{string}	string	"Bad request"
//	@failure	403			{string}	string	"Join the call first"
//	@failure	404			{string}	string	"Call or stream not found"
//	@failure	503			{string}	string	"No stream service available"
//	@failure	500			{string}	string	"Internal server error"
//	@Router		/user/me/channels/{channel_id}/call/streams/{stream_id}/join [post]
func (e *entity) JoinDMCallStream(c *fiber.Ctx) error {
	channelID, err := e.parseDMCallChannelID(c)
	if err != nil {
		return err
	}
	streamID, err := e.parseDMCallStreamID(c)
	if err != nil {
		return err
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	if _, err := e.validateActiveDMCallParticipant(c.UserContext(), channelID, user.Id); err != nil {
		return err
	}
	var meta streammeta.Metadata
	if e.cache == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "unable to get stream")
	}
	if err := e.cache.GetJSON(c.UserContext(), streammeta.MetaKey(streamID), &meta); err != nil || meta.ID == 0 || meta.ChannelID != channelID {
		return fiber.NewError(fiber.StatusNotFound, "stream not found")
	}
	binding, err := e.dmStreamBinding(c.UserContext(), streamID, meta.Region)
	if err != nil || binding.URL == "" {
		return fiber.NewError(fiber.StatusServiceUnavailable, "no stream service available")
	}
	token, err := e.issueDMStreamToken(user.Id, meta, streammeta.RoleViewer, binding)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to issue stream token")
	}
	return c.JSON(JoinDMCallStreamResponse{StreamID: meta.ID, StreamURL: binding.URL, StreamToken: token})
}

// StopDMCallStream
//
//	@Summary	Stop an owned screen share in a direct-message voice call
//	@Produce	json
//	@Tags		User
//	@Param		channel_id	path	int64	true	"Direct DM channel ID"
//	@Param		stream_id	path	int64	true	"Stream ID"
//	@Success	204
//	@failure	400	{string}	string	"Bad request"
//	@failure	403	{string}	string	"Join the call first or not stream owner"
//	@failure	404	{string}	string	"Call not found"
//	@failure	500	{string}	string	"Internal server error"
//	@Router		/user/me/channels/{channel_id}/call/streams/{stream_id} [delete]
func (e *entity) StopDMCallStream(c *fiber.Ctx) error {
	channelID, err := e.parseDMCallChannelID(c)
	if err != nil {
		return err
	}
	streamID, err := e.parseDMCallStreamID(c)
	if err != nil {
		return err
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	call, err := e.validateActiveDMCallParticipant(c.UserContext(), channelID, user.Id)
	if err != nil {
		return err
	}
	var meta streammeta.Metadata
	if e.cache == nil {
		return c.SendStatus(fiber.StatusNoContent)
	}
	if err := e.cache.GetJSON(c.UserContext(), streammeta.MetaKey(streamID), &meta); err != nil || meta.ID == 0 || meta.ChannelID != channelID {
		return c.SendStatus(fiber.StatusNoContent)
	}
	if meta.OwnerUserID != user.Id {
		return fiber.NewError(fiber.StatusForbidden, "not stream owner")
	}
	_ = e.cache.Delete(c.UserContext(), streammeta.MetaKey(streamID))
	_ = e.cache.Delete(c.UserContext(), streammeta.RouteKey(streamID))
	_ = e.cache.Delete(c.UserContext(), dmCallStreamUserKey(user.Id))
	_ = e.cache.HDel(c.UserContext(), streammeta.ChannelKey(channelID), strconv.FormatInt(streamID, 10))
	e.publishDMCall(c.UserContext(), call, func(uid int64) mqmsg.EventDataMessage {
		return mqmsg.NewDMCallStreamStopped(e.dmCallSummaryForUser(call, uid), user.Id, streamID, "owner_stop")
	})
	return c.SendStatus(fiber.StatusNoContent)
}
