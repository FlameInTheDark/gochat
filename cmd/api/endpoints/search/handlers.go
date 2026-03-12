package search

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/embed"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/msgsearch"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

// SearchChannel
//
//	@Summary	Search messages in a channel
//	@Produce	json
//	@Tags		Search
//	@Param		request	body		MessageSearchRequest	true	"Search request data"
//	@Success	200		{array}		MessageSearchResponse	"Messages"
//	@failure	400		{string}	string					"Bad request"
//	@failure	401		{string}	string					"Unauthorized"
//	@failure	403		{string}	string					"Forbidden"
//	@failure	500		{string}	string					"Internal server error"
//	@Router		/search/messages [post]
func (e *entity) SearchChannel(c *fiber.Ctx) error {
	req, user, err := e.parseSearchRequest(c)
	if err != nil {
		return err
	}

	if _, err := e.authorizeSearchScope(c.UserContext(), req.ChannelId, user.Id, nil); err != nil {
		return err
	}

	return e.executeSearch(c, req, nil)
}

// Search
//
//	@Summary	Search messages
//	@Produce	json
//	@Tags		Search
//	@Param		guild_id	path		int64					true	"Guild id"	example(2230469276416868352)
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

	req, user, err := e.parseSearchRequest(c)
	if err != nil {
		return err
	}

	searchGuildID, err := e.authorizeSearchScope(c.UserContext(), req.ChannelId, user.Id, &guildId)
	if err != nil {
		return err
	}

	return e.executeSearch(c, req, searchGuildID)
}

func (e *entity) parseSearchRequest(c *fiber.Ctx) (*MessageSearchRequest, *helper.JWTUser, error) {
	var req MessageSearchRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	if err := req.Validate(); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusUnauthorized, "unable to get user token")
	}

	return &req, user, nil
}

func (e *entity) authorizeSearchScope(ctx context.Context, channelID, userID int64, requestedGuildID *int64) (*int64, error) {
	channel, err := e.ch.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "channel not found")
	}

	switch channel.Type {
	case model.ChannelTypeGuild, model.ChannelTypeThread:
		guildID := requestedGuildID
		if guildID == nil {
			guildChannel, err := e.gc.GetGuildByChannel(ctx, channelID)
			if err != nil {
				return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to get guild channel")
			}
			guildID = &guildChannel.GuildId
		}

		_, gc, _, canRead, err := e.perm.ChannelPerm(
			ctx,
			*guildID,
			channelID,
			userID,
			permissions.PermServerViewChannels,
			permissions.PermTextReadMessageHistory,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
			}
			return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
		}
		if !canRead || gc == nil {
			return nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
		}

		resolvedGuildID := gc.GuildId
		return &resolvedGuildID, nil

	case model.ChannelTypeDM, model.ChannelTypeGroupDM:
		if requestedGuildID != nil {
			return nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
		}

		_, _, _, canRead, err := e.perm.ChannelPerm(
			ctx,
			0,
			channelID,
			userID,
			permissions.PermServerViewChannels,
			permissions.PermTextReadMessageHistory,
		)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
		}
		if !canRead {
			return nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
		}

		return nil, nil

	default:
		return nil, fiber.NewError(fiber.StatusBadRequest, ErrUnsupportedChannel)
	}
}

func (e *entity) executeSearch(c *fiber.Ctx, req *MessageSearchRequest, guildID *int64) error {
	res, err := e.search.Search(c.UserContext(), msgsearch.SearchRequest{
		GuildId:   guildID,
		ChannelId: req.ChannelId,
		UserId:    req.AuthorId,
		Content:   req.Content,
		Mentions:  []int64(req.Mentions),
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

	var attIDs []int64
	attSeen := make(map[int64]struct{})
	for _, m := range msgs {
		for _, aid := range m.Attachments {
			if _, ok := attSeen[aid]; ok {
				continue
			}
			attSeen[aid] = struct{}{}
			attIDs = append(attIDs, aid)
		}
	}

	attMap := make(map[int64]model.Attachment)
	if len(attIDs) > 0 {
		ats, err := e.at.SelectAttachmentsByChannel(c.UserContext(), req.ChannelId, attIDs)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetMessages)
		}
		for _, a := range ats {
			attMap[a.Id] = a
		}
	}

	uidSeen := make(map[int64]struct{})
	var userIDs []int64
	for _, m := range msgs {
		if _, ok := uidSeen[m.UserId]; ok {
			continue
		}
		uidSeen[m.UserId] = struct{}{}
		userIDs = append(userIDs, m.UserId)
	}

	users, err := e.user.GetUsersList(c.UserContext(), userIDs)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUsers)
	}

	userMap := make(map[int64]*model.User)
	for i := range users {
		userMap[users[i].Id] = &users[i]
	}

	resp := MessageSearchResponse{Pages: (res.Total + 10 - 1) / 10}
	for _, m := range msgs {
		var dtoAts []dto.Attachment
		if len(m.Attachments) > 0 {
			for _, aid := range m.Attachments {
				if a, ok := attMap[aid]; ok {
					var full string
					if a.URL != nil {
						full = *a.URL
					}
					var previewFull *string
					if a.PreviewURL != nil && *a.PreviewURL != "" {
						previewFull = a.PreviewURL
					}
					dtoAts = append(dtoAts, dto.Attachment{
						ContentType: a.ContentType,
						Filename:    a.Name,
						Height:      a.Height,
						Width:       a.Width,
						URL:         full,
						PreviewURL:  previewFull,
						Size:        a.FileSize,
					})
				}
			}
		}

		flags := model.NormalizeMessageFlags(m.Flags)
		embeds, err := embed.ParseMergedEmbeds(m.EmbedsJSON, m.AutoEmbedsJSON, model.HasMessageFlag(flags, model.MessageFlagSuppressEmbeds))
		if err != nil {
			if e.log != nil {
				e.log.Error("failed to decode message embeds", "message_id", m.Id, "error", err.Error())
			}
			embeds = nil
		}

		if u, ok := userMap[m.UserId]; ok {
			resp.Messages = append(resp.Messages, dto.Message{
				Id:        m.Id,
				ChannelId: m.ChannelId,
				Author: dto.User{
					Id:            u.Id,
					Name:          u.Name,
					Discriminator: "",
				},
				Content:     m.Content,
				Position:    optionalInt64(m.Position),
				Attachments: dtoAts,
				Embeds:      embeds,
				Flags:       flags,
				Type:        m.Type,
				Reference:   optionalInt64(m.Reference),
				ThreadId:    optionalInt64(m.Thread),
				UpdatedAt:   m.EditedAt,
			})
			continue
		}

		resp.Messages = append(resp.Messages, dto.Message{
			Id:          m.Id,
			ChannelId:   m.ChannelId,
			Author:      dto.User{},
			Content:     m.Content,
			Position:    optionalInt64(m.Position),
			Attachments: dtoAts,
			Embeds:      embeds,
			Flags:       flags,
			Type:        m.Type,
			Reference:   optionalInt64(m.Reference),
			ThreadId:    optionalInt64(m.Thread),
			UpdatedAt:   m.EditedAt,
		})
	}

	return c.JSON(resp)
}

func optionalInt64(value int64) *int64 {
	if value == 0 {
		return nil
	}
	v := value
	return &v
}
