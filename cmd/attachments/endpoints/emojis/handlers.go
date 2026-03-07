package emojis

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/dto"
	emojiutil "github.com/FlameInTheDark/gochat/internal/emoji"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

// Upload
//
//	@Summary		Upload guild emoji image
//	@Description	Uploads and finalizes the binary payload for a previously created guild emoji placeholder.
//	@Tags			Upload
//	@Accept			application/octet-stream
//	@Produce		json
//	@Param			guild_id	path		int64	true	"Guild ID"
//	@Param			emoji_id	path		int64	true	"Emoji ID"
//	@Param			file		body		[]byte	true	"Binary emoji image"
//	@Success		201			{string}	string	"Created"
//	@Success		204			{string}	string	"No Content (already uploaded)"
//	@failure		400			{string}	string	"Bad request"
//	@failure		401			{string}	string	"Unauthorized"
//	@failure		403			{string}	string	"Forbidden"
//	@failure		404			{string}	string	"Emoji not found"
//	@failure		409			{string}	string	"Conflict"
//	@failure		410			{string}	string	"Upload expired"
//	@failure		413			{string}	string	"File too large"
//	@failure		415			{string}	string	"Unsupported media type"
//	@failure		500			{string}	string	"Internal server error"
//	@Router			/upload/emojis/{guild_id}/{emoji_id} [post]
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
		if e.log != nil {
			e.log.Error("emoji upload failed",
				slog.String("error", err.Error()),
				slog.Int64("guild_id", guildId),
				slog.Int64("emoji_id", emojiId),
				slog.Int64("user_id", user.Id),
			)
		}
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
