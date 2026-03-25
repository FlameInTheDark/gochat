package emoji

import "github.com/gofiber/fiber/v2"

const (
	ErrIncorrectEmojiID     = "incorrect emoji ID"
	ErrUnableToGetUserToken = "unable to get user token"
	ErrUnableToGetEmojiInfo = "unable to get emoji info"
	ErrEmojiNotFound        = "emoji not found"
)

func (e *entity) parseEmojiID(c *fiber.Ctx) (int64, error) {
	emojiID, err := c.ParamsInt("emoji_id")
	if err != nil || emojiID <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectEmojiID)
	}
	return int64(emojiID), nil
}
