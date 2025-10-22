package sfu

import (
	"log/slog"
	"time"

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
