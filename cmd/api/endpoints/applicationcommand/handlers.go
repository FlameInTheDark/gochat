package applicationcommand

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	appcmddispatch "github.com/FlameInTheDark/gochat/internal/applicationcommanddispatch"
	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/gofiber/fiber/v2"
)

func (e *entity) ListVisibleCommands(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unable to get user")
	}
	channelID, guildID, err := e.channelAndGuild(c)
	if err != nil {
		return err
	}
	if guildID == nil {
		return c.JSON([]appcmd.ApplicationCommand{})
	}
	if _, _, _, ok, err := e.perm.ChannelPerm(c.UserContext(), *guildID, channelID, user.Id, permissions.PermServerViewChannels, permissions.PermUseApplicationCommands); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to check command permissions")
	} else if !ok {
		return c.JSON([]appcmd.ApplicationCommand{})
	}
	commandType := appcmd.CommandType(c.QueryInt("type", int(appcmd.CommandTypeChatInput)))
	if commandType == 0 {
		commandType = appcmd.CommandTypeChatInput
	}
	commands, err := e.appcmd.ListVisibleGuildCommands(c.UserContext(), *guildID, commandType, c.Query("query"), 100)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to list application commands")
	}
	effective, err := e.perm.GetChannelPermissions(c.UserContext(), *guildID, channelID, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to resolve permissions")
	}
	out := commands[:0]
	for _, command := range commands {
		if !commandAllowedInGuild(command, effective) {
			continue
		}
		out = append(out, command)
	}
	return c.JSON(out)
}

func (e *entity) InvokeCommand(c *fiber.Ctx) error {
	return e.invoke(c, appcmd.InteractionTypeApplicationCommand)
}

func (e *entity) AutocompleteCommand(c *fiber.Ctx) error {
	return e.invoke(c, appcmd.InteractionTypeAutocomplete)
}

func (e *entity) invoke(c *fiber.Ctx, interactionType appcmd.InteractionType) error {
	actor, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unable to get user")
	}
	var req appcmd.InvokeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid interaction body")
	}
	if req.CommandID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "command_id is required")
	}
	if req.ChannelID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "channel_id is required")
	}
	command, err := e.appcmd.GetCommand(c.UserContext(), req.CommandID)
	if err != nil {
		return commandLookupError(err)
	}
	guildID := req.GuildID
	if guildID == nil {
		_, resolvedGuildID, err := e.channelAndGuildFromID(c, req.ChannelID)
		if err != nil {
			return err
		}
		guildID = resolvedGuildID
	}
	if guildID == nil {
		return fiber.NewError(fiber.StatusBadRequest, "application commands are currently available in guild channels")
	}
	if command.GuildID != nil && *command.GuildID != *guildID {
		return fiber.NewError(fiber.StatusNotFound, "application command not found")
	}
	if _, err := e.bot.GetBotGuild(c.UserContext(), command.ApplicationID, *guildID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "application command not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "unable to verify bot install")
	}
	if _, _, _, ok, err := e.perm.ChannelPerm(c.UserContext(), *guildID, req.ChannelID, actor.Id, permissions.PermServerViewChannels, permissions.PermUseApplicationCommands); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to check command permissions")
	} else if !ok {
		return fiber.NewError(fiber.StatusForbidden, "you cannot use application commands in this channel")
	}
	effective, err := e.perm.GetChannelPermissions(c.UserContext(), *guildID, req.ChannelID, actor.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to resolve permissions")
	}
	if !commandAllowedInGuild(command, effective) {
		return fiber.NewError(fiber.StatusForbidden, "you cannot use this command")
	}
	appPerms, err := e.perm.GetChannelPermissions(c.UserContext(), *guildID, req.ChannelID, command.ApplicationID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to resolve bot permissions")
	}
	token, tokenPrefix, tokenHash, err := appcmd.GenerateInteractionToken()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to create interaction token")
	}
	interactionID := idgen.Next()
	contextType := appcmd.InteractionContextGuild
	data := appcmd.ApplicationCommandInteractionData{
		ID:       command.ID,
		Name:     command.Name,
		Type:     command.Type,
		Options:  req.Options,
		TargetID: req.TargetID,
		GuildID:  guildID,
	}
	record := appcmd.InteractionRecord{
		ID:             interactionID,
		ApplicationID:  command.ApplicationID,
		CommandID:      command.ID,
		Type:           interactionType,
		GuildID:        guildID,
		ChannelID:      req.ChannelID,
		InvokerUserID:  actor.Id,
		TokenHash:      tokenHash,
		TokenPrefix:    tokenPrefix,
		AppPermissions: appPerms,
		Locale:         defaultString(req.Locale, "en-US"),
		GuildLocale:    "en-US",
		Context:        contextType,
		AckState:       appcmd.AckStatePending,
		ExpiresAt:      time.Now().Add(appcmd.InteractionTokenTTL),
	}
	if err := e.payload.CreateInteractionPayload(c.UserContext(), appcmd.InteractionPayloadRecord{
		ID:            interactionID,
		ApplicationID: command.ApplicationID,
		CommandID:     command.ID,
		Type:          interactionType,
		GuildID:       guildID,
		ChannelID:     req.ChannelID,
		InvokerUserID: actor.Id,
		DataJSON:      appcmd.CommandDataJSON(data),
		CreatedAt:     time.Now(),
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to store interaction payload")
	}
	if err := e.appcmd.CreateInteraction(c.UserContext(), record); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to create interaction")
	}
	member, userDTO := e.interactionActor(c, actor.Id)
	interaction := appcmd.Interaction{
		ID:             interactionID,
		ApplicationID:  command.ApplicationID,
		Type:           interactionType,
		Data:           &data,
		GuildID:        guildID,
		ChannelID:      &req.ChannelID,
		Member:         member,
		User:           userDTO,
		AppPermissions: strconv.FormatInt(appPerms, 10),
		Locale:         record.Locale,
		GuildLocale:    record.GuildLocale,
		Context:        &contextType,
		Token:          token,
		Version:        1,
		AuthorizingIntegrationOwners: map[string]string{
			"0": strconv.FormatInt(*guildID, 10),
		},
	}
	if err := appcmddispatch.DispatchInteraction(c.UserContext(), e.nc, e.registry, command.ApplicationID, guildID, interaction); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to dispatch interaction")
	}
	state := "dispatched"
	if interactionType == appcmd.InteractionTypeAutocomplete {
		state = "autocomplete_dispatched"
	}
	return c.Status(fiber.StatusAccepted).JSON(appcmd.InvokeResponse{InteractionID: interactionID, State: state})
}

func (e *entity) channelAndGuild(c *fiber.Ctx) (int64, *int64, error) {
	channelID, err := strconv.ParseInt(c.Query("channel_id"), 10, 64)
	if err != nil || channelID <= 0 {
		return 0, nil, fiber.NewError(fiber.StatusBadRequest, "channel_id is required")
	}
	return e.channelAndGuildFromID(c, channelID)
}

func (e *entity) channelAndGuildFromID(c *fiber.Ctx, channelID int64) (int64, *int64, error) {
	if raw := strings.TrimSpace(c.Query("guild_id")); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			return 0, nil, fiber.NewError(fiber.StatusBadRequest, "invalid guild_id")
		}
		return channelID, &id, nil
	}
	gc, err := e.gc.GetGuildByChannel(c.UserContext(), channelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return channelID, nil, nil
		}
		return 0, nil, fiber.NewError(fiber.StatusInternalServerError, "unable to resolve channel guild")
	}
	return channelID, &gc.GuildId, nil
}

func (e *entity) interactionActor(c *fiber.Ctx, userID int64) (any, any) {
	u, err := e.user.GetUserById(c.UserContext(), userID)
	if err != nil {
		return nil, nil
	}
	var disc string
	if d, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), userID); err == nil {
		disc = d.Discriminator
	}
	userDTO := dto.User{
		Id:            u.Id,
		Name:          u.Name,
		Discriminator: disc,
		Bio:           u.Bio,
		BannerColor:   u.BannerColor,
		PanelColor:    u.PanelColor,
		IsBot:         u.IsBot(),
	}
	return map[string]any{"user": userDTO}, userDTO
}

func commandAllowedInGuild(command appcmd.ApplicationCommand, effective int64) bool {
	if !hasContext(command.Contexts, appcmd.InteractionContextGuild) {
		return false
	}
	if command.DefaultMemberPermissions == nil {
		return true
	}
	return permissions.CheckPermissions(effective, permissions.PermAdministrator) || effective&*command.DefaultMemberPermissions == *command.DefaultMemberPermissions
}

func hasContext(contexts []appcmd.InteractionContextType, target appcmd.InteractionContextType) bool {
	for _, context := range contexts {
		if context == target {
			return true
		}
	}
	return false
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func commandLookupError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fiber.NewError(fiber.StatusNotFound, "application command not found")
	}
	return fiber.NewError(fiber.StatusInternalServerError, "unable to get application command")
}
