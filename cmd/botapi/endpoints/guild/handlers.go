package guild

import (
	"sort"
	"strconv"

	"github.com/FlameInTheDark/gochat/internal/botauth"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/gofiber/fiber/v2"
)

// List
//
//	@Summary	List guilds where the bot is installed
//	@Produce	json
//	@Tags		Bot Guild
//	@Security	BotToken
//	@Success	200	{array}		Response
//	@Failure	401	{string}	string	"Unauthorized"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/bot/api/v1/guild [get]
func (e *Entity) List(c *fiber.Ctx) error {
	principal, ok := botauth.FromFiber(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "missing bot principal")
	}
	installs, err := e.bot.ListBotGuilds(c.UserContext(), principal.BotUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get bot guilds")
	}
	out := make([]Response, 0, len(installs))
	for _, install := range installs {
		g, err := e.g.GetGuildById(c.UserContext(), install.GuildId)
		if err != nil {
			continue
		}
		out = append(out, Response{
			Id:                 g.Id,
			Name:               g.Name,
			Owner:              g.OwnerId,
			Public:             g.Public,
			GrantedPermissions: install.GrantedPermissions,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Id < out[j].Id })
	return c.JSON(out)
}

// Channels
//
//	@Summary	List channels visible to the bot in a guild
//	@Produce	json
//	@Tags		Bot Guild
//	@Security	BotToken
//	@Param		guild_id	path		int64	true	"Guild id"
//	@Success	200			{array}		dto.Channel
//	@Failure	400			{string}	string	"Bad request"
//	@Failure	401			{string}	string	"Unauthorized"
//	@Failure	403			{string}	string	"Forbidden"
//	@Failure	500			{string}	string	"Internal server error"
//	@Router		/bot/api/v1/guild/{guild_id}/channels [get]
func (e *Entity) Channels(c *fiber.Ctx) error {
	principal, ok := botauth.FromFiber(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "missing bot principal")
	}
	guildID, err := strconv.ParseInt(c.Params("guild_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid guild id")
	}
	if _, err := e.bot.GetBotGuild(c.UserContext(), principal.BotUserID, guildID); err != nil {
		return fiber.NewError(fiber.StatusForbidden, "bot is not installed in this guild")
	}
	guildChannels, err := e.gc.GetGuildChannels(c.UserContext(), guildID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get guild channels")
	}
	channelIDs := make([]int64, 0, len(guildChannels))
	positions := make(map[int64]int, len(guildChannels))
	for _, gc := range guildChannels {
		channelIDs = append(channelIDs, gc.ChannelId)
		positions[gc.ChannelId] = gc.Position
	}
	channels, err := e.ch.GetChannelsBulk(c.UserContext(), channelIDs)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get channels")
	}
	out := make([]dto.Channel, 0, len(channels))
	for _, ch := range channels {
		if ch.Type != model.ChannelTypeGuild && ch.Type != model.ChannelTypeGuildVoice && ch.Type != model.ChannelTypeGuildCategory && ch.Type != model.ChannelTypeThread {
			continue
		}
		_, _, _, allowed, err := e.perm.ChannelPerm(c.UserContext(), guildID, ch.Id, principal.BotUserID, permissions.PermServerViewChannels)
		if err != nil || !allowed {
			continue
		}
		gid := guildID
		last := ch.LastMessage
		out = append(out, dto.Channel{
			Id:            ch.Id,
			Type:          ch.Type,
			GuildId:       &gid,
			Name:          ch.Name,
			ParentId:      ch.ParentID,
			CreatorId:     ch.CreatorID,
			Position:      positions[ch.Id],
			Topic:         ch.Topic,
			Permissions:   ch.Permissions,
			Private:       ch.Private,
			Closed:        ch.Closed,
			LastMessageId: last,
			VoiceRegion:   ch.VoiceRegion,
			CreatedAt:     ch.CreatedAt,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return c.JSON(out)
}
