package message

import (
	"strings"

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

func badRequestQueryParseError() error {
	return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseQuery)
}

func badRequestEmbedError(err error) error {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = ErrInvalidEmbeds
	}
	return fiber.NewError(fiber.StatusBadRequest, message)
}
