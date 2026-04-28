package stream

import (
	"context"
	"encoding/json"
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
	hdrToken                = "X-Webhook-Token"
	streamRouteTTLSeconds   = int64(180)
	streamPresenceTTLSecond = int64(180)
)

func badRequestValidationError(err error) error {
	return fiber.NewError(fiber.StatusBadRequest, err.Error())
}

func refreshStreamRoute(ctx context.Context, cache cache.Cache, streamID int64, routeID, routeURL, region string) error {
	if cache == nil || streamID == 0 || routeID == "" || routeURL == "" {
		return nil
	}
	return cache.SetTimedJSON(ctx, streammeta.RouteKey(streamID), streammeta.RouteBinding{
		ID:     routeID,
		URL:    routeURL,
		Region: region,
	}, streamRouteTTLSeconds)
}

func (e *entity) publishPresence(ctx context.Context, userID int64) {
	if e.pstore == nil {
		return
	}
	_, _ = presence.Refresh(ctx, e.pstore, e.nats, userID, streamPresenceTTLSecond)
}

func (e *entity) publishStart(ctx context.Context, meta streammeta.Metadata) {
	_ = mq.SendGuildUpdate(ctx, e.mqt, meta.GuildID, &mqmsg.GuildMemberStartStream{
		GuildId:   meta.GuildID,
		ChannelId: meta.ChannelID,
		UserId:    meta.OwnerUserID,
		Stream:    meta.ActiveStream,
	})
}

func (e *entity) publishStop(ctx context.Context, meta streammeta.Metadata, reason string) {
	_ = mq.SendGuildUpdate(ctx, e.mqt, meta.GuildID, &mqmsg.GuildMemberStopStream{
		GuildId:   meta.GuildID,
		ChannelId: meta.ChannelID,
		UserId:    meta.OwnerUserID,
		StreamId:  meta.ID,
		Reason:    reason,
	})
}

func (e *entity) loadMetadata(ctx context.Context, streamID int64) (streammeta.Metadata, bool) {
	if e.cache == nil {
		return streammeta.Metadata{}, false
	}
	var meta streammeta.Metadata
	if err := e.cache.GetJSON(ctx, streammeta.MetaKey(streamID), &meta); err != nil || meta.ID == 0 {
		return streammeta.Metadata{}, false
	}
	return meta, true
}

// Heartbeat registers or refreshes a stream instance for discovery.
func (e *entity) Heartbeat(c *fiber.Ctx) error {
	log := observability.LoggerFromFiber(c, e.log)

	var req HeartbeatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return badRequestValidationError(err)
	}

	if !e.tokens.Validate("stream", req.ID, c.Get(hdrToken)) {
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
		log.Error("stream discovery register failed", slog.String("error", err.Error()), slog.String("id", req.ID), slog.String("region", req.Region))
		return fiber.NewError(fiber.StatusBadGateway, "discovery register failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Start marks a stream active after the publisher media session is established.
func (e *entity) Start(c *fiber.Ctx) error {
	var req StartRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return badRequestValidationError(err)
	}
	if !e.tokens.Validate("stream", "", c.Get(hdrToken)) {
		return fiber.ErrUnauthorized
	}

	meta, ok := e.loadMetadata(c.UserContext(), req.StreamID)
	if !ok {
		meta = streammeta.Metadata{
			ActiveStream: streammeta.ActiveStream{
				ID:         req.StreamID,
				ChannelID:  req.ChannelID,
				SourceType: req.SourceType,
				AudioMode:  req.AudioMode,
				StartedAt:  req.StartedAt,
			},
			GuildID:     req.GuildID,
			OwnerUserID: req.OwnerUserID,
		}
	}
	if meta.StartedAt == 0 {
		meta.StartedAt = time.Now().Unix()
	}
	meta.GuildID = req.GuildID
	meta.OwnerUserID = req.OwnerUserID
	meta.ChannelID = req.ChannelID
	meta.SourceType = req.SourceType
	meta.AudioMode = req.AudioMode
	meta.Region = req.Region
	meta.RouteID = req.RouteID
	meta.RouteURL = req.RouteURL
	meta.PublisherSID = req.PublisherSessionID

	raw, err := json.Marshal(meta)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to marshal stream")
	}

	if e.cache != nil {
		_ = refreshStreamRoute(c.UserContext(), e.cache, req.StreamID, req.RouteID, req.RouteURL, req.Region)
		_ = e.cache.SetTimedJSON(c.UserContext(), streammeta.MetaKey(req.StreamID), meta, streamRouteTTLSeconds)
		_ = e.cache.SetTimedJSON(c.UserContext(), streammeta.UserKey(req.OwnerUserID), meta, streamRouteTTLSeconds)
		_ = e.cache.HSet(c.UserContext(), streammeta.ChannelKey(req.ChannelID), streammeta_fmtInt64(req.StreamID), string(raw))
		_ = e.cache.SetTTL(c.UserContext(), streammeta.ChannelKey(req.ChannelID), streamRouteTTLSeconds)
	}
	if e.pstore != nil {
		_ = e.pstore.SetActiveStream(c.UserContext(), req.OwnerUserID, meta.ActiveStream, streamPresenceTTLSecond)
	}
	e.publishStart(c.UserContext(), meta)
	e.publishPresence(c.UserContext(), req.OwnerUserID)
	return c.SendStatus(fiber.StatusOK)
}

// Stop removes active stream state.
func (e *entity) Stop(c *fiber.Ctx) error {
	var req StopRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return badRequestValidationError(err)
	}
	if !e.tokens.Validate("stream", "", c.Get(hdrToken)) {
		return fiber.ErrUnauthorized
	}

	meta, ok := e.loadMetadata(c.UserContext(), req.StreamID)
	if !ok {
		meta = streammeta.Metadata{
			ActiveStream: streammeta.ActiveStream{
				ID:        req.StreamID,
				ChannelID: req.ChannelID,
			},
			GuildID:     req.GuildID,
			OwnerUserID: req.OwnerUserID,
		}
	}
	if req.Reason == "" {
		req.Reason = "ended"
	}
	if meta.OwnerUserID != 0 && req.OwnerUserID != 0 && meta.OwnerUserID != req.OwnerUserID {
		return c.SendStatus(fiber.StatusOK)
	}
	if meta.PublisherSID != "" && req.PublisherSessionID != "" && meta.PublisherSID != req.PublisherSessionID {
		return c.SendStatus(fiber.StatusOK)
	}

	if e.cache != nil {
		_ = e.cache.Delete(c.UserContext(), streammeta.MetaKey(req.StreamID))
		_ = e.cache.Delete(c.UserContext(), streammeta.RouteKey(req.StreamID))
		_ = e.cache.Delete(c.UserContext(), streammeta.UserKey(meta.OwnerUserID))
		_ = e.cache.Delete(c.UserContext(), streammeta.RebindKey(req.StreamID))
		_ = e.cache.HDel(c.UserContext(), streammeta.ChannelKey(req.ChannelID), streammeta_fmtInt64(req.StreamID))
	}
	if e.pstore != nil {
		_ = e.pstore.ClearActiveStream(c.UserContext(), meta.OwnerUserID)
	}
	e.publishStop(c.UserContext(), meta, req.Reason)
	e.publishPresence(c.UserContext(), meta.OwnerUserID)
	return c.SendStatus(fiber.StatusOK)
}

// Alive refreshes stream TTLs while the stream is active.
func (e *entity) Alive(c *fiber.Ctx) error {
	var req AliveRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return badRequestValidationError(err)
	}
	if !e.tokens.Validate("stream", "", c.Get(hdrToken)) {
		return fiber.ErrUnauthorized
	}

	meta, ok := e.loadMetadata(c.UserContext(), req.StreamID)
	if !ok {
		return c.SendStatus(fiber.StatusOK)
	}

	if req.RouteID != "" && req.RouteURL != "" && e.cache != nil {
		meta.RouteID = req.RouteID
		meta.RouteURL = req.RouteURL
		if req.Region != "" {
			meta.Region = req.Region
		}
		_ = refreshStreamRoute(c.UserContext(), e.cache, req.StreamID, meta.RouteID, meta.RouteURL, meta.Region)
		_ = e.cache.SetTimedJSON(c.UserContext(), streammeta.MetaKey(req.StreamID), meta, streamRouteTTLSeconds)
	}
	if e.cache != nil {
		_ = e.cache.SetTTL(c.UserContext(), streammeta.MetaKey(req.StreamID), streamRouteTTLSeconds)
		_ = e.cache.SetTTL(c.UserContext(), streammeta.UserKey(meta.OwnerUserID), streamRouteTTLSeconds)
		_ = e.cache.SetTTL(c.UserContext(), streammeta.ChannelKey(meta.ChannelID), streamRouteTTLSeconds)
	}
	if e.pstore != nil {
		_ = e.pstore.SetActiveStream(c.UserContext(), meta.OwnerUserID, meta.ActiveStream, streamPresenceTTLSecond)
	}
	return c.SendStatus(fiber.StatusOK)
}

func streammeta_fmtInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}
