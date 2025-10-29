package attachments

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	hdrToken = "X-Webhook-Token"
	reqTO    = 5 * time.Second
)

// Finalize
//
//	@Summary		Finalize attachment metadata
//	@Description	Persist completed attachment metadata after upload
//	@Tags			Webhook
//	@Accept			json
//	@Produce		json
//	@Param			X-Webhook-Token	header	string			true	"JWT token"
//	@Param			request			body	FinalizeRequest	true	"Finalize payload"
//	@Success		204
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		401	{string}	string	"Unauthorized"
//	@Failure		503	{string}	string	"Service unavailable"
//	@Failure		502	{string}	string	"Bad gateway"
//	@Router			/webhook/attachments/finalize [post]
func (e *entity) Finalize(c *fiber.Ctx) error {
	if !e.tokens.Validate("attachments", "", c.Get(hdrToken)) {
		return fiber.ErrUnauthorized
	}
	if e.att == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "attachments store not configured")
	}

	var req FinalizeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), reqTO)
	defer cancel()
	if err := e.att.DoneAttachment(ctx, req.ID, req.ChannelID, req.ContentType, req.URL, req.PreviewURL, req.Height, req.Width, req.FileSize, req.Name, req.AuthorID); err != nil {
		e.log.Error("attachment finalize failed", slog.String("error", err.Error()), slog.Int64("id", req.ID), slog.Int64("channel", req.ChannelID))
		return fiber.NewError(fiber.StatusBadGateway, "finalize failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
