package message

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"sync"
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

// Performance optimization constants
const (
	MaxBatchSize = 50 // Maximum messages to process at once
)

// Object pools for memory optimization
var (
	messagePool = sync.Pool{
		New: func() interface{} {
			return &dto.Message{}
		},
	}

	attachmentPool = sync.Pool{
		New: func() interface{} {
			return make([]dto.Attachment, 0, 5)
		},
	}
)

// Send
//
//	@Summary	Send message
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id	path		int64				true	"Channel id"
//	@Param		request		body		SendMessageRequest	true	"Message data"
//	@Success	200			{object}	dto.Message			"Message"
//	@failure	400			{string}	string				"Bad request"
//	@failure	401			{string}	string				"Unauthorized"
//	@failure	403			{string}	string				"Forbidden"
//	@failure	500			{string}	string				"Internal server error"
//	@Router		/message/channel/{channel_id} [post]
func (e *entity) Send(c *fiber.Ctx) error {
	// Parse and validate request
	req, user, channelId, err := e.parseSendMessageRequest(c)
	if err != nil {
		return err
	}

	// Validate channel and permissions
	channel, guildId, err := e.validateSendPermissions(c, channelId, user.Id)
	if err != nil {
		return err
	}

	// Create and send message
	message, err := e.createAndSendMessage(c, req, user, channel, guildId)
	if err != nil {
		return err
	}

	return c.JSON(message)
}

// parseSendMessageRequest handles request parsing and user authentication
func (e *entity) parseSendMessageRequest(c *fiber.Ctx) (*SendMessageRequest, *helper.JWTUser, int64, error) {
	var req SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	if err := req.Validate(); err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	channelIdStr := c.Params("channel_id")
	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return &req, user, channelId, nil
}

// validateSendPermissions checks if user can send messages to the channel
func (e *entity) validateSendPermissions(c *fiber.Ctx, channelId, userId int64) (*model.Channel, *int64, error) {
	channel, err := e.ch.GetChannel(c.UserContext(), channelId)
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusNotFound, "channel not found")
	}

	// Check if channel type supports messaging
	if channel.Type == model.ChannelTypeGuildCategory || channel.Type == model.ChannelTypeGuildVoice {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToSentToThisChannel)
	}

	// Check guild permissions if it's a guild channel
	if channel.Type == model.ChannelTypeGuild {
		guildChannel, err := e.gc.GetGuildByChannel(c.UserContext(), channelId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to get guild channel")
		}

		if !errors.Is(err, sql.ErrNoRows) {
			_, _, _, canSend, err := e.perm.ChannelPerm(c.UserContext(), guildChannel.GuildId, guildChannel.ChannelId, userId, permissions.PermTextSendMessage)
			if err != nil {
				return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
			}
			if !canSend {
				return nil, nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
			}
			return &channel, &guildChannel.GuildId, nil
		}
	}

	return &channel, nil, nil
}

// createAndSendMessage creates the message and handles all related operations
func (e *entity) createAndSendMessage(c *fiber.Ctx, req *SendMessageRequest, jwtUser *helper.JWTUser, channel *model.Channel, guildId *int64) (dto.Message, error) {
	// Fetch user data concurrently
	userData, err := e.fetchUserDataForMessage(c, jwtUser.Id)
	if err != nil {
		return dto.Message{}, err
	}

	// Create message with transaction-like behavior
	messageId := idgen.Next()
	if err := e.createMessageWithCleanup(c, messageId, channel.Id, jwtUser.Id, req); err != nil {
		return dto.Message{}, err
	}

	// Build response message
	message, err := e.buildMessageResponse(c, messageId, channel, userData, req)
	if err != nil {
		// Cleanup on failure
		_ = e.msg.DeleteMessage(c.UserContext(), messageId, channel.Id)
		return dto.Message{}, err
	}

	// Send events (non-blocking)
	go e.sendMessageEvents(channel.Id, guildId, message, userData, req)

	return message, nil
}

// fetchUserDataForMessage fetches user and discriminator data concurrently
func (e *entity) fetchUserDataForMessage(c *fiber.Ctx, userId int64) (*messageUserData, error) {
	type userResult struct {
		user *model.User
		err  error
	}
	type discResult struct {
		disc *model.Discriminator
		err  error
	}

	userCh := make(chan userResult, 1)
	discCh := make(chan discResult, 1)

	go func() {
		user, err := e.user.GetUserById(c.UserContext(), userId)
		userCh <- userResult{&user, err}
	}()

	go func() {
		disc, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), userId)
		discCh <- discResult{&disc, err}
	}()

	userRes := <-userCh
	discRes := <-discCh

	if userRes.err != nil {
		return nil, helper.HttpDbError(userRes.err, ErrUnableToGetUser)
	}
	if discRes.err != nil {
		return nil, helper.HttpDbError(discRes.err, ErrUnableToGetUserDiscriminator)
	}

	return &messageUserData{
		User:          userRes.user,
		Discriminator: discRes.disc,
	}, nil
}

// messageUserData holds user data for message creation
type messageUserData struct {
	User          *model.User
	Discriminator *model.Discriminator
}

// createMessageWithCleanup creates message and updates channel with proper error handling
func (e *entity) createMessageWithCleanup(c *fiber.Ctx, messageId, channelId, userId int64, req *SendMessageRequest) error {
	// Create the message
	if err := e.msg.CreateMessage(c.UserContext(), messageId, channelId, userId, req.Content, req.Attachments); err != nil {
		return helper.HttpDbError(err, ErrUnableToSendMessage)
	}

	// Update channel's last message
	if err := e.ch.SetLastMessage(c.UserContext(), channelId, messageId); err != nil {
		// Cleanup message if channel update fails
		_ = e.msg.DeleteMessage(c.UserContext(), messageId, channelId)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update channel")
	}

	return nil
}

// buildMessageResponse constructs the message response DTO
func (e *entity) buildMessageResponse(c *fiber.Ctx, messageId int64, channel *model.Channel, userData *messageUserData, req *SendMessageRequest) (dto.Message, error) {
	// Fetch attachments if any
	var attachments []dto.Attachment
	if len(req.Attachments) > 0 {
		ats, err := e.at.SelectAttachmentByIDs(c.UserContext(), req.Attachments)
		if err != nil {
			return dto.Message{}, helper.HttpDbError(err, ErrUnableToGetAttachements)
		}

		for _, at := range ats {
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

	return dto.Message{
		Id:        messageId,
		ChannelId: channel.Id,
		Author: dto.User{
			Id:            userData.User.Id,
			Name:          userData.User.Name,
			Discriminator: userData.Discriminator.Discriminator,
			Avatar:        userData.User.Avatar,
		},
		Content:     req.Content,
		Attachments: attachments,
	}, nil
}

// sendMessageEvents sends message and indexing events asynchronously
func (e *entity) sendMessageEvents(channelId int64, guildId *int64, message dto.Message, userData *messageUserData, req *SendMessageRequest) {
	// Send message event
	if err := e.mqt.SendChannelMessage(channelId, &mqmsg.CreateMessage{
		GuildId: guildId,
		Message: message,
	}); err != nil {
		e.log.Error("failed to send message event",
			"message_id", message.Id,
			"channel_id", channelId,
			"error", err.Error())
	}

	// Send indexing event
	var hasTypes []string
	if HasURL(req.Content) {
		hasTypes = append(hasTypes, "url")
	}
	for _, attachment := range message.Attachments {
		if attachment.ContentType != nil {
			hasTypes = append(hasTypes, GetAttachmentType(*attachment.ContentType))
		}
	}

	if err := e.imq.IndexMessage(dto.IndexMessage{
		MessageId: message.Id,
		UserId:    userData.User.Id,
		ChannelId: channelId,
		GuildId:   guildId,
		Mentions:  req.Mentions,
		Has:       UniqueAttachmentTypes(hasTypes),
		Content:   message.Content,
	}); err != nil {
		e.log.Error("failed to send index message event",
			"message_id", message.Id,
			"error", err.Error())
	}
}

// GetMessages
//
//	@Summary	Get messages
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id	path		int64		true	"Channel id"
//	@Param		from		query		int64		false	"Start point for messages"
//	@Param		direction	query		string		false	"Select direction"
//	@Param		limit		query		int			false	"Message count limit"
//	@Success	200			{array}		dto.Message	"Messages"
//	@failure	400			{string}	string		"Bad request"
//	@failure	403			{string}	string		"Forbidden"
//	@failure	404			{string}	string		"Not found"
//	@failure	500			{string}	string		"Internal server error"
//	@Router		/message/channel/{channel_id} [get]
func (e *entity) GetMessages(c *fiber.Ctx) error {
	// Parse and validate request
	req, user, channelId, err := e.parseGetMessagesRequest(c)
	if err != nil {
		return err
	}

	// Validate channel and permissions
	channel, guildId, err := e.validateReadPermissions(c, channelId, user.Id)
	if err != nil {
		return err
	}

	// Handle empty channel
	if channel.LastMessage == 0 {
		return c.JSON([]dto.Message{})
	}

	// Fetch and build messages
	messages, err := e.fetchAndBuildMessages(c, req, channel, guildId)
	if err != nil {
		return err
	}

	return c.JSON(messages)
}

// parseGetMessagesRequest handles request parsing and validation with batch size limits
func (e *entity) parseGetMessagesRequest(c *fiber.Ctx) (*GetMessagesRequest, *helper.JWTUser, int64, error) {
	var req GetMessagesRequest
	if err := c.QueryParser(&req); err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := req.Validate(); err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Set defaults
	if req.Limit == nil {
		limit := DefaultLimit
		req.Limit = &limit
	}
	if req.Direction == nil {
		direction := DirectionBefore
		req.Direction = &direction
	}

	// Enforce maximum batch size for performance
	if *req.Limit > MaxBatchSize {
		*req.Limit = MaxBatchSize
	}

	channelIdStr := c.Params("channel_id")
	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return &req, user, channelId, nil
}

// validateReadPermissions checks if user can read messages from the channel
func (e *entity) validateReadPermissions(c *fiber.Ctx, channelId, userId int64) (*model.Channel, *int64, error) {
	channel, err := e.ch.GetChannel(c.UserContext(), channelId)
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusNotFound, "channel not found")
	}

	// Check if channel type supports reading
	if channel.Type == model.ChannelTypeGuildCategory || channel.Type == model.ChannelTypeGuildVoice {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToReadFromThisChannel)
	}

	// Check guild permissions
	if channel.Type == model.ChannelTypeGuild {
		guildChannel, err := e.gc.GetGuildByChannel(c.UserContext(), channelId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to get guild channel")
		}

		if !errors.Is(err, sql.ErrNoRows) {
			_, _, _, canRead, err := e.perm.ChannelPerm(
				c.UserContext(),
				guildChannel.GuildId,
				guildChannel.ChannelId,
				userId,
				permissions.PermServerViewChannels,
				permissions.PermTextReadMessageHistory,
			)
			if err != nil {
				return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
			}
			if !canRead {
				return nil, nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
			}
			return &channel, &guildChannel.GuildId, nil
		}
	}

	// TODO: Implement DM and GroupDM permission checks
	return &channel, nil, nil
}

// fetchAndBuildMessages fetches messages and builds DTOs with all related data
func (e *entity) fetchAndBuildMessages(c *fiber.Ctx, req *GetMessagesRequest, channel *model.Channel, guildId *int64) ([]dto.Message, error) {
	// Fetch raw messages
	rawMessages, userIds, err := e.fetchRawMessages(c, req, channel)
	if err != nil {
		return nil, err
	}

	if len(rawMessages) == 0 {
		return []dto.Message{}, nil
	}

	// Fetch all related data concurrently
	messageData, err := e.fetchMessageRelatedData(c, rawMessages, userIds, guildId)
	if err != nil {
		return nil, err
	}

	// Build message DTOs with memory optimization
	messages := e.buildMessageDTOsOptimized(rawMessages, messageData)

	return messages, nil
}

// fetchRawMessages retrieves messages based on direction and parameters
func (e *entity) fetchRawMessages(c *fiber.Ctx, req *GetMessagesRequest, channel *model.Channel) ([]model.Message, []int64, error) {
	var rawMessages []model.Message
	var userIds []int64
	var err error

	switch *req.Direction {
	case DirectionBefore:
		fromId := req.From
		if fromId == nil {
			fromId = &channel.LastMessage
		}
		rawMessages, userIds, err = e.msg.GetMessagesBefore(c.UserContext(), channel.Id, *fromId, *req.Limit)
	case DirectionAfter:
		if req.From == nil {
			return nil, nil, fiber.NewError(fiber.StatusBadRequest, "from parameter required for after direction")
		}
		rawMessages, userIds, err = e.msg.GetMessagesAfter(c.UserContext(), channel.Id, *req.From, channel.LastMessage, *req.Limit)
	}

	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to fetch messages")
	}

	return rawMessages, userIds, nil
}

// messageRelatedData holds all data needed to build message DTOs
type messageRelatedData struct {
	Users       map[int64]*model.User
	Members     map[int64]*model.Member
	Attachments map[int64]*model.Attachment
}

// fetchMessageRelatedData fetches users, members, and attachments concurrently
func (e *entity) fetchMessageRelatedData(c *fiber.Ctx, messages []model.Message, userIds []int64, guildId *int64) (*messageRelatedData, error) {
	type usersResult struct {
		users []model.User
		err   error
	}
	type membersResult struct {
		members []model.Member
		err     error
	}
	type attachmentsResult struct {
		attachments []model.Attachment
		err         error
	}

	usersCh := make(chan usersResult, 1)
	membersCh := make(chan membersResult, 1)
	attachmentsCh := make(chan attachmentsResult, 1)

	// Fetch users
	go func() {
		users, err := e.user.GetUsersList(c.UserContext(), userIds)
		usersCh <- usersResult{users, err}
	}()

	// Fetch members if guild channel
	go func() {
		if guildId != nil {
			members, err := e.m.GetMembersList(c.UserContext(), *guildId, userIds)
			membersCh <- membersResult{members, err}
		} else {
			membersCh <- membersResult{nil, nil}
		}
	}()

	// Fetch attachments
	go func() {
		attachmentIds := e.extractAttachmentIds(messages)
		if len(attachmentIds) > 0 {
			attachments, err := e.at.SelectAttachmentByIDs(c.UserContext(), attachmentIds)
			attachmentsCh <- attachmentsResult{attachments, err}
		} else {
			attachmentsCh <- attachmentsResult{nil, nil}
		}
	}()

	// Collect results
	usersRes := <-usersCh
	membersRes := <-membersCh
	attachmentsRes := <-attachmentsCh

	// Check for errors
	if usersRes.err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to fetch users")
	}
	if membersRes.err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to fetch members")
	}
	if attachmentsRes.err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to fetch attachments")
	}

	// Build maps
	data := &messageRelatedData{
		Users:       make(map[int64]*model.User),
		Members:     make(map[int64]*model.Member),
		Attachments: make(map[int64]*model.Attachment),
	}

	for i := range usersRes.users {
		data.Users[usersRes.users[i].Id] = &usersRes.users[i]
	}

	if membersRes.members != nil {
		for i := range membersRes.members {
			data.Members[membersRes.members[i].UserId] = &membersRes.members[i]
		}
	}

	if attachmentsRes.attachments != nil {
		for i := range attachmentsRes.attachments {
			data.Attachments[attachmentsRes.attachments[i].Id] = &attachmentsRes.attachments[i]
		}
	}

	return data, nil
}

// extractAttachmentIds extracts unique attachment IDs from messages
func (e *entity) extractAttachmentIds(messages []model.Message) []int64 {
	attachmentMap := make(map[int64]bool)
	for _, message := range messages {
		for _, attachmentId := range message.Attachments {
			attachmentMap[attachmentId] = true
		}
	}

	attachmentIds := make([]int64, 0, len(attachmentMap))
	for id := range attachmentMap {
		attachmentIds = append(attachmentIds, id)
	}

	return attachmentIds
}

// buildMessageDTOs constructs message DTOs from raw data
func (e *entity) buildMessageDTOs(messages []model.Message, data *messageRelatedData) []dto.Message {
	result := make([]dto.Message, len(messages))

	for i, message := range messages {
		// Build author
		author := e.buildAuthor(message.UserId, data)

		// Build attachments
		attachments := e.buildAttachments(message.Attachments, data)

		result[i] = dto.Message{
			Id:          message.Id,
			ChannelId:   message.ChannelId,
			Author:      author,
			Content:     message.Content,
			Attachments: attachments,
			UpdatedAt:   message.EditedAt,
		}
	}

	return result
}

// buildAuthor constructs author DTO with member override if available
func (e *entity) buildAuthor(userId int64, data *messageRelatedData) dto.User {
	user := data.Users[userId]
	member := data.Members[userId]

	author := dto.User{
		Id:     userId,
		Name:   user.Name,
		Avatar: user.Avatar,
	}

	// Override with member data if available
	if member != nil {
		if member.Username != nil {
			author.Name = *member.Username
		}
		if member.Avatar != nil {
			author.Avatar = member.Avatar
		}
	}

	return author
}

// buildAttachments constructs attachment DTOs
func (e *entity) buildAttachments(attachmentIds []int64, data *messageRelatedData) []dto.Attachment {
	if len(attachmentIds) == 0 {
		return nil
	}

	attachments := make([]dto.Attachment, 0, len(attachmentIds))
	for _, id := range attachmentIds {
		if attachment, exists := data.Attachments[id]; exists {
			attachments = append(attachments, dto.Attachment{
				ContentType: attachment.ContentType,
				Filename:    attachment.Name,
				Height:      attachment.Height,
				Width:       attachment.Width,
				URL:         fmt.Sprintf("media/%d/%d/%s", attachment.ChannelId, attachment.Id, attachment.Name),
				Size:        attachment.FileSize,
			})
		}
	}

	return attachments
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
//	@failure	400			{string}	string					"Bad request"
//	@failure	403			{string}	string					"Forbidden"
//	@failure	404			{string}	string					"Not found"
//	@failure	500			{string}	string					"Internal server error"
//	@Router		/message/channel/{channel_id}/{message_id} [patch]
func (e *entity) Update(c *fiber.Ctx) error {
	// Parse and validate request
	req, user, channelId, messageId, err := e.parseUpdateMessageRequest(c)
	if err != nil {
		return err
	}

	// Validate message ownership and get message data
	message, guildId, err := e.validateMessageOwnership(c, messageId, channelId, user.Id)
	if err != nil {
		return err
	}

	// Update message and build response
	updatedMessage, err := e.updateMessageAndBuildResponse(c, req, message, user, guildId)
	if err != nil {
		return err
	}

	// Send update event
	go e.sendUpdateEvent(channelId, guildId, updatedMessage)

	return c.JSON(updatedMessage)
}

// parseUpdateMessageRequest handles request parsing and validation
func (e *entity) parseUpdateMessageRequest(c *fiber.Ctx) (*UpdateMessageRequest, *helper.JWTUser, int64, int64, error) {
	var req UpdateMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, nil, 0, 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	if err := req.Validate(); err != nil {
		return nil, nil, 0, 0, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	channelIdStr := c.Params("channel_id")
	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		return nil, nil, 0, 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}

	messageIdStr := c.Params("message_id")
	messageId, err := strconv.ParseInt(messageIdStr, 10, 64)
	if err != nil {
		return nil, nil, 0, 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectMessageID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return nil, nil, 0, 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return &req, user, channelId, messageId, nil
}

// validateMessageOwnership checks if user owns the message and returns message data
func (e *entity) validateMessageOwnership(c *fiber.Ctx, messageId, channelId, userId int64) (*model.Message, *int64, error) {
	message, err := e.msg.GetMessage(c.UserContext(), messageId, channelId)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, nil, fiber.NewError(fiber.StatusNotFound, "message not found")
		}
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to get message")
	}

	if message.UserId != userId {
		return nil, nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	// Get guild ID if it's a guild channel
	var guildId *int64
	guildChannel, err := e.gc.GetGuildByChannel(c.UserContext(), channelId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to get guild channel")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		guildId = &guildChannel.GuildId
	}

	return &message, guildId, nil
}

// updateMessageAndBuildResponse updates the message and builds the response DTO
func (e *entity) updateMessageAndBuildResponse(c *fiber.Ctx, req *UpdateMessageRequest, message *model.Message, jwtUser *helper.JWTUser, guildId *int64) (dto.Message, error) {
	// Update the message
	if err := e.msg.UpdateMessage(c.UserContext(), message.Id, message.ChannelId, req.Content); err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "failed to update message")
	}

	// Fetch user data for response
	userData, err := e.fetchUserDataForUpdate(c, jwtUser.Id, guildId)
	if err != nil {
		return dto.Message{}, err
	}

	// Build response
	updatedAt := time.Now()
	return dto.Message{
		Id:        message.Id,
		ChannelId: message.ChannelId,
		Author: dto.User{
			Id:            userData.User.Id,
			Name:          userData.DisplayName,
			Discriminator: userData.Discriminator.Discriminator,
			Avatar:        userData.Avatar,
		},
		Content:   req.Content,
		UpdatedAt: &updatedAt,
	}, nil
}

// updateUserData holds user data for message updates
type updateUserData struct {
	User          *model.User
	Discriminator *model.Discriminator
	DisplayName   string
	Avatar        *int64
}

// fetchUserDataForUpdate fetches user data with guild member overrides
func (e *entity) fetchUserDataForUpdate(c *fiber.Ctx, userId int64, guildId *int64) (*updateUserData, error) {
	type userResult struct {
		user *model.User
		err  error
	}
	type discResult struct {
		disc *model.Discriminator
		err  error
	}
	type memberResult struct {
		member *model.Member
		err    error
	}

	userCh := make(chan userResult, 1)
	discCh := make(chan discResult, 1)
	memberCh := make(chan memberResult, 1)

	// Fetch user data
	go func() {
		user, err := e.user.GetUserById(c.UserContext(), userId)
		userCh <- userResult{&user, err}
	}()

	// Fetch discriminator
	go func() {
		disc, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), userId)
		discCh <- discResult{&disc, err}
	}()

	// Fetch member data if guild channel
	go func() {
		if guildId != nil {
			member, err := e.m.GetMember(c.UserContext(), userId, *guildId)
			memberCh <- memberResult{&member, err}
		} else {
			memberCh <- memberResult{nil, nil}
		}
	}()

	// Collect results
	userRes := <-userCh
	discRes := <-discCh
	memberRes := <-memberCh

	// Check for errors
	if userRes.err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUser)
	}
	if discRes.err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUserDiscriminator)
	}
	if memberRes.err != nil && !errors.Is(memberRes.err, sql.ErrNoRows) {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to get member")
	}

	// Build user data with member overrides
	userData := &updateUserData{
		User:          userRes.user,
		Discriminator: discRes.disc,
		DisplayName:   userRes.user.Name,
		Avatar:        userRes.user.Avatar,
	}

	// Apply member overrides if available
	if memberRes.member != nil {
		if memberRes.member.Username != nil {
			userData.DisplayName = *memberRes.member.Username
		}
		if memberRes.member.Avatar != nil {
			userData.Avatar = memberRes.member.Avatar
		}
	}

	return userData, nil
}

// sendUpdateEvent sends the message update event asynchronously
func (e *entity) sendUpdateEvent(channelId int64, guildId *int64, message dto.Message) {
	if err := e.mqt.SendChannelMessage(channelId, &mqmsg.UpdateMessage{
		GuildId: guildId,
		Message: message,
	}); err != nil {
		e.log.Error("failed to send message update event",
			"message_id", message.Id,
			"channel_id", channelId,
			"error", err.Error())
	}

	var hasTypes []string
	if HasURL(message.Content) {
		hasTypes = append(hasTypes, "url")
	}
	for _, attachment := range message.Attachments {
		if attachment.ContentType != nil {
			hasTypes = append(hasTypes, GetAttachmentType(*attachment.ContentType))
		}
	}

	if err := e.imq.UpdateMessage(dto.IndexMessage{
		MessageId: message.Id,
		UserId:    message.Author.Id,
		ChannelId: channelId,
		GuildId:   guildId,
		Mentions:  nil,
		Has:       UniqueAttachmentTypes(hasTypes),
		Content:   message.Content,
	}); err != nil {
		e.log.Error("failed to send update message event",
			"message_id", message.Id,
			"channel_id", channelId,
			"error", err.Error())
	}
}

// Delete
//
//	@Summary	Delete message
//	@Produce	json
//	@Tags		Message
//	@Param		message_id	path		int64	true	"Message id"
//	@Param		channel_id	path		int64	true	"Channel id"
//	@Success	200			{string}	string	"OK"
//	@failure	400			{string}	string	"Bad request"
//	@failure	403			{string}	string	"Forbidden"
//	@failure	404			{string}	string	"Not found"
//	@failure	500			{string}	string	"Internal server error"
//	@Router		/message/channel/{channel_id}/{message_id} [delete]
func (e *entity) Delete(c *fiber.Ctx) error {
	// Parse request parameters
	user, channelId, messageId, err := e.parseDeleteMessageRequest(c)
	if err != nil {
		return err
	}

	// Validate message ownership
	message, err := e.validateDeletePermission(c, messageId, channelId, user.Id)
	if err != nil {
		return err
	}

	// Delete message and send event
	if err := e.deleteMessageAndNotify(c, message); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// parseDeleteMessageRequest handles request parsing and user authentication
func (e *entity) parseDeleteMessageRequest(c *fiber.Ctx) (*helper.JWTUser, int64, int64, error) {
	channelIdStr := c.Params("channel_id")
	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		return nil, 0, 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}

	messageIdStr := c.Params("message_id")
	messageId, err := strconv.ParseInt(messageIdStr, 10, 64)
	if err != nil {
		return nil, 0, 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectMessageID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return nil, 0, 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return user, channelId, messageId, nil
}

// validateDeletePermission checks if user can delete the message
func (e *entity) validateDeletePermission(c *fiber.Ctx, messageId, channelId, userId int64) (*model.Message, error) {
	message, err := e.msg.GetMessage(c.UserContext(), messageId, channelId)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "message not found")
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to get message")
	}

	if message.UserId != userId {
		return nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	return &message, nil
}

// deleteMessageAndNotify deletes the message and sends notification event
func (e *entity) deleteMessageAndNotify(c *fiber.Ctx, message *model.Message) error {
	// Delete the message
	if err := e.msg.DeleteMessage(c.UserContext(), message.Id, message.ChannelId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete message")
	}

	// Send delete event asynchronously
	go e.sendDeleteEvent(message.ChannelId, message.Id)

	return nil
}

// sendDeleteEvent sends the message delete event asynchronously
func (e *entity) sendDeleteEvent(channelId, messageId int64) {
	if err := e.mqt.SendChannelMessage(channelId, &mqmsg.DeleteMessage{
		MessageId: messageId,
		ChannelId: channelId,
	}); err != nil {
		e.log.Error("failed to send message delete event",
			"message_id", messageId,
			"channel_id", channelId,
			"error", err.Error())
	}

	if err := e.imq.IndexDeleteMessage(dto.IndexDeleteMessage{
		MessageId: messageId,
		ChannelId: channelId,
	}); err != nil {
		e.log.Error("failed to send index delete message event",
			"message_id", messageId,
			"error", err.Error())
	}
}

// Attachment
//
//	@Summary	Create attachment
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id	path		int64					true	"Channel id"
//	@Param		request		body		UploadAttachmentRequest	true	"Attachment data"
//	@Success	200			{object}	dto.AttachmentUpload	"Attachment upload data"
//	@failure	400			{string}	string					"Bad request"
//	@failure	403			{string}	string					"Forbidden"
//	@failure	413			{string}	string					"File too large"
//	@failure	500			{string}	string					"Internal server error"
//	@Router		/message/channel/{channel_id}/attachment [post]
func (e *entity) Attachment(c *fiber.Ctx) error {
	// Parse and validate request
	req, user, channelId, err := e.parseAttachmentRequest(c)
	if err != nil {
		return err
	}

	// Validate file size limits
	if err := e.validateFileSizeLimit(c, user.Id, req.FileSize); err != nil {
		return err
	}

	// Validate channel and upload permissions
	if err := e.validateUploadPermissions(c, channelId, user.Id); err != nil {
		return err
	}

	// Create attachment and upload URL
	attachment, err := e.createAttachmentUpload(c, req, channelId)
	if err != nil {
		return err
	}

	return c.JSON(attachment)
}

// parseAttachmentRequest handles request parsing and user authentication
func (e *entity) parseAttachmentRequest(c *fiber.Ctx) (*UploadAttachmentRequest, *helper.JWTUser, int64, error) {
	channelIdStr := c.Params("channel_id")
	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}

	var req UploadAttachmentRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	if err := req.Validate(); err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return nil, nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return &req, user, channelId, nil
}

// validateFileSizeLimit checks if the file size is within user's upload limit
func (e *entity) validateFileSizeLimit(c *fiber.Ctx, userId int64, fileSize int64) error {
	user, err := e.user.GetUserById(c.UserContext(), userId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUser)
	}

	var uploadLimit int64
	if user.UploadLimit == nil {
		uploadLimit = e.uploadLimit
	} else {
		uploadLimit = *user.UploadLimit
	}

	if fileSize > uploadLimit {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, ErrFileIsTooBig)
	}

	return nil
}

// validateUploadPermissions checks if user can upload files to the channel
func (e *entity) validateUploadPermissions(c *fiber.Ctx, channelId, userId int64) error {
	channel, err := e.ch.GetChannel(c.UserContext(), channelId)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "channel not found")
	}

	// Check guild permissions
	if channel.Type == model.ChannelTypeGuild {
		guildChannel, err := e.gc.GetGuildByChannel(c.UserContext(), channelId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to get guild channel")
		}

		if !errors.Is(err, sql.ErrNoRows) {
			_, _, _, canAttach, err := e.perm.ChannelPerm(c.UserContext(), guildChannel.GuildId, guildChannel.ChannelId, userId, permissions.PermTextAttachFiles)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
			}
			if !canAttach {
				return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
			}
		}
	} else if channel.Type == model.ChannelTypeGroupDM {
		// Check if user is participant in group DM
		if channel.ParentID != nil && *channel.ParentID != userId {
			// TODO: Implement proper group DM participant check
			return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
		}
	}

	return nil
}

// createAttachmentUpload creates the attachment record and generates upload URL
func (e *entity) createAttachmentUpload(c *fiber.Ctx, req *UploadAttachmentRequest, channelId int64) (dto.AttachmentUpload, error) {
	attachmentId := idgen.Next()

	// Generate upload URL
	uploadURL, err := e.storage.MakeUploadAttachment(c.UserContext(), channelId, attachmentId, req.FileSize, req.Filename)
	if err != nil {
		e.log.Error("failed to create upload URL",
			"channel_id", channelId,
			"attachment_id", attachmentId,
			"filename", req.Filename,
			"error", err.Error())
		return dto.AttachmentUpload{}, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateUploadURL)
	}

	// Create attachment record
	if err := e.at.CreateAttachment(c.UserContext(), attachmentId, channelId, req.FileSize, req.Height, req.Width, req.Filename); err != nil {
		return dto.AttachmentUpload{}, helper.HttpDbError(err, ErrUnableToCreateAttachment)
	}

	return dto.AttachmentUpload{
		Id:        attachmentId,
		ChannelId: channelId,
		FileName:  req.Filename,
		UploadURL: uploadURL,
	}, nil
}

// Performance optimization functions for GetMessages

// buildMessageDTOsOptimized constructs message DTOs with memory optimization using object pools
func (e *entity) buildMessageDTOsOptimized(messages []model.Message, data *messageRelatedData) []dto.Message {
	result := make([]dto.Message, len(messages))

	for i, message := range messages {
		// Get message from pool
		msg := messagePool.Get().(*dto.Message)

		// Reset the message
		*msg = dto.Message{}

		// Build author
		author := e.buildAuthorOptimized(message.UserId, data)

		// Build attachments with pool
		attachments := e.buildAttachmentsOptimized(message.Attachments, data)

		// Set message fields
		msg.Id = message.Id
		msg.ChannelId = message.ChannelId
		msg.Author = author
		msg.Content = message.Content
		msg.Attachments = attachments
		msg.UpdatedAt = message.EditedAt

		result[i] = *msg

		// Return message to pool
		messagePool.Put(msg)
	}

	return result
}

// buildAuthorOptimized constructs author DTO with member override if available (optimized version)
func (e *entity) buildAuthorOptimized(userId int64, data *messageRelatedData) dto.User {
	user, userExists := data.Users[userId]
	if !userExists {
		// Fallback for missing user data
		return dto.User{
			Id:   userId,
			Name: "Unknown User",
		}
	}

	author := dto.User{
		Id:     userId,
		Name:   user.Name,
		Avatar: user.Avatar,
	}

	// Override with member data if available
	if member, memberExists := data.Members[userId]; memberExists {
		if member.Username != nil {
			author.Name = *member.Username
		}
		if member.Avatar != nil {
			author.Avatar = member.Avatar
		}
	}

	return author
}

// buildAttachmentsOptimized constructs attachment DTOs using object pools
func (e *entity) buildAttachmentsOptimized(attachmentIds []int64, data *messageRelatedData) []dto.Attachment {
	if len(attachmentIds) == 0 {
		return nil
	}

	// Get attachment slice from pool
	attachments := attachmentPool.Get().([]dto.Attachment)
	attachments = attachments[:0] // Reset length but keep capacity

	for _, id := range attachmentIds {
		if attachment, exists := data.Attachments[id]; exists {
			attachments = append(attachments, dto.Attachment{
				ContentType: attachment.ContentType,
				Filename:    attachment.Name,
				Height:      attachment.Height,
				Width:       attachment.Width,
				URL:         fmt.Sprintf("media/%d/%d/%s", attachment.ChannelId, attachment.Id, attachment.Name),
				Size:        attachment.FileSize,
			})
		}
	}

	// Create result copy and return slice to pool
	result := make([]dto.Attachment, len(attachments))
	copy(result, attachments)

	// Return to pool
	attachmentPool.Put(attachments)

	return result
}
