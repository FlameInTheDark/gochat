package message

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
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
	chid, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	ch, err := e.ch.GetChannel(c.UserContext(), chid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if ch.Type == model.ChannelTypeGuildCategory || ch.Type == model.ChannelTypeGuildVoice {
		return fiber.NewError(fiber.StatusUnauthorized, ErrUnableToSentToThisChannel)
	}
	var ok = true
	var guildId *int64
	if ch.Type == model.ChannelTypeGuild {
		gc, err := e.gc.GetGuildByChannel(c.UserContext(), chid)
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
		err = e.msg.CreateMessage(c.UserContext(), msgid, chid, u.Id, req.Content, req.Attachments)
		if err := helper.HttpDbError(err, ErrUnableToSendMessage); err != nil {
			return err
		}

		err = e.ch.SetLastMessage(c.UserContext(), chid, msgid)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		ats, err := e.at.SelectAttachmentByIDs(c.UserContext(), req.Attachments)
		if err := helper.HttpDbError(err, ErrUnableToGetAttachements); err != nil {
			return err
		}

		resp := dto.Message{
			Id:        msgid,
			ChannelId: ch.Id,
			Author: dto.User{
				Id:            user.Id,
				Name:          user.Name,
				Discriminator: disc.Discriminator,
				Avatar:        user.Avatar,
			},
			Content: req.Content,
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

		err = e.mqt.SendChannelMessage(chid, &mqmsg.CreateMessage{
			GuildId: guildId,
			Message: resp,
		})
		if err != nil {
			remerr := e.msg.DeleteMessage(c.UserContext(), msgid, chid)
			e.log.Error("unable to send message event", slog.String("error", errors.Join(err, remerr).Error()))
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
		}

		return c.JSON(resp)
	}
	return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
}

// GetMessages
//
//	@Summary	Get message
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id	path		int64		true	"Channel id"
//	@Param		from		query		int64		false	"Start point for messages"
//	@Param		direction	query		string		false	"Select direction"
//	@Param		limit		query		int			false	"Message count limit"
//	@Success	200			{object}	dto.Message	"Message"
//	@failure	400			{string}	string		"Incorrect request body"
//	@failure	401			{string}	string		"Unauthorized"
//	@failure	500			{string}	string		"Something bad happened"
//	@Router		/message/channel/{channel_id} [get]
func (e *entity) GetMessages(c *fiber.Ctx) error {
	var req GetMessagesRequest
	err := c.QueryParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if req.Limit == nil {
		var limit = DefaultLimit
		req.Limit = &limit
	}
	if req.Direction == nil {
		var dir = DirectionBefore
		req.Direction = &dir
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
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToReadFromThisChannel)
	}
	if ch.LastMessage == 0 {
		return c.JSON([]dto.Message{})
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
			_, _, _, ok, err = e.perm.ChannelPerm(c.UserContext(), gc.GuildId, gc.ChannelId, u.Id, permissions.PermServerViewChannels)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
		}
	} else if ch.Type == model.ChannelTypeDM {
		// TODO: to be implemented
	} else if ch.Type == model.ChannelTypeGroupDM {
		// TODO: to be implemented
	}
	if ok {
		var rawmsgs []model.Message
		var uids []int64
		var err error
		switch *req.Direction {
		case DirectionBefore:
			if req.From == nil {
				req.From = &ch.LastMessage
			}
			rawmsgs, uids, err = e.msg.GetMessagesBefore(c.UserContext(), ch.Id, *req.From, *req.Limit)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
		case DirectionAfter:
			if req.From == nil {
				return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetMessage)
			}
			rawmsgs, uids, err = e.msg.GetMessagesAfter(c.UserContext(), ch.Id, *req.From, ch.LastMessage, *req.Limit)
		}

		users, err := e.user.GetUsersList(c.UserContext(), uids)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		var um = make(map[int64]*model.User)
		for i, u := range users {
			um[u.Id] = &users[i]
		}

		var mm = make(map[int64]*model.Member)
		if guildId != nil {
			members, err := e.m.GetMembersList(c.UserContext(), *guildId, uids)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
			for i, m := range members {
				mm[m.UserId] = &members[i]
			}
		}

		var am = make(map[int64]bool)
		for _, m := range rawmsgs {
			for _, aid := range m.Attachments {
				am[aid] = true
			}
		}

		var atids []int64
		for id, _ := range am {
			atids = append(atids, id)
		}

		ats, err := e.at.SelectAttachmentByIDs(c.UserContext(), atids)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		var atms = make(map[int64]*model.Attachment)
		for _, a := range ats {
			atms[a.Id] = &a
		}

		var messages []dto.Message
		for _, m := range rawmsgs {
			var author dto.User
			if memb, ok := mm[m.UserId]; ok {
				author.Id = memb.UserId
				if memb.Username != nil {
					author.Name = *memb.Username
				} else {
					author.Name = um[m.UserId].Name
				}
				if memb.Avatar != nil {
					author.Avatar = memb.Avatar
				} else {
					author.Avatar = um[m.UserId].Avatar
				}
			} else {
				author.Id = um[m.UserId].Id
				author.Name = um[m.UserId].Name
				author.Avatar = um[m.UserId].Avatar
			}

			var attachments []dto.Attachment
			for _, a := range m.Attachments {
				if at, ok := atms[a]; ok {
					attachments = append(attachments, dto.Attachment{
						ContentType: at.ContentType,
						Filename:    at.Name,
						Height:      at.Height,
						Width:       at.Width,
						URL:         fmt.Sprintf("media/%d/%d/%s", at.ChannelId, at.Id, at.Name),
						Size:        at.FileSize,
					})
				}
			}

			messages = append(messages, dto.Message{
				Id:          m.Id,
				ChannelId:   m.ChannelId,
				Author:      author,
				Content:     m.Content,
				Attachments: attachments,
				UpdatedAt:   nil,
			})
		}
		return c.JSON(messages)
	}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
}

// Update
//
//	@Summary	Update message
//	@Produce	json
//	@Tags		Message
//	@Param		message_id	path		int64					true	"Message id"
//	@Param		channel_id	path		int64					true	"Channel id"
//	@Param		request		body		UpdateMessageRequest	true	"Message data"
//	@Success	200			{object}	dto.Message				"Message"
//	@failure	400			{string}	string					"Incorrect request body"
//	@failure	401			{string}	string					"Unauthorized"
//	@failure	500			{string}	string					"Something bad happened"
//	@Router		/message/channel/{channel_id}/{message_id} [patch]
func (e *entity) Update(c *fiber.Ctx) error {
	var req UpdateMessageRequest
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	if req.Content == "" {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	channelId := c.Params("channel_id")
	chid, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	messageId := c.Params("message_id")
	msgid, err := strconv.ParseInt(messageId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectMessageID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	user, err := e.user.GetUserById(c.UserContext(), u.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUser)
	}
	msg, err := e.msg.GetMessage(c.UserContext(), msgid, chid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if msg.UserId == u.Id {
		var username = user.Name
		var gid *int64
		gc, err := e.gc.GetGuildByChannel(c.UserContext(), chid)
		if err != nil && !errors.Is(err, gocql.ErrNotFound) {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuild)
		}
		if !errors.Is(err, gocql.ErrNotFound) {
			gid = &gc.GuildId
			m, err := e.m.GetMember(c.UserContext(), u.Id, gc.GuildId)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
			if m.Username != nil {
				username = *m.Username
			}
		}
		d, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), u.Id)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		err = e.msg.UpdateMessage(c.UserContext(), msg.Id, chid, req.Content)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		upd := time.Now()
		resp := dto.Message{
			Id:        msg.Id,
			ChannelId: msg.ChannelId,
			Author: dto.User{
				Id:            user.Id,
				Name:          username,
				Discriminator: d.Discriminator,
				Avatar:        user.Avatar,
			},
			Content:   req.Content,
			UpdatedAt: &upd,
		}
		err = e.mqt.SendChannelMessage(chid, &mqmsg.UpdateMessage{
			GuildId: gid,
			Message: resp,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(resp)
	}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
}

// Delete
//
//	@Summary	Delete message
//	@Produce	json
//	@Tags		Message
//	@Param		message_id	path		int64		true	"Message id"
//	@Param		channel_id	path		int64		true	"Channel id"
//	@Success	200			{object}	dto.Message	"Message"
//	@failure	400			{string}	string		"Incorrect request body"
//	@failure	401			{string}	string		"Unauthorized"
//	@failure	500			{string}	string		"Something bad happened"
//	@Router		/message/channel/{channel_id}/{message_id} [delete]
func (e *entity) Delete(c *fiber.Ctx) error {
	channelId := c.Params("channel_id")
	chid, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	messageId := c.Params("message_id")
	msgid, err := strconv.ParseInt(messageId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectMessageID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	msg, err := e.msg.GetMessage(c.UserContext(), msgid, chid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if msg.UserId == u.Id {
		err = e.msg.DeleteMessage(c.UserContext(), msg.Id, chid)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		err = e.mqt.SendChannelMessage(chid, &mqmsg.DeleteMessage{
			MessageId: msgid,
			ChannelId: chid,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.SendStatus(fiber.StatusOK)
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
