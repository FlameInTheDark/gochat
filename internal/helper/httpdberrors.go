package helper

import (
	"errors"

	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
)

func HttpDbError(err error, msg string) error {
	switch errors.Unwrap(err) {
	case nil:
		return nil
	case gocql.ErrNotFound:
		return fiber.NewError(fiber.StatusNotFound, msg)
	default:
		return fiber.NewError(fiber.StatusInternalServerError, msg+": "+err.Error())
	}
}
