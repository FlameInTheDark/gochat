package user

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
)

// Request for creating avatar metadata
type CreateAvatarRequest struct {
	FileSize    int64  `json:"file_size" example:"120000"`
	ContentType string `json:"content_type" example:"image/png"`
}

// Validate constraints: <= 250KB and image/*
func (r CreateAvatarRequest) Validate() error {
	if r.FileSize <= 0 || r.FileSize > 250*1024 {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "file is too big")
	}
	if len(r.ContentType) < 6 || r.ContentType[:6] != "image/" {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, "unsupported content type")
	}
	return nil
}

// CreateAvatar
//
//	@Summary		Create avatar metadata
//	@Description	Creates an avatar placeholder and returns upload info. Upload the binary to attachments service.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateAvatarRequest	true	"Avatar creation request"
//	@Success		200		{object}	dto.AvatarUpload	"Avatar upload data"
//	@Router			/user/me/avatar [post]
func (e *entity) CreateAvatar(c *fiber.Ctx) error {
	var req CreateAvatarRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "unable to parse body")
	}
	if err := req.Validate(); err != nil {
		return err
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "unable to get user token")
	}
	id := idgen.Next()

	// Create placeholder avatar in Cassandra
	if err := e.av.CreateAvatar(c.UserContext(), id, u.Id, e.attachTTL, req.FileSize); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to create avatar")
	}

	return c.JSON(dto.AvatarUpload{
		Id:     id,
		UserId: u.Id,
	})
}

// ListAvatars
//
//	@Summary		List my avatars
//	@Description	Returns a list of previously created avatars for the authenticated user.
//	@Tags			User
//	@Produce		json
//	@Success		200	{array}		dto.Avatar	"List of avatars"
//	@failure		401	{string}	string		"Unauthorized"
//	@Router			/user/me/avatars [get]
func (e *entity) ListAvatars(c *fiber.Ctx) error {
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	items, err := e.av.GetAvatarsByUserId(c.UserContext(), u.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUser)
	}
	// Build response, include only finalized (done) with URL
	var resp []dto.Avatar
	resp = make([]dto.Avatar, 0, len(items))
	for _, a := range items {
		if !a.Done || a.URL == nil {
			continue
		}
		url := ""
		if a.URL != nil {
			url = *a.URL
		}
		resp = append(resp, dto.Avatar{
			Id:          a.Id,
			URL:         url,
			ContentType: a.ContentType,
			Width:       a.Width,
			Height:      a.Height,
			Size:        a.FileSize,
		})
	}
	// Sort by ID desc (newest first) for convenience
	sort.Slice(resp, func(i, j int) bool { return resp[i].Id > resp[j].Id })
	return c.JSON(resp)
}

// DeleteAvatar
//
//	@Summary	Delete my avatar by ID
//	@Tags		User
//	@Param		avatar_id	path		int64	true	"Avatar ID"
//	@Success	200			{string}	string	"OK"
//	@failure	400			{string}	string	"Bad request"
//	@failure	401			{string}	string	"Unauthorized"
//	@failure	500			{string}	string	"Internal server error"
//	@Router		/user/me/avatars/{avatar_id} [delete]
func (e *entity) DeleteAvatar(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	avatarIdStr := c.Params("avatar_id")
	avatarId, err := strconv.ParseInt(avatarIdStr, 10, 64)
	if err != nil || avatarId <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseID)
	}

	// Prevent deleting active avatar
	u, err := e.user.GetUserById(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUser)
	}
	if u.Avatar != nil && *u.Avatar == avatarId {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrUnableToDeleteActiveAvatar)
	}

	if err := e.av.RemoveAvatar(c.UserContext(), avatarId, user.Id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToModifyUser)
	}

	// Clean cached avatar data if present
	_ = e.cache.Delete(c.UserContext(), fmt.Sprintf("avatars:%d:%d", user.Id, avatarId))
	return c.SendStatus(fiber.StatusOK)
}
