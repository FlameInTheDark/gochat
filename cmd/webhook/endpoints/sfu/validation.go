package sfu

import (
	"github.com/FlameInTheDark/gochat/internal/validationutil"
	"github.com/gofiber/fiber/v2"
)

const errInvalidRequest = "invalid request"

func badRequestValidationError(err error) error {
	message := validationutil.PublicMessage(err)
	if message == "" {
		message = errInvalidRequest
	}
	return fiber.NewError(fiber.StatusBadRequest, message)
}
