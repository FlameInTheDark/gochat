package observability

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

func LoggerWithContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if logger == nil {
		logger = Logger()
	}
	if logger == nil || ctx == nil {
		return logger
	}
	return helper.WithContext(logger, ctx)
}

func LoggerFromFiber(c *fiber.Ctx, logger *slog.Logger) *slog.Logger {
	if c == nil {
		return LoggerWithContext(nil, logger)
	}
	return LoggerWithContext(c.UserContext(), logger)
}
