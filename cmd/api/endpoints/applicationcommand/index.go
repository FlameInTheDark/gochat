package applicationcommand

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/gofiber/fiber/v2"
)

const applicationCommandIndexCacheTTL = int64(60)

func (e *entity) GuildApplicationCommandIndex(c *fiber.Ctx) error {
	actor, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unable to get user")
	}
	guildID, err := strconv.ParseInt(c.Params("guild_id"), 10, 64)
	if err != nil || guildID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid guild_id")
	}
	if _, ok, err := e.perm.GuildPerm(c.UserContext(), guildID, actor.Id, permissions.PermServerViewChannels, permissions.PermUseApplicationCommands); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to check guild permissions")
	} else if !ok {
		return c.JSON(emptyCommandIndex())
	}
	index, err := e.cachedGuildApplicationCommandIndex(c.UserContext(), guildID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to load application command index")
	}
	return c.JSON(index)
}

func (e *entity) cachedGuildApplicationCommandIndex(ctx context.Context, guildID int64) (appcmd.ApplicationCommandIndex, error) {
	key := applicationCommandIndexCacheKey(guildID)
	if e.cache != nil {
		var cached appcmd.ApplicationCommandIndex
		if err := e.cache.GetJSON(ctx, key, &cached); err == nil {
			normalizeCommandIndex(&cached)
			return cached, nil
		}
	}
	index, err := e.buildGuildApplicationCommandIndex(ctx, guildID)
	if err != nil {
		return appcmd.ApplicationCommandIndex{}, err
	}
	if e.cache != nil {
		_ = e.cache.SetTimedJSON(ctx, key, index, applicationCommandIndexCacheTTL)
	}
	return index, nil
}

func (e *entity) buildGuildApplicationCommandIndex(ctx context.Context, guildID int64) (appcmd.ApplicationCommandIndex, error) {
	commands, err := e.appcmd.ListGuildCommandIndex(ctx, guildID)
	if err != nil {
		return appcmd.ApplicationCommandIndex{}, err
	}
	version := int64(0)
	applicationIDs := make([]int64, 0)
	seen := make(map[int64]struct{})
	for _, command := range commands {
		if command.Version > version {
			version = command.Version
		}
		applicationIDs = appendApplicationID(applicationIDs, seen, command.ApplicationID)
	}
	installs, err := e.bot.ListGuildBots(ctx, guildID)
	if err != nil {
		return appcmd.ApplicationCommandIndex{}, err
	}
	for _, install := range installs {
		applicationIDs = appendApplicationID(applicationIDs, seen, install.BotUserId)
	}
	if len(applicationIDs) == 0 {
		return emptyCommandIndex(), nil
	}
	sort.Slice(applicationIDs, func(i, j int) bool { return applicationIDs[i] < applicationIDs[j] })
	users, err := e.user.GetUsersList(ctx, applicationIDs)
	if err != nil {
		return appcmd.ApplicationCommandIndex{}, err
	}
	bots, err := e.bot.GetBotsByIDs(ctx, applicationIDs)
	if err != nil {
		return appcmd.ApplicationCommandIndex{}, err
	}
	userByID := make(map[int64]model.User, len(users))
	for _, u := range users {
		userByID[u.Id] = u
	}
	botByID := make(map[int64]model.Bot, len(bots))
	for _, b := range bots {
		botByID[b.BotUserId] = b
	}
	applications := make([]appcmd.ApplicationCommandIndexApplication, 0, len(applicationIDs))
	for _, id := range applicationIDs {
		u, ok := userByID[id]
		if !ok {
			continue
		}
		b, ok := botByID[id]
		if !ok || b.Disabled {
			continue
		}
		var icon *string
		if u.Avatar != nil {
			raw := strconv.FormatInt(*u.Avatar, 10)
			icon = &raw
		}
		applications = append(applications, appcmd.ApplicationCommandIndexApplication{
			ID:          id,
			Name:        u.Name,
			Description: b.Description,
			Icon:        icon,
			BotID:       id,
			Flags:       "0",
		})
	}
	sort.SliceStable(applications, func(i, j int) bool {
		if applications[i].Name == applications[j].Name {
			return applications[i].ID < applications[j].ID
		}
		return applications[i].Name < applications[j].Name
	})
	index := appcmd.ApplicationCommandIndex{
		Applications:        applications,
		ApplicationCommands: commands,
		Version:             version,
	}
	normalizeCommandIndex(&index)
	return index, nil
}

func appendApplicationID(ids []int64, seen map[int64]struct{}, id int64) []int64 {
	if id == 0 {
		return ids
	}
	if _, ok := seen[id]; ok {
		return ids
	}
	seen[id] = struct{}{}
	return append(ids, id)
}

func applicationCommandIndexCacheKey(guildID int64) string {
	return fmt.Sprintf("appcmd:index:guild:%d", guildID)
}

func emptyCommandIndex() appcmd.ApplicationCommandIndex {
	return appcmd.ApplicationCommandIndex{
		Applications:        []appcmd.ApplicationCommandIndexApplication{},
		ApplicationCommands: []appcmd.ApplicationCommand{},
		Version:             0,
	}
}

func normalizeCommandIndex(index *appcmd.ApplicationCommandIndex) {
	if index.Applications == nil {
		index.Applications = []appcmd.ApplicationCommandIndexApplication{}
	}
	if index.ApplicationCommands == nil {
		index.ApplicationCommands = []appcmd.ApplicationCommand{}
	}
}
