package attachments

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

// Upload
//
//	@Summary		Upload attachment
//	@Description	Uploads a file for an existing attachment. Stores the original as-is and generates a WebP preview for images/videos. Finalizes the attachment metadata.
//	@Tags			Upload
//	@Accept			application/octet-stream
//	@Produce		json
//	@Param			channel_id		path		int64	true	"Channel ID"
//	@Param			attachment_id	path		int64	true	"Attachment ID"
//	@Param			file			body		[]byte	true	"Binary file to upload"
//	@Success		201				{string}	string	"Created"
//	@Success		204				{string}	string	"No Content (already uploaded)"
//	@failure		400				{string}	string	"Bad request"
//	@failure		401				{string}	string	"Unauthorized"
//	@failure		403				{string}	string	"Forbidden"
//	@failure		404				{string}	string	"Attachment not found"
//	@failure		413				{string}	string	"File too large"
//	@failure		415				{string}	string	"Unsupported media type"
//	@failure		500				{string}	string	"Internal server error"
//	@Router			/upload/attachments/{channel_id}/{attachment_id} [post]
func (e *entity) Upload(c *fiber.Ctx) error {
	channelId, err := strconv.ParseInt(c.Params("channel_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}

	attachmentId, err := strconv.ParseInt(c.Params("attachment_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectAttachmentID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	body, err := requestBodyReader(c)
	if err != nil {
		return err
	}

	result, err := e.uploader.Upload(c.UserContext(), user.Id, channelId, attachmentId, body)
	if err != nil {
		return attachmentUploadError(err)
	}
	if result.AlreadyDone {
		return c.SendStatus(fiber.StatusNoContent)
	}

	kind := result.Kind
	if kind == "" {
		kind = "other"
	}
	incTransferred(kind, result.Size)

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

func attachmentUploadError(err error) error {
	switch {
	case errors.Is(err, upload.ErrPlaceholderNotFound):
		return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetAttachment)
	case errors.Is(err, upload.ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, ErrForbiddenToUpload)
	case errors.Is(err, upload.ErrEmptyBody), errors.Is(err, upload.ErrSizeMismatch):
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToReadBody)
	case errors.Is(err, upload.ErrTooLarge):
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
	case errors.Is(err, upload.ErrUnsupportedMedia):
		return fiber.NewError(fiber.StatusUnsupportedMediaType, ErrUnsupportedContentType)
	case errors.Is(err, upload.ErrMediaProcess):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToProcessMedia)
	case errors.Is(err, upload.ErrStorage):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUploadToStorage)
	case errors.Is(err, upload.ErrFinalize):
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeAttachment)
	default:
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFinalizeAttachment)
	}
}
