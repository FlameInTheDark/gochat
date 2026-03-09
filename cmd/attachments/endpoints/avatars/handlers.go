package avatars

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
//	@Summary		Upload user avatar
//	@Description	Uploads an avatar image. Resizes to max 128x128 and converts to WebP <= 250KB. Finalizes avatar metadata.
//	@Tags			Upload
//	@Accept			application/octet-stream
//	@Produce		json
//	@Param			user_id		path		int64	true	"User ID"
//	@Param			avatar_id	path		int64	true	"Avatar ID"
//	@Param			file		body		[]byte	true	"Binary image payload"
//	@Success		201			{string}	string	"Created"
//	@Success		204			{string}	string	"No Content (already uploaded)"
//	@failure		400			{string}	string	"Bad request"
//	@failure		401			{string}	string	"Unauthorized"
//	@failure		403			{string}	string	"Forbidden"
//	@failure		404			{string}	string	"Avatar not found"
//	@failure		413			{string}	string	"File too large"
//	@failure		415			{string}	string	"Unsupported Media Type"
//	@failure		500			{string}	string	"Internal server error"
//	@Router			/upload/avatars/{user_id}/{avatar_id} [post]
func (e *entity) Upload(c *fiber.Ctx) error {
	userId, err := strconv.ParseInt(c.Params("user_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectUserID)
	}
	avatarId, err := strconv.ParseInt(c.Params("avatar_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectAvatarID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	body, err := requestBodyReader(c)
	if err != nil {
		return err
	}

	result, err := e.uploader.Upload(c.UserContext(), user.Id, userId, avatarId, body)
	if err != nil {
		return avatarUploadError(err)
	}
	if result.AlreadyDone {
		return c.SendStatus(fiber.StatusNoContent)
	}

	go e.finalizeAvatarSideEffects(userId, avatarId, result)

	return c.SendStatus(fiber.StatusCreated)
}

func (e *entity) finalizeAvatarSideEffects(userId, avatarId int64, result *upload.AvatarResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := upload.Retry(ctx, 3, 100*time.Millisecond, func(ctx context.Context) error {
		return e.usr.SetUserAvatar(ctx, userId, avatarId)
	}); err != nil {
		e.log.Error("failed to activate uploaded avatar", "user_id", userId, "avatar_id", avatarId, "error", err)
		return
	}

	if err := upload.Retry(ctx, 3, 100*time.Millisecond, func(ctx context.Context) error {
		u, err := e.usr.GetUserById(ctx, userId)
		if err != nil {
			return err
		}
		contentType := result.ContentType
		width := result.Width
		height := result.Height
		ad := dto.AvatarData{URL: result.URL, ContentType: &contentType, Width: &width, Height: &height, Size: result.Size}
		upd := mqmsg.UpdateUser{User: dto.User{Id: u.Id, Name: u.Name, Discriminator: "", Avatar: &ad}}
		return e.mqt.SendUserUpdate(userId, &upd)
	}); err != nil {
		e.log.Error("failed to publish avatar upload update", "user_id", userId, "avatar_id", avatarId, "error", err)
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

func avatarUploadError(err error) error {
	switch {
	case errors.Is(err, upload.ErrPlaceholderNotFound):
		return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetAvatar)
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
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeAvatar)
	default:
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeAvatar)
	}
}
