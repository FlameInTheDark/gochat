package sfu

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/presence"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
	"github.com/gofiber/fiber/v2"
)

const (
	hdrToken                   = "X-Webhook-Token"
	voiceRouteActiveTTLSeconds = 180
	voiceClientsTTLSeconds     = 120
	dmCallTTLSeconds           = int64(60 * 60 * 6)
	dmCallUserIndexTTLSeconds  = int64(60 * 60 * 6)
)

type voiceRouteBinding struct {
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
	Participants map[int64]int64 `json:"participants,omitempty"`
	DismissedBy  map[int64]bool  `json:"dismissed_by,omitempty"`
	StartedAt    int64           `json:"started_at"`
	UpdatedAt    int64           `json:"updated_at"`
	SoloSince    int64           `json:"solo_since,omitempty"`
	EndedAt      int64           `json:"ended_at,omitempty"`
}

func dmCallStateKey(channelID int64) string {
	return "dmcall:state:" + strconv.FormatInt(channelID, 10)
}

func dmCallRouteKey(channelID int64) string {
	return "dmcall:route:" + strconv.FormatInt(channelID, 10)
}

func dmCallUserIndexKey(userID int64) string {
	return "dmcall:user:" + strconv.FormatInt(userID, 10)
}

func voiceClientsKey(channelID int64) string {
	return fmt.Sprintf("voice:clients:%d", channelID)
}

func voiceRouteKey(channelID int64) string {
	return fmt.Sprintf("voice:route:%d", channelID)
}

func voiceRebindKey(channelID int64) string {
	return fmt.Sprintf("voice:rebind:%d", channelID)
}

func refreshVoiceRoute(ctx context.Context, cache cache.Cache, channelID int64, routeID, routeURL, region string) error {
	if cache == nil || channelID == 0 || routeID == "" || routeURL == "" {
		return nil
	}
	if raw, err := cache.Get(ctx, fmt.Sprintf("voice:rebind:%d", channelID)); err == nil && raw != "" {
		var marker voiceRouteBinding
		if json.Unmarshal([]byte(raw), &marker) == nil && marker.ID != "" && marker.ID != routeID {
			return nil
		}
	}
	return cache.SetTimedJSON(ctx, fmt.Sprintf("voice:route:%d", channelID), voiceRouteBinding{
		ID:     routeID,
		URL:    routeURL,
		Region: region,
	}, voiceRouteActiveTTLSeconds)
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
	if err := e.cache.SetTimedJSON(ctx, dmCallStateKey(call.ChannelID), call, dmCallTTLSeconds); err != nil {
		return err
	}
	for _, userID := range []int64{call.CallerID, call.RecipientID} {
		_ = e.cache.HSet(ctx, dmCallUserIndexKey(userID), strconv.FormatInt(call.ChannelID, 10), "1")
		_ = e.cache.SetTTL(ctx, dmCallUserIndexKey(userID), dmCallUserIndexTTLSeconds)
	}
	return nil
}

func (e *entity) activeDMVoiceParticipants(ctx context.Context, call dmCallState) map[int64]int64 {
	active := map[int64]int64{}
	if e.cache == nil {
		return active
	}
	clients, err := e.cache.HGetAll(ctx, voiceClientsKey(call.ChannelID))
	if err != nil {
		return active
	}
	now := time.Now().Unix()
	for rawUserID := range clients {
		userID, err := strconv.ParseInt(rawUserID, 10, 64)
		if err != nil {
			continue
		}
		if userID == call.CallerID || userID == call.RecipientID {
			active[userID] = now
		}
	}
	return active
}

func (e *entity) handleDMVoiceLeave(ctx context.Context, channelID, userID int64) {
	if e.cache == nil {
		return
	}
	var call dmCallState
	if err := e.cache.GetJSON(ctx, dmCallStateKey(channelID), &call); err != nil || call.CallID == 0 || call.EndedAt != 0 {
		return
	}
	if call.CallerID != userID && call.RecipientID != userID {
		return
	}

	active := e.activeDMVoiceParticipants(ctx, call)
	now := time.Now().Unix()
	call.Participants = active
	if len(active) == 0 {
		call.EndedAt = now
		e.clearDMCall(ctx, call, "empty")
		return
	}
	if len(active) == 1 {
		call.SoloSince = now
	} else {
		call.SoloSince = 0
	}
	if err := e.saveDMCall(ctx, call); err == nil {
		e.publishDMCall(ctx, call, func(uid int64) mqmsg.EventDataMessage {
			return mqmsg.NewDMCallLeft(e.dmCallSummaryForUser(call, uid), userID)
		})
	}
}

func (e *entity) clearOwnedStream(ctx context.Context, userID, channelID int64, reason string) {
	if e.cache == nil || userID == 0 || channelID == 0 {
		return
	}

	var meta streammeta.Metadata
	if err := e.cache.GetJSON(ctx, streammeta.UserKey(userID), &meta); err != nil || meta.ID == 0 || meta.ChannelID != channelID {
		return
	}

	_ = e.cache.Delete(ctx, streammeta.MetaKey(meta.ID))
	_ = e.cache.Delete(ctx, streammeta.RouteKey(meta.ID))
	_ = e.cache.Delete(ctx, streammeta.UserKey(userID))
	_ = e.cache.Delete(ctx, streammeta.RebindKey(meta.ID))
	_ = e.cache.HDel(ctx, streammeta.ChannelKey(channelID), strconv.FormatInt(meta.ID, 10))
	if e.pstore != nil {
		_ = e.pstore.ClearActiveStream(ctx, userID)
		_, _ = presence.Refresh(ctx, e.pstore, e.nats, userID, voiceRouteActiveTTLSeconds)
	}
	if meta.GuildID != 0 {
		_ = mq.SendGuildUpdate(ctx, e.mqt, meta.GuildID, &mqmsg.GuildMemberStopStream{
			GuildId:   meta.GuildID,
			ChannelId: meta.ChannelID,
			UserId:    meta.OwnerUserID,
			StreamId:  meta.ID,
			Reason:    reason,
		})
	}
}

// Heartbeat
//
//	@Summary		SFU heartbeat
//	@Description	Register or refresh SFU instance for discovery
//	@Tags			Webhook
//	@Accept			json
//	@Produce		json
//	@Param			X-Webhook-Token	header	string				true	"JWT token"
//	@Param			request			body	HeartbeatRequest	true	"Heartbeat payload"
//	@Success		204
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		401	{string}	string	"Unauthorized"
//	@Failure		503	{string}	string	"Service unavailable"
//	@Failure		502	{string}	string	"Bad gateway"
//	@Router			/webhook/sfu/heartbeat [post]
func (e *entity) Heartbeat(c *fiber.Ctx) error {
	log := observability.LoggerFromFiber(c, e.log)

	var req HeartbeatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return badRequestValidationError(err)
	}

	if !e.tokens.Validate("sfu", req.ID, c.Get(hdrToken)) {
		return fiber.ErrUnauthorized
	}
	if e.disco == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "discovery manager not configured")
	}

	inst := discovery.Instance{
		ID:        req.ID,
		Region:    req.Region,
		URL:       req.URL,
		Load:      req.Load,
		UpdatedAt: time.Now().Unix(),
	}
	if err := e.disco.Register(c.UserContext(), req.Region, inst); err != nil {
		log.Error("discovery register failed", slog.String("error", err.Error()), slog.String("id", req.ID), slog.String("region", req.Region))
		return fiber.NewError(fiber.StatusBadGateway, "discovery register failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Heartbeat
//
//	@Summary		SFU voice join
//	@Description	Add client to voice channel participants list
//	@Tags			Webhook
//	@Accept			json
//	@Produce		json
//	@Param			X-Webhook-Token	header	string			true	"JWT token"
//	@Param			request			body	ChannelUserJoin	true	"Client join data"
//	@Success		200
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		401	{string}	string	"Unauthorized"
//	@Failure		500	{string}	string	"Unable to set status"
//	@Failure		503	{string}	string	"Service unavailable"
//	@Failure		502	{string}	string	"Bad gateway"
//	@Router			/webhook/sfu/voice/join [post]
func (e *entity) ChannelUserJoin(c *fiber.Ctx) error {
	log := observability.LoggerFromFiber(c, e.log)

	var req ChannelUserJoin
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return badRequestValidationError(err)
	}

	log.Info("voice join received",
		slog.Int64("channel_id", req.ChannelId),
		slog.Int64("user_id", req.UserId),
		slog.Bool("has_guild_id", req.GuildId != nil))

	if !e.tokens.Validate("sfu", "", c.Get(hdrToken)) {
		return fiber.ErrUnauthorized
	}
	err := e.cache.HSet(
		c.UserContext(),
		voiceClientsKey(req.ChannelId),
		strconv.FormatInt(req.UserId, 10), "true")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to update channel state")
	}
	if err := refreshVoiceRoute(c.UserContext(), e.cache, req.ChannelId, req.RouteID, req.RouteURL, req.Region); err != nil {
		log.Error("unable to refresh voice route",
			slog.Int64("channel_id", req.ChannelId),
			slog.String("error", err.Error()))
	}
	ttlErr := e.cache.SetTTL(c.UserContext(), fmt.Sprintf("voice:clients:%d", req.ChannelId), voiceClientsTTLSeconds)
	if ttlErr != nil {
		log.Error("unable to set ttl for voice clients",
			slog.Int64("channel_id", req.ChannelId),
			slog.String("error", ttlErr.Error()))
	}
	if req.GuildId != nil {
		ctx := observability.BackgroundFromContext(c.UserContext())
		asyncLog := observability.LoggerWithContext(ctx, log)
		go func() {
			if err := mq.SendGuildUpdate(ctx, e.mqt, *req.GuildId, &mqmsg.GuildMemberJoinVoice{
				GuildId:   *req.GuildId,
				UserId:    req.UserId,
				ChannelId: req.ChannelId,
			}); err != nil {
				asyncLog.Error("unable to send guild voice join update", slog.String("error", err.Error()))
			}
		}()
	}
	return c.SendStatus(fiber.StatusOK)
}

// Heartbeat
//
//	@Summary		SFU voice leave
//	@Description	Remove client from voice channel participants list
//	@Tags			Webhook
//	@Accept			json
//	@Produce		json
//	@Param			X-Webhook-Token	header	string				true	"JWT token"
//	@Param			request			body	ChannelUserLeave	true	"Client join data"
//	@Success		200
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		401	{string}	string	"Unauthorized"
//	@Failure		500	{string}	string	"Unable to set status"
//	@Failure		503	{string}	string	"Service unavailable"
//	@Failure		502	{string}	string	"Bad gateway"
//	@Router			/webhook/sfu/voice/leave [post]
func (e *entity) ChannelUserLeave(c *fiber.Ctx) error {
	log := observability.LoggerFromFiber(c, e.log)

	var req ChannelUserLeave
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return badRequestValidationError(err)
	}

	if !e.tokens.Validate("sfu", "", c.Get(hdrToken)) {
		return fiber.ErrUnauthorized
	}
	err := e.cache.HDel(
		c.UserContext(),
		voiceClientsKey(req.ChannelId),
		strconv.FormatInt(req.UserId, 10))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to update channel state")
	}
	ttlErr := e.cache.SetTTL(c.UserContext(), fmt.Sprintf("voice:clients:%d", req.ChannelId), voiceClientsTTLSeconds)
	if ttlErr != nil {
		log.Error("unable to set ttl for voice clients",
			slog.Int64("channel_id", req.ChannelId),
			slog.String("error", ttlErr.Error()))
	}
	e.clearOwnedStream(c.UserContext(), req.UserId, req.ChannelId, "voice_left")
	if req.GuildId == nil {
		e.handleDMVoiceLeave(c.UserContext(), req.ChannelId, req.UserId)
	}
	if req.GuildId != nil {
		ctx := observability.BackgroundFromContext(c.UserContext())
		asyncLog := observability.LoggerWithContext(ctx, log)
		go func() {
			if err := mq.SendGuildUpdate(ctx, e.mqt, *req.GuildId, &mqmsg.GuildMemberLeaveVoice{
				GuildId:   *req.GuildId,
				UserId:    req.UserId,
				ChannelId: req.ChannelId,
			}); err != nil {
				asyncLog.Error("unable to send guild voice leave update", slog.String("error", err.Error()))
			}
		}()
	}
	return c.SendStatus(fiber.StatusOK)
}

// Heartbeat
//
//	@Summary		SFU update channel TTL
//	@Description	Updates channel TTL to keep it alive in system cache for next connections
//	@Tags			Webhook
//	@Accept			json
//	@Produce		json
//	@Param			X-Webhook-Token	header	string			true	"JWT token"
//	@Param			request			body	ChannelAlive	true	"Channel liveness data"
//	@Success		200
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		401	{string}	string	"Unauthorized"
//	@Failure		500	{string}	string	"Unable to set status"
//	@Failure		503	{string}	string	"Service unavailable"
//	@Failure		502	{string}	string	"Bad gateway"
//	@Router			/webhook/sfu/channel/alive [post]
func (e *entity) ChannelAlive(c *fiber.Ctx) error {
	log := observability.LoggerFromFiber(c, e.log)

	var req ChannelAlive
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return badRequestValidationError(err)
	}

	if !e.tokens.Validate("sfu", "", c.Get(hdrToken)) {
		return fiber.ErrUnauthorized
	}

	if err := refreshVoiceRoute(c.UserContext(), e.cache, req.ChannelId, req.RouteID, req.RouteURL, req.Region); err != nil {
		log.Error("unable to refresh voice route",
			slog.Int64("channel_id", req.ChannelId),
			slog.String("error", err.Error()))
	}

	ttlErr := e.cache.SetTTL(c.UserContext(), fmt.Sprintf("voice:clients:%d", req.ChannelId), voiceClientsTTLSeconds)
	if ttlErr != nil {
		log.Error("unable to set ttl for channel users",
			slog.Int64("channel_id", req.ChannelId),
			slog.String("error", ttlErr.Error()))
	}

	return c.SendStatus(fiber.StatusOK)
}
