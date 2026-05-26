package message

import (
	"errors"
	"strconv"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/embed"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	reactionutil "github.com/FlameInTheDark/gochat/internal/reaction"
	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
)

// Send
//
//	@Summary	Send a message as the bot
//	@Accept		json
//	@Produce	json
//	@Tags		Bot Message
//	@Security	BotToken
//	@Param		channel_id	path		int64		true	"Channel id"
//	@Param		request		body		SendRequest	true	"Message body"
//	@Success	200			{object}	dto.Message
//	@Failure	400			{string}	string	"Bad request"
//	@Failure	401			{string}	string	"Unauthorized"
//	@Failure	403			{string}	string	"Forbidden"
//	@Failure	500			{string}	string	"Internal server error"
//	@Router		/bot/api/v1/message/channel/{channel_id} [post]
func (e *Entity) Send(c *fiber.Ctx) error {
	principal, channel, guildID, err := e.requireChannel(c, permissions.PermServerViewChannels, permissions.PermTextSendMessage)
	if err != nil {
		return err
	}
	var req SendRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid message body")
	}
	if req.Content == "" && len(req.Attachments) == 0 && len(req.Embeds) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "message content, attachment, or embed is required")
	}
	embedsJSON, err := embed.MarshalEmbeds(req.Embeds)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	autoEmbedsJSON, err := embed.MarshalEmbeds(nil)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	position, err := e.ch.ReserveMessagePositions(c.UserContext(), channel.Id, 1)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to allocate message position")
	}
	msgID := idgen.Next()
	msgType := model.MessageTypeChat
	var reference int64
	var referenceChannel int64
	if req.Reference != nil && *req.Reference != 0 {
		reference = *req.Reference
		referenceChannel = channel.Id
		msgType = model.MessageTypeReply
	}
	if err := e.msg.CreateMessageWithMeta(c.UserContext(), msgID, channel.Id, principal.BotUserID, req.Content, req.Attachments, embedsJSON, autoEmbedsJSON, 0, msgType, referenceChannel, reference, 0, position); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to create message")
	}
	if err := e.ch.SetLastMessage(c.UserContext(), channel.Id, msgID); err != nil {
		_ = e.msg.DeleteMessage(c.UserContext(), msgID, channel.Id)
		return fiber.NewError(fiber.StatusInternalServerError, "unable to update channel last message")
	}
	if guildID != nil {
		_ = e.gclm.SetChannelLastMessage(c.UserContext(), *guildID, channel.Id, msgID)
	}
	_ = e.rs.SetReadState(c.UserContext(), principal.BotUserID, channel.Id, msgID)
	out, err := e.messageDTO(c, model.Message{
		Id:               msgID,
		ChannelId:        channel.Id,
		UserId:           principal.BotUserID,
		Content:          req.Content,
		Position:         position,
		Attachments:      req.Attachments,
		EmbedsJSON:       &embedsJSON,
		AutoEmbedsJSON:   &autoEmbedsJSON,
		Type:             int(msgType),
		ReferenceChannel: referenceChannel,
		Reference:        reference,
	})
	if err != nil {
		return err
	}
	_ = mq.SendChannelMessage(c.UserContext(), e.mqt, channel.Id, &mqmsg.CreateMessage{GuildId: guildID, Message: out})
	return c.JSON(out)
}

// List
//
//	@Summary	List messages visible to the bot
//	@Produce	json
//	@Tags		Bot Message
//	@Security	BotToken
//	@Param		channel_id	path		int64	true	"Channel id"
//	@Param		from		query		int64	false	"Message id cursor"
//	@Param		limit		query		int		false	"Page size"
//	@Param		direction	query		string	false	"before, after, or around"
//	@Success	200			{array}		dto.Message
//	@Failure	400			{string}	string	"Bad request"
//	@Failure	401			{string}	string	"Unauthorized"
//	@Failure	403			{string}	string	"Forbidden"
//	@Failure	500			{string}	string	"Internal server error"
//	@Router		/bot/api/v1/message/channel/{channel_id} [get]
func (e *Entity) List(c *fiber.Ctx) error {
	_, channel, _, err := e.requireChannel(c, permissions.PermServerViewChannels, permissions.PermTextReadMessageHistory)
	if err != nil {
		return err
	}
	limit := c.QueryInt("limit", 50)
	if limit <= 0 || limit > 100 {
		return fiber.NewError(fiber.StatusBadRequest, "limit must be between 1 and 100")
	}
	from, _ := strconv.ParseInt(c.Query("from"), 10, 64)
	if from == 0 {
		from = channel.LastMessage
	}
	if from == 0 {
		return c.JSON([]dto.Message{})
	}
	var messages []model.Message
	switch c.Query("direction", "before") {
	case "before":
		messages, _, err = e.msg.GetMessagesBefore(c.UserContext(), channel.Id, from, limit)
	case "after":
		messages, _, err = e.msg.GetMessagesAfter(c.UserContext(), channel.Id, from, channel.LastMessage, limit)
	case "around":
		messages, _, err = e.msg.GetMessagesAround(c.UserContext(), channel.Id, from, channel.LastMessage, limit)
	default:
		return fiber.NewError(fiber.StatusBadRequest, "direction must be before, after, or around")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get messages")
	}
	out := make([]dto.Message, 0, len(messages))
	for _, msg := range messages {
		built, err := e.messageDTO(c, msg)
		if err != nil {
			return err
		}
		out = append(out, built)
	}
	return c.JSON(out)
}

// Update
//
//	@Summary	Edit a message as the bot
//	@Accept		json
//	@Produce	json
//	@Tags		Bot Message
//	@Security	BotToken
//	@Param		channel_id	path		int64			true	"Channel id"
//	@Param		message_id	path		int64			true	"Message id"
//	@Param		request		body		UpdateRequest	true	"Message update"
//	@Success	200			{object}	dto.Message
//	@Failure	400			{string}	string	"Bad request"
//	@Failure	401			{string}	string	"Unauthorized"
//	@Failure	403			{string}	string	"Forbidden"
//	@Failure	404			{string}	string	"Not found"
//	@Failure	500			{string}	string	"Internal server error"
//	@Router		/bot/api/v1/message/channel/{channel_id}/{message_id} [patch]
func (e *Entity) Update(c *fiber.Ctx) error {
	principal, channel, guildID, err := e.requireChannel(c, permissions.PermServerViewChannels, permissions.PermTextReadMessageHistory)
	if err != nil {
		return err
	}
	msgID, err := parseParamInt64(c, "message_id")
	if err != nil {
		return err
	}
	msg, err := e.msg.GetMessage(c.UserContext(), msgID, channel.Id)
	if err != nil {
		return messageLookupError(err)
	}
	if !model.IsEditableMessageType(model.MessageType(msg.Type)) {
		return fiber.NewError(fiber.StatusBadRequest, "message type is not editable")
	}
	if msg.UserId != principal.BotUserID {
		if guildID == nil {
			return fiber.NewError(fiber.StatusForbidden, "bot cannot edit this message")
		}
		_, _, _, allowed, err := e.perm.ChannelPerm(c.UserContext(), *guildID, channel.Id, principal.BotUserID, permissions.PermTextManageMessages)
		if err != nil || !allowed {
			return fiber.NewError(fiber.StatusForbidden, "bot cannot edit this message")
		}
	}
	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid message body")
	}
	content := msg.Content
	if req.Content != nil {
		content = *req.Content
	}
	flags := model.NormalizeMessageFlags(msg.Flags)
	if req.Flags != nil {
		flags = *req.Flags
	}
	manualEmbeds := msg.EmbedsJSON
	if req.Embeds != nil {
		embedsJSON, err := embed.MarshalEmbeds(*req.Embeds)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		manualEmbeds = &embedsJSON
	}
	if err := e.msg.UpdateMessage(c.UserContext(), msg.Id, msg.ChannelId, content, stringValue(manualEmbeds), stringValue(msg.AutoEmbedsJSON), flags); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to update message")
	}
	updated, err := e.msg.GetMessage(c.UserContext(), msgID, channel.Id)
	if err != nil {
		return messageLookupError(err)
	}
	out, err := e.messageDTO(c, updated)
	if err != nil {
		return err
	}
	_ = mq.SendChannelMessage(c.UserContext(), e.mqt, channel.Id, &mqmsg.UpdateMessage{GuildId: guildID, Message: out})
	return c.JSON(out)
}

// Delete
//
//	@Summary	Delete a message as the bot
//	@Produce	json
//	@Tags		Bot Message
//	@Security	BotToken
//	@Param		channel_id	path	int64	true	"Channel id"
//	@Param		message_id	path	int64	true	"Message id"
//	@Success	204
//	@Failure	400	{string}	string	"Bad request"
//	@Failure	401	{string}	string	"Unauthorized"
//	@Failure	403	{string}	string	"Forbidden"
//	@Failure	404	{string}	string	"Not found"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/bot/api/v1/message/channel/{channel_id}/{message_id} [delete]
func (e *Entity) Delete(c *fiber.Ctx) error {
	principal, channel, guildID, err := e.requireChannel(c, permissions.PermServerViewChannels, permissions.PermTextReadMessageHistory)
	if err != nil {
		return err
	}
	msgID, err := parseParamInt64(c, "message_id")
	if err != nil {
		return err
	}
	msg, err := e.msg.GetMessage(c.UserContext(), msgID, channel.Id)
	if err != nil {
		return messageLookupError(err)
	}
	if msg.UserId != principal.BotUserID {
		if guildID == nil {
			return fiber.NewError(fiber.StatusForbidden, "bot cannot delete this message")
		}
		_, _, _, allowed, err := e.perm.ChannelPerm(c.UserContext(), *guildID, channel.Id, principal.BotUserID, permissions.PermTextManageMessages)
		if err != nil || !allowed {
			return fiber.NewError(fiber.StatusForbidden, "bot cannot delete this message")
		}
	}
	if err := e.msg.DeleteMessage(c.UserContext(), msg.Id, channel.Id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to delete message")
	}
	if channel.LastMessage == msg.Id {
		e.repairLastMessage(c, channel.Id, guildID, msg.Id)
	}
	_ = mq.SendChannelMessage(c.UserContext(), e.mqt, channel.Id, &mqmsg.DeleteMessage{GuildId: guildID, ChannelId: channel.Id, MessageId: msg.Id})
	return c.SendStatus(fiber.StatusNoContent)
}

// Ack
//
//	@Summary	Mark a channel read as the bot
//	@Produce	json
//	@Tags		Bot Message
//	@Security	BotToken
//	@Param		channel_id	path	int64	true	"Channel id"
//	@Param		message_id	path	int64	true	"Message id"
//	@Success	204
//	@Failure	400	{string}	string	"Bad request"
//	@Failure	401	{string}	string	"Unauthorized"
//	@Failure	403	{string}	string	"Forbidden"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/bot/api/v1/message/channel/{channel_id}/{message_id}/ack [post]
func (e *Entity) Ack(c *fiber.Ctx) error {
	principal, channel, _, err := e.requireChannel(c, permissions.PermServerViewChannels, permissions.PermTextReadMessageHistory)
	if err != nil {
		return err
	}
	msgID, err := parseParamInt64(c, "message_id")
	if err != nil {
		return err
	}
	if err := e.rs.SetReadState(c.UserContext(), principal.BotUserID, channel.Id, msgID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to update read state")
	}
	_ = mq.SendUserUpdate(c.UserContext(), e.mqt, principal.BotUserID, &mqmsg.UpdateReadState{ChannelId: channel.Id, MessageId: msgID})
	return c.SendStatus(fiber.StatusNoContent)
}

// Typing
//
//	@Summary	Send a typing indicator as the bot
//	@Produce	json
//	@Tags		Bot Message
//	@Security	BotToken
//	@Param		channel_id	path	int64	true	"Channel id"
//	@Success	204
//	@Failure	400	{string}	string	"Bad request"
//	@Failure	401	{string}	string	"Unauthorized"
//	@Failure	403	{string}	string	"Forbidden"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/bot/api/v1/message/channel/{channel_id}/typing [post]
func (e *Entity) Typing(c *fiber.Ctx) error {
	principal, channel, guildID, err := e.requireChannel(c, permissions.PermServerViewChannels, permissions.PermTextSendMessage)
	if err != nil {
		return err
	}
	_ = mq.SendChannelMessage(c.UserContext(), e.mqt, channel.Id, &mqmsg.ChannelUserTyping{GuildId: guildID, ChannelId: channel.Id, UserId: principal.BotUserID})
	return c.SendStatus(fiber.StatusNoContent)
}

// AddReaction
//
//	@Summary	Add the bot reaction to a message
//	@Produce	json
//	@Tags		Bot Message
//	@Security	BotToken
//	@Param		channel_id		path		int64	true	"Channel id"
//	@Param		message_id		path		int64	true	"Message id"
//	@Param		reaction_name	path		string	true	"Reaction name"
//	@Success	200				{object}	dto.MessageReaction
//	@Failure	400				{string}	string	"Bad request"
//	@Failure	401				{string}	string	"Unauthorized"
//	@Failure	403				{string}	string	"Forbidden"
//	@Failure	404				{string}	string	"Not found"
//	@Failure	500				{string}	string	"Internal server error"
//	@Router		/bot/api/v1/message/channel/{channel_id}/{message_id}/reactions/{reaction_name} [put]
func (e *Entity) AddReaction(c *fiber.Ctx) error {
	principal, channel, guildID, err := e.requireChannel(c, permissions.PermServerViewChannels, permissions.PermTextReadMessageHistory, permissions.PermTextAddReactions)
	if err != nil {
		return err
	}
	msgID, err := parseParamInt64(c, "message_id")
	if err != nil {
		return err
	}
	if _, err := e.msg.GetMessage(c.UserContext(), msgID, channel.Id); err != nil {
		return messageLookupError(err)
	}
	parsed, err := reactionutil.ParseReactionName(c.Params("reaction_name"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid reaction name")
	}
	reaction := model.Reaction{MessageId: msgID, ReactionId: idgen.Next(), UserId: principal.BotUserID, BucketKey: parsed.BucketKey, Custom: parsed.Custom, EmojiId: parsed.EmojiId, EmojiName: parsed.EmojiName}
	if err := e.react.UpsertReaction(c.UserContext(), reaction); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to add reaction")
	}
	summary := e.reactionSummary(c, msgID, parsed.BucketKey, true)
	_ = mq.SendChannelMessage(c.UserContext(), e.mqt, channel.Id, &mqmsg.MessageReactionAdd{GuildId: guildID, ChannelId: channel.Id, MessageId: msgID, Reaction: summary})
	return c.JSON(summary)
}

// RemoveReaction
//
//	@Summary	Remove the bot reaction from a message
//	@Produce	json
//	@Tags		Bot Message
//	@Security	BotToken
//	@Param		channel_id		path	int64	true	"Channel id"
//	@Param		message_id		path	int64	true	"Message id"
//	@Param		reaction_name	path	string	true	"Reaction name"
//	@Success	204
//	@Failure	400	{string}	string	"Bad request"
//	@Failure	401	{string}	string	"Unauthorized"
//	@Failure	403	{string}	string	"Forbidden"
//	@Failure	404	{string}	string	"Not found"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/bot/api/v1/message/channel/{channel_id}/{message_id}/reactions/{reaction_name} [delete]
func (e *Entity) RemoveReaction(c *fiber.Ctx) error {
	principal, channel, guildID, err := e.requireChannel(c, permissions.PermServerViewChannels, permissions.PermTextReadMessageHistory)
	if err != nil {
		return err
	}
	msgID, err := parseParamInt64(c, "message_id")
	if err != nil {
		return err
	}
	parsed, err := reactionutil.ParseReactionName(c.Params("reaction_name"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid reaction name")
	}
	reaction, err := e.react.GetUserReaction(c.UserContext(), msgID, principal.BotUserID, parsed.BucketKey)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get reaction")
	}
	if err := e.react.DeleteReaction(c.UserContext(), reaction); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to remove reaction")
	}
	summary := e.reactionSummary(c, msgID, parsed.BucketKey, false)
	_ = mq.SendChannelMessage(c.UserContext(), e.mqt, channel.Id, &mqmsg.MessageReactionRemove{GuildId: guildID, ChannelId: channel.Id, MessageId: msgID, Reaction: summary})
	return c.SendStatus(fiber.StatusNoContent)
}

// GetReactionUsers
//
//	@Summary	List users who reacted with a specific reaction
//	@Produce	json
//	@Tags		Bot Message
//	@Security	BotToken
//	@Param		channel_id		path		int64	true	"Channel id"
//	@Param		message_id		path		int64	true	"Message id"
//	@Param		reaction_name	path		string	true	"Reaction name"
//	@Param		after			query		int64	false	"Reaction ID cursor"
//	@Param		limit			query		int		false	"Page size"
//	@Success	200				{object}	dto.MessageReactionUsersPage
//	@Failure	400				{string}	string	"Bad request"
//	@Failure	401				{string}	string	"Unauthorized"
//	@Failure	403				{string}	string	"Forbidden"
//	@Failure	404				{string}	string	"Not found"
//	@Failure	500				{string}	string	"Internal server error"
//	@Router		/bot/api/v1/message/channel/{channel_id}/{message_id}/reactions/{reaction_name} [get]
func (e *Entity) GetReactionUsers(c *fiber.Ctx) error {
	_, channel, _, err := e.requireChannel(c, permissions.PermServerViewChannels, permissions.PermTextReadMessageHistory)
	if err != nil {
		return err
	}
	msgID, err := parseParamInt64(c, "message_id")
	if err != nil {
		return err
	}
	if _, err := e.msg.GetMessage(c.UserContext(), msgID, channel.Id); err != nil {
		return messageLookupError(err)
	}
	parsed, err := reactionutil.ParseReactionName(c.Params("reaction_name"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid reaction name")
	}
	var req GetReactionUsersRequest
	if err := c.QueryParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid reaction query")
	}
	if req.After != nil && *req.After <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "after must be greater than zero")
	}
	limit := reactionutil.UsersPageDefaultLimit
	if req.Limit != nil {
		limit = reactionutil.ParseLimit(*req.Limit)
	}
	rows, err := e.react.ListBucketReactions(c.UserContext(), msgID, parsed.BucketKey, req.After, limit+1)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get reaction users")
	}
	var nextAfter *int64
	if len(rows) > limit {
		next := rows[limit].ReactionId
		nextAfter = &next
		rows = rows[:limit]
	}
	items := make([]dto.User, 0, len(rows))
	for _, row := range rows {
		user, err := e.publicUser(c, row.UserId)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "unable to build reaction users")
		}
		items = append(items, user)
	}
	return c.JSON(dto.MessageReactionUsersPage{Items: items, NextAfter: nextAfter})
}
