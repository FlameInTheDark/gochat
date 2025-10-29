package guild

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
)

// CreateIcon
//
//	@Summary		Create guild icon metadata
//	@Description	Creates an icon placeholder and returns upload info. Only guild owner may create.
//	@Tags			Guild
//	@Accept			json
//	@Produce		json
//	@Param			guild_id	path		int64				true	"Guild ID"
//	@Param			request		body		CreateIconRequest	true	"Icon creation request"
//	@Success		200			{object}	dto.IconUpload		"Icon upload data"
//	@failure		400			{string}	string				"Incorrect request body"
//	@failure		401			{string}	string				"Unauthorized"
//	@failure		403			{string}	string				"Forbidden"
//	@failure		500			{string}	string				"Something bad happened"
//	@Router			/guild/{guild_id}/icon [post]
func (e *entity) CreateIcon(c *fiber.Ctx) error {
	var req CreateIconRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	if err := req.Validate(); err != nil {
		return err
	}

	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	g, err := e.g.GetGuildById(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}
	if g.OwnerId != u.Id {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	id := idgen.Next()
	if err := e.icon.CreateIcon(c.UserContext(), id, guildId, e.attachTTL, req.FileSize); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateAttachment)
	}

	return c.JSON(dto.IconUpload{Id: id, GuildId: guildId})
}

// ListIcons
//
//	@Summary		List guild icons
//	@Description	Returns a list of previously created icons for a guild. Only the guild owner may access this list.
//	@Tags			Guild
//	@Produce		json
//	@Param			guild_id	path		int64		true	"Guild ID"
//	@Success		200			{array}		dto.Icon	"List of icons"
//	@failure		401			{string}	string		"Unauthorized"
//	@failure		403			{string}	string		"Forbidden"
//	@failure		500			{string}	string		"Internal server error"
//	@Router			/guild/{guild_id}/icons [get]
func (e *entity) ListIcons(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	g, err := e.g.GetGuildById(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}
	if g.OwnerId != u.Id {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	icons, err := e.icon.GetIconsByGuildId(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get icons")
	}

	resp := make([]dto.Icon, 0, len(icons))
	for _, ic := range icons {
		if !ic.Done || ic.URL == nil {
			continue
		}
		var urlStr string
		if ic.URL != nil {
			urlStr = *ic.URL
		}
		var w, h int64
		if ic.Width != nil {
			w = *ic.Width
		}
		if ic.Height != nil {
			h = *ic.Height
		}
		resp = append(resp, dto.Icon{
			Id:       ic.Id,
			URL:      urlStr,
			Filesize: ic.FileSize,
			Width:    w,
			Height:   h,
		})
	}

	sort.Slice(resp, func(i, j int) bool { return resp[i].Id > resp[j].Id })
	return c.JSON(resp)
}

// DeleteIcon
//
//	@Summary		Delete guild icon by ID
//	@Description	Deletes a guild icon. Only the guild owner may delete.
//	@Tags			Guild
//	@Param			guild_id	path		int64	true	"Guild ID"
//	@Param			icon_id		path		int64	true	"Icon ID"
//	@Success		200			{string}	string	"OK"
//	@failure		401			{string}	string	"Unauthorized"
//	@failure		403			{string}	string	"Forbidden"
//	@failure		500			{string}	string	"Internal server error"
//	@Router			/guild/{guild_id}/icons/{icon_id} [delete]
func (e *entity) DeleteIcon(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	iconId, err := e.parseIconID(c)
	if err != nil {
		return err
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	g, err := e.g.GetGuildById(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}
	if g.OwnerId != u.Id {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}
	// Prevent deleting currently selected icon
	if g.Icon != nil && *g.Icon == iconId {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrUnableToDeleteActiveIcon)
	}
	if err := e.icon.RemoveIcon(c.UserContext(), iconId, guildId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateAttachment)
	}
	_ = e.cache.Delete(c.UserContext(), fmt.Sprintf("icons:%d:%d", guildId, iconId))
	return c.SendStatus(fiber.StatusOK)
}

func (e *entity) parseIconID(c *fiber.Ctx) (int64, error) {
	iconIdStr := c.Params("icon_id")
	iconId, err := strconv.ParseInt(iconIdStr, 10, 64)
	if err != nil || iconId <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectIconID)
	}
	return iconId, nil
}
