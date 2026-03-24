package message

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	reactionutil "github.com/FlameInTheDark/gochat/internal/reaction"
)

// AddReaction
//
//	@Summary	Add message reaction
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id		path		int64	true	"Channel id"
//	@Param		message_id		path		int64	true	"Message id"
//	@Param		reaction_name	path		string	true	"Reaction name"
//	@Success	200				{string}	string	"OK"
//	@failure	400				{string}	string	"Bad request"
//	@failure	403				{string}	string	"Forbidden"
//	@failure	404				{string}	string	"Not found"
//	@failure	500				{string}	string	"Internal server error"
//	@Router		/message/channel/{channel_id}/{message_id}/reactions/{reaction_name} [put]
func (e *entity) AddReaction(c *fiber.Ctx) error {
	user, channelId, messageId, parsed, err := e.parseReactionRoute(c)
	if err != nil {
		return err
	}

	_, guildId, err := e.validateReactionChannel(c, channelId, user.Id, true)
	if err != nil {
		return err
	}

	if _, err := e.msg.GetMessage(c.UserContext(), messageId, channelId); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "message not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetMessage)
	}

	resolved, err := e.resolveReactionForAdd(c.UserContext(), user.Id, parsed)
	if err != nil {
		return err
	}

	created, summary, _, err := e.addReactionHot(c.UserContext(), messageId, user.Id, resolved)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to add reaction")
	}
	if created {
		go e.sendReactionUpdateEvent(observability.BackgroundFromContext(c.UserContext()), channelId, guildId, messageId, summary, true)
		// Stale cached DTO would show wrong reaction counts; evict so next window
		// fetch rebuilds with fresh data.
		go e.evictMessageDTOFromCache(observability.BackgroundFromContext(c.UserContext()), channelId, messageId)
	}

	return c.SendStatus(fiber.StatusOK)
}

// RemoveReaction
//
//	@Summary	Remove message reaction
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id		path		int64	true	"Channel id"
//	@Param		message_id		path		int64	true	"Message id"
//	@Param		reaction_name	path		string	true	"Reaction name"
//	@Success	200				{string}	string	"OK"
//	@failure	400				{string}	string	"Bad request"
//	@failure	403				{string}	string	"Forbidden"
//	@failure	404				{string}	string	"Not found"
//	@failure	500				{string}	string	"Internal server error"
//	@Router		/message/channel/{channel_id}/{message_id}/reactions/{reaction_name} [delete]
func (e *entity) RemoveReaction(c *fiber.Ctx) error {
	user, channelId, messageId, parsed, err := e.parseReactionRoute(c)
	if err != nil {
		return err
	}

	_, guildId, err := e.validateReactionChannel(c, channelId, user.Id, false)
	if err != nil {
		return err
	}

	if _, err := e.msg.GetMessage(c.UserContext(), messageId, channelId); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "message not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetMessage)
	}

	removed, summary, err := e.removeReactionHot(c.UserContext(), messageId, user.Id, parsed.BucketKey)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to remove reaction")
	}
	if removed {
		go e.sendReactionUpdateEvent(observability.BackgroundFromContext(c.UserContext()), channelId, guildId, messageId, summary, false)
		go e.evictMessageDTOFromCache(observability.BackgroundFromContext(c.UserContext()), channelId, messageId)
	}

	return c.SendStatus(fiber.StatusOK)
}

// GetReactionUsers
//
//	@Summary	List users who reacted with a specific reaction
//	@Produce	json
//	@Tags		Message
//	@Param		channel_id		path		int64	true	"Channel id"
//	@Param		message_id		path		int64	true	"Message id"
//	@Param		reaction_name	path		string	true	"Reaction name"
//	@Param		after			query		int64	false	"Reaction ID cursor"
//	@Param		limit			query		int		false	"Page size"
//	@Success	200				{object}	dto.MessageReactionUsersPage
//	@failure	400				{string}	string	"Bad request"
//	@failure	403				{string}	string	"Forbidden"
//	@failure	404				{string}	string	"Not found"
//	@failure	500				{string}	string	"Internal server error"
//	@Router		/message/channel/{channel_id}/{message_id}/reactions/{reaction_name} [get]
func (e *entity) GetReactionUsers(c *fiber.Ctx) error {
	user, channelId, messageId, parsed, err := e.parseReactionRoute(c)
	if err != nil {
		return err
	}

	req, err := e.parseGetReactionUsersRequest(c)
	if err != nil {
		return err
	}

	_, _, err = e.validateReactionChannel(c, channelId, user.Id, false)
	if err != nil {
		return err
	}

	if _, err := e.msg.GetMessage(c.UserContext(), messageId, channelId); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "message not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetMessage)
	}

	limit := reactionutil.ParseLimit(*req.Limit)
	rows, nextAfter, err := e.listReactionUsersCacheFirst(c.UserContext(), messageId, parsed.BucketKey, req.After, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get reaction users")
	}

	var items []dto.User
	items, err = e.buildReactionUsers(c.UserContext(), rows)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to build reaction users")
	}

	return c.JSON(dto.MessageReactionUsersPage{
		Items:     items,
		NextAfter: nextAfter,
	})
}

func (e *entity) parseReactionRoute(c *fiber.Ctx) (*helper.JWTUser, int64, int64, reactionutil.ParsedName, error) {
	channelIdStr := c.Params("channel_id")
	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		return nil, 0, 0, reactionutil.ParsedName{}, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}

	messageIdStr := c.Params("message_id")
	messageId, err := strconv.ParseInt(messageIdStr, 10, 64)
	if err != nil {
		return nil, 0, 0, reactionutil.ParsedName{}, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectMessageID)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return nil, 0, 0, reactionutil.ParsedName{}, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	parsed, err := reactionutil.ParseReactionName(c.Params("reaction_name"))
	if err != nil {
		return nil, 0, 0, reactionutil.ParsedName{}, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectReactionName)
	}

	return user, channelId, messageId, parsed, nil
}

func (e *entity) parseGetReactionUsersRequest(c *fiber.Ctx) (*GetReactionUsersRequest, error) {
	var req GetReactionUsersRequest
	if err := c.QueryParser(&req); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := req.Validate(); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if req.Limit == nil {
		limit := reactionutil.UsersPageDefaultLimit
		req.Limit = &limit
	}
	return &req, nil
}

func (e *entity) validateReactionChannel(c *fiber.Ctx, channelId, userId int64, requireAdd bool) (*model.Channel, *int64, error) {
	channel, err := e.ch.GetChannel(c.UserContext(), channelId)
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusNotFound, "channel not found")
	}
	if channel.Type == model.ChannelTypeGuildCategory || channel.Type == model.ChannelTypeGuildVoice {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToReadFromThisChannel)
	}
	if requireAdd && channel.Type == model.ChannelTypeThread && channel.Closed {
		return nil, nil, fiber.NewError(fiber.StatusForbidden, ErrThreadClosed)
	}

	if channel.Type == model.ChannelTypeGuild || channel.Type == model.ChannelTypeThread {
		guildChannel, err := e.gc.GetGuildByChannel(c.UserContext(), channelId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil, fiber.NewError(fiber.StatusNotFound, "channel not found")
			}
			return nil, nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuild)
		}

		perms := []permissions.RolePermission{
			permissions.PermServerViewChannels,
			permissions.PermTextReadMessageHistory,
		}
		if requireAdd {
			perms = append(perms, permissions.PermTextAddReactions)
		}
		_, _, _, ok, err := e.perm.ChannelPerm(c.UserContext(), guildChannel.GuildId, channelId, userId, perms...)
		if err != nil {
			return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
		}
		if !ok {
			return nil, nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
		}
		return &channel, &guildChannel.GuildId, nil
	}

	_, _, _, ok, err := e.perm.ChannelPerm(c.UserContext(), 0, channelId, userId)
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
	}
	if !ok {
		return nil, nil, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}
	return &channel, nil, nil
}

func (e *entity) resolveReactionForAdd(ctx context.Context, userId int64, parsed reactionutil.ParsedName) (reactionutil.ParsedName, error) {
	if !parsed.Custom {
		return parsed, nil
	}

	lookup, err := e.lookupEmojiCached(ctx, parsed.EmojiId)
	if err != nil {
		return parsed, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetMessage)
	}
	if lookup == nil || !lookup.Done {
		return parsed, fiber.NewError(fiber.StatusNotFound, "emoji not found")
	}

	ok, err := e.m.IsGuildMember(ctx, lookup.GuildId, userId)
	if err != nil {
		return parsed, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuild)
	}
	if !ok {
		return parsed, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	parsed.EmojiId = lookup.Id
	parsed.EmojiName = lookup.Name
	parsed.BucketKey = reactionutil.CustomBucketKey(lookup.Id)
	return parsed, nil
}

func (e *entity) addReactionHot(ctx context.Context, messageId, userId int64, parsed reactionutil.ParsedName) (bool, model.ReactionSummary, model.Reaction, error) {
	existing, warm, err := e.loadUserReactionCache(ctx, messageId, userId)
	if err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}
	if !warm {
		rows, err := e.react.GetUserReactions(ctx, messageId, userId)
		if err != nil {
			return false, model.ReactionSummary{}, model.Reaction{}, err
		}
		if err := e.backfillUserReactionsCache(ctx, messageId, userId, rows); err != nil {
			return false, model.ReactionSummary{}, model.Reaction{}, err
		}
		existing, _, err = e.loadUserReactionCache(ctx, messageId, userId)
		if err != nil {
			return false, model.ReactionSummary{}, model.Reaction{}, err
		}
	}

	if current, ok := existing[parsed.BucketKey]; ok {
		summary, err := e.loadBucketSummary(ctx, messageId, parsed.BucketKey)
		if err != nil {
			return false, model.ReactionSummary{}, model.Reaction{}, err
		}
		if summary.EmojiName == "" {
			summary = model.ReactionSummary{
				MessageId: messageId,
				BucketKey: current.BucketKey,
				Custom:    current.Custom,
				EmojiId:   current.EmojiId,
				EmojiName: current.EmojiName,
				Count:     summary.Count,
			}
		}
		return false, summary, current, nil
	}

	reaction := model.Reaction{
		MessageId:  messageId,
		ReactionId: idgen.Next(),
		UserId:     userId,
		BucketKey:  parsed.BucketKey,
		Custom:     parsed.Custom,
		EmojiId:    parsed.EmojiId,
		EmojiName:  parsed.EmojiName,
	}
	summary := model.ReactionSummary{
		MessageId: messageId,
		BucketKey: parsed.BucketKey,
		Custom:    parsed.Custom,
		EmojiId:   parsed.EmojiId,
		EmojiName: parsed.EmojiName,
		Count:     1,
	}
	previousSummary, err := e.loadBucketSummary(ctx, messageId, parsed.BucketKey)
	if err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}

	if e.cache == nil {
		if err := e.react.UpsertReaction(ctx, reaction); err != nil {
			return false, model.ReactionSummary{}, model.Reaction{}, err
		}
		summary.Count = previousSummary.Count + 1
		if err := e.react.SetSummary(ctx, summary); err != nil {
			return false, model.ReactionSummary{}, model.Reaction{}, err
		}
		return true, summary, reaction, nil
	}

	userKey := reactionutil.UserKey(messageId, userId)
	if err := e.cache.HSet(ctx, userKey, reactionutil.WarmMarkerField, "1"); err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}
	if err := e.cache.HSet(ctx, userKey, parsed.BucketKey, strconv.FormatInt(reaction.ReactionId, 10)); err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}
	if err := e.cache.HSet(ctx, reactionutil.SummaryMetaKey(messageId), reactionutil.WarmMarkerField, "1"); err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}

	metaJSON, err := json.Marshal(reactionutil.CacheSummaryMeta{
		Custom:    summary.Custom,
		EmojiId:   summary.EmojiId,
		EmojiName: summary.EmojiName,
	})
	if err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}
	if err := e.cache.HSet(ctx, reactionutil.SummaryMetaKey(messageId), parsed.BucketKey, string(metaJSON)); err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}
	count, err := e.cache.HIncrBy(ctx, reactionutil.SummaryCountKey(messageId), parsed.BucketKey, 1)
	if err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}
	summary.Count = int(count)

	if err := e.cache.ZAdd(ctx, reactionutil.BucketSortedSetKey(messageId, parsed.BucketKey), float64(reaction.ReactionId), strconv.FormatInt(reaction.ReactionId, 10)); err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}
	rowJSON, err := reactionutil.EncodeReactionRow(reaction)
	if err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}
	if err := e.cache.HSet(ctx, reactionutil.BucketRowsKey(messageId, parsed.BucketKey), strconv.FormatInt(reaction.ReactionId, 10), rowJSON); err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}
	if err := e.enqueueReactionFlush(ctx, reactionutil.FlushEvent{
		Op:        "add",
		MessageId: messageId,
		Reaction:  reaction,
		Count:     summary.Count,
	}); err != nil {
		return false, model.ReactionSummary{}, model.Reaction{}, err
	}

	return true, summary, reaction, nil
}

func (e *entity) removeReactionHot(ctx context.Context, messageId, userId int64, bucketKey string) (bool, model.ReactionSummary, error) {
	existing, warm, err := e.loadUserReactionCache(ctx, messageId, userId)
	if err != nil {
		return false, model.ReactionSummary{}, err
	}
	if !warm {
		rows, err := e.react.GetUserReactions(ctx, messageId, userId)
		if err != nil {
			return false, model.ReactionSummary{}, err
		}
		if err := e.backfillUserReactionsCache(ctx, messageId, userId, rows); err != nil {
			return false, model.ReactionSummary{}, err
		}
		existing, _, err = e.loadUserReactionCache(ctx, messageId, userId)
		if err != nil {
			return false, model.ReactionSummary{}, err
		}
	}

	reaction, ok := existing[bucketKey]
	if !ok {
		return false, model.ReactionSummary{}, nil
	}

	summary, err := e.loadBucketSummary(ctx, messageId, bucketKey)
	if err != nil {
		return false, model.ReactionSummary{}, err
	}
	if summary.EmojiName == "" {
		summary = model.ReactionSummary{
			MessageId: messageId,
			BucketKey: reaction.BucketKey,
			Custom:    reaction.Custom,
			EmojiId:   reaction.EmojiId,
			EmojiName: reaction.EmojiName,
			Count:     summary.Count,
		}
	}

	if e.cache == nil {
		if err := e.react.DeleteReaction(ctx, reaction); err != nil {
			return false, model.ReactionSummary{}, err
		}
		if summary.Count > 0 {
			summary.Count--
		}
		if err := e.react.SetSummary(ctx, summary); err != nil {
			return false, model.ReactionSummary{}, err
		}
		return true, summary, nil
	}

	userKey := reactionutil.UserKey(messageId, userId)
	if err := e.cache.HSet(ctx, userKey, reactionutil.WarmMarkerField, "1"); err != nil {
		return false, model.ReactionSummary{}, err
	}
	if err := e.cache.HDel(ctx, userKey, bucketKey); err != nil {
		return false, model.ReactionSummary{}, err
	}
	if err := e.cache.ZRem(ctx, reactionutil.BucketSortedSetKey(messageId, bucketKey), strconv.FormatInt(reaction.ReactionId, 10)); err != nil {
		return false, model.ReactionSummary{}, err
	}
	if err := e.cache.HDel(ctx, reactionutil.BucketRowsKey(messageId, bucketKey), strconv.FormatInt(reaction.ReactionId, 10)); err != nil {
		return false, model.ReactionSummary{}, err
	}
	count, err := e.cache.HIncrBy(ctx, reactionutil.SummaryCountKey(messageId), bucketKey, -1)
	if err != nil {
		return false, model.ReactionSummary{}, err
	}
	if count < 0 {
		count = 0
		if err := e.cache.HSet(ctx, reactionutil.SummaryCountKey(messageId), bucketKey, "0"); err != nil {
			return false, model.ReactionSummary{}, err
		}
	}
	summary.Count = int(count)
	if summary.Count <= 0 {
		if err := e.cache.HDel(ctx, reactionutil.SummaryCountKey(messageId), bucketKey); err != nil {
			return false, model.ReactionSummary{}, err
		}
		if err := e.cache.HDel(ctx, reactionutil.SummaryMetaKey(messageId), bucketKey); err != nil {
			return false, model.ReactionSummary{}, err
		}
	}
	if err := e.enqueueReactionFlush(ctx, reactionutil.FlushEvent{
		Op:        "remove",
		MessageId: messageId,
		Reaction:  reaction,
		Count:     summary.Count,
	}); err != nil {
		return false, model.ReactionSummary{}, err
	}

	return true, summary, nil
}

func (e *entity) enqueueReactionFlush(ctx context.Context, event reactionutil.FlushEvent) error {
	if e.cache == nil {
		return nil
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return e.cache.XAdd(ctx, reactionutil.FlushStreamKey, reactionutil.FlushStreamApproxMaxLen, true, map[string]interface{}{
		"event": string(payload),
		"ts":    time.Now().UTC().UnixMilli(),
	})
}

func (e *entity) loadUserReactionCache(ctx context.Context, messageId, userId int64) (map[string]model.Reaction, bool, error) {
	if e.cache == nil {
		return map[string]model.Reaction{}, false, nil
	}
	fields, err := e.cache.HGetAll(ctx, reactionutil.UserKey(messageId, userId))
	if err != nil {
		return nil, false, err
	}
	rows := make(map[string]model.Reaction)
	warm := false
	for bucketKey, reactionIdStr := range fields {
		if bucketKey == reactionutil.WarmMarkerField {
			warm = reactionIdStr == "1"
			continue
		}
		reactionId, err := strconv.ParseInt(reactionIdStr, 10, 64)
		if err != nil {
			continue
		}
		custom, emojiId, emojiName := reactionutil.BucketKeyEmoji(bucketKey)
		rows[bucketKey] = model.Reaction{
			MessageId:  messageId,
			ReactionId: reactionId,
			UserId:     userId,
			BucketKey:  bucketKey,
			Custom:     custom,
			EmojiId:    emojiId,
			EmojiName:  emojiName,
		}
	}
	return rows, warm, nil
}

func (e *entity) backfillUserReactionsCache(ctx context.Context, messageId, userId int64, rows []model.Reaction) error {
	if e.cache == nil {
		return nil
	}
	userKey := reactionutil.UserKey(messageId, userId)
	if err := e.cache.HSet(ctx, userKey, reactionutil.WarmMarkerField, "1"); err != nil {
		return err
	}
	for _, row := range rows {
		if err := e.cache.HSet(ctx, userKey, row.BucketKey, strconv.FormatInt(row.ReactionId, 10)); err != nil {
			return err
		}
	}
	return nil
}

func (e *entity) loadMessageSummariesCache(ctx context.Context, messageId int64) (map[string]model.ReactionSummary, bool, error) {
	if e.cache == nil {
		return map[string]model.ReactionSummary{}, false, nil
	}
	counts, err := e.cache.HGetAll(ctx, reactionutil.SummaryCountKey(messageId))
	if err != nil {
		return nil, false, err
	}
	metaFields, err := e.cache.HGetAll(ctx, reactionutil.SummaryMetaKey(messageId))
	if err != nil {
		return nil, false, err
	}

	summaries := make(map[string]model.ReactionSummary)
	warm := metaFields[reactionutil.WarmMarkerField] == "1"
	for bucketKey, countStr := range counts {
		if bucketKey == reactionutil.WarmMarkerField {
			continue
		}
		count, err := strconv.Atoi(countStr)
		if err != nil || count <= 0 {
			continue
		}
		summary := model.ReactionSummary{
			MessageId: messageId,
			BucketKey: bucketKey,
			Count:     count,
		}
		if metaJSON, ok := metaFields[bucketKey]; ok && metaJSON != "" {
			var meta reactionutil.CacheSummaryMeta
			if err := json.Unmarshal([]byte(metaJSON), &meta); err == nil {
				summary.Custom = meta.Custom
				summary.EmojiId = meta.EmojiId
				summary.EmojiName = meta.EmojiName
			}
		}
		if summary.EmojiName == "" {
			summary.Custom, summary.EmojiId, summary.EmojiName = reactionutil.BucketKeyEmoji(bucketKey)
		}
		summaries[bucketKey] = summary
	}
	return summaries, warm, nil
}

func (e *entity) backfillMessageSummariesCache(ctx context.Context, messageId int64, rows []model.ReactionSummary) error {
	if e.cache == nil {
		return nil
	}
	if err := e.cache.HSet(ctx, reactionutil.SummaryMetaKey(messageId), reactionutil.WarmMarkerField, "1"); err != nil {
		return err
	}
	for _, row := range rows {
		if err := e.cache.HSet(ctx, reactionutil.SummaryCountKey(messageId), row.BucketKey, strconv.Itoa(row.Count)); err != nil {
			return err
		}
		metaJSON, err := json.Marshal(reactionutil.CacheSummaryMeta{
			Custom:    row.Custom,
			EmojiId:   row.EmojiId,
			EmojiName: row.EmojiName,
		})
		if err != nil {
			return err
		}
		if err := e.cache.HSet(ctx, reactionutil.SummaryMetaKey(messageId), row.BucketKey, string(metaJSON)); err != nil {
			return err
		}
	}
	return nil
}

func (e *entity) loadBucketSummary(ctx context.Context, messageId int64, bucketKey string) (model.ReactionSummary, error) {
	summaries, warm, err := e.loadMessageSummariesCache(ctx, messageId)
	if err != nil {
		return model.ReactionSummary{}, err
	}
	if !warm {
		rows, err := e.react.ListMessageSummaries(ctx, messageId)
		if err != nil {
			return model.ReactionSummary{}, err
		}
		if err := e.backfillMessageSummariesCache(ctx, messageId, rows); err != nil {
			return model.ReactionSummary{}, err
		}
		summaries = make(map[string]model.ReactionSummary, len(rows))
		for _, row := range rows {
			summaries[row.BucketKey] = row
		}
	}
	if summary, ok := summaries[bucketKey]; ok {
		return summary, nil
	}
	custom, emojiId, emojiName := reactionutil.BucketKeyEmoji(bucketKey)
	return model.ReactionSummary{
		MessageId: messageId,
		BucketKey: bucketKey,
		Custom:    custom,
		EmojiId:   emojiId,
		EmojiName: emojiName,
		Count:     0,
	}, nil
}

func (e *entity) listReactionUsersCacheFirst(ctx context.Context, messageId int64, bucketKey string, after *int64, limit int) ([]model.Reaction, *int64, error) {
	summary, err := e.loadBucketSummary(ctx, messageId, bucketKey)
	if err != nil {
		return nil, nil, err
	}
	if summary.Count <= 0 {
		return []model.Reaction{}, nil, nil
	}

	rows, hot, err := e.listBucketReactionsHot(ctx, messageId, bucketKey, after, limit+1)
	if err != nil {
		return nil, nil, err
	}
	if !hot {
		rows, err = e.react.ListBucketReactions(ctx, messageId, bucketKey, after, limit+1)
		if err != nil {
			return nil, nil, err
		}
		if err := e.backfillBucketReactionRows(ctx, rows); err != nil {
			return nil, nil, err
		}
	}

	var nextAfter *int64
	if len(rows) > limit {
		cursor := rows[limit-1].ReactionId
		nextAfter = &cursor
		rows = rows[:limit]
	}
	return rows, nextAfter, nil
}

func (e *entity) listBucketReactionsHot(ctx context.Context, messageId int64, bucketKey string, after *int64, limit int) ([]model.Reaction, bool, error) {
	if e.cache == nil {
		return nil, false, nil
	}
	summaries, warm, err := e.loadMessageSummariesCache(ctx, messageId)
	if err != nil {
		return nil, false, err
	}
	if !warm {
		return nil, false, nil
	}
	summary, ok := summaries[bucketKey]
	if !ok || summary.Count <= 0 {
		return []model.Reaction{}, true, nil
	}

	max := "+inf"
	if after != nil {
		max = fmt.Sprintf("(%d", *after)
	}
	ids, err := e.cache.ZRevRangeByScore(ctx, reactionutil.BucketSortedSetKey(messageId, bucketKey), max, "-inf", 0, int64(limit))
	if err != nil {
		return nil, false, err
	}
	if len(ids) == 0 {
		return nil, false, nil
	}

	rows := make([]model.Reaction, 0, len(ids))
	for _, idStr := range ids {
		rowJSON, err := e.cache.HGet(ctx, reactionutil.BucketRowsKey(messageId, bucketKey), idStr)
		if err != nil || rowJSON == "" {
			return nil, false, nil
		}
		var row model.Reaction
		if err := json.Unmarshal([]byte(rowJSON), &row); err != nil {
			return nil, false, nil
		}
		rows = append(rows, row)
	}
	return rows, true, nil
}

func (e *entity) backfillBucketReactionRows(ctx context.Context, rows []model.Reaction) error {
	if e.cache == nil {
		return nil
	}
	for _, row := range rows {
		if err := e.cache.HSet(ctx, reactionutil.UserKey(row.MessageId, row.UserId), reactionutil.WarmMarkerField, "1"); err != nil {
			return err
		}
		if err := e.cache.HSet(ctx, reactionutil.UserKey(row.MessageId, row.UserId), row.BucketKey, strconv.FormatInt(row.ReactionId, 10)); err != nil {
			return err
		}
		if err := e.cache.ZAdd(ctx, reactionutil.BucketSortedSetKey(row.MessageId, row.BucketKey), float64(row.ReactionId), strconv.FormatInt(row.ReactionId, 10)); err != nil {
			return err
		}
		rowJSON, err := reactionutil.EncodeReactionRow(row)
		if err != nil {
			return err
		}
		if err := e.cache.HSet(ctx, reactionutil.BucketRowsKey(row.MessageId, row.BucketKey), strconv.FormatInt(row.ReactionId, 10), rowJSON); err != nil {
			return err
		}
	}
	return nil
}

func (e *entity) loadMessageReactions(ctx context.Context, messageId, userId int64) ([]dto.MessageReaction, error) {
	summaries, warm, err := e.loadMessageSummariesCache(ctx, messageId)
	if err != nil {
		return nil, err
	}
	if !warm {
		rows, err := e.react.ListMessageSummaries(ctx, messageId)
		if err != nil {
			return nil, err
		}
		if err := e.backfillMessageSummariesCache(ctx, messageId, rows); err != nil {
			return nil, err
		}
		summaries = make(map[string]model.ReactionSummary, len(rows))
		for _, row := range rows {
			summaries[row.BucketKey] = row
		}
	}

	userReactions, warm, err := e.loadUserReactionCache(ctx, messageId, userId)
	if err != nil {
		return nil, err
	}
	if !warm && len(summaries) > 0 {
		rows, err := e.react.GetUserReactions(ctx, messageId, userId)
		if err != nil {
			return nil, err
		}
		if err := e.backfillUserReactionsCache(ctx, messageId, userId, rows); err != nil {
			return nil, err
		}
		userReactions = make(map[string]model.Reaction, len(rows))
		for _, row := range rows {
			userReactions[row.BucketKey] = row
		}
	}

	result := make([]dto.MessageReaction, 0, len(summaries))
	for _, summary := range summaries {
		if summary.Count <= 0 {
			continue
		}
		_, me := userReactions[summary.BucketKey]
		result = append(result, e.reactionSummaryToDTO(ctx, summary, me))
	}
	return result, nil
}

func (e *entity) reactionSummaryToDTO(ctx context.Context, summary model.ReactionSummary, me bool) dto.MessageReaction {
	name := summary.EmojiName
	var emojiID *int64
	if summary.Custom && summary.EmojiId != 0 {
		currentName := name
		if lookup, err := e.lookupEmojiCached(ctx, summary.EmojiId); err == nil && lookup != nil && lookup.Name != "" {
			currentName = lookup.Name
		}
		name = currentName
		id := summary.EmojiId
		emojiID = &id
	}
	return dto.MessageReaction{
		Count: summary.Count,
		Me:    me,
		Emoji: dto.MessageReactionEmoji{
			Id:   emojiID,
			Name: name,
		},
	}
}

func (e *entity) sendReactionUpdateEvent(ctx context.Context, channelId int64, guildId *int64, messageId int64, summary model.ReactionSummary, added bool) {
	if e.mqt == nil {
		return
	}
	reaction := e.reactionSummaryToDTO(ctx, summary, false)
	if added {
		_ = mq.SendChannelMessage(ctx, e.mqt, channelId, &mqmsg.MessageReactionAdd{
			GuildId:   guildId,
			ChannelId: channelId,
			MessageId: messageId,
			Reaction:  reaction,
		})
		return
	}
	_ = mq.SendChannelMessage(ctx, e.mqt, channelId, &mqmsg.MessageReactionRemove{
		GuildId:   guildId,
		ChannelId: channelId,
		MessageId: messageId,
		Reaction:  reaction,
	})
}

func (e *entity) buildReactionUsers(ctx context.Context, reactions []model.Reaction) ([]dto.User, error) {
	if len(reactions) == 0 {
		return []dto.User{}, nil
	}
	userIDs := reactionUserIDs(reactions)
	users, err := e.user.GetUsersList(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	discriminators, err := e.disc.GetDiscriminatorsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	usersByID := make(map[int64]model.User, len(users))
	for _, user := range users {
		usersByID[user.Id] = user
	}
	discriminatorsByID := make(map[int64]string, len(discriminators))
	for _, discriminator := range discriminators {
		discriminatorsByID[discriminator.UserId] = discriminator.Discriminator
	}

	result := make([]dto.User, 0, len(reactions))
	for _, reaction := range reactions {
		user, ok := usersByID[reaction.UserId]
		if !ok {
			result = append(result, dto.User{
				Id:            reaction.UserId,
				Name:          "Unknown User",
				Discriminator: discriminatorsByID[reaction.UserId],
			})
			continue
		}
		userDTO := publicUserDTO(user, user.Name, discriminatorsByID[user.Id])
		if user.Avatar != nil {
			if avatarData, err := e.getAvatarDataCached(ctx, user.Id, *user.Avatar); err == nil && avatarData != nil {
				userDTO.Avatar = avatarData
			}
		}
		result = append(result, userDTO)
	}
	return result, nil
}

func (e *entity) buildReactionMembers(ctx context.Context, guildId int64, reactions []model.Reaction) ([]dto.Member, error) {
	if len(reactions) == 0 {
		return []dto.Member{}, nil
	}
	userIDs := reactionUserIDs(reactions)
	users, err := e.user.GetUsersList(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	discriminators, err := e.disc.GetDiscriminatorsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	members, err := e.m.GetMembersList(ctx, guildId, userIDs)
	if err != nil {
		return nil, err
	}
	roles, err := e.ur.GetUsersRolesByGuild(ctx, guildId, userIDs)
	if err != nil {
		return nil, err
	}

	usersByID := make(map[int64]model.User, len(users))
	for _, user := range users {
		usersByID[user.Id] = user
	}
	discriminatorsByID := make(map[int64]string, len(discriminators))
	for _, discriminator := range discriminators {
		discriminatorsByID[discriminator.UserId] = discriminator.Discriminator
	}
	membersByID := make(map[int64]model.Member, len(members))
	for _, member := range members {
		membersByID[member.UserId] = member
	}
	rolesByID := make(map[int64][]int64, len(roles))
	for _, roleSet := range roles {
		rolesByID[roleSet.UserId] = append([]int64(nil), roleSet.Roles...)
	}

	result := make([]dto.Member, 0, len(reactions))
	for _, reaction := range reactions {
		user, ok := usersByID[reaction.UserId]
		if !ok {
			result = append(result, dto.Member{
				User: dto.User{
					Id:            reaction.UserId,
					Name:          "Unknown User",
					Discriminator: discriminatorsByID[reaction.UserId],
				},
				Roles: rolesByID[reaction.UserId],
			})
			continue
		}

		userDTO := publicUserDTO(user, user.Name, discriminatorsByID[user.Id])
		if user.Avatar != nil {
			if avatarData, err := e.getAvatarDataCached(ctx, user.Id, *user.Avatar); err == nil && avatarData != nil {
				userDTO.Avatar = avatarData
			}
		}

		memberDTO := dto.Member{
			User:  userDTO,
			Roles: rolesByID[reaction.UserId],
		}
		if member, ok := membersByID[reaction.UserId]; ok {
			memberDTO.Username = member.Username
			memberDTO.Avatar = member.Avatar
			memberDTO.JoinAt = member.JoinAt
		}
		result = append(result, memberDTO)
	}
	return result, nil
}

func reactionUserIDs(reactions []model.Reaction) []int64 {
	ids := make([]int64, 0, len(reactions))
	for _, reaction := range reactions {
		ids = append(ids, reaction.UserId)
	}
	return ids
}
