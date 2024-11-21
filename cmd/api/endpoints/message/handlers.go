package message

import (
	"errors"
	"fmt"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"log/slog"
	"strconv"

	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

// Send
//
//	@Summary	Get message
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id	path		int64				true	"Channel id"
//	@Param		request		body		SendMessageRequest	true	"Message data"
//	@Success	200			{object}	dto.Message			"Message"
//	@failure	400			{string}	string				"Incorrect request body"
//	@failure	401			{string}	string				"Unauthorized"
//	@failure	500			{string}	string				"Something bad happened"
//	@Router		/message/channel/{channel_id} [post]
func (e *entity) Send(c *fiber.Ctx) error {
	var req SendMessageRequest
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	channelId := c.Params("channel_id")
	id, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	ch, err := e.ch.GetChannel(c.UserContext(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if ch.Type == model.ChannelTypeGuildCategory || ch.Type == model.ChannelTypeGuildVoice {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSentToThisChannel)
	}
	var ok = true
	var guildId *int64
	if ch.Type == model.ChannelTypeGuild {
		gc, err := e.gc.GetGuildByChannel(c.UserContext(), id)
		if err != nil && !errors.Is(err, gocql.ErrNotFound) {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if !errors.Is(err, gocql.ErrNotFound) {
			guildId = &gc.GuildId
			_, _, _, ok, err = e.perm.ChannelPerm(c.UserContext(), gc.GuildId, gc.ChannelId, u.Id, permissions.PermTextSendMessage)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
		}
	}
	if ok {
		user, err := e.user.GetUserById(c.UserContext(), u.Id)
		if err := helper.HttpDbError(err, ErrUnableToGetUser); err != nil {
			return err
		}
		disc, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), u.Id)
		if err := helper.HttpDbError(err, ErrUnableToGetUserDiscriminator); err != nil {
			return err
		}
		msgid := idgen.Next()
		err = e.msg.CreateMessage(c.UserContext(), msgid, id, u.Id, req.Content, req.Attachments)
		if err := helper.HttpDbError(err, ErrUnableToSendMessage); err != nil {
			return err
		}

		ats, err := e.at.SelectAttachemntsByIDs(c.UserContext(), req.Attachments)
		if err := helper.HttpDbError(err, ErrUnableToGetAttachements); err != nil {
			return err
		}

		resp := dto.Message{
			Id:        msgid,
			ChannelId: id,
			Author: dto.User{
				Id:            user.Id,
				Name:          user.Name,
				Discriminator: disc.Discriminator,
				Avatar:        user.Avatar,
			},
			Content:     req.Content,
			Attachments: nil,
		}

		for _, at := range ats {
			resp.Attachments = append(resp.Attachments, dto.Attachment{
				ContentType: at.ContentType,
				Filename:    at.Name,
				Height:      at.Height,
				Width:       at.Width,
				URL:         fmt.Sprintf("media/%d/%d/%s", at.ChannelId, at.Id, at.Name),
				Size:        at.FileSize,
			})
		}

		err = e.mqt.SendChannelMessage(id, &mqmsg.CreateMessage{
			GuildId: guildId,
			Message: resp,
		})
		if err != nil {
			remerr := e.msg.DeleteMessage(c.UserContext(), msgid)
			e.log.Error("unable to send message event", slog.String("error", errors.Join(err, remerr).Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
		}

		return c.JSON(resp)
	}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
}

// Update
//
//	@Summary	Update message
//	@Produce	json
//	@Tags		Message
//	@Param		message_id	path		int64					true	"Message id"
//	@Param		request		body		UpdateMessageRequest	true	"Message data"
//	@Success	200			{object}	dto.Message				"Message"
//	@failure	400			{string}	string					"Incorrect request body"
//	@failure	401			{string}	string					"Unauthorized"
//	@failure	500			{string}	string					"Something bad happened"
//	@Router		/message/{message_id} [post]
func (e *entity) Update(c *fiber.Ctx) error {
	var req UpdateMessageRequest
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	channelId := c.Params("channel_id")
	id, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	ch, err := e.ch.GetChannel(c.UserContext(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	var ok = true
	var guildId *int64
	if ch.Type == model.ChannelTypeGuild {
		gc, err := e.gc.GetGuildByChannel(c.UserContext(), id)
		if err != nil && !errors.Is(err, gocql.ErrNotFound) {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if !errors.Is(err, gocql.ErrNotFound) {
			guildId = &gc.GuildId
			_, _, _, ok, err = e.perm.ChannelPerm(c.UserContext(), gc.GuildId, gc.ChannelId, u.Id, permissions.PermServerManageChannels)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
		}
	} else if ch.Type == model.ChannelTypeGroupDM && ch.ParentID != nil && *ch.ParentID != u.Id {
		ok = false
	}
	if ok {
		user, err := e.user.GetUserById(c.UserContext(), u.Id)
		if err := helper.HttpDbError(err, ErrUnableToGetUser); err != nil {
			return err
		}
		disc, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), u.Id)
		if err := helper.HttpDbError(err, ErrUnableToGetUserDiscriminator); err != nil {
			return err
		}
		msgid := idgen.Next()
		err = e.msg.CreateMessage(c.UserContext(), msgid, id, u.Id, req.Content, req.Attachments)
		if err := helper.HttpDbError(err, ErrUnableToSendMessage); err != nil {
			return err
		}

		ats, err := e.at.SelectAttachemntsByIDs(c.UserContext(), req.Attachments)
		if err := helper.HttpDbError(err, ErrUnableToGetAttachements); err != nil {
			return err
		}

		resp := dto.Message{
			Id:        msgid,
			ChannelId: id,
			Author: dto.User{
				Id:            user.Id,
				Name:          user.Name,
				Discriminator: disc.Discriminator,
				Avatar:        user.Avatar,
			},
			Content:     req.Content,
			Attachments: nil,
		}

		for _, at := range ats {
			resp.Attachments = append(resp.Attachments, dto.Attachment{
				ContentType: at.ContentType,
				Filename:    at.Name,
				Height:      at.Height,
				Width:       at.Width,
				URL:         fmt.Sprintf("media/%d/%d/%s", at.ChannelId, at.Id, at.Name),
				Size:        at.FileSize,
			})
		}

		err = e.mqt.SendChannelMessage(id, &mqmsg.CreateMessage{
			GuildId: guildId,
			Message: resp,
		})
		if err != nil {
			remerr := e.msg.DeleteMessage(c.UserContext(), msgid)
			e.log.Error("unable to send message event", slog.String("error", errors.Join(err, remerr).Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
		}

		return c.JSON(resp)
	}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
}

// Attachment
//
//	@Summary	Create attachment
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id	path		int64					true	"Channel id"
//	@Param		request		body		UploadAttachmentRequest	true	"Attachment data"
//	@Success	200			{object}	dto.AttachmentUpload	"Attachment upload data"
//	@failure	400			{string}	string					"Incorrect request body"
//	@failure	401			{string}	string					"Unauthorized"
//	@failure	500			{string}	string					"Something bad happened"
//	@Router		/message/channel/{channel_id}/attachment [post]
func (e *entity) Attachment(c *fiber.Ctx) error {
	channelId := c.Params("channel_id")
	id, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	var req UploadAttachmentRequest
	err = c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	dbu, err := e.user.GetUserById(c.UserContext(), u.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUser)
	}
	switch dbu.UploadLimit == nil {
	case true:
		if req.FileSize > e.uploadLimit {
			return fiber.NewError(fiber.StatusNotAcceptable, ErrFileIsTooBig)
		}
	case false:
		if req.FileSize > *dbu.UploadLimit {
			return fiber.NewError(fiber.StatusNotAcceptable, ErrFileIsTooBig)
		}
	}

	ch, err := e.ch.GetChannel(c.UserContext(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	var ok = true
	if ch.Type == model.ChannelTypeGuild {
		gc, err := e.gc.GetGuildByChannel(c.UserContext(), id)
		if err != nil && !errors.Is(err, gocql.ErrNotFound) {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if !errors.Is(err, gocql.ErrNotFound) {
			_, _, _, ok, err = e.perm.ChannelPerm(c.UserContext(), gc.GuildId, gc.ChannelId, u.Id, permissions.PermTextAttachFiles)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
		}
	} else if ch.Type == model.ChannelTypeGroupDM && ch.ParentID != nil && *ch.ParentID != u.Id {
		ok = false
	}

	if ok {
		atid := idgen.Next()
		url, err := e.storage.MakeUploadAttachment(c.UserContext(), id, atid, req.FileSize, req.Filename)
		if err != nil {
			e.log.Error(err.Error())
			return fiber.NewError(fiber.StatusNotAcceptable, ErrUnableToCreateUploadURL)
		}
		err = e.at.CreateAttachment(c.UserContext(), atid, id, req.FileSize, req.Height, req.Width, req.Filename)
		if err := helper.HttpDbError(err, ErrUnableToCreateAttachment); err != nil {
			return err
		}
		return c.JSON(dto.AttachmentUpload{
			Id:        atid,
			ChannelId: id,
			FileName:  req.Filename,
			UploadURL: url,
		})
	}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
}
