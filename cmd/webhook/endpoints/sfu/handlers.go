package sfu

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
	"github.com/gofiber/fiber/v2"
)

const (
	hdrToken = "X-Webhook-Token"
)

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
	var req HeartbeatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
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
		e.log.Error("discovery register failed", slog.String("error", err.Error()), slog.String("id", req.ID), slog.String("region", req.Region))
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
	var req ChannelUserJoin
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	slog.Info("Join data", slog.Any("data", req))

	if !e.tokens.Validate("sfu", "", c.Get(hdrToken)) {
		return fiber.ErrUnauthorized
	}
	err := e.cache.HSet(
		c.UserContext(),
		fmt.Sprintf("voice:clients:%d", req.ChannelId),
		strconv.FormatInt(req.UserId, 10), "true")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to update channel state")
	}
	ttlErr := e.cache.SetTTL(c.UserContext(), fmt.Sprintf("voice:clients:%d", req.ChannelId), 120)
	if ttlErr != nil {
		slog.Error("unable to set TTL for channel",
			slog.Int64("channel_id", req.ChannelId),
			slog.String("error", ttlErr.Error()))
	}
	if req.GuildId != nil {
		go func() {
			if err := e.mqt.SendGuildUpdate(*req.GuildId, &mqmsg.GuildMemberJoinVoice{
				GuildId:   *req.GuildId,
				UserId:    req.UserId,
				ChannelId: req.ChannelId,
			}); err != nil {
				slog.Error("unable to send guild update", slog.String("error", err.Error()))
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
	var req ChannelUserLeave
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if !e.tokens.Validate("sfu", "", c.Get(hdrToken)) {
		return fiber.ErrUnauthorized
	}
	err := e.cache.HDel(
		c.UserContext(),
		fmt.Sprintf("voice:clients:%d", req.ChannelId),
		strconv.FormatInt(req.UserId, 10))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to update channel state")
	}
	ttlErr := e.cache.SetTTL(c.UserContext(), fmt.Sprintf("voice:clients:%d", req.ChannelId), 120)
	if ttlErr != nil {
		slog.Error("unable to set TTL for channel",
			slog.Int64("channel_id", req.ChannelId),
			slog.String("error", ttlErr.Error()))
	}
	if req.GuildId != nil {
		go func() {
			if err := e.mqt.SendGuildUpdate(*req.GuildId, &mqmsg.GuildMemberLeaveVoice{
				GuildId:   *req.GuildId,
				UserId:    req.UserId,
				ChannelId: req.ChannelId,
			}); err != nil {
				slog.Error("unable to send guild update", slog.String("error", err.Error()))
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
	var req ChannelAlive
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if !e.tokens.Validate("sfu", "", c.Get(hdrToken)) {
		return fiber.ErrUnauthorized
	}

	ttlErr := e.cache.SetTTL(c.UserContext(), fmt.Sprintf("voice:clients:%d", req.ChannelId), 120)
	if ttlErr != nil {
		slog.Error("unable to set TTL for channel users",
			slog.Int64("channel_id", req.ChannelId),
			slog.String("error", ttlErr.Error()))
	}
	ttlErr = e.cache.SetTTL(c.UserContext(), fmt.Sprintf("voice:route:%d", req.ChannelId), 120)
	if ttlErr != nil {
		slog.Error("unable to set TTL for channel",
			slog.Int64("channel_id", req.ChannelId),
			slog.String("error", ttlErr.Error()))
	}

	return c.SendStatus(fiber.StatusOK)
}
