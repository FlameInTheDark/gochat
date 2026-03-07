package icons

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

// Upload
//
//	@Summary		Upload guild icon
//	@Description	Uploads a guild icon. Resizes to max 128x128 and converts to WebP <= 250KB. Only guild owner can upload. Sets guild icon and emits update.
//	@Tags			Upload
//	@Accept			application/octet-stream
//	@Produce		json
//	@Param			guild_id	path		int64	true	"Guild ID"
//	@Param			icon_id		path		int64	true	"Icon ID"
//	@Param			file		body		[]byte	true	"Binary image payload"
//	@Success		201			{string}	string	"Created"
//	@Success		204			{string}	string	"No Content (already uploaded)"
//	@failure		400			{string}	string	"Bad request"
//	@failure		401			{string}	string	"Unauthorized"
//	@failure		403			{string}	string	"Forbidden"
//	@failure		404			{string}	string	"Icon not found"
//	@failure		413			{string}	string	"File too large"
//	@failure		415			{string}	string	"Unsupported Media Type"
//	@failure		500			{string}	string	"Internal server error"
//	@Router			/upload/icons/{guild_id}/{icon_id} [post]
func (e *entity) Upload(c *fiber.Ctx) error {
	guildId, err := strconv.ParseInt(c.Params("guild_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}
	iconId, err := strconv.ParseInt(c.Params("icon_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectIconID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	g, err := e.gld.GetGuildById(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuild)
	}
	if g.OwnerId != user.Id {
		return fiber.NewError(fiber.StatusForbidden, ErrForbiddenToUpload)
	}

	body, err := requestBodyReader(c)
	if err != nil {
		return err
	}

	result, err := e.uploader.Upload(c.UserContext(), guildId, iconId, body)
	if err != nil {
		return iconUploadError(err)
	}
	if result.AlreadyDone {
		return c.SendStatus(fiber.StatusNoContent)
	}

	go e.finalizeIconSideEffects(guildId, iconId, result)

	return c.SendStatus(fiber.StatusCreated)
}

func (e *entity) finalizeIconSideEffects(guildId, iconId int64, result *upload.IconResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := upload.Retry(ctx, 3, 100*time.Millisecond, func(ctx context.Context) error {
		return e.gld.SetGuildIcon(ctx, guildId, iconId)
	}); err != nil {
		e.log.Error("failed to activate uploaded guild icon", "guild_id", guildId, "icon_id", iconId, "error", err)
		return
	}

	if err := upload.Retry(ctx, 3, 100*time.Millisecond, func(ctx context.Context) error {
		guild, err := e.gld.GetGuildById(ctx, guildId)
		if err != nil {
			return err
		}
		icon := dto.Icon{Id: iconId, URL: result.URL, Filesize: result.Size, Width: result.Width, Height: result.Height}
		upd := mqmsg.UpdateGuild{Guild: dto.Guild{Id: guild.Id, Name: guild.Name, Icon: &icon, Owner: guild.OwnerId, Public: guild.Public, Permissions: guild.Permissions}}
		return e.mqt.SendGuildUpdate(guildId, &upd)
	}); err != nil {
		e.log.Error("failed to publish guild icon update", "guild_id", guildId, "icon_id", iconId, "error", err)
	}
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

func iconUploadError(err error) error {
	switch {
	case errors.Is(err, upload.ErrPlaceholderNotFound):
		return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetIcon)
	case errors.Is(err, upload.ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, ErrForbiddenToUpload)
	case errors.Is(err, upload.ErrEmptyBody), errors.Is(err, upload.ErrSizeMismatch):
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToReadBody)
	case errors.Is(err, upload.ErrTooLarge):
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
	case errors.Is(err, upload.ErrUnsupportedMedia):
		return fiber.NewError(fiber.StatusUnsupportedMediaType, ErrUnsupportedContentType)
	case errors.Is(err, upload.ErrMediaProcess):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToProcessImage)
	case errors.Is(err, upload.ErrStorage):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
	case errors.Is(err, upload.ErrFinalize):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeIcon)
	default:
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeIcon)
	}
}
