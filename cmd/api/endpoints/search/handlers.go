package search

import (
	"github.com/gofiber/fiber/v2"
)

func (e *entity) Search(c *fiber.Ctx) error {
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
}
