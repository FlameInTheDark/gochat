package message

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/embed"
	"github.com/FlameInTheDark/gochat/internal/embedmq"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/messageposition"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/threadcount"
)

const MaxBatchSize = 50

const (
	messageNonceTTLSeconds        = 5 * 60
	messageNonceLockTTLSeconds    = 15
	messageNonceLookupAttempts    = 20
	messageNonceLookupDelay       = 25 * time.Millisecond
	ErrNonceAlreadyBeingProcessed = "message with this nonce is already being created"
	ErrReplyMustBeInSameChannel   = "reply target must exist in the same channel"
)

type enforcedNonceRecord struct {
	ChannelID int64 `json:"channel_id"`
	MessageID int64 `json:"message_id"`
}

type enforcedNonceReservation struct {
	key      string
	lockKey  string
	record   *enforcedNonceRecord
	acquired bool
}

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

	validatedAttachments, err := e.validateMessageAttachments(c.UserContext(), channel.Id, user.Id, []int64(req.Attachments))
	if err != nil {
		return err
	}
	if err := e.validateReplyReference(c.UserContext(), channel.Id, req); err != nil {
		return err
	}

	req.Content, err = e.sanitizeEmojiContent(c.UserContext(), user.Id, req.Content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	}

	nonceReservation, err := e.beginEnforcedNonce(c.UserContext(), user.Id, channel.Id, req.Nonce, req.EnforceNonce)
	if err != nil {
		return err
	}
	if nonceReservation != nil {
		defer e.releaseEnforcedNonce(c.UserContext(), nonceReservation)
		if message, found, err := e.loadNonceDuplicateMessage(c, nonceReservation, req.Nonce); err != nil {
			return err
		} else if found {
			if err := e.rs.SetReadState(c.UserContext(), user.Id, channelId, message.Id); err != nil {
				e.log.Error("unable to set read state after nonce replay", slog.String("error", err.Error()))
			}
			return c.JSON(message)
		}
	}

	// Create and send message
	message, err := e.createAndSendMessage(c, req, user, channel, guildId, validatedAttachments)
	if err != nil {
		return err
	}

	if err := e.persistEnforcedNonce(c.UserContext(), nonceReservation, channel.Id, message.Id); err != nil && e.log != nil {
		e.log.Error("failed to persist enforced message nonce",
			"user_id", user.Id,
			"channel_id", channel.Id,
			"message_id", message.Id,
			"error", err.Error())
	}

	if err := e.rs.SetReadState(c.UserContext(), user.Id, channelId, message.Id); err != nil {
		e.log.Error("unable to set read state after message sent", slog.String("error", err.Error()))
	}

	return c.JSON(message)
}

// CreateThread
//
//	@Summary	Create thread from message
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id	path		int64				true	"Parent channel id"
//	@Param		message_id	path		int64				true	"Source message id"
//	@Param		request		body		CreateThreadRequest	true	"Thread data"
//	@Success	201			{object}	dto.Channel			"Thread channel"
//	@failure	400			{string}	string				"Bad request"
//	@failure	403			{string}	string				"Forbidden"
//	@failure	404			{string}	string				"Not found"
//	@failure	409			{string}	string				"Conflict"
//	@failure	500			{string}	string				"Internal server error"
//	@Router		/message/channel/{channel_id}/{message_id}/thread [post]
func (e *entity) CreateThread(c *fiber.Ctx) error {
	req, user, parentChannelID, sourceMessageID, err := e.parseThreadRequest(c)
	if err != nil {
		return err
	}

	parentChannel, guildID, guildChannel, err := e.validateThreadCreation(c, parentChannelID, user.Id)
	if err != nil {
		return err
	}

	sourceMessage, err := e.msg.GetMessage(c.UserContext(), sourceMessageID, parentChannelID)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "message not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetMessage)
	}
	if sourceMessage.Thread != 0 {
		return fiber.NewError(fiber.StatusConflict, ErrThreadAlreadyExists)
	}

	validatedAttachments, err := e.validateMessageAttachments(c.UserContext(), parentChannel.Id, user.Id, []int64(req.Attachments))
	if err != nil {
		return err
	}

	req.Content, err = e.sanitizeEmojiContent(c.UserContext(), user.Id, req.Content)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}

	userData, err := e.fetchUserDataForMessage(c, user.Id)
	if err != nil {
		return err
	}

	result, err := e.createThreadFromMessage(c, req, user.Id, guildID, guildChannel, parentChannel, &sourceMessage, validatedAttachments, userData)
	if err != nil {
		return err
	}

	go e.sendThreadCreateEvents(guildID, parentChannel, result, userData)

	return c.Status(fiber.StatusCreated).JSON(e.dtoThreadChannel(result.Channel, guildID, result.Position, result.Member, result.MemberIds))
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

func (e *entity) parseThreadRequest(c *fiber.Ctx) (*CreateThreadRequest, *helper.JWTUser, int64, int64, error) {
	var req CreateThreadRequest
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

func (e *entity) validateThreadCreation(c *fiber.Ctx, channelId, userId int64) (*model.Channel, int64, *model.GuildChannel, error) {
	channel, err := e.ch.GetChannel(c.UserContext(), channelId)
	if err != nil {
		return nil, 0, nil, fiber.NewError(fiber.StatusNotFound, "channel not found")
	}
	if channel.Type == model.ChannelTypeThread {
		return nil, 0, nil, fiber.NewError(fiber.StatusBadRequest, ErrThreadNestingForbidden)
	}
	if channel.Type != model.ChannelTypeGuild {
		return nil, 0, nil, fiber.NewError(fiber.StatusBadRequest, ErrThreadSourceInvalid)
	}

	guildChannel, err := e.gc.GetGuildByChannel(c.UserContext(), channelId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, nil, fiber.NewError(fiber.StatusBadRequest, ErrThreadSourceInvalid)
		}
		return nil, 0, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to get guild channel")
	}

	isMember, err := e.m.IsGuildMember(c.UserContext(), guildChannel.GuildId, userId)
	if err != nil {
		return nil, 0, nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuild)
	}
	if !isMember {
		return nil, 0, nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	_, _, _, canCreate, err := e.perm.ChannelPerm(c.UserContext(), guildChannel.GuildId, channelId, userId, permissions.PermTextCreateThreads)
	if err != nil {
		return nil, 0, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
	}
	if !canCreate {
		return nil, 0, nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	return &channel, guildChannel.GuildId, &guildChannel, nil
}

func requestedReferenceID(req *SendMessageRequest) int64 {
	if req == nil || req.Reference == nil {
		return 0
	}
	return int64(*req.Reference)
}

func requestedMessageType(req *SendMessageRequest) model.MessageType {
	if requestedReferenceID(req) != 0 {
		return model.MessageTypeReply
	}
	return model.MessageTypeChat
}

func (e *entity) validateReplyReference(ctx context.Context, channelID int64, req *SendMessageRequest) error {
	referenceID := requestedReferenceID(req)
	if referenceID == 0 {
		return nil
	}

	_, err := e.msg.GetMessage(ctx, referenceID, channelID)
	if err == nil {
		return nil
	}
	if errors.Is(err, gocql.ErrNotFound) {
		return fiber.NewError(fiber.StatusBadRequest, ErrReplyMustBeInSameChannel)
	}
	return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetMessage)
}

func messageNonceCacheKey(userID, channelID int64, nonce *helper.MessageNonce) string {
	return fmt.Sprintf("message:nonce:%d:%d:%s", userID, channelID, nonce.CacheKeyPart())
}

func (e *entity) beginEnforcedNonce(ctx context.Context, userID, channelID int64, nonce *helper.MessageNonce, enforce bool) (*enforcedNonceReservation, error) {
	if !enforce || nonce == nil || nonce.IsZero() {
		return nil, nil
	}
	if e.cache == nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	}

	key := messageNonceCacheKey(userID, channelID, nonce)
	record, found, err := e.loadEnforcedNonceRecord(ctx, key)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	}
	if found {
		return &enforcedNonceReservation{key: key, record: &record}, nil
	}

	lockKey := key + ":lock"
	acquired, err := e.cache.SetTimedJSONNX(ctx, lockKey, map[string]int64{"user_id": userID}, messageNonceLockTTLSeconds)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	}
	if !acquired {
		record, found, err := e.waitForEnforcedNonceRecord(ctx, key)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
		}
		if found {
			return &enforcedNonceReservation{key: key, record: &record}, nil
		}
		return nil, fiber.NewError(fiber.StatusConflict, ErrNonceAlreadyBeingProcessed)
	}

	return &enforcedNonceReservation{key: key, lockKey: lockKey, acquired: true}, nil
}

func (e *entity) loadEnforcedNonceRecord(ctx context.Context, key string) (enforcedNonceRecord, bool, error) {
	var record enforcedNonceRecord
	err := e.cache.GetJSON(ctx, key, &record)
	if err == nil {
		return record, true, nil
	}
	if errors.Is(err, redis.Nil) {
		return enforcedNonceRecord{}, false, nil
	}
	return enforcedNonceRecord{}, false, err
}

func (e *entity) waitForEnforcedNonceRecord(ctx context.Context, key string) (enforcedNonceRecord, bool, error) {
	for attempt := 0; attempt < messageNonceLookupAttempts; attempt++ {
		record, found, err := e.loadEnforcedNonceRecord(ctx, key)
		if err != nil {
			return enforcedNonceRecord{}, false, err
		}
		if found {
			return record, true, nil
		}
		select {
		case <-ctx.Done():
			return enforcedNonceRecord{}, false, ctx.Err()
		case <-time.After(messageNonceLookupDelay):
		}
	}
	return enforcedNonceRecord{}, false, nil
}

func (e *entity) releaseEnforcedNonce(ctx context.Context, reservation *enforcedNonceReservation) {
	if reservation == nil || !reservation.acquired || reservation.lockKey == "" || e.cache == nil {
		return
	}
	_ = e.cache.Delete(ctx, reservation.lockKey)
	reservation.acquired = false
}

func (e *entity) persistEnforcedNonce(ctx context.Context, reservation *enforcedNonceReservation, channelID, messageID int64) error {
	if reservation == nil || !reservation.acquired || reservation.key == "" || e.cache == nil {
		return nil
	}
	record := enforcedNonceRecord{ChannelID: channelID, MessageID: messageID}
	if err := e.cache.SetTimedJSON(ctx, reservation.key, record, messageNonceTTLSeconds); err != nil {
		return err
	}
	e.releaseEnforcedNonce(ctx, reservation)
	return nil
}

func (e *entity) loadNonceDuplicateMessage(c *fiber.Ctx, reservation *enforcedNonceReservation, nonce *helper.MessageNonce) (dto.Message, bool, error) {
	if reservation == nil || reservation.record == nil {
		return dto.Message{}, false, nil
	}

	channel, err := e.ch.GetChannel(c.UserContext(), reservation.record.ChannelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = e.cache.Delete(c.UserContext(), reservation.key)
			return dto.Message{}, false, nil
		}
		return dto.Message{}, false, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	}

	rawMessage, err := e.msg.GetMessage(c.UserContext(), reservation.record.MessageID, reservation.record.ChannelID)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			_ = e.cache.Delete(c.UserContext(), reservation.key)
			return dto.Message{}, false, nil
		}
		return dto.Message{}, false, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	}

	guildID, err := e.guildIDByChannel(c.UserContext(), channel.Id)
	if err != nil {
		return dto.Message{}, false, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	}

	message, err := e.buildStoredMessageResponse(c, &channel, guildID, rawMessage)
	if err != nil {
		return dto.Message{}, false, err
	}
	message.Nonce = nonce.Clone()
	return message, true, nil
}

func (e *entity) guildIDByChannel(ctx context.Context, channelID int64) (*int64, error) {
	guildChannel, err := e.gc.GetGuildByChannel(ctx, channelID)
	if err == nil {
		return &guildChannel.GuildId, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return nil, err
}

type threadCreateResult struct {
	Channel       *model.Channel
	Position      int
	Member        *dto.ThreadMember
	MemberIds     []int64
	Initial       dto.Message
	Starter       dto.Message
	StarterReq    *SendMessageRequest
	Followup      dto.Message
	SourceMessage dto.Message
}

func (e *entity) createThreadFromMessage(c *fiber.Ctx, req *CreateThreadRequest, creatorID, guildID int64, guildChannel *model.GuildChannel, parentChannel *model.Channel, sourceMessage *model.Message, validatedAttachments []model.Attachment, userData *messageUserData) (*threadCreateResult, error) {
	threadID := idgen.Next()
	threadName := deriveThreadName(req.Name, sourceMessage, req.Content)
	threadPosition := guildChannel.Position
	creator := creatorID

	if err := e.gc.AddChannel(
		c.UserContext(),
		guildID,
		threadID,
		threadName,
		model.ChannelTypeThread,
		&parentChannel.Id,
		parentChannel.Private,
		threadPosition,
		nil,
		&creator,
		false,
	); err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}

	threadChannel, err := e.ch.GetChannel(c.UserContext(), threadID)
	if err != nil {
		_ = e.gc.RemoveChannel(c.UserContext(), guildID, threadID)
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	threadMember, err := e.tm.AddThreadMember(c.UserContext(), threadID, creatorID)
	if err != nil {
		_ = e.gc.RemoveChannel(c.UserContext(), guildID, threadID)
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}

	var initialMessageID int64
	var starterMessageID int64
	var followupMessageID int64
	var initialAttachmentIDs []int64
	var starterAttachmentIDs []int64
	var followupMessageRefCreated bool
	var threadClaimed bool
	var sourceThreadAttached bool
	var expectedSourceThread int64
	cleanupThread := func() {
		if sourceThreadAttached {
			if err := e.msg.SetThread(context.Background(), sourceMessage.Id, parentChannel.Id, expectedSourceThread); err != nil && e.log != nil {
				e.log.Error("failed to rollback source message thread after thread create failure",
					"message_id", sourceMessage.Id,
					"channel_id", parentChannel.Id,
					"error", err.Error())
			}
		}
		if threadClaimed {
			if err := e.msg.ReleaseThreadClaim(context.Background(), parentChannel.Id, sourceMessage.Id); err != nil && e.log != nil {
				e.log.Error("failed to release source message thread claim after thread create failure",
					"message_id", sourceMessage.Id,
					"channel_id", parentChannel.Id,
					"error", err.Error())
			}
		}
		if followupMessageRefCreated {
			if err := e.msg.DeleteThreadCreatedMessageRef(context.Background(), threadID); err != nil && e.log != nil {
				e.log.Error("failed to delete thread-created message ref after thread create failure",
					"thread_id", threadID,
					"error", err.Error())
			}
		}
		if followupMessageID != 0 {
			_ = e.msg.DeleteMessage(context.Background(), followupMessageID, parentChannel.Id)
		}
		if initialMessageID != 0 {
			_ = e.msg.DeleteMessage(context.Background(), initialMessageID, threadID)
		}
		if starterMessageID != 0 {
			_ = e.msg.DeleteMessage(context.Background(), starterMessageID, threadID)
		}
		for _, attachmentID := range initialAttachmentIDs {
			_ = e.at.RemoveAttachment(context.Background(), attachmentID, threadID)
		}
		for _, attachmentID := range starterAttachmentIDs {
			_ = e.at.RemoveAttachment(context.Background(), attachmentID, threadID)
		}
		_ = e.gc.RemoveChannel(context.Background(), guildID, threadID)
	}

	sourceAttachments := e.loadAttachments(c.UserContext(), parentChannel.Id, sourceMessage.Attachments)
	initialAttachmentIDs, initialAttachments, err := e.cloneAttachmentsToChannel(c.UserContext(), parentChannel.Id, threadID, sourceMessage.Attachments, sourceMessage.UserId)
	if err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}

	starterAttachmentIDs, starterAttachments, err := e.cloneAttachmentsToChannel(c.UserContext(), parentChannel.Id, threadID, []int64(req.Attachments), creatorID)
	if err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}

	emptyEmbedsJSON, err := embed.MarshalEmbeds(nil)
	if err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	manualEmbedsJSON, err := embed.MarshalEmbeds(req.Embeds)
	if err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	initialMessagePosition, err := e.allocateMessagePosition(c.UserContext(), threadID)
	if err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	initialMessageID = idgen.Next()
	if err := e.msg.CreateMessageWithMeta(
		c.UserContext(),
		initialMessageID,
		threadID,
		sourceMessage.UserId,
		sourceMessage.Content,
		initialAttachmentIDs,
		stringValue(sourceMessage.EmbedsJSON),
		stringValue(sourceMessage.AutoEmbedsJSON),
		model.NormalizeMessageFlags(sourceMessage.Flags),
		model.MessageTypeThreadInitial,
		parentChannel.Id,
		sourceMessage.Id,
		0,
		initialMessagePosition,
	); err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	initialDTO := e.buildThreadMessageDTO(
		c,
		initialMessageID,
		threadID,
		sourceMessage.UserId,
		guildID,
		sourceMessage.Content,
		initialMessagePosition,
		initialAttachmentIDs,
		initialAttachments,
		sourceMessage.EmbedsJSON,
		sourceMessage.AutoEmbedsJSON,
		model.NormalizeMessageFlags(sourceMessage.Flags),
		int(model.MessageTypeThreadInitial),
		parentChannel.Id,
		sourceMessage.Id,
		0,
		sourceMessage.EditedAt,
	)

	starterReq := req.MessageRequest(starterAttachmentIDs)

	starterMessagePosition, err := e.allocateMessagePosition(c.UserContext(), threadID)
	if err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	starterMessageID = idgen.Next()
	if err := e.msg.CreateMessageWithMeta(
		c.UserContext(),
		starterMessageID,
		threadID,
		creatorID,
		req.Content,
		starterAttachmentIDs,
		manualEmbedsJSON,
		emptyEmbedsJSON,
		0,
		model.MessageTypeChat,
		0,
		0,
		0,
		starterMessagePosition,
	); err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	if err := e.ch.SetLastMessage(c.UserContext(), threadID, starterMessageID); err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	if err := e.gclm.SetChannelLastMessage(c.UserContext(), guildID, threadID, starterMessageID); err != nil {
		slog.Error("unable to set thread last message id", slog.String("error", err.Error()))
	}

	starterDTO, err := e.buildMessageResponse(c, starterMessageID, &threadChannel, starterMessagePosition, userData, starterReq, starterAttachments)
	if err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}

	expectedSourceThread = sourceMessage.Thread
	applied, _, err := e.msg.ClaimThread(c.UserContext(), parentChannel.Id, sourceMessage.Id, threadID)
	if err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	if !applied {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusConflict, ErrThreadAlreadyExists)
	}
	threadClaimed = true
	if err := e.msg.SetThread(c.UserContext(), sourceMessage.Id, parentChannel.Id, threadID); err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	sourceThreadAttached = true
	sourceMessage.Thread = threadID
	sourceDTO := e.buildThreadMessageDTO(
		c,
		sourceMessage.Id,
		parentChannel.Id,
		sourceMessage.UserId,
		guildID,
		sourceMessage.Content,
		sourceMessage.Position,
		sourceMessage.Attachments,
		sourceAttachments,
		sourceMessage.EmbedsJSON,
		sourceMessage.AutoEmbedsJSON,
		model.NormalizeMessageFlags(sourceMessage.Flags),
		sourceMessage.Type,
		sourceMessageReferenceChannel(sourceMessage),
		sourceMessage.Reference,
		threadID,
		sourceMessage.EditedAt,
	)

	followupMessagePosition, err := e.allocateMessagePosition(c.UserContext(), parentChannel.Id)
	if err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	followupMessageID = idgen.Next()
	followupContent := deriveThreadCreationMessageContent(threadName)
	if err := e.msg.CreateMessageWithMeta(
		c.UserContext(),
		followupMessageID,
		parentChannel.Id,
		creatorID,
		followupContent,
		nil,
		emptyEmbedsJSON,
		emptyEmbedsJSON,
		0,
		model.MessageTypeThreadCreated,
		parentChannel.Id,
		sourceMessage.Id,
		threadID,
		followupMessagePosition,
	); err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	if err := e.ch.SetLastMessage(c.UserContext(), parentChannel.Id, followupMessageID); err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	if err := e.gclm.SetChannelLastMessage(c.UserContext(), guildID, parentChannel.Id, followupMessageID); err != nil {
		slog.Error("unable to set parent channel last message id", slog.String("error", err.Error()))
	}
	if err := e.msg.CreateThreadCreatedMessageRef(c.UserContext(), threadID, parentChannel.Id, followupMessageID); err != nil {
		cleanupThread()
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateThread)
	}
	followupMessageRefCreated = true
	followupDTO := e.buildThreadMessageDTO(c, followupMessageID, parentChannel.Id, creatorID, guildID, followupContent, followupMessagePosition, nil, nil, nil, nil, 0, int(model.MessageTypeThreadCreated), parentChannel.Id, sourceMessage.Id, threadID, nil)

	for _, attachment := range validatedAttachments {
		if err := e.at.RemoveAttachment(c.UserContext(), attachment.Id, parentChannel.Id); err != nil && e.log != nil {
			e.log.Error("failed to remove temporary parent attachment after thread create",
				"attachment_id", attachment.Id,
				"channel_id", parentChannel.Id,
				"error", err.Error())
		}
	}
	if err := e.rs.SetReadState(c.UserContext(), creatorID, threadID, starterMessageID); err != nil && e.log != nil {
		e.log.Error("unable to set thread read state after thread create", slog.String("error", err.Error()))
	}
	if err := e.rs.SetReadState(c.UserContext(), creatorID, parentChannel.Id, followupMessageID); err != nil && e.log != nil {
		e.log.Error("unable to set parent read state after thread create", slog.String("error", err.Error()))
	}
	if err := e.ch.AdjustMessageCount(c.UserContext(), threadID, 2); err != nil {
		if e.log != nil {
			e.log.Error("unable to persist initial thread message count",
				"thread_id", threadID,
				"error", err.Error())
		}
		e.bumpThreadMessageCount(c.UserContext(), threadID, 2)
	}

	threadChannel.LastMessage = starterMessageID
	threadChannel.MessageCount = 2
	threadChannel.Permissions = parentChannel.Permissions
	threadMetadata := e.dtoThreadChannel(&threadChannel, guildID, threadPosition, nil, []int64{creatorID})
	sourceDTO.Thread = cloneChannelDTO(&threadMetadata)
	followupDTO.Thread = cloneChannelDTO(&threadMetadata)

	return &threadCreateResult{
		Channel:       &threadChannel,
		Position:      threadPosition,
		Member:        buildMessageThreadMemberDTO(&threadMember),
		MemberIds:     []int64{creatorID},
		Initial:       initialDTO,
		Starter:       starterDTO,
		StarterReq:    starterReq,
		Followup:      followupDTO,
		SourceMessage: sourceDTO,
	}, nil
}

func (e *entity) cloneAttachmentsToChannel(ctx context.Context, sourceChannelID, targetChannelID int64, attachmentIDs []int64, fallbackAuthorID int64) ([]int64, []model.Attachment, error) {
	if len(attachmentIDs) == 0 {
		return nil, nil, nil
	}

	sourceAttachments, err := e.at.SelectAttachmentsByChannel(ctx, sourceChannelID, attachmentIDs)
	if err != nil {
		return nil, nil, err
	}
	if len(sourceAttachments) != len(attachmentIDs) {
		return nil, nil, fmt.Errorf("missing source attachments")
	}

	byID := make(map[int64]model.Attachment, len(sourceAttachments))
	for _, attachment := range sourceAttachments {
		byID[attachment.Id] = attachment
	}

	clonedAttachmentIDs := make([]int64, 0, len(attachmentIDs))
	clonedAttachments := make([]model.Attachment, 0, len(attachmentIDs))
	cleanupCreatedAttachments := func() {
		for _, attachmentID := range clonedAttachmentIDs {
			_ = e.at.RemoveAttachment(ctx, attachmentID, targetChannelID)
		}
	}

	for _, attachmentID := range attachmentIDs {
		sourceAttachment, ok := byID[attachmentID]
		if !ok {
			cleanupCreatedAttachments()
			return nil, nil, fmt.Errorf("source attachment %d not found", attachmentID)
		}

		newAttachmentID := idgen.Next()
		authorID := fallbackAuthorID
		if sourceAttachment.AuthorId != nil {
			authorID = *sourceAttachment.AuthorId
		}
		if err := e.at.CreateAttachment(ctx, newAttachmentID, targetChannelID, authorID, e.attachTTL, sourceAttachment.FileSize, sourceAttachment.Name); err != nil {
			cleanupCreatedAttachments()
			return nil, nil, err
		}

		fileSize := sourceAttachment.FileSize
		name := sourceAttachment.Name
		doneAuthorID := authorID
		if err := e.at.DoneAttachment(
			ctx,
			newAttachmentID,
			targetChannelID,
			sourceAttachment.ContentType,
			sourceAttachment.URL,
			sourceAttachment.PreviewURL,
			sourceAttachment.Height,
			sourceAttachment.Width,
			&fileSize,
			&name,
			&doneAuthorID,
		); err != nil {
			_ = e.at.RemoveAttachment(ctx, newAttachmentID, targetChannelID)
			cleanupCreatedAttachments()
			return nil, nil, err
		}

		clonedAttachmentIDs = append(clonedAttachmentIDs, newAttachmentID)
		attachmentAuthorID := doneAuthorID
		clonedAttachments = append(clonedAttachments, model.Attachment{
			Id:          newAttachmentID,
			ChannelId:   targetChannelID,
			Name:        sourceAttachment.Name,
			FileSize:    sourceAttachment.FileSize,
			ContentType: sourceAttachment.ContentType,
			Height:      sourceAttachment.Height,
			Width:       sourceAttachment.Width,
			URL:         sourceAttachment.URL,
			PreviewURL:  sourceAttachment.PreviewURL,
			AuthorId:    &attachmentAuthorID,
			Done:        true,
		})
	}

	return clonedAttachmentIDs, clonedAttachments, nil
}

func (e *entity) loadAttachments(ctx context.Context, channelID int64, attachmentIDs []int64) []model.Attachment {
	if len(attachmentIDs) == 0 {
		return nil
	}

	attachments, err := e.at.SelectAttachmentsByChannel(ctx, channelID, attachmentIDs)
	if err != nil {
		if e.log != nil {
			e.log.Error("failed to load attachments",
				"channel_id", channelID,
				"error", err.Error())
		}
		return nil
	}

	return attachments
}

func deriveThreadName(explicitName string, sourceMessage *model.Message, starterContent string) string {
	name := strings.TrimSpace(explicitName)
	if name == "" {
		name = strings.Join(strings.Fields(sourceMessage.Content), " ")
	}
	if name == "" {
		name = strings.Join(strings.Fields(starterContent), " ")
	}
	if name == "" {
		name = fmt.Sprintf("thread-%d", sourceMessage.Id)
	}
	runes := []rune(name)
	if len(runes) > maxThreadNameLength {
		name = string(runes[:maxThreadNameLength])
	}
	return name
}

func deriveThreadCreationMessageContent(threadName string) string {
	return strings.TrimSpace(threadName)
}

func (e *entity) buildThreadMessageDTO(c *fiber.Ctx, messageID, channelID, authorID, guildID int64, content string, messagePosition int64, attachmentIDs []int64, attachments []model.Attachment, manualEmbedsJSON, autoEmbedsJSON *string, flags, msgType int, referenceChannel, reference, thread int64, updatedAt *time.Time) dto.Message {
	return dto.Message{
		Id:                 messageID,
		ChannelId:          channelID,
		Author:             e.buildMessageAuthor(c, authorID, &guildID),
		Content:            content,
		Position:           optionalInt64(messagePosition),
		Attachments:        e.buildAttachmentDTOs(attachmentIDs, attachments),
		Embeds:             e.mergedMessageEmbeds(messageID, manualEmbedsJSON, autoEmbedsJSON, flags),
		Flags:              flags,
		Type:               msgType,
		Reference:          optionalInt64(reference),
		ReferenceChannelId: optionalReferenceChannelID(channelID, referenceChannel, reference),
		ThreadId:           optionalInt64(thread),
		Thread:             nil,
		UpdatedAt:          updatedAt,
	}
}

func (e *entity) buildMessageAuthor(c *fiber.Ctx, userID int64, guildID *int64) dto.User {
	userData, err := e.fetchUserDataForUpdate(c, userID, guildID)
	if err != nil {
		return dto.User{Id: userID, Name: strconv.FormatInt(userID, 10)}
	}

	author := dto.User{
		Id:            userData.User.Id,
		Name:          userData.DisplayName,
		Discriminator: userData.Discriminator.Discriminator,
	}
	if userData.Avatar != nil {
		if ad, err := e.getAvatarDataCached(c.UserContext(), userData.User.Id, *userData.Avatar); err == nil && ad != nil {
			author.Avatar = ad
		}
	} else if userData.User.Avatar != nil {
		if ad, err := e.getAvatarDataCached(c.UserContext(), userData.User.Id, *userData.User.Avatar); err == nil && ad != nil {
			author.Avatar = ad
		}
	}
	return author
}

func buildMessageThreadMemberDTO(member *model.ThreadMember) *dto.ThreadMember {
	if member == nil {
		return nil
	}
	return &dto.ThreadMember{
		UserId:        member.UserId,
		JoinTimestamp: member.JoinAt,
		Flags:         member.Flags,
	}
}

func buildMessageThreadMemberIDs(members []model.ThreadMember) []int64 {
	if len(members) == 0 {
		return nil
	}
	userIDs := make([]int64, 0, len(members))
	seen := make(map[int64]struct{}, len(members))
	for _, member := range members {
		if _, ok := seen[member.UserId]; ok {
			continue
		}
		seen[member.UserId] = struct{}{}
		userIDs = append(userIDs, member.UserId)
	}
	return userIDs
}

func buildMessageThreadMemberIDsByThread(members []model.ThreadMember) map[int64][]int64 {
	if len(members) == 0 {
		return map[int64][]int64{}
	}
	memberIDsByThread := make(map[int64][]int64)
	seenByThread := make(map[int64]map[int64]struct{})
	for _, member := range members {
		seen, ok := seenByThread[member.ThreadId]
		if !ok {
			seen = make(map[int64]struct{})
			seenByThread[member.ThreadId] = seen
		}
		if _, ok := seen[member.UserId]; ok {
			continue
		}
		seen[member.UserId] = struct{}{}
		memberIDsByThread[member.ThreadId] = append(memberIDsByThread[member.ThreadId], member.UserId)
	}
	return memberIDsByThread
}

func (e *entity) applyThreadMessageCount(ctx context.Context, channel *model.Channel) {
	if channel == nil || channel.Type != model.ChannelTypeThread || e.cache == nil {
		return
	}
	delta, err := e.cache.GetInt64(ctx, threadcount.DeltaKey(channel.Id))
	if err != nil || delta <= 0 {
		return
	}
	channel.MessageCount += delta
}

func (e *entity) bumpThreadMessageCount(ctx context.Context, threadID int64, delta int64) {
	if threadID == 0 || delta <= 0 || e.cache == nil {
		return
	}
	for i := int64(0); i < delta; i++ {
		if _, err := e.cache.Incr(ctx, threadcount.DeltaKey(threadID)); err != nil {
			if e.log != nil {
				e.log.Error("failed to increment cached thread message count",
					"thread_id", threadID,
					"delta", delta,
					"error", err.Error())
			}
			return
		}
	}
	if err := e.cache.SetTTL(ctx, threadcount.DeltaKey(threadID), threadcount.DeltaTTLSeconds); err != nil && e.log != nil {
		e.log.Error("failed to refresh cached thread message count ttl",
			"thread_id", threadID,
			"error", err.Error())
	}
}

func optionalMessageThreadCount(channel *model.Channel) *int64 {
	if channel == nil || channel.Type != model.ChannelTypeThread {
		return nil
	}
	count := channel.MessageCount
	return &count
}

func (e *entity) dtoThreadChannel(channel *model.Channel, guildID int64, position int, member *dto.ThreadMember, memberIDs []int64) dto.Channel {
	return dto.Channel{
		Id:            channel.Id,
		Type:          channel.Type,
		GuildId:       &guildID,
		CreatorId:     channel.CreatorID,
		Member:        member,
		MemberIds:     memberIDs,
		Name:          channel.Name,
		ParentId:      channel.ParentID,
		Position:      position,
		Topic:         channel.Topic,
		Permissions:   channel.Permissions,
		Private:       channel.Private,
		Closed:        channel.Closed,
		LastMessageId: channel.LastMessage,
		MessageCount:  optionalMessageThreadCount(channel),
		VoiceRegion:   channel.VoiceRegion,
		CreatedAt:     channel.CreatedAt,
	}
}

func (e *entity) sendThreadCreateEvents(guildID int64, parentChannel *model.Channel, result *threadCreateResult, userData *messageUserData) {
	threadDTO := e.dtoThreadChannel(result.Channel, guildID, result.Position, nil, result.MemberIds)
	if err := e.mqt.SendGuildUpdate(guildID, &mqmsg.CreateChannel{
		GuildId: &guildID,
		Channel: threadDTO,
	}); err != nil {
		e.log.Error("failed to send channel create event for thread",
			"thread_id", result.Channel.Id,
			"error", err.Error())
	}
	if err := e.mqt.SendGuildUpdate(guildID, &mqmsg.CreateThread{
		GuildId: &guildID,
		Thread:  threadDTO,
	}); err != nil {
		e.log.Error("failed to send thread create event",
			"thread_id", result.Channel.Id,
			"error", err.Error())
	}

	e.sendUpdateEvent(parentChannel.Id, &guildID, result.SourceMessage)
	e.sendMessageCreateEvent(result.Channel, &guildID, result.Initial)
	e.dispatchMessageSideEffects(result.Channel, &guildID, result.Starter, userData, result.StarterReq)
	e.sendMessageCreateEvent(parentChannel, &guildID, result.Followup)
	if err := e.cache.Delete(context.Background(), fmt.Sprintf("guild:%d:channels", guildID)); err != nil && e.log != nil {
		e.log.Error("failed to invalidate guild channel cache after thread create",
			"guild_id", guildID,
			"error", err.Error())
	}
}

func optionalInt64(value int64) *int64 {
	if value == 0 {
		return nil
	}
	v := value
	return &v
}

func (e *entity) allocateMessagePosition(ctx context.Context, channelID int64) (int64, error) {
	return messageposition.Next(ctx, e.cache, e.ch, channelID)
}

func optionalReferenceChannelID(messageChannelID, referenceChannelID, referenceID int64) *int64 {
	if referenceID == 0 {
		return nil
	}
	if referenceChannelID == 0 {
		referenceChannelID = messageChannelID
	}
	return optionalInt64(referenceChannelID)
}

func cloneChannelDTO(channel *dto.Channel) *dto.Channel {
	if channel == nil {
		return nil
	}
	cloned := *channel
	if cloned.GuildId != nil {
		guildID := *cloned.GuildId
		cloned.GuildId = &guildID
	}
	if cloned.ParticipantId != nil {
		participantID := *cloned.ParticipantId
		cloned.ParticipantId = &participantID
	}
	if cloned.CreatorId != nil {
		creatorID := *cloned.CreatorId
		cloned.CreatorId = &creatorID
	}
	if cloned.Member != nil {
		member := *cloned.Member
		cloned.Member = &member
	}
	if cloned.MemberIds != nil {
		cloned.MemberIds = append([]int64(nil), cloned.MemberIds...)
	}
	if cloned.MessageCount != nil {
		messageCount := *cloned.MessageCount
		cloned.MessageCount = &messageCount
	}
	if cloned.ParentId != nil {
		parentID := *cloned.ParentId
		cloned.ParentId = &parentID
	}
	if cloned.Topic != nil {
		topic := *cloned.Topic
		cloned.Topic = &topic
	}
	if cloned.Permissions != nil {
		permissions := *cloned.Permissions
		cloned.Permissions = &permissions
	}
	if cloned.VoiceRegion != nil {
		voiceRegion := *cloned.VoiceRegion
		cloned.VoiceRegion = &voiceRegion
	}
	if cloned.Roles != nil {
		cloned.Roles = append([]int64(nil), cloned.Roles...)
	}
	return &cloned
}

func (e *entity) lookupThreadMetadata(ctx context.Context, threadID int64) *dto.Channel {
	if threadID == 0 {
		return nil
	}

	threadChannel, err := e.ch.GetChannel(ctx, threadID)
	if err != nil || threadChannel.Type != model.ChannelTypeThread {
		return nil
	}
	e.applyThreadMessageCount(ctx, &threadChannel)

	guildChannel, err := e.gc.GetGuildByChannel(ctx, threadID)
	if err != nil {
		return nil
	}

	threadMembers, err := e.tm.GetThreadMembers(ctx, threadID)
	if err != nil {
		return nil
	}

	threadDTO := e.dtoThreadChannel(&threadChannel, guildChannel.GuildId, guildChannel.Position, nil, buildMessageThreadMemberIDs(threadMembers))
	return cloneChannelDTO(&threadDTO)
}

func (e *entity) threadMetadataFromCache(threadID int64, data *messageRelatedData) *dto.Channel {
	if threadID == 0 || data == nil {
		return nil
	}
	thread, ok := data.Threads[threadID]
	if !ok {
		return nil
	}
	return cloneChannelDTO(&thread)
}

func sourceMessageReferenceChannel(message *model.Message) int64 {
	if message == nil {
		return 0
	}
	if message.Reference != 0 && message.ReferenceChannel != 0 {
		return message.ReferenceChannel
	}
	return message.ChannelId
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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
	if channel.Type == model.ChannelTypeGuild || channel.Type == model.ChannelTypeThread {
		guildChannel, err := e.gc.GetGuildByChannel(c.UserContext(), channelId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to get guild channel")
		}

		if !errors.Is(err, sql.ErrNoRows) {
			isMember, err := e.m.IsGuildMember(c.UserContext(), guildChannel.GuildId, userId)
			if err != nil {
				return nil, nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuild)
			}
			if !isMember {
				return nil, nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
			}

			requiredPerm := permissions.PermTextSendMessage
			if channel.Type == model.ChannelTypeThread {
				requiredPerm = permissions.PermTextSendMessageInThreads
				if channel.Closed {
					return nil, nil, fiber.NewError(fiber.StatusForbidden, ErrThreadClosed)
				}
			}
			_, _, _, canSend, err := e.perm.ChannelPerm(c.UserContext(), guildChannel.GuildId, guildChannel.ChannelId, userId, requiredPerm)
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
func (e *entity) createAndSendMessage(c *fiber.Ctx, req *SendMessageRequest, jwtUser *helper.JWTUser, channel *model.Channel, guildId *int64, validatedAttachments []model.Attachment) (dto.Message, error) {
	// Fetch user data concurrently
	userData, err := e.fetchUserDataForMessage(c, jwtUser.Id)
	if err != nil {
		return dto.Message{}, err
	}

	if channel.Type == model.ChannelTypeThread {
		if _, err := e.tm.AddThreadMember(c.UserContext(), channel.Id, jwtUser.Id); err != nil {
			return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
		}
	}

	// Create message with transaction-like behavior
	messageId := idgen.Next()
	position, err := e.allocateMessagePosition(c.UserContext(), channel.Id)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	}
	if err := e.createMessageWithCleanup(c, messageId, channel.Id, jwtUser.Id, position, req); err != nil {
		return dto.Message{}, err
	}

	// Build response message
	message, err := e.buildMessageResponse(c, messageId, channel, position, userData, req, validatedAttachments)
	if err != nil {
		// Cleanup on failure
		_ = e.msg.DeleteMessage(c.UserContext(), messageId, channel.Id)
		return dto.Message{}, err
	}
	if channel.Type == model.ChannelTypeThread {
		e.bumpThreadMessageCount(c.UserContext(), channel.Id, 1)
	}

	if guildId != nil {
		err := e.gclm.SetChannelLastMessage(c.UserContext(), *guildId, channel.Id, messageId)
		if err != nil {
			slog.Error("unable to set guild channel last message id", slog.String("error", err.Error()))
		}
	}

	e.dispatchMessageSideEffects(channel, guildId, message, userData, req)

	return message, nil
}

func (e *entity) dispatchMessageSideEffects(channel *model.Channel, guildId *int64, message dto.Message, userData *messageUserData, req *SendMessageRequest) {
	if channel == nil || userData == nil || req == nil {
		return
	}

	go e.sendMessageEvents(channel, guildId, message, userData, req)

	if !model.IsEditableMessageType(model.MessageType(message.Type)) {
		return
	}

	if HasURL(message.Content) {
		go e.enqueueMakeEmbed(guildId, message)
	}

	users, roles, everyone, here := MentionsExtractor(req.Content)
	if users == nil && roles == nil && !everyone && !here {
		return
	}

	go func() {
		for _, u := range users {
			switch channel.Type {
			case model.ChannelTypeGuild, model.ChannelTypeThread:
				if guildId != nil {
					if ok, err := e.m.IsGuildMember(context.Background(), *guildId, u); err == nil && ok {
						if err := e.mention.AddMention(context.Background(), u, channel.Id, message.Id, message.Author.Id); err != nil {
							e.log.Error("unable to save mention", slog.String("error", err.Error()))
						}
						e.sendMentionUserUpdate(u, guildId, channel.Id, message.Id, message.Author.Id, model.ChannelMentionUser)
					}
				}
			default:
				if ok, err := e.fr.IsFriend(context.Background(), u, message.Author.Id); err == nil && ok {
					if err := e.mention.AddMention(context.Background(), u, channel.Id, message.Id, message.Author.Id); err != nil {
						e.log.Error("unable to save mention", slog.String("error", err.Error()))
					}
					e.sendMentionUserUpdate(u, nil, channel.Id, message.Id, message.Author.Id, model.ChannelMentionUser)
				}
			}
		}
		if guildId != nil {
			threadMentionRecipients := []int64(nil)
			if channel.Type == model.ChannelTypeThread {
				recipients, err := e.threadMemberUserIDs(context.Background(), channel.Id, message.Author.Id)
				if err != nil {
					e.log.Error("unable to get thread mention recipients", slog.String("error", err.Error()))
				} else {
					threadMentionRecipients = recipients
				}
			}
			for _, r := range roles {
				if err := e.mention.AddChannelMention(
					context.Background(),
					*guildId,
					channel.Id,
					message.Id,
					message.Author.Id,
					&r,
					model.ChannelMentionRole); err != nil {
					e.log.Error("unable to save role mention", slog.String("error", err.Error()))
				}
				if channel.Type == model.ChannelTypeThread {
					recipients, err := e.threadRoleMentionRecipients(context.Background(), *guildId, threadMentionRecipients, r)
					if err != nil {
						e.log.Error("unable to resolve thread role mention recipients", slog.String("error", err.Error()))
						continue
					}
					for _, userID := range recipients {
						e.sendMentionUserUpdate(userID, guildId, channel.Id, message.Id, message.Author.Id, model.ChannelMentionRole)
					}
				} else if err := e.mqt.SendGuildUpdate(*guildId, &mqmsg.Mention{
					GuildId:   guildId,
					ChannelId: channel.Id,
					MessageId: message.Id,
					AuthorId:  message.Author.Id,
					Type:      int(model.ChannelMentionRole),
				}); err != nil {
					e.log.Error("unable to send role mention notification", slog.String("error", err.Error()))
				}
			}
			if everyone {
				if err := e.mention.AddChannelMention(
					context.Background(),
					*guildId,
					channel.Id,
					message.Id,
					message.Author.Id,
					nil,
					model.ChannelMentionEveryone); err != nil {
					e.log.Error("unable to save role mention", slog.String("error", err.Error()))
				}
				if channel.Type == model.ChannelTypeThread {
					for _, userID := range threadMentionRecipients {
						e.sendMentionUserUpdate(userID, guildId, channel.Id, message.Id, message.Author.Id, model.ChannelMentionEveryone)
					}
				} else if err := e.mqt.SendGuildUpdate(*guildId, &mqmsg.Mention{
					GuildId:   guildId,
					ChannelId: channel.Id,
					MessageId: message.Id,
					AuthorId:  message.Author.Id,
					Type:      int(model.ChannelMentionEveryone),
				}); err != nil {
					e.log.Error("unable to send role mention notification", slog.String("error", err.Error()))
				}
			}
			if here {
				if channel.Type == model.ChannelTypeThread {
					for _, userID := range threadMentionRecipients {
						e.sendMentionUserUpdate(userID, guildId, channel.Id, message.Id, message.Author.Id, model.ChannelMentionHere)
					}
				} else if err := e.mqt.SendGuildUpdate(*guildId, &mqmsg.Mention{
					GuildId:   guildId,
					ChannelId: channel.Id,
					MessageId: message.Id,
					AuthorId:  message.Author.Id,
					Type:      int(model.ChannelMentionHere),
				}); err != nil {
					e.log.Error("unable to send role mention notification", slog.String("error", err.Error()))
				}
			}
		}
	}()
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
func (e *entity) createMessageWithCleanup(c *fiber.Ctx, messageId, channelId, userId, position int64, req *SendMessageRequest) error {
	manualEmbedsJSON, err := embed.MarshalEmbeds(req.Embeds)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	autoEmbedsJSON, err := embed.MarshalEmbeds(nil)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	referenceID := requestedReferenceID(req)
	msgType := requestedMessageType(req)
	referenceChannelID := int64(0)
	if referenceID != 0 {
		referenceChannelID = channelId
	}

	// Create the message
	if err := e.msg.CreateMessageWithMeta(
		c.UserContext(),
		messageId,
		channelId,
		userId,
		req.Content,
		[]int64(req.Attachments),
		manualEmbedsJSON,
		autoEmbedsJSON,
		0,
		msgType,
		referenceChannelID,
		referenceID,
		0,
		position,
	); err != nil {
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
func (e *entity) validateMessageAttachments(ctx context.Context, channelId, userId int64, attachmentIds []int64) ([]model.Attachment, error) {
	if len(attachmentIds) == 0 {
		return nil, nil
	}

	requested := make(map[int64]struct{}, len(attachmentIds))
	for _, attachmentId := range attachmentIds {
		if _, exists := requested[attachmentId]; exists {
			return nil, fiber.NewError(fiber.StatusBadRequest, ErrInvalidAttachments)
		}
		requested[attachmentId] = struct{}{}
	}

	attachments, err := e.at.SelectAttachmentsByChannel(ctx, channelId, attachmentIds)
	if err != nil {
		return nil, helper.HttpDbError(err, ErrUnableToGetAttachements)
	}
	if len(attachments) != len(requested) {
		return nil, fiber.NewError(fiber.StatusBadRequest, ErrInvalidAttachments)
	}

	attachmentMap := make(map[int64]model.Attachment, len(attachments))
	for _, attachment := range attachments {
		if attachment.ChannelId != channelId || !attachment.Done || attachment.AuthorId == nil || *attachment.AuthorId != userId {
			return nil, fiber.NewError(fiber.StatusBadRequest, ErrInvalidAttachments)
		}
		attachmentMap[attachment.Id] = attachment
	}

	result := make([]model.Attachment, 0, len(attachmentIds))
	for _, attachmentId := range attachmentIds {
		attachment, exists := attachmentMap[attachmentId]
		if !exists {
			return nil, fiber.NewError(fiber.StatusBadRequest, ErrInvalidAttachments)
		}
		result = append(result, attachment)
	}

	return result, nil
}

func (e *entity) buildAttachmentDTOs(attachmentIds []int64, attachments []model.Attachment) []dto.Attachment {
	if len(attachmentIds) == 0 || len(attachments) == 0 {
		return nil
	}

	attachmentMap := make(map[int64]model.Attachment, len(attachments))
	for _, attachment := range attachments {
		attachmentMap[attachment.Id] = attachment
	}

	result := make([]dto.Attachment, 0, len(attachmentIds))
	for _, attachmentId := range attachmentIds {
		attachment, exists := attachmentMap[attachmentId]
		if !exists {
			continue
		}
		var url string
		if attachment.URL != nil {
			url = *attachment.URL
		}
		result = append(result, dto.Attachment{
			ContentType: attachment.ContentType,
			Filename:    attachment.Name,
			Height:      attachment.Height,
			Width:       attachment.Width,
			URL:         url,
			PreviewURL:  attachment.PreviewURL,
			Size:        attachment.FileSize,
		})
	}
	return result
}

// buildMessageResponse constructs the message response DTO
func (e *entity) buildMessageResponse(c *fiber.Ctx, messageId int64, channel *model.Channel, position int64, userData *messageUserData, req *SendMessageRequest, validatedAttachments []model.Attachment) (dto.Message, error) {
	attachments := e.buildAttachmentDTOs([]int64(req.Attachments), validatedAttachments)

	// Build author with avatar data if present
	author := dto.User{
		Id:            userData.User.Id,
		Name:          userData.User.Name,
		Discriminator: userData.Discriminator.Discriminator,
	}
	if userData.User.Avatar != nil {
		if ad, err := e.getAvatarDataCached(c.UserContext(), userData.User.Id, *userData.User.Avatar); err == nil && ad != nil {
			author.Avatar = ad
		}
	}

	return dto.Message{
		Id:                 messageId,
		ChannelId:          channel.Id,
		Author:             author,
		Content:            req.Content,
		Position:           optionalInt64(position),
		Nonce:              cloneMessageNonce(req.Nonce),
		Attachments:        attachments,
		Embeds:             req.Embeds,
		Flags:              0,
		Type:               int(requestedMessageType(req)),
		Reference:          optionalInt64(requestedReferenceID(req)),
		ReferenceChannelId: optionalReferenceChannelID(channel.Id, channel.Id, requestedReferenceID(req)),
		ThreadId:           nil,
	}, nil
}

func cloneMessageNonce(nonce *helper.MessageNonce) *helper.MessageNonce {
	if nonce == nil {
		return nil
	}
	return nonce.Clone()
}

func (e *entity) buildStoredMessageResponse(c *fiber.Ctx, channel *model.Channel, guildId *int64, message model.Message) (dto.Message, error) {
	data, err := e.fetchMessageRelatedData(c, []model.Message{message}, []int64{message.UserId}, guildId)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	}

	messages := e.buildMessageDTOsOptimized([]model.Message{message}, data)
	if len(messages) != 1 {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendMessage)
	}

	if channel != nil && channel.Type == model.ChannelTypeThread {
		e.applyThreadMessageCount(c.UserContext(), channel)
		if messages[0].Thread != nil {
			if cloned := cloneChannelDTO(messages[0].Thread); cloned != nil {
				messages[0].Thread = cloned
			}
		}
	}

	return messages[0], nil
}

func (e *entity) parseMessageEmbeds(messageId int64, raw *string) []embed.Embed {
	embeds, err := embed.ParseEmbeds(raw)
	if err != nil {
		if e.log != nil {
			e.log.Error("failed to decode message embeds",
				"message_id", messageId,
				"error", err.Error())
		}
		return nil
	}

	return embeds
}

func (e *entity) mergedMessageEmbeds(messageId int64, manualRaw, autoRaw *string, flags int) []embed.Embed {
	embeds, err := embed.ParseMergedEmbeds(manualRaw, autoRaw, model.HasMessageFlag(flags, model.MessageFlagSuppressEmbeds))
	if err != nil {
		if e.log != nil {
			e.log.Error("failed to decode message embeds",
				"message_id", messageId,
				"error", err.Error())
		}
		return nil
	}

	return embeds
}

func (e *entity) enqueueMakeEmbed(guildId *int64, message dto.Message) {
	if e.emq == nil {
		return
	}
	if err := e.emq.MakeEmbed(embedmq.MakeEmbedRequest{GuildId: guildId, Message: message}); err != nil && e.log != nil {
		e.log.Error("failed to enqueue embed generation",
			"message_id", message.Id,
			"channel_id", message.ChannelId,
			"error", err.Error())
	}
}

// sendMessageEvents sends message and indexing events asynchronously
func (e *entity) sendMessageEvents(channel *model.Channel, guildId *int64, message dto.Message, userData *messageUserData, req *SendMessageRequest) {
	e.sendMessageCreateEvent(channel, guildId, message)

	if userData == nil || req == nil || !model.IsEditableMessageType(model.MessageType(message.Type)) {
		return
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
		ChannelId: channel.Id,
		GuildId:   guildId,
		Mentions:  req.Mentions,
		Has:       UniqueAttachmentTypes(hasTypes),
		Type:      message.Type,
		Content:   message.Content,
	}); err != nil {
		e.log.Error("failed to send index message event",
			"message_id", message.Id,
			"error", err.Error())
	}

	// Notify DM recipient via user topic for 1:1 DMs
	if guildId == nil {
		// Determine channel type and DM participants
		ch, err := e.ch.GetChannel(context.Background(), channel.Id)
		if err == nil && ch.Type == model.ChannelTypeDM {
			// Fetch both rows for this DM channel and pick the other user
			rows, rerr := e.dmc.GetDmChannelByChannelId(context.Background(), channel.Id)
			if rerr == nil {
				var recipientId int64
				for _, r := range rows {
					if r.UserId != message.Author.Id {
						recipientId = r.UserId
						break
					}
				}
				if recipientId != 0 {
					var ad *dto.AvatarData
					if userData.User.Avatar != nil {
						if v, err := e.getAvatarDataCached(context.Background(), userData.User.Id, *userData.User.Avatar); err == nil {
							ad = v
						}
					}
					_ = e.mqt.SendUserUpdate(recipientId, &mqmsg.DMMessage{
						ChannelId: channel.Id,
						MessageId: message.Id,
						From:      mqmsg.UserBrief{Id: userData.User.Id, Name: userData.User.Name, Discriminator: userData.Discriminator.Discriminator, Avatar: userData.User.Avatar, AvatarData: ad},
					})
				}
			}
		}
	}
}

func (e *entity) sendMessageCreateEvent(channel *model.Channel, guildId *int64, message dto.Message) {
	if channel == nil {
		return
	}
	channelId := channel.Id
	if err := e.mqt.SendChannelMessage(channelId, &mqmsg.CreateMessage{
		GuildId: guildId,
		Message: message,
	}); err != nil {
		e.log.Error("failed to send message event",
			"message_id", message.Id,
			"channel_id", channelId,
			"error", err.Error())
	}

	if guildId != nil {
		if channel.Type == model.ChannelTypeThread {
			e.sendThreadActivityEvent(*guildId, channelId, message.Id, message.Author.Id)
			return
		}
		if err := e.mqt.SendGuildUpdate(*guildId, &mqmsg.GuildChannelMessage{
			GuildId:   guildId,
			ChannelId: channelId,
			MessageId: message.Id,
		}); err != nil {
			e.log.Error("failed to send guild message event",
				slog.String("error", err.Error()))
		}
	}
}

func (e *entity) threadMemberUserIDs(ctx context.Context, threadID, excludeUserID int64) ([]int64, error) {
	members, err := e.tm.GetThreadMembers(ctx, threadID)
	if err != nil {
		return nil, err
	}
	userIDs := make([]int64, 0, len(members))
	seen := make(map[int64]struct{}, len(members))
	for _, member := range members {
		if member.UserId == excludeUserID {
			continue
		}
		if _, ok := seen[member.UserId]; ok {
			continue
		}
		seen[member.UserId] = struct{}{}
		userIDs = append(userIDs, member.UserId)
	}
	return userIDs, nil
}

func (e *entity) threadRoleMentionRecipients(ctx context.Context, guildID int64, memberUserIDs []int64, roleID int64) ([]int64, error) {
	if len(memberUserIDs) == 0 {
		return nil, nil
	}
	userRoles, err := e.ur.GetUsersRolesByGuild(ctx, guildID, memberUserIDs)
	if err != nil {
		return nil, err
	}
	recipients := make([]int64, 0, len(userRoles))
	for _, userRole := range userRoles {
		for _, rid := range userRole.Roles {
			if rid == roleID {
				recipients = append(recipients, userRole.UserId)
				break
			}
		}
	}
	return recipients, nil
}

func (e *entity) sendThreadActivityEvent(guildID, threadID, messageID, authorID int64) {
	userIDs, err := e.threadMemberUserIDs(context.Background(), threadID, authorID)
	if err != nil {
		e.log.Error("failed to load thread members for activity event",
			"thread_id", threadID,
			"error", err.Error())
		return
	}

	for _, userID := range userIDs {
		if err := e.mqt.SendUserUpdate(userID, &mqmsg.GuildChannelMessage{
			GuildId:   &guildID,
			ChannelId: threadID,
			MessageId: messageID,
		}); err != nil {
			e.log.Error("failed to send thread activity event",
				"user_id", userID,
				"thread_id", threadID,
				"message_id", messageID,
				"error", err.Error())
		}
	}
}

func (e *entity) sendMentionUserUpdate(userID int64, guildID *int64, channelID, messageID, authorID int64, mentionType model.ChannelMentionType) {
	if err := e.mqt.SendUserUpdate(userID, &mqmsg.Mention{
		GuildId:   guildID,
		ChannelId: channelID,
		MessageId: messageID,
		AuthorId:  authorID,
		Type:      int(mentionType),
	}); err != nil {
		e.log.Error("unable to send mention notification",
			slog.String("error", err.Error()),
			slog.Int64("user_id", userID),
			slog.Int64("channel_id", channelID),
			slog.Int64("message_id", messageID))
	}
}

const avatarCacheTTLSeconds = 3600 // 1 hour

func (e *entity) getAvatarDataCached(ctx context.Context, userId, avatarId int64) (*dto.AvatarData, error) {
	key := fmt.Sprintf("avatars:%d:%d", userId, avatarId)
	var ad dto.AvatarData
	if e.cache != nil {
		if err := e.cache.GetJSON(ctx, key, &ad); err == nil && ad.URL != "" {
			return &ad, nil
		}
	}
	av, err := e.av.GetAvatar(ctx, avatarId, userId)
	if err != nil {
		return nil, err
	}
	if av.URL == nil || *av.URL == "" {
		return nil, nil
	}
	ad = dto.AvatarData{
		URL:         *av.URL,
		ContentType: av.ContentType,
		Width:       av.Width,
		Height:      av.Height,
		Size:        av.FileSize,
	}
	if e.cache != nil {
		_ = e.cache.SetTimedJSON(ctx, key, ad, avatarCacheTTLSeconds)
	}
	return &ad, nil
}

// GetMessages
//
//	@Summary		Get messages
//	@Description	Response order depends on `direction`.
//	@Description	`before`: newest to oldest, including the `from` message when it exists. If `from` is omitted, the server starts from the channel's current `last_message_id`.
//	@Description	`after`: oldest to newest, including the `from` message.
//	@Description	`around`: the `from` message first, then older messages in descending order, then newer messages in ascending order.
//	@Produce		json
//	@Tags			Message
//	@Param			channel_id	path		int64		true	"Channel id"																															example(2230469276416868352)
//	@Param			from		query		int64		false	"Start point for messages. Included in the response when it exists."																	example(2230469276416868352)
//	@Param			direction	query		string		false	"Select direction and response order: before=newest->oldest, after=oldest->newest, around=from first then older desc then newer asc."	Enums(before, after, around)	example(before)
//	@Param			limit		query		int			false	"Message count limit"																													example(30)
//	@Success		200			{array}		dto.Message	"Messages returned in the order implied by `direction`."
//	@failure		400			{string}	string		"Bad request"
//	@failure		403			{string}	string		"Forbidden"
//	@failure		404			{string}	string		"Not found"
//	@failure		500			{string}	string		"Internal server error"
//	@Router			/message/channel/{channel_id} [get]
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
	if channel.Type == model.ChannelTypeGuild || channel.Type == model.ChannelTypeThread {
		guildChannel, err := e.gc.GetGuildByChannel(c.UserContext(), channelId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to get guild channel")
		}

		if !errors.Is(err, sql.ErrNoRows) {
			isMember, err := e.m.IsGuildMember(c.UserContext(), guildChannel.GuildId, userId)
			if err != nil {
				return nil, nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuild)
			}
			if !isMember {
				return nil, nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
			}

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
	if guildId != nil {
		if err := e.redactBannedMessages(c.UserContext(), *guildId, rawMessages, messages); err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to apply banned message visibility")
		}
	}

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
	case DirectionAround:
		if req.From == nil {
			return nil, nil, fiber.NewError(fiber.StatusBadRequest, "from parameter required for around direction")
		}
		rawMessages, userIds, err = e.msg.GetMessagesAround(c.UserContext(), channel.Id, *req.From, channel.LastMessage, *req.Limit)
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
	AvData      map[int64]*dto.AvatarData
	Threads     map[int64]dto.Channel
}

// fetchMessageRelatedData fetches users, members, attachments, and thread metadata concurrently
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
	type threadsResult struct {
		channels      []model.Channel
		guildChannels []model.GuildChannel
		members       []model.ThreadMember
		err           error
	}

	usersCh := make(chan usersResult, 1)
	membersCh := make(chan membersResult, 1)
	attachmentsCh := make(chan attachmentsResult, 1)
	threadsCh := make(chan threadsResult, 1)

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
		if len(attachmentIds) > 0 && len(messages) > 0 {
			attachments, err := e.at.SelectAttachmentsByChannel(c.UserContext(), messages[0].ChannelId, attachmentIds)
			attachmentsCh <- attachmentsResult{attachments, err}
		} else {
			attachmentsCh <- attachmentsResult{nil, nil}
		}
	}()

	go func() {
		threadIDs := e.extractThreadIDs(messages)
		if len(threadIDs) == 0 {
			threadsCh <- threadsResult{}
			return
		}

		threadChannels, err := e.ch.GetChannelsBulk(c.UserContext(), threadIDs)
		if err != nil {
			threadsCh <- threadsResult{err: err}
			return
		}
		guildChannels, err := e.gc.GetGuildChannelsByChannelIDs(c.UserContext(), threadIDs)
		if err != nil {
			threadsCh <- threadsResult{err: err}
			return
		}
		threadMembers, err := e.tm.GetThreadMembersBulk(c.UserContext(), threadIDs)
		if err != nil {
			threadsCh <- threadsResult{err: err}
			return
		}

		threadsCh <- threadsResult{
			channels:      threadChannels,
			guildChannels: guildChannels,
			members:       threadMembers,
		}
	}()

	// Collect results
	usersRes := <-usersCh
	membersRes := <-membersCh
	attachmentsRes := <-attachmentsCh
	threadsRes := <-threadsCh

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
	if threadsRes.err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "failed to fetch thread metadata")
	}

	// Build maps
	data := &messageRelatedData{
		Users:       make(map[int64]*model.User),
		Members:     make(map[int64]*model.Member),
		Attachments: make(map[int64]*model.Attachment),
		AvData:      make(map[int64]*dto.AvatarData),
		Threads:     make(map[int64]dto.Channel),
	}

	for i := range usersRes.users {
		u := usersRes.users[i]
		data.Users[u.Id] = &u
		if u.Avatar != nil {
			if ad, err := e.getAvatarDataCached(c.UserContext(), u.Id, *u.Avatar); err == nil && ad != nil {
				data.AvData[u.Id] = ad
			}
		}
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
	if threadsRes.channels != nil {
		guildChannelsByID := make(map[int64]model.GuildChannel, len(threadsRes.guildChannels))
		threadMemberIDsByThread := buildMessageThreadMemberIDsByThread(threadsRes.members)
		for _, guildChannel := range threadsRes.guildChannels {
			guildChannelsByID[guildChannel.ChannelId] = guildChannel
		}

		for i := range threadsRes.channels {
			threadChannel := threadsRes.channels[i]
			if threadChannel.Type != model.ChannelTypeThread {
				continue
			}
			e.applyThreadMessageCount(c.UserContext(), &threadChannel)
			guildChannel, ok := guildChannelsByID[threadChannel.Id]
			if !ok {
				continue
			}
			data.Threads[threadChannel.Id] = e.dtoThreadChannel(&threadChannel, guildChannel.GuildId, guildChannel.Position, nil, threadMemberIDsByThread[threadChannel.Id])
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

func (e *entity) extractThreadIDs(messages []model.Message) []int64 {
	threadMap := make(map[int64]bool)
	for _, message := range messages {
		if message.Thread != 0 {
			threadMap[message.Thread] = true
		}
	}

	threadIDs := make([]int64, 0, len(threadMap))
	for id := range threadMap {
		threadIDs = append(threadIDs, id)
	}

	return threadIDs
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

	currentFlags := model.NormalizeMessageFlags(message.Flags)
	suppressWasEnabled := model.HasMessageFlag(currentFlags, model.MessageFlagSuppressEmbeds)

	if req.Content != nil {
		sanitizedContent, err := e.sanitizeEmojiContent(c.UserContext(), user.Id, *req.Content)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUpdateMessage)
		}
		req.Content = &sanitizedContent
	}

	// Update message and build response
	updatedMessage, err := e.updateMessageAndBuildResponse(c, req, message, user, guildId)
	if err != nil {
		return err
	}

	// Send update event
	go e.sendUpdateEvent(channelId, guildId, updatedMessage)

	contentChanged := req.Content != nil && *req.Content != message.Content
	suppressIsEnabled := model.HasMessageFlag(updatedMessage.Flags, model.MessageFlagSuppressEmbeds)
	suppressLifted := suppressWasEnabled && !suppressIsEnabled
	if !suppressIsEnabled && HasURL(updatedMessage.Content) && (contentChanged || suppressLifted) {
		go e.enqueueMakeEmbed(guildId, updatedMessage)
	}

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
	channel, err := e.ch.GetChannel(c.UserContext(), channelId)
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusNotFound, "channel not found")
	}
	if channel.Type == model.ChannelTypeThread && channel.Closed {
		return nil, nil, fiber.NewError(fiber.StatusForbidden, ErrThreadClosed)
	}

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
	if !model.IsEditableMessageType(model.MessageType(message.Type)) {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, ErrMessageNotEditable)
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
	updatedContent := message.Content
	if req.Content != nil {
		updatedContent = *req.Content
	}

	updatedEmbeds, err := embed.ParseEmbeds(message.EmbedsJSON)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "failed to decode message embeds")
	}
	if req.Embeds != nil {
		updatedEmbeds = append([]embed.Embed(nil), (*req.Embeds)...)
	}

	updatedAutoEmbeds, err := embed.ParseEmbeds(message.AutoEmbedsJSON)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "failed to decode generated message embeds")
	}

	updatedFlags := model.NormalizeMessageFlags(message.Flags)
	if req.Flags != nil {
		updatedFlags = *req.Flags
	}

	contentChanged := req.Content != nil && *req.Content != message.Content
	if contentChanged || model.HasMessageFlag(updatedFlags, model.MessageFlagSuppressEmbeds) {
		updatedAutoEmbeds = nil
	}

	if updatedContent == "" && len(message.Attachments) == 0 && len(updatedEmbeds) == 0 {
		return dto.Message{}, fiber.NewError(fiber.StatusBadRequest, ErrMessagePayloadRequired)
	}

	embedsJSON, err := embed.MarshalEmbeds(updatedEmbeds)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	autoEmbedsJSON, err := embed.MarshalEmbeds(updatedAutoEmbeds)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Update the message
	if err := e.msg.UpdateMessage(c.UserContext(), message.Id, message.ChannelId, updatedContent, embedsJSON, autoEmbedsJSON, updatedFlags); err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "failed to update message")
	}

	var validatedAttachments []model.Attachment
	if len(message.Attachments) > 0 {
		validatedAttachments, err = e.at.SelectAttachmentsByChannel(c.UserContext(), message.ChannelId, message.Attachments)
		if err != nil {
			return dto.Message{}, helper.HttpDbError(err, ErrUnableToGetAttachements)
		}
	}

	// Fetch user data for response
	userData, err := e.fetchUserDataForUpdate(c, jwtUser.Id, guildId)
	if err != nil {
		return dto.Message{}, err
	}

	// Build response
	updatedAt := time.Now()
	author := dto.User{
		Id:            userData.User.Id,
		Name:          userData.DisplayName,
		Discriminator: userData.Discriminator.Discriminator,
	}
	if userData.Avatar != nil {
		if ad, err := e.getAvatarDataCached(c.UserContext(), userData.User.Id, *userData.Avatar); err == nil && ad != nil {
			author.Avatar = ad
		}
	} else if userData.User.Avatar != nil {
		if ad, err := e.getAvatarDataCached(c.UserContext(), userData.User.Id, *userData.User.Avatar); err == nil && ad != nil {
			author.Avatar = ad
		}
	}

	responseEmbeds := embed.MergeEmbeds(updatedEmbeds)
	if !model.HasMessageFlag(updatedFlags, model.MessageFlagSuppressEmbeds) {
		responseEmbeds = embed.MergeEmbeds(updatedEmbeds, updatedAutoEmbeds)
	}

	return dto.Message{
		Id:                 message.Id,
		ChannelId:          message.ChannelId,
		Author:             author,
		Content:            updatedContent,
		Position:           optionalInt64(message.Position),
		Attachments:        e.buildAttachmentDTOs(message.Attachments, validatedAttachments),
		Embeds:             responseEmbeds,
		Flags:              updatedFlags,
		Type:               message.Type,
		Reference:          optionalInt64(message.Reference),
		ReferenceChannelId: optionalReferenceChannelID(message.ChannelId, message.ReferenceChannel, message.Reference),
		ThreadId:           optionalInt64(message.Thread),
		Thread:             e.lookupThreadMetadata(c.UserContext(), message.Thread),
		UpdatedAt:          &updatedAt,
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

	if e.imq != nil {
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
	channel, err := e.ch.GetChannel(c.UserContext(), channelId)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "channel not found")
	}
	if channel.Type == model.ChannelTypeThread && channel.Closed {
		return nil, fiber.NewError(fiber.StatusForbidden, ErrThreadClosed)
	}

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

	switch channel.Type {
	case model.ChannelTypeGuild, model.ChannelTypeThread:
		guildChannel, err := e.gc.GetGuildByChannel(c.UserContext(), channelId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to get guild channel")
		}

		if !errors.Is(err, sql.ErrNoRows) {
			isMember, err := e.m.IsGuildMember(c.UserContext(), guildChannel.GuildId, userId)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuild)
			}
			if !isMember {
				return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
			}
			if channel.Type == model.ChannelTypeThread && channel.Closed {
				return fiber.NewError(fiber.StatusForbidden, ErrThreadClosed)
			}
			_, _, _, canAttach, err := e.perm.ChannelPerm(c.UserContext(), guildChannel.GuildId, guildChannel.ChannelId, userId, permissions.PermTextAttachFiles)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
			}
			if !canAttach {
				return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
			}
		}
	case model.ChannelTypeGroupDM:
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

	user, err := helper.GetUser(c)
	if err != nil {
		return dto.AttachmentUpload{}, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	if err := e.at.CreateAttachment(c.UserContext(), attachmentId, channelId, user.Id, e.attachTTL, req.FileSize, req.Filename); err != nil {
		return dto.AttachmentUpload{}, helper.HttpDbError(err, ErrUnableToCreateAttachment)
	}

	return dto.AttachmentUpload{
		Id:        attachmentId,
		ChannelId: channelId,
		FileName:  req.Filename,
	}, nil
}

// buildMessageDTOsOptimized constructs message DTOs efficiently
func (e *entity) buildMessageDTOsOptimized(messages []model.Message, data *messageRelatedData) []dto.Message {
	result := make([]dto.Message, len(messages))

	for i, message := range messages {
		flags := model.NormalizeMessageFlags(message.Flags)
		result[i] = dto.Message{
			Id:                 message.Id,
			ChannelId:          message.ChannelId,
			Author:             e.buildAuthorOptimized(message.UserId, data),
			Content:            message.Content,
			Position:           optionalInt64(message.Position),
			Attachments:        e.buildAttachmentsOptimized(message.Attachments, data),
			Embeds:             e.mergedMessageEmbeds(message.Id, message.EmbedsJSON, message.AutoEmbedsJSON, flags),
			Flags:              flags,
			UpdatedAt:          message.EditedAt,
			Type:               message.Type,
			Reference:          optionalInt64(message.Reference),
			ReferenceChannelId: optionalReferenceChannelID(message.ChannelId, message.ReferenceChannel, message.Reference),
			ThreadId:           optionalInt64(message.Thread),
			Thread:             e.threadMetadataFromCache(message.Thread, data),
		}
	}

	return result
}

// buildAuthorOptimized constructs author DTO with member override if available
func (e *entity) buildAuthorOptimized(userId int64, data *messageRelatedData) dto.User {
	user, userExists := data.Users[userId]
	if !userExists {
		return dto.User{
			Id:   userId,
			Name: "Unknown User",
		}
	}

	author := dto.User{
		Id:   userId,
		Name: user.Name,
	}
	if ad, ok := data.AvData[userId]; ok {
		author.Avatar = ad
	}

	if member, memberExists := data.Members[userId]; memberExists {
		if member.Username != nil {
			author.Name = *member.Username
		}
		if member.Avatar != nil {
			if ad, err := e.getAvatarDataCached(context.Background(), userId, *member.Avatar); err == nil && ad != nil {
				author.Avatar = ad
			}
		}
	}

	return author
}

// buildAttachmentsOptimized constructs attachment DTOs efficiently
func (e *entity) buildAttachmentsOptimized(attachmentIds []int64, data *messageRelatedData) []dto.Attachment {
	if len(attachmentIds) == 0 {
		return nil
	}

	// Pre-allocate with exact capacity to avoid reallocation
	attachments := make([]dto.Attachment, 0, len(attachmentIds))

	for _, id := range attachmentIds {
		if attachment, exists := data.Attachments[id]; exists {
			var full string
			if attachment.URL != nil {
				full = *attachment.URL
			}
			attachments = append(attachments, dto.Attachment{
				ContentType: attachment.ContentType,
				Filename:    attachment.Name,
				Height:      attachment.Height,
				Width:       attachment.Width,
				URL:         full,
				PreviewURL:  attachment.PreviewURL,
				Size:        attachment.FileSize,
			})
		}
	}

	return attachments
}

// SetReadState
//
//	@Summary	Set channel read state for current user
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id	path		int64	true	"Channel id"
//	@Param		message_id	path		int64	true	"Message id"
//	@Success	200			{string}	string	"Read state updated"
//	@failure	400			{string}	string	"Bad request"
//	@failure	500			{string}	string	"Internal server error"
//	@Router		/message/channel/{channel_id}/{message_id}/ack [post]
func (e *entity) SetReadState(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	channelIdStr := c.Params("channel_id")
	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}

	messageIdStr := c.Params("message_id")
	messageId, err := strconv.ParseInt(messageIdStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}

	// Validate channel and permissions
	channel, _, err := e.validateReadPermissions(c, channelId, user.Id)
	if err != nil {
		return err
	}

	err = e.rs.SetReadState(c.UserContext(), user.Id, channel.Id, messageId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSetReadState)
	}

	go func() {
		if err := e.mqt.SendUserUpdate(user.Id, &mqmsg.UpdateReadState{
			ChannelId: channelId,
			MessageId: messageId,
		}); err != nil {
			slog.Error("unable to send user update read state event",
				slog.String("error", err.Error()))
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

// Typing
//
//	@Summary	Send user typing event in the channel
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id	path		int64	true	"Channel id"
//	@Success	200			{string}	string	"typing status sent"
//	@failure	400			{string}	string	"Bad request"
//	@failure	500			{string}	string	"Internal server error"
//	@Router		/message/channel/{channel_id}/typing [post]
func (e *entity) Typing(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	channelIdStr := c.Params("channel_id")
	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	// Validate channel and permissions
	channel, _, err := e.validateSendPermissions(c, channelId, user.Id)
	if err != nil {
		return err
	}
	err = e.mqt.SendChannelMessage(channelId, &mqmsg.ChannelUserTyping{
		ChannelId: channel.Id,
		UserId:    user.Id,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendTypingEvent)
	}
	return c.SendStatus(fiber.StatusOK)
}
