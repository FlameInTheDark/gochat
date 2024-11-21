package search

import (
	"github.com/gofiber/fiber/v2"
)

func (e *entity) Search(c *fiber.Ctx) error {
	//var req SendMessageRequest
	//err := c.BodyParser(&req)
	//if err != nil {
	//	return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	//}
	//channelId := c.Params("channel_id")
	//id, err := strconv.ParseInt(channelId, 10, 64)
	//if err != nil {
	//	return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	//}
	//u, err := helper.GetUser(c)
	//if err != nil {
	//	return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	//}
	//if e.checkChannelPermissions(c.UserContext(), u.Id, id, permissions.PermissionViewChannel) {
	//	var guildId *int64
	//	gc, err := e.gc.GetGuildByChannel(c.UserContext(), id)
	//	if errors.Is(err, gocql.ErrNotFound) {
	//		guildId = nil
	//	} else if err == nil {
	//		guildId = &gc.GuildId
	//	} else {
	//		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	//	}
	//	user, err := e.user.GetUserById(c.UserContext(), u.Id)
	//	if err := helper.HttpDbError(err, ErrUnableToGetUser); err != nil {
	//		return err
	//	}
	//	disc, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), u.Id)
	//	if err := helper.HttpDbError(err, ErrUnableToGetUserDiscriminator); err != nil {
	//		return err
	//	}
	//	msgid := idgen.Next()
	//	err = e.msg.CreateMessage(c.UserContext(), msgid, id, u.Id, req.Content, req.Attachments)
	//	if err := helper.HttpDbError(err, ErrUnableToSendMessage); err != nil {
	//		return err
	//	}
	//
	//	ats, err := e.at.SelectAttachemntsByIDs(c.UserContext(), req.Attachments)
	//	if err := helper.HttpDbError(err, ErrUnableToGetAttachements); err != nil {
	//		return err
	//	}
	//
	//	resp := dto.Message{
	//		Id:        msgid,
	//		ChannelId: id,
	//		Author: dto.User{
	//			Id:            user.Id,
	//			Name:          user.Name,
	//			Discriminator: disc.Discriminator,
	//			Avatar:        user.Avatar,
	//		},
	//		Content:     req.Content,
	//		Attachments: nil,
	//	}
	//
	//	for _, at := range ats {
	//		resp.Attachments = append(resp.Attachments, dto.Attachment{
	//			ContentType: at.ContentType,
	//			Filename:    at.Name,
	//			Height:      at.Height,
	//			Width:       at.Width,
	//			URL:         fmt.Sprintf("media/%d/%d/%s", at.ChannelId, at.Id, at.Name),
	//			Size:        at.FileSize,
	//		})
	//	}
	//
	//	err = e.msgmq.PublishMessage(id, &mqmsg.CreateMessage{
	//		GuildId: guildId,
	//		Message: resp,
	//	})
	//	if err != nil {
	//		remerr := e.msg.DeleteMessage(c.UserContext(), msgid)
	//		e.log.Error("unable to send message event", slog.String("error", errors.Join(err, remerr).Error()))
	//		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	//	}
	//
	//	return c.JSON(resp)
	//}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
}

func (e *entity) Update(c *fiber.Ctx) error {
	//var req SendMessageRequest
	//err := c.BodyParser(&req)
	//if err != nil {
	//	return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	//}
	//channelId := c.Params("channel_id")
	//id, err := strconv.ParseInt(channelId, 10, 64)
	//if err != nil {
	//	return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	//}
	//u, err := helper.GetUser(c)
	//if err != nil {
	//	return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	//}
	//if e.checkChannelPermissions(c.UserContext(), u.Id, id, permissions.PermissionSendMessages) {
	//	var guildId *int64
	//	gc, err := e.gc.GetGuildByChannel(c.UserContext(), id)
	//	if errors.Is(err, gocql.ErrNotFound) {
	//		guildId = nil
	//	} else if err == nil {
	//		guildId = &gc.GuildId
	//	} else {
	//		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	//	}
	//	user, err := e.user.GetUserById(c.UserContext(), u.Id)
	//	if err := helper.HttpDbError(err, ErrUnableToGetUser); err != nil {
	//		return err
	//	}
	//	disc, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), u.Id)
	//	if err := helper.HttpDbError(err, ErrUnableToGetUserDiscriminator); err != nil {
	//		return err
	//	}
	//	msgid := idgen.Next()
	//	err = e.msg.CreateMessage(c.UserContext(), msgid, id, u.Id, req.Content, req.Attachments)
	//	if err := helper.HttpDbError(err, ErrUnableToSendMessage); err != nil {
	//		return err
	//	}
	//
	//	ats, err := e.at.SelectAttachemntsByIDs(c.UserContext(), req.Attachments)
	//	if err := helper.HttpDbError(err, ErrUnableToGetAttachements); err != nil {
	//		return err
	//	}
	//
	//	resp := dto.Message{
	//		Id:        msgid,
	//		ChannelId: id,
	//		Author: dto.User{
	//			Id:            user.Id,
	//			Name:          user.Name,
	//			Discriminator: disc.Discriminator,
	//			Avatar:        user.Avatar,
	//		},
	//		Content:     req.Content,
	//		Attachments: nil,
	//	}
	//
	//	for _, at := range ats {
	//		resp.Attachments = append(resp.Attachments, dto.Attachment{
	//			ContentType: at.ContentType,
	//			Filename:    at.Name,
	//			Height:      at.Height,
	//			Width:       at.Width,
	//			URL:         fmt.Sprintf("media/%d/%d/%s", at.ChannelId, at.Id, at.Name),
	//			Size:        at.FileSize,
	//		})
	//	}
	//
	//	err = e.msgmq.PublishMessage(id, &mqmsg.CreateMessage{
	//		GuildId: guildId,
	//		Message: resp,
	//	})
	//	if err != nil {
	//		remerr := e.msg.DeleteMessage(c.UserContext(), msgid)
	//		e.log.Error("unable to send message event", slog.String("error", errors.Join(err, remerr).Error()))
	//		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	//	}
	//
	//	return c.JSON(resp)
	//}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
}

func (e *entity) Attachment(c *fiber.Ctx) error {
	//channelId := c.Params("channel_id")
	//id, err := strconv.ParseInt(channelId, 10, 64)
	//if err != nil {
	//	return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	//}
	//var req UploadAttachmentRequest
	//err = c.BodyParser(&req)
	//if err != nil {
	//	return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	//}
	//user, err := helper.GetUser(c)
	//if err != nil {
	//	return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	//}
	//
	//dbu, err := e.user.GetUserById(c.UserContext(), user.Id)
	//if err != nil {
	//	return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUser)
	//}
	//
	//switch dbu.UploadLimit == nil {
	//case true:
	//	if req.FileSize > e.uploadLimit {
	//		return fiber.NewError(fiber.StatusNotAcceptable, ErrFileIsTooBig)
	//	}
	//case false:
	//	if req.FileSize > *dbu.UploadLimit {
	//		return fiber.NewError(fiber.StatusNotAcceptable, ErrFileIsTooBig)
	//	}
	//}
	//
	//if e.checkChannelPermissions(c.UserContext(), user.Id, id, permissions.PermissionSendMessages) {
	//	atid := idgen.Next()
	//	url, err := e.storage.MakeUploadAttachment(c.UserContext(), id, atid, req.FileSize, req.Filename)
	//	if err != nil {
	//		e.log.Error(err.Error())
	//		return fiber.NewError(fiber.StatusNotAcceptable, ErrUnableToCreateUploadURL)
	//	}
	//	err = e.at.CreateAttachment(c.UserContext(), atid, id, req.FileSize, req.Height, req.Width, req.Filename)
	//	if err := helper.HttpDbError(err, ErrUnableToCreateAttachment); err != nil {
	//		return err
	//	}
	//	return c.JSON(dto.AttachmentUpload{
	//		Id:        atid,
	//		ChannelId: id,
	//		FileName:  req.Filename,
	//		UploadURL: url,
	//	})
	//}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
}
