package developer

import (
	"strings"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/gofiber/fiber/v2"
)

const (
	profileAvatarMaxSizeBytes = 250 * 1024
	profileBannerMaxSizeBytes = 10 * 1024 * 1024
)

type CreateBotAvatarRequest struct {
	FileSize    int64  `json:"file_size" example:"120000"`
	ContentType string `json:"content_type" example:"image/png"`
}

func (r CreateBotAvatarRequest) Validate() error {
	if r.FileSize <= 0 || r.FileSize > profileAvatarMaxSizeBytes {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "file is too big")
	}
	if !strings.HasPrefix(strings.ToLower(r.ContentType), "image/") {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, "unsupported content type")
	}
	return nil
}

type CreateBotBannerRequest struct {
	FileSize    int64  `json:"file_size" example:"1048576"`
	ContentType string `json:"content_type" example:"image/png"`
}

func (r CreateBotBannerRequest) Validate() error {
	if r.FileSize <= 0 || r.FileSize > profileBannerMaxSizeBytes {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "file is too big")
	}
	if !strings.HasPrefix(strings.ToLower(r.ContentType), "image/") {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, "unsupported content type")
	}
	return nil
}

// CreateBotAvatar
//
//	@Summary	Create bot avatar upload metadata
//	@Accept		json
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		bot_id	path		int64					true	"Bot user id"
//	@Param		request	body		CreateBotAvatarRequest	true	"Avatar upload request"
//	@Success	200		{object}	dto.AvatarUpload
//	@Failure	400		{string}	string	"Bad request"
//	@Failure	401		{string}	string	"Unauthorized"
//	@Failure	404		{string}	string	"Not found"
//	@Failure	413		{string}	string	"File too large"
//	@Failure	415		{string}	string	"Unsupported media type"
//	@Failure	500		{string}	string	"Internal server error"
//	@Router		/developer/bots/{bot_id}/avatar [post]
func (e *entity) CreateBotAvatar(c *fiber.Ctx) error {
	bot, err := e.requireOwnerBot(c)
	if err != nil {
		return err
	}
	var req CreateBotAvatarRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errBadRequest)
	}
	if err := req.Validate(); err != nil {
		return err
	}

	avatarID := idgen.Next()
	if err := e.av.CreateAvatar(c.UserContext(), avatarID, bot.BotUserId, e.attachTTL, req.FileSize); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to create bot avatar")
	}
	return c.JSON(dto.AvatarUpload{Id: avatarID, UserId: bot.BotUserId})
}

// CreateBotBanner
//
//	@Summary	Create bot banner upload metadata
//	@Accept		json
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		bot_id	path		int64					true	"Bot user id"
//	@Param		request	body		CreateBotBannerRequest	true	"Banner upload request"
//	@Success	200		{object}	dto.BannerUpload
//	@Failure	400		{string}	string	"Bad request"
//	@Failure	401		{string}	string	"Unauthorized"
//	@Failure	404		{string}	string	"Not found"
//	@Failure	413		{string}	string	"File too large"
//	@Failure	415		{string}	string	"Unsupported media type"
//	@Failure	500		{string}	string	"Internal server error"
//	@Router		/developer/bots/{bot_id}/banner [post]
func (e *entity) CreateBotBanner(c *fiber.Ctx) error {
	bot, err := e.requireOwnerBot(c)
	if err != nil {
		return err
	}
	var req CreateBotBannerRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errBadRequest)
	}
	if err := req.Validate(); err != nil {
		return err
	}

	bannerID := idgen.Next()
	if err := e.bn.CreateBanner(c.UserContext(), bannerID, bot.BotUserId, e.attachTTL, req.FileSize); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to create bot banner")
	}
	return c.JSON(dto.BannerUpload{Id: bannerID, UserId: bot.BotUserId})
}

func (e *entity) avatarData(c *fiber.Ctx, userID int64, avatarID *int64) *dto.AvatarData {
	if avatarID == nil {
		return nil
	}
	avatar, err := e.av.GetAvatar(c.UserContext(), *avatarID, userID)
	if err != nil || !avatar.Done || avatar.URL == nil {
		return nil
	}
	return &dto.AvatarData{
		Id:          avatar.Id,
		URL:         *avatar.URL,
		ContentType: avatar.ContentType,
		Width:       avatar.Width,
		Height:      avatar.Height,
		Size:        avatar.FileSize,
	}
}

func (e *entity) bannerData(c *fiber.Ctx, userID int64, bannerID *int64) *dto.BannerData {
	if bannerID == nil {
		return &dto.BannerData{Exists: false}
	}
	banner, err := e.bn.GetBanner(c.UserContext(), *bannerID, userID)
	if err != nil || !banner.Done || banner.URL == nil {
		return &dto.BannerData{Exists: false}
	}
	return &dto.BannerData{
		Exists:      true,
		Id:          banner.Id,
		URL:         *banner.URL,
		ContentType: banner.ContentType,
		Width:       banner.Width,
		Height:      banner.Height,
		Size:        banner.FileSize,
	}
}
