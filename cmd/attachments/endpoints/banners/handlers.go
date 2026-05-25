package banners

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

// Upload
//
//	@Summary		Upload user profile banner
//	@Description	Uploads a profile banner image. Requires source dimensions at least 680x240 and converts to WebP under 10MB.
//	@Tags			Upload
//	@Accept			application/octet-stream
//	@Produce		json
//	@Param			user_id		path		int64	true	"User ID"
//	@Param			banner_id	path		int64	true	"Banner ID"
//	@Param			crop_x		query		int64	false	"Crop X coordinate in source pixels"
//	@Param			crop_y		query		int64	false	"Crop Y coordinate in source pixels"
//	@Param			crop_width	query		int64	false	"Crop width in source pixels"
//	@Param			crop_height	query		int64	false	"Crop height in source pixels"
//	@Param			file		body		[]byte	true	"Binary image payload"
//	@Success		201			{string}	string	"Created"
//	@Success		204			{string}	string	"No Content (already uploaded)"
//	@failure		400			{string}	string	"Bad request"
//	@failure		401			{string}	string	"Unauthorized"
//	@failure		403			{string}	string	"Forbidden"
//	@failure		404			{string}	string	"Banner not found"
//	@failure		413			{string}	string	"File too large"
//	@failure		415			{string}	string	"Unsupported Media Type"
//	@failure		500			{string}	string	"Internal server error"
//	@Router			/upload/profile-covers/{user_id}/{banner_id} [post]
func (e *entity) Upload(c *fiber.Ctx) error {
	reqLog := observability.LoggerFromFiber(c, e.log)

	userId, err := strconv.ParseInt(c.Params("user_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectUserID)
	}
	bannerId, err := strconv.ParseInt(c.Params("banner_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectBannerID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	crop, err := parseBannerCrop(c)
	if err != nil {
		return err
	}

	body, err := requestBodyReader(c)
	if err != nil {
		return err
	}

	result, err := e.uploader.Upload(c.UserContext(), user.Id, userId, bannerId, body, crop)
	if err != nil {
		return bannerUploadError(err)
	}
	if result.AlreadyDone {
		return c.SendStatus(fiber.StatusNoContent)
	}

	go e.finalizeBannerSideEffects(observability.BackgroundFromContext(c.UserContext()), userId, bannerId, result, reqLog)

	return c.SendStatus(fiber.StatusCreated)
}

func parseBannerCrop(c *fiber.Ctx) (*upload.CropArea, error) {
	keys := []string{"crop_x", "crop_y", "crop_width", "crop_height"}
	hasCrop := false
	for _, key := range keys {
		if c.Query(key) != "" {
			hasCrop = true
			break
		}
	}
	if !hasCrop {
		return nil, nil
	}

	values := make([]int64, len(keys))
	for i, key := range keys {
		raw := c.Query(key)
		if raw == "" {
			return nil, fiber.NewError(fiber.StatusBadRequest, ErrInvalidDimensions)
		}
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, ErrInvalidDimensions)
		}
		values[i] = value
	}
	return &upload.CropArea{
		X:      values[0],
		Y:      values[1],
		Width:  values[2],
		Height: values[3],
	}, nil
}

func (e *entity) finalizeBannerSideEffects(ctx context.Context, userId, bannerId int64, result *upload.BannerResult, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	log := observability.LoggerWithContext(ctx, logger)

	if err := upload.Retry(ctx, 3, 100*time.Millisecond, func(ctx context.Context) error {
		return e.usr.SetUserBanner(ctx, userId, bannerId)
	}); err != nil {
		log.Error("failed to activate uploaded banner", "user_id", userId, "banner_id", bannerId, "error", err)
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
		bd := dto.BannerData{Exists: true, Id: bannerId, URL: result.URL, ContentType: &contentType, Width: &width, Height: &height, Size: result.Size}
		upd := mqmsg.UpdateUser{User: dto.User{
			Id:            u.Id,
			Name:          u.Name,
			Discriminator: "",
			Bio:           u.Bio,
			BannerColor:   u.BannerColor,
			PanelColor:    u.PanelColor,
			Banner:        &bd,
		}}
		return mq.SendUserUpdate(ctx, e.mqt, userId, &upd)
	}); err != nil {
		log.Error("failed to publish banner upload update", "user_id", userId, "banner_id", bannerId, "error", err)
	}
}

func bannerUploadError(err error) error {
	switch {
	case errors.Is(err, upload.ErrPlaceholderNotFound):
		return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetBanner)
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
	case errors.Is(err, upload.ErrMediaProcess):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToProcessImage)
	case errors.Is(err, upload.ErrStorage):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
	case errors.Is(err, upload.ErrFinalize):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeBanner)
	default:
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeBanner)
	}
}
