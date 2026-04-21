package guild

import (
	"log/slog"
	"net/http"

	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/gofiber/fiber/v2"
)

func (e *entity) publicError(c *fiber.Ctx, status int, err error) error {
	if err != nil {
		observability.LoggerFromFiber(c, e.log).Error(
			"guild endpoint failed",
			slog.Int("status", status),
			slog.String("error", err.Error()),
		)
	}

	message := http.StatusText(status)
	if message == "" {
		message = fiber.ErrInternalServerError.Message
	}
	return fiber.NewError(status, message)
}
