package search

import (
	"github.com/FlameInTheDark/gochat/internal/validationutil"
	"github.com/gofiber/fiber/v2"
)

func badRequestValidationError(err error) error {
	message := validationutil.PublicMessage(err)
	if message == "" {
		message = ErrUnableToParseBody
	}
	return fiber.NewError(fiber.StatusBadRequest, message)
}
