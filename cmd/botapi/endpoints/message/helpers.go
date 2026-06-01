package message

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/FlameInTheDark/gochat/internal/botauth"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/embed"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	reactionutil "github.com/FlameInTheDark/gochat/internal/reaction"
	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
)

func (e *Entity) requireChannel(c *fiber.Ctx, perms ...permissions.RolePermission) (*botauth.Principal, *model.Channel, *int64, error) {
	principal, ok := botauth.FromFiber(c)
	if !ok {
		return nil, nil, nil, fiber.NewError(fiber.StatusUnauthorized, "missing bot principal")
	}
	channelID, err := parseParamInt64(c, "channel_id")
	if err != nil {
		return nil, nil, nil, err
	}
	guildChannel, err := e.gc.GetGuildByChannel(c.UserContext(), channelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			channel, _, _, allowed, err := e.perm.ChannelPerm(c.UserContext(), 0, channelID, principal.BotUserID, perms...)
			if err != nil {
				return nil, nil, nil, fiber.NewError(fiber.StatusInternalServerError, "unable to check channel permissions")
			}
			if !allowed || channel == nil {
				return nil, nil, nil, fiber.NewError(fiber.StatusForbidden, "bot cannot access this channel")
			}
			return principal, channel, nil, nil
		}
		return nil, nil, nil, fiber.NewError(fiber.StatusInternalServerError, "unable to resolve channel guild")
	}
	channel, _, _, allowed, err := e.perm.ChannelPerm(c.UserContext(), guildChannel.GuildId, channelID, principal.BotUserID, perms...)
	if err != nil {
		return nil, nil, nil, fiber.NewError(fiber.StatusInternalServerError, "unable to check channel permissions")
	}
	if !allowed || channel == nil {
		return nil, nil, nil, fiber.NewError(fiber.StatusForbidden, "bot cannot access this channel")
	}
	guildID := guildChannel.GuildId
	return principal, channel, &guildID, nil
}

func (e *Entity) messageDTO(c *fiber.Ctx, msg model.Message) (dto.Message, error) {
	author, err := e.publicUser(c, msg.UserId)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "unable to get message author")
	}
	flags := model.NormalizeMessageFlags(msg.Flags)
	embeds, err := embed.ParseMergedEmbeds(msg.EmbedsJSON, msg.AutoEmbedsJSON, model.HasMessageFlag(flags, model.MessageFlagSuppressEmbeds))
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "unable to parse message embeds")
	}
	position := msg.Position
	out := dto.Message{
		Id:          msg.Id,
		ChannelId:   msg.ChannelId,
		Author:      author,
		Content:     msg.Content,
		Position:    &position,
		Embeds:      embeds,
		Flags:       flags,
		Type:        msg.Type,
		UpdatedAt:   msg.EditedAt,
		Attachments: nil,
		Reactions:   e.messageReactions(c, msg.Id, author.Id),
	}
	if msg.Reference != 0 {
		out.Reference = &msg.Reference
	}
	if msg.ReferenceChannel != 0 {
		out.ReferenceChannelId = &msg.ReferenceChannel
	}
	if msg.Thread != 0 {
		out.ThreadId = &msg.Thread
	}
	if msg.InteractionID != nil && msg.InteractionApplicationID != nil && msg.InteractionCommandID != nil && msg.InteractionCommandName != nil && msg.InteractionUserID != nil {
		out.Interaction = &dto.MessageInteraction{
			Id:            *msg.InteractionID,
			ApplicationId: *msg.InteractionApplicationID,
			CommandId:     *msg.InteractionCommandID,
			CommandName:   *msg.InteractionCommandName,
			UserId:        *msg.InteractionUserID,
		}
	}
	return out, nil
}

func (e *Entity) publicUser(c *fiber.Ctx, userID int64) (dto.User, error) {
	u, err := e.user.GetUserById(c.UserContext(), userID)
	if err != nil {
		return dto.User{}, err
	}
	var disc string
	if d, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), userID); err == nil {
		disc = d.Discriminator
	}
	return dto.User{
		Id:            u.Id,
		Name:          u.Name,
		Discriminator: disc,
		Bio:           u.Bio,
		BannerColor:   u.BannerColor,
		PanelColor:    u.PanelColor,
		IsBot:         u.IsBot(),
	}, nil
}

func (e *Entity) messageReactions(c *fiber.Ctx, messageID, viewerID int64) []dto.MessageReaction {
	summaries, err := e.react.ListMessageSummaries(c.UserContext(), messageID)
	if err != nil {
		return nil
	}
	out := make([]dto.MessageReaction, 0, len(summaries))
	for _, summary := range summaries {
		if summary.Count <= 0 {
			continue
		}
		me := false
		if _, err := e.react.GetUserReaction(c.UserContext(), messageID, viewerID, summary.BucketKey); err == nil {
			me = true
		}
		out = append(out, reactionSummaryToDTO(summary, me))
	}
	return out
}

func (e *Entity) reactionSummary(c *fiber.Ctx, messageID int64, bucketKey string, me bool) dto.MessageReaction {
	summaries, err := e.react.ListMessageSummaries(c.UserContext(), messageID)
	if err != nil {
		return dto.MessageReaction{Me: me, Emoji: dto.MessageReactionEmoji{Name: bucketKey}}
	}
	for _, summary := range summaries {
		if summary.BucketKey == bucketKey {
			return reactionSummaryToDTO(summary, me)
		}
	}
	custom, emojiID, emojiName := reactionutil.BucketKeyEmoji(bucketKey)
	var id *int64
	if custom && emojiID != 0 {
		id = &emojiID
	}
	return dto.MessageReaction{Me: me, Emoji: dto.MessageReactionEmoji{Id: id, Name: emojiName}}
}

func reactionSummaryToDTO(summary model.ReactionSummary, me bool) dto.MessageReaction {
	var emojiID *int64
	if summary.Custom && summary.EmojiId != 0 {
		id := summary.EmojiId
		emojiID = &id
	}
	return dto.MessageReaction{
		Count: summary.Count,
		Me:    me,
		Emoji: dto.MessageReactionEmoji{Id: emojiID, Name: summary.EmojiName},
	}
}

func (e *Entity) repairLastMessage(c *fiber.Ctx, channelID int64, guildID *int64, deletedID int64) {
	nextLast := int64(0)
	if messages, _, err := e.msg.GetMessagesBefore(c.UserContext(), channelID, deletedID-1, 1); err == nil && len(messages) > 0 {
		nextLast = messages[0].Id
	}
	_ = e.ch.SetLastMessage(c.UserContext(), channelID, nextLast)
	if guildID == nil {
		return
	}
	if nextLast == 0 {
		_ = e.gclm.ClearChannelLastMessage(c.UserContext(), *guildID, channelID)
		return
	}
	_ = e.gclm.SetChannelLastMessage(c.UserContext(), *guildID, channelID, nextLast)
}

func parseParamInt64(c *fiber.Ctx, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Params(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return id, nil
}

func messageLookupError(err error) error {
	if errors.Is(err, gocql.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "message not found")
	}
	return fiber.NewError(fiber.StatusInternalServerError, "unable to get message")
}

func stringValue(value *string) string {
	if value == nil {
		return "[]"
	}
	return *value
}
