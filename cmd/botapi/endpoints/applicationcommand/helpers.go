package applicationcommand

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
	"github.com/FlameInTheDark/gochat/internal/botauth"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/embed"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

const ephemeralResponseKeyPrefix = "appcmd:ephemeral:"
const autocompleteResponseKeyPrefix = "appcmd:autocomplete:"
const avatarCacheTTLSeconds = 3600

func parseInt64Param(c *fiber.Ctx, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Params(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return id, nil
}

func (e *Entity) requireApplication(c *fiber.Ctx) (*botauth.Principal, int64, error) {
	principal, ok := botauth.FromFiber(c)
	if !ok {
		return nil, 0, fiber.NewError(fiber.StatusUnauthorized, "missing bot principal")
	}
	applicationID, err := parseInt64Param(c, "application_id")
	if err != nil {
		return nil, 0, err
	}
	if applicationID != principal.BotUserID {
		return nil, 0, fiber.NewError(fiber.StatusForbidden, "application id does not match bot token")
	}
	return principal, applicationID, nil
}

func guildIDParam(c *fiber.Ctx) (*int64, error) {
	raw := c.Params("guild_id")
	if raw == "" {
		return nil, nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid guild_id")
	}
	return &id, nil
}

func prepareCommand(applicationID int64, guildID *int64, cmd appcmd.ApplicationCommand, existingID int64) (appcmd.ApplicationCommand, error) {
	cmd = appcmd.NormalizeCommand(cmd)
	cmd.ApplicationID = applicationID
	cmd.GuildID = guildID
	if existingID != 0 {
		cmd.ID = existingID
	} else if cmd.ID == 0 {
		cmd.ID = idgen.Next()
	}
	cmd.Version = idgen.Next()
	if err := appcmd.ValidateCommand(cmd); err != nil {
		return appcmd.ApplicationCommand{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return cmd, nil
}

func (e *Entity) loadInteraction(c *fiber.Ctx, applicationID int64, token string) (appcmd.InteractionRecord, error) {
	record, err := e.appcmd.GetInteractionByToken(c.UserContext(), applicationID, appcmd.HashInteractionToken(token))
	if err != nil {
		return appcmd.InteractionRecord{}, fiber.NewError(fiber.StatusNotFound, "unknown interaction")
	}
	if time.Now().After(record.ExpiresAt) {
		return appcmd.InteractionRecord{}, fiber.NewError(fiber.StatusNotFound, "interaction token expired")
	}
	return record, nil
}

func (e *Entity) publicUser(ctx context.Context, userID int64) (dto.User, error) {
	u, err := e.user.GetUserById(ctx, userID)
	if err != nil {
		return dto.User{}, err
	}
	var disc string
	if d, err := e.disc.GetDiscriminatorByUserId(ctx, userID); err == nil {
		disc = d.Discriminator
	}
	out := dto.User{
		Id:            u.Id,
		Name:          u.Name,
		Discriminator: disc,
		Bio:           u.Bio,
		BannerColor:   u.BannerColor,
		PanelColor:    u.PanelColor,
		IsBot:         u.IsBot(),
	}
	if u.Avatar != nil {
		if avatar, err := e.getAvatarDataCached(ctx, u.Id, *u.Avatar); err == nil && avatar != nil {
			out.Avatar = avatar
		}
	}
	return out, nil
}

func (e *Entity) getAvatarDataCached(ctx context.Context, userID, avatarID int64) (*dto.AvatarData, error) {
	key := fmt.Sprintf("avatars:%d:%d", userID, avatarID)
	var cached dto.AvatarData
	if e.cache != nil {
		if err := e.cache.GetJSON(ctx, key, &cached); err == nil && cached.URL != "" {
			return &cached, nil
		}
	}

	if e.av == nil {
		return nil, nil
	}
	avatar, err := e.av.GetAvatar(ctx, avatarID, userID)
	if err != nil {
		return nil, err
	}
	if !avatar.Done || avatar.URL == nil || *avatar.URL == "" {
		return nil, nil
	}

	resolved := dto.AvatarData{
		Id:          avatar.Id,
		URL:         *avatar.URL,
		ContentType: avatar.ContentType,
		Width:       avatar.Width,
		Height:      avatar.Height,
		Size:        avatar.FileSize,
	}
	if e.cache != nil {
		_ = e.cache.SetTimedJSON(ctx, key, resolved, avatarCacheTTLSeconds)
	}
	return &resolved, nil
}

func (e *Entity) messageDTO(ctx context.Context, msg model.Message) (dto.Message, error) {
	author, err := e.publicUser(ctx, msg.UserId)
	if err != nil {
		return dto.Message{}, err
	}
	flags := model.NormalizeMessageFlags(msg.Flags)
	embeds, err := embed.ParseMergedEmbeds(msg.EmbedsJSON, msg.AutoEmbedsJSON, model.HasMessageFlag(flags, model.MessageFlagSuppressEmbeds))
	if err != nil {
		return dto.Message{}, err
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

func (e *Entity) createPublicInteractionMessage(c *fiber.Ctx, record appcmd.InteractionRecord, data *appcmd.InteractionResponseData) (dto.Message, error) {
	if data == nil {
		data = &appcmd.InteractionResponseData{}
	}
	embedsJSON, err := embed.MarshalEmbeds(nil)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	autoEmbedsJSON, err := embed.MarshalEmbeds(nil)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	position, err := e.ch.ReserveMessagePositions(c.UserContext(), record.ChannelID, 1)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "unable to allocate message position")
	}
	msgID := idgen.Next()
	flags := data.Flags &^ (appcmd.MessageFlagEphemeral | appcmd.MessageFlagLoading)
	payload, err := e.payload.GetInteractionPayload(c.UserContext(), record.ID)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "unable to load interaction payload")
	}
	commandName := ""
	if parsed := appcmd.ParseCommandData(payload.DataJSON); parsed != nil {
		commandName = parsed.Name
	}
	if err := e.msg.CreateMessageWithInteraction(c.UserContext(), msgID, record.ChannelID, record.ApplicationID, data.Content, data.Attachments, embedsJSON, autoEmbedsJSON, flags, model.MessageTypeChat, 0, 0, 0, position, record.ID, record.ApplicationID, record.CommandID, record.InvokerUserID, commandName); err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "unable to create interaction response")
	}
	if err := e.ch.SetLastMessage(c.UserContext(), record.ChannelID, msgID); err != nil {
		_ = e.msg.DeleteMessage(c.UserContext(), msgID, record.ChannelID)
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "unable to update channel last message")
	}
	if record.GuildID != nil {
		_ = e.gclm.SetChannelLastMessage(c.UserContext(), *record.GuildID, record.ChannelID, msgID)
	}
	_ = e.rs.SetReadState(c.UserContext(), record.ApplicationID, record.ChannelID, msgID)
	msg := model.Message{
		Id:                       msgID,
		ChannelId:                record.ChannelID,
		UserId:                   record.ApplicationID,
		Content:                  data.Content,
		Position:                 position,
		Attachments:              data.Attachments,
		EmbedsJSON:               &embedsJSON,
		AutoEmbedsJSON:           &autoEmbedsJSON,
		Flags:                    &flags,
		Type:                     int(model.MessageTypeChat),
		InteractionID:            &record.ID,
		InteractionApplicationID: &record.ApplicationID,
		InteractionCommandID:     &record.CommandID,
		InteractionCommandName:   &commandName,
		InteractionUserID:        &record.InvokerUserID,
	}
	out, err := e.messageDTO(c.UserContext(), msg)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "unable to build response message")
	}
	_ = mq.SendChannelMessage(c.UserContext(), e.mqt, record.ChannelID, &mqmsg.CreateMessage{GuildId: record.GuildID, Message: out})
	return out, nil
}

func (e *Entity) setEphemeralResponse(ctx context.Context, interactionID int64, data *appcmd.InteractionResponseData) error {
	if e.cache == nil {
		return nil
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return e.cache.Client().Set(ctx, ephemeralResponseKeyPrefix+strconv.FormatInt(interactionID, 10), raw, appcmd.InteractionTokenTTL).Err()
}

func (e *Entity) getEphemeralResponse(ctx context.Context, interactionID int64) (*appcmd.InteractionResponseData, bool, error) {
	if e.cache == nil {
		return nil, false, nil
	}
	raw, err := e.cache.Client().Get(ctx, ephemeralResponseKeyPrefix+strconv.FormatInt(interactionID, 10)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var data appcmd.InteractionResponseData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, false, err
	}
	return &data, true, nil
}

func (e *Entity) deleteEphemeralResponse(ctx context.Context, interactionID int64) {
	if e.cache != nil {
		_ = e.cache.Client().Del(ctx, ephemeralResponseKeyPrefix+strconv.FormatInt(interactionID, 10)).Err()
	}
}

func (e *Entity) notifyInteractionStatus(ctx context.Context, record appcmd.InteractionRecord, state string, data *appcmd.InteractionResponseData, message *dto.Message) {
	commandName := ""
	if payload, err := e.payload.GetInteractionPayload(ctx, record.ID); err == nil {
		if parsed := appcmd.ParseCommandData(payload.DataJSON); parsed != nil {
			commandName = parsed.Name
		}
	}
	_ = mq.SendUserUpdate(ctx, e.mqt, record.InvokerUserID, &mqmsg.ApplicationCommandInteractionStatus{
		InteractionID: record.ID,
		ApplicationID: record.ApplicationID,
		CommandID:     record.CommandID,
		CommandName:   commandName,
		ChannelID:     record.ChannelID,
		GuildID:       record.GuildID,
		UserID:        record.InvokerUserID,
		State:         state,
		Response:      data,
		Message:       message,
	})
}

func (e *Entity) setAutocompleteResponse(ctx context.Context, interactionID int64, choices []appcmd.ApplicationCommandChoice) error {
	if e.cache == nil {
		return nil
	}
	raw, err := json.Marshal(choices)
	if err != nil {
		return err
	}
	return e.cache.Client().Set(ctx, autocompleteResponseKeyPrefix+strconv.FormatInt(interactionID, 10), raw, 10*time.Second).Err()
}

func interactionLookupError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fiber.NewError(fiber.StatusNotFound, "interaction response not found")
	}
	if errors.Is(err, gocql.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "interaction response message not found")
	}
	return fiber.NewError(fiber.StatusInternalServerError, "unable to load interaction response")
}

func commandLookupError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fiber.NewError(fiber.StatusNotFound, "application command not found")
	}
	return fiber.NewError(fiber.StatusInternalServerError, "unable to get application command")
}

func discordError(code int, message string) fiber.Map {
	return fiber.Map{"code": code, "message": message}
}

func requireInteractionID(c *fiber.Ctx, record appcmd.InteractionRecord) error {
	interactionID, err := parseInt64Param(c, "interaction_id")
	if err != nil {
		return err
	}
	if interactionID != record.ID {
		return fiber.NewError(fiber.StatusNotFound, "unknown interaction")
	}
	return nil
}

func (e *Entity) getMessageResponse(c *fiber.Ctx, record appcmd.InteractionRecord, messageID int64) (dto.Message, error) {
	msg, err := e.msg.GetMessage(c.UserContext(), messageID, record.ChannelID)
	if err != nil {
		return dto.Message{}, interactionLookupError(err)
	}
	return e.messageDTO(c.UserContext(), msg)
}

func (e *Entity) editMessageResponse(c *fiber.Ctx, record appcmd.InteractionRecord, messageID int64, data *appcmd.InteractionResponseData) (dto.Message, error) {
	msg, err := e.msg.GetMessage(c.UserContext(), messageID, record.ChannelID)
	if err != nil {
		return dto.Message{}, interactionLookupError(err)
	}
	content := msg.Content
	flags := model.NormalizeMessageFlags(msg.Flags)
	if data != nil {
		content = data.Content
		flags = data.Flags &^ (appcmd.MessageFlagEphemeral | appcmd.MessageFlagLoading)
	}
	if err := e.msg.UpdateMessage(c.UserContext(), msg.Id, msg.ChannelId, content, stringValue(msg.EmbedsJSON), stringValue(msg.AutoEmbedsJSON), flags); err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "unable to edit interaction response")
	}
	updated, err := e.msg.GetMessage(c.UserContext(), messageID, record.ChannelID)
	if err != nil {
		return dto.Message{}, interactionLookupError(err)
	}
	out, err := e.messageDTO(c.UserContext(), updated)
	if err != nil {
		return dto.Message{}, fiber.NewError(fiber.StatusInternalServerError, "unable to build response message")
	}
	_ = mq.SendChannelMessage(c.UserContext(), e.mqt, record.ChannelID, &mqmsg.UpdateMessage{GuildId: record.GuildID, Message: out})
	return out, nil
}

func (e *Entity) deleteMessageResponse(c *fiber.Ctx, record appcmd.InteractionRecord, messageID int64) error {
	if err := e.msg.DeleteMessage(c.UserContext(), messageID, record.ChannelID); err != nil {
		return interactionLookupError(err)
	}
	_ = mq.SendChannelMessage(c.UserContext(), e.mqt, record.ChannelID, &mqmsg.DeleteMessage{GuildId: record.GuildID, ChannelId: record.ChannelID, MessageId: messageID})
	return nil
}

func stringValue(value *string) string {
	if value == nil {
		return "[]"
	}
	return *value
}

func duplicateAckError() error {
	return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("%d: interaction has already been acknowledged", 40060))
}
