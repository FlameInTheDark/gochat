package auth

import (
	"log/slog"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/validationutil"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type validatableRequest interface {
	Validate() error
}

func (e *entity) parseAndValidate(c *fiber.Ctx, operation string, req validatableRequest) error {
	log := observability.LoggerFromFiber(c, e.log)

	if err := c.BodyParser(req); err != nil {
		return rejectBadRequest(c, log, operation, "parse", parseBodyPublicMessage(err), err)
	}

	if err := req.Validate(); err != nil {
		publicMessage := validationutil.PublicMessage(err)
		if publicMessage == "" {
			publicMessage = ErrUnableToParseBody
		}
		return rejectBadRequest(c, log, operation, "validation", publicMessage, err)
	}

	return nil
}

func parseBodyPublicMessage(err error) string {
	if err == nil {
		return ErrUnableToParseBody
	}

	message := strings.TrimSpace(err.Error())
	if message == "" {
		return ErrUnableToParseBody
	}

	return ErrUnableToParseBody + ": " + message
}

func rejectBadRequest(c *fiber.Ctx, log *slog.Logger, operation, stage, publicMessage string, err error) error {
	logAttrs := []slog.Attr{
		slog.String("auth_operation", operation),
		slog.String("stage", stage),
		slog.String("content_type", c.Get(fiber.HeaderContentType)),
		slog.Int("body_size", len(c.Body())),
	}
	if err != nil {
		logAttrs = append(logAttrs, slog.String("error", err.Error()))
	}

	if log != nil {
		args := make([]any, 0, len(logAttrs))
		for _, attr := range logAttrs {
			args = append(args, attr)
		}
		log.Warn("auth request rejected", args...)
	}

	span := trace.SpanFromContext(c.UserContext())
	spanAttrs := []attribute.KeyValue{
		attribute.String("auth.operation", operation),
		attribute.String("auth.error_stage", stage),
		attribute.String("http.request.content_type", c.Get(fiber.HeaderContentType)),
		attribute.Int("http.request.body_size", len(c.Body())),
		attribute.String("http.route", c.Path()),
		attribute.String("error.message", publicMessage),
	}
	span.SetAttributes(spanAttrs...)
	span.AddEvent("auth.request.rejected", trace.WithAttributes(spanAttrs...))
	if err != nil {
		span.RecordError(err)
	}

	return fiber.NewError(fiber.StatusBadRequest, publicMessage)
}
