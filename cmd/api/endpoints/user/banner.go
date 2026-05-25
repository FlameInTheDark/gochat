package user

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
)

const profileBannerMaxSizeBytes = 10 * 1024 * 1024

type CreateBannerRequest struct {
	FileSize    int64  `json:"file_size" example:"1048576"`
	ContentType string `json:"content_type" example:"image/png"`
}

func (r CreateBannerRequest) Validate() error {
	if r.FileSize <= 0 || r.FileSize > profileBannerMaxSizeBytes {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
	}
	if !strings.HasPrefix(strings.ToLower(r.ContentType), "image/") {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, ErrUnsupportedContentType)
	}
	return nil
}

// CreateBanner
//
//	@Summary		Create profile banner metadata
//	@Description	Creates a profile banner placeholder and returns upload info. Upload the binary to attachments service.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateBannerRequest	true	"Banner creation request"
//	@Success		200		{object}	dto.BannerUpload	"Banner upload data"
//	@Router			/user/me/banner [post]
func (e *entity) CreateBanner(c *fiber.Ctx) error {
	var req CreateBannerRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}
	if err := req.Validate(); err != nil {
		return err
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	id := idgen.Next()
	if err := e.bn.CreateBanner(c.UserContext(), id, u.Id, e.attachTTL, req.FileSize); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateBanner)
	}
	return c.JSON(dto.BannerUpload{Id: id, UserId: u.Id})
}
