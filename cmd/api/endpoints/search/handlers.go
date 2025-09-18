package search

import (
	"strconv"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
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

	_, gc, _, canRead, err := e.perm.ChannelPerm(c.UserContext(), guildId, req.ChannelId, user.Id, permissions.PermServerViewChannels, permissions.PermTextReadMessageHistory)
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
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFindMessages)
	}

	msgs, err := e.msg.GetChannelMessagesByIDs(c.UserContext(), req.ChannelId, res.Ids)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetMessages)
	}

	var uidsmap = make(map[int64]bool)
	var uids []int64

	for _, m := range msgs {
		if _, ok := uidsmap[m.UserId]; !ok {
			uids = append(uids, m.UserId)
			uidsmap[m.UserId] = true
		}
	}

	users, err := e.user.GetUsersList(c.UserContext(), uids)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUsers)
	}

	var umap = make(map[int64]*model.User)
	for _, u := range users {
		umap[u.Id] = &u
	}

	var resp = MessageSearchResponse{Pages: (res.Total + 10 - 1) / 10}
	for _, m := range msgs {
		if u, ok := umap[m.UserId]; ok {
			resp.Messages = append(resp.Messages, dto.Message{
				Id:        m.Id,
				ChannelId: m.ChannelId,
				Author: dto.User{
					Id:            u.Id,
					Name:          u.Name,
					Discriminator: "",
					Avatar:        u.Avatar,
				},
				Content:     m.Content,
				Attachments: nil,
				UpdatedAt:   m.EditedAt,
			})
			continue
		}
		resp.Messages = append(resp.Messages, dto.Message{
			Id:          m.Id,
			ChannelId:   m.ChannelId,
			Author:      dto.User{},
			Content:     m.Content,
			Attachments: nil,
			UpdatedAt:   m.EditedAt,
		})
	}

	return c.JSON(resp)
}
