package search

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/msgsearch"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

// Search
//
//	@Summary	Search messages
//	@Produce	json
//	@Tags		Search
//	@Param		guild_id	path		int64					true	"Channel id"
//	@Param		request		body		MessageSearchRequest	true	"Search request data"
//	@Success	200			{array}		MessageSearchResponse	"Messages"
//	@failure	400			{string}	string					"Bad request"
//	@failure	401			{string}	string					"Unauthorized"
//	@failure	403			{string}	string					"Forbidden"
//	@failure	500			{string}	string					"Internal server error"
//	@Router		/search/{guild_id}/messages [post]
func (e *entity) Search(c *fiber.Ctx) error {
	gidStr := c.Params("guild_id")
	guildId, err := strconv.ParseInt(gidStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}

	var req MessageSearchRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unable to get user token")
	}

	_, gc, _, canRead, err := e.perm.ChannelPerm(c.UserContext(), guildId, *req.ChannelId, user.Id, permissions.PermServerViewChannels, permissions.PermTextReadMessageHistory)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
	}
	if !canRead || gc == nil {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	res, err := e.search.Search(c.UserContext(), msgsearch.SearchRequest{
		GuildId:   guildId,
		ChannelId: req.ChannelId,
		UserId:    req.AuthorId,
		Content:   req.Content,
		Mentions:  req.Mentions,
		Has:       req.Has,
		From:      req.Page * 10,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(MessageSearchResponse{
		Ids:   res.Ids,
		Pages: res.Total / 10,
	})
}
