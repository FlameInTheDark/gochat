package helper

import (
	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
)

func HttpDbError(c *fiber.Ctx, err error) error {
	switch err {
	case nil:
		return nil
	case gocql.ErrNotFound:
		return c.SendStatus(fiber.StatusNotFound)
	default:
		return c.SendStatus(fiber.StatusInternalServerError)
	}
}
