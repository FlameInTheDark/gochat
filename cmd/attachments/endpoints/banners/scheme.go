package banners

import (
	"bytes"
	"io"

	"github.com/gofiber/fiber/v2"
)

const (
	ErrUnableToGetUserToken    = "unable to get user token"
	ErrIncorrectUserID         = "incorrect user ID"
	ErrIncorrectBannerID       = "incorrect banner ID"
	ErrUnableToGetBanner       = "unable to get banner"
	ErrForbiddenToUpload       = "forbidden to upload"
	ErrUnableToReadBody        = "unable to read request body"
	ErrFileIsTooBig            = "file is too big"
	ErrUnsupportedContentType  = "unsupported content type"
	ErrInvalidDimensions       = "banner must be at least 680x240"
	ErrUnableToProcessImage    = "unable to process image"
	ErrUnableToUploadToStorage = "unable to upload to storage"
	ErrUnableToFinalizeBanner  = "unable to finalize banner"
)

const (
	bannerMaxSizeBytes = 10 * 1024 * 1024
	bannerMaxDim       = 1920
	bannerMinWidth     = 680
	bannerMinHeight    = 240
)

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
