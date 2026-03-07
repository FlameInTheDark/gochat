package emojis

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/dto"
	emojiutil "github.com/FlameInTheDark/gochat/internal/emoji"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

func (e *entity) Upload(c *fiber.Ctx) error {
	guildId, err := strconv.ParseInt(c.Params("guild_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}
	emojiId, err := strconv.ParseInt(c.Params("emoji_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectEmojiID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	_, ok, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermCreateExpressions)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetEmoji)
	}
	if !ok {
		return fiber.NewError(fiber.StatusForbidden, ErrForbiddenToUpload)
	}

	body, err := requestBodyReader(c)
	if err != nil {
		return err
	}

	result, err := e.uploader.Upload(c.UserContext(), guildId, emojiId, body)
	if err != nil {
		return emojiUploadError(err)
	}
	if result.AlreadyDone {
		return c.SendStatus(fiber.StatusNoContent)
	}

	if e.cache != nil {
		_ = e.cache.Delete(c.UserContext(), emojiutil.LookupCacheKey(emojiId))
		_ = e.cache.Delete(c.UserContext(), emojiutil.GuildCacheKey(guildId))
	}
	go func() {
		_ = e.mqt.SendGuildUpdate(guildId, &mqmsg.CreateGuildEmoji{Emoji: dto.GuildEmoji{Id: emojiId, GuildId: guildId, Name: result.Name, Animated: result.Animated}})
	}()
	return c.SendStatus(fiber.StatusCreated)
}

func requestBodyReader(c *fiber.Ctx) (io.Reader, error) {
	if r := c.Context().RequestBodyStream(); r != nil {
		return r, nil
	}
	body := c.Body()
	if len(body) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToReadBody)
	}
	return bytes.NewReader(body), nil
}

func emojiUploadError(err error) error {
	switch {
	case errors.Is(err, upload.ErrPlaceholderNotFound):
		return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetEmoji)
	case errors.Is(err, upload.ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, ErrForbiddenToUpload)
	case errors.Is(err, upload.ErrEmptyBody), errors.Is(err, upload.ErrSizeMismatch):
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToReadBody)
	case errors.Is(err, upload.ErrTooLarge):
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
	case errors.Is(err, upload.ErrUnsupportedMedia):
		return fiber.NewError(fiber.StatusUnsupportedMediaType, ErrUnsupportedContentType)
	case errors.Is(err, upload.ErrInvalidDimensions):
		return fiber.NewError(fiber.StatusBadRequest, ErrInvalidDimensions)
	case errors.Is(err, upload.ErrQuotaExceeded):
		return fiber.NewError(fiber.StatusConflict, ErrEmojiQuotaExceeded)
	case errors.Is(err, upload.ErrUploadExpired):
		return fiber.NewError(fiber.StatusGone, ErrEmojiUploadExpired)
	case errors.Is(err, upload.ErrMediaProcess):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeEmoji)
	case errors.Is(err, upload.ErrStorage):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
	case errors.Is(err, upload.ErrFinalize):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeEmoji)
	default:
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeEmoji)
	}
}
