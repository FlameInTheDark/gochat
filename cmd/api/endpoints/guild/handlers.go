package guild

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/embed"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/threadcount"
	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
)

// Helper functions for common operations

// parseGuildID extracts and validates guild ID from URL parameters
func (e *entity) parseGuildID(c *fiber.Ctx) (int64, error) {
	guildIdStr := c.Params("guild_id")
	guildId, err := strconv.ParseInt(guildIdStr, 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}
	return guildId, nil
}

// parseChannelID extracts and validates channel ID from URL parameters
func (e *entity) parseChannelID(c *fiber.Ctx) (int64, error) {
	channelIdStr := c.Params("channel_id")
	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	return channelId, nil
}

// parseCategoryID extracts and validates category ID from URL parameters
func (e *entity) parseCategoryID(c *fiber.Ctx) (int64, error) {
	categoryIdStr := c.Params("category_id")
	categoryId, err := strconv.ParseInt(categoryIdStr, 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	return categoryId, nil
}

func (e *entity) parseUserID(c *fiber.Ctx) (int64, error) {
	memberIdStr := c.Params("user_id")
	memberId, err := strconv.ParseInt(memberIdStr, 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectMemberID)
	}
	return memberId, nil
}

// parseRoleID extracts and validates role ID from URL parameters
func (e *entity) parseRoleID(c *fiber.Ctx) (int64, error) {
	roleIdStr := c.Params("role_id")
	roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, ErrIncorrectRoleID)
	}
	return roleId, nil
}

// validateGuildAccess validates user access to guild and returns guild context
func (e *entity) validateGuildAccess(c *fiber.Ctx, guildId int64) (*guildContext, error) {
	user, err := helper.GetUser(c)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	member, err := e.memb.GetMember(c.UserContext(), user.Id, guildId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fiber.NewError(fiber.StatusForbidden, ErrNotAMember)
		}
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMember)
	}

	guild, err := e.g.GetGuildById(c.UserContext(), member.GuildId)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return &guildContext{
		User:   user,
		Guild:  &guild,
		Member: &member,
	}, nil
}

// getUserRoles fetches user roles for a guild
func (e *entity) getUserRoles(c *fiber.Ctx, guildId, userId int64) (map[int64]*model.Role, error) {
	userRoles, err := e.ur.GetUserRoles(c.UserContext(), guildId, userId)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var roleIds []int64
	for _, ur := range userRoles {
		roleIds = append(roleIds, ur.RoleId)
	}

	roles, err := e.role.GetRolesBulk(c.UserContext(), guildId, roleIds)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	roleMap := make(map[int64]*model.Role)
	for i, role := range roles {
		roleMap[role.Id] = &roles[i]
	}

	return roleMap, nil
}

func buildThreadMemberDTO(member *model.ThreadMember) *dto.ThreadMember {
	if member == nil {
		return nil
	}
	return &dto.ThreadMember{
		UserId:        member.UserId,
		JoinTimestamp: member.JoinAt,
		Flags:         member.Flags,
	}
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

func (e *entity) applyThreadMessageCounts(ctx context.Context, channels []model.Channel) {
	for i := range channels {
		e.applyThreadMessageCount(ctx, &channels[i])
	}
}

func buildThreadMemberIDs(members []model.ThreadMember) []int64 {
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

func (e *entity) currentUserThreadMember(ctx context.Context, userID int64, channel *model.Channel) (*dto.ThreadMember, []int64, error) {
	if channel == nil || channel.Type != model.ChannelTypeThread {
		return nil, nil, nil
	}

	members, err := e.tm.GetThreadMembers(ctx, channel.Id)
	if err != nil {
		return nil, nil, err
	}
	var currentMember *dto.ThreadMember
	for i := range members {
		if members[i].UserId == userID {
			currentMember = buildThreadMemberDTO(&members[i])
			break
		}
	}
	return currentMember, buildThreadMemberIDs(members), nil
}

func (e *entity) currentUserThreadMembers(ctx context.Context, userID int64, channels []model.Channel) (map[int64]*dto.ThreadMember, map[int64][]int64, error) {
	threadIDs := make([]int64, 0, len(channels))
	for _, channel := range channels {
		if channel.Type == model.ChannelTypeThread {
			threadIDs = append(threadIDs, channel.Id)
		}
	}
	if len(threadIDs) == 0 {
		return map[int64]*dto.ThreadMember{}, map[int64][]int64{}, nil
	}

	members, err := e.tm.GetThreadMembersBulk(ctx, threadIDs)
	if err != nil {
		return nil, nil, err
	}
	byThreadID := make(map[int64]*dto.ThreadMember, len(members))
	memberIDsByThread := make(map[int64][]int64, len(threadIDs))
	seenByThread := make(map[int64]map[int64]struct{}, len(threadIDs))
	for i := range members {
		member := members[i]
		seen, ok := seenByThread[member.ThreadId]
		if !ok {
			seen = make(map[int64]struct{})
			seenByThread[member.ThreadId] = seen
		}
		if _, ok := seen[member.UserId]; !ok {
			seen[member.UserId] = struct{}{}
			memberIDsByThread[member.ThreadId] = append(memberIDsByThread[member.ThreadId], member.UserId)
		}
		if member.UserId == userID {
			byThreadID[member.ThreadId] = buildThreadMemberDTO(&member)
		}
	}
	return byThreadID, memberIDsByThread, nil
}

// checkChannelPermissions validates if user can view a private channel
func (e *entity) checkChannelPermissions(c *fiber.Ctx, channel *model.Channel, guild *model.Guild, user *helper.JWTUser, roles map[int64]*model.Role) (bool, error) {
	_, _, _, canView, err := e.perm.ChannelPerm(c.UserContext(), guild.Id, channel.Id, user.Id, permissions.PermServerViewChannels)
	if err != nil {
		return false, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return canView, nil
}

// createDefaultChannels creates default text category and general channel for new guild
func (e *entity) createDefaultChannels(c *fiber.Ctx, guildId int64, isPublic bool) (int64, error) {
	categoryId := idgen.Next()
	channelId := idgen.Next()

	if err := e.gc.AddChannel(c.UserContext(), guildId, categoryId, "text", model.ChannelTypeGuildCategory, nil, isPublic, 0, nil, nil, false); err != nil {
		return 0, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if err := e.gc.AddChannel(c.UserContext(), guildId, channelId, "general", model.ChannelTypeGuild, &categoryId, isPublic, 0, nil, nil, false); err != nil {
		return 0, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return channelId, nil
}

// Get
//
//	@Summary	Get guild
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64		true	"Guild id"	example(2230469276416868352)
//	@Success	200			{object}	dto.Guild	"Guild"
//	@failure	400			{string}	string		"Incorrect request body"
//	@failure	401			{string}	string		"Unauthorized"
//	@failure	500			{string}	string		"Something bad happened"
//	@Router		/guild/{guild_id} [get]
func (e *entity) Get(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	guildCtx, err := e.validateGuildAccess(c, guildId)
	if err != nil {
		return err
	}

	return c.JSON(e.dtoGuildWithIcon(c, guildCtx.Guild))
}

// GetChannels
//
//	@Summary	Get guild channels
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64		true	"Guild id"	example(2230469276416868352)
//	@Success	200			{array}		dto.Channel	"List of channels"
//	@failure	400			{string}	string		"Incorrect request body"
//	@failure	401			{string}	string		"Unauthorized"
//	@failure	500			{string}	string		"Something bad happened"
//	@Router		/guild/{guild_id}/channel [get]
func (e *entity) GetChannels(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	guildCtx, err := e.validateGuildAccess(c, guildId)
	if err != nil {
		return err
	}

	return e.fetchAndFilterChannels(c, guildCtx)
}

// deriveChannelParents assigns ParentId to guild/voice channels based on positional order.
// After sorting by Position, each guild or voice channel inherits the id of the last
// category above it. Channels before any category have a nil parent.
// Thread channels keep their existing ParentId unchanged.
func deriveChannelParents(channels []dto.Channel) {
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Position < channels[j].Position
	})
	var currentCategoryId *int64
	for i := range channels {
		switch channels[i].Type {
		case model.ChannelTypeGuildCategory:
			id := channels[i].Id
			currentCategoryId = &id
			channels[i].ParentId = nil
		case model.ChannelTypeGuild, model.ChannelTypeGuildVoice:
			channels[i].ParentId = currentCategoryId
		}
	}
}

// fetchAndFilterChannels retrieves guild channels and filters based on permissions
func (e *entity) fetchAndFilterChannels(c *fiber.Ctx, guildCtx *guildContext) error {
	var cachedChannels []dto.Channel
	err := e.cache.GetJSON(c.UserContext(), fmt.Sprintf("guild:%d:channels", guildCtx.Guild.Id), cachedChannels)
	if err == nil {
		return c.JSON(cachedChannels)
	}

	guildChannels, err := e.gc.GetGuildChannels(c.UserContext(), guildCtx.Guild.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var channelIds = make([]int64, len(guildChannels))
	for i, gch := range guildChannels {
		channelIds[i] = gch.ChannelId
	}

	channels, err := e.ch.GetChannelsBulk(c.UserContext(), channelIds)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	croles, err := e.rperm.GetChannelRolesBulk(c.UserContext(), channelIds)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	channelsData := make([]dto.Channel, 0, len(channels))
	for i, ch := range channels {
		if ch.Type == model.ChannelTypeThread {
			continue
		}
		if ch.Permissions == nil {
			ch.Permissions = &guildCtx.Guild.Permissions
		}
		channelsData = append(channelsData, channelModelToDTO(&ch, &guildCtx.Guild.Id, guildChannels[i].Position, croles[i].Roles))
	}

	deriveChannelParents(channelsData)

	go func() {
		if err := e.cache.SetTimedJSON(
			context.Background(),
			fmt.Sprintf("guild:%d:channels", guildCtx.Guild.Id),
			channelsData,
			3600); err != nil {
			slog.Error("unable to set cached response for guild channels list", slog.String("error", err.Error()))
		}
	}()

	return c.JSON(channelsData)
}

// GetChannel
//
//	@Summary	Get guild channel
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64		true	"Guild id"		example(2230469276416868352)
//	@Param		channel_id	path		int64		true	"Channel id"	example(2230469276416868352)
//	@Success	200			{object}	dto.Channel	"Channel"
//	@failure	400			{string}	string		"Incorrect request body"
//	@failure	401			{string}	string		"Unauthorized"
//	@failure	500			{string}	string		"Something bad happened"
//	@Router		/guild/{guild_id}/channel/{channel_id} [get]
func (e *entity) GetChannel(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}

	guildCtx, err := e.validateGuildAccess(c, guildId)
	if err != nil {
		return err
	}

	roles, err := e.getUserRoles(c, guildCtx.Guild.Id, guildCtx.User.Id)
	if err != nil {
		return err
	}

	return e.fetchSingleChannel(c, guildCtx, channelId, roles)
}

// GetChannelThreads
//
//	@Summary	Get channel threads
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64		true	"Guild id"		example(2230469276416868352)
//	@Param		channel_id	path		int64		true	"Channel id"	example(2230469276416868352)
//	@Success	200			{array}		dto.Channel	"List of threads"
//	@failure	400			{string}	string		"Incorrect request body"
//	@failure	401			{string}	string		"Unauthorized"
//	@failure	500			{string}	string		"Something bad happened"
//	@Router		/guild/{guild_id}/channel/{channel_id}/threads [get]
func (e *entity) GetChannelThreads(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}

	guildCtx, err := e.validateGuildAccess(c, guildId)
	if err != nil {
		return err
	}

	parentChannel, _, _, canView, err := e.perm.ChannelPerm(c.UserContext(), guildCtx.Guild.Id, channelId, guildCtx.User.Id, permissions.PermServerViewChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !canView || parentChannel == nil {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}
	if parentChannel.Type != model.ChannelTypeGuild {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetChannel)
	}

	threads, err := e.ch.GetChannelThreads(c.UserContext(), channelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if len(threads) == 0 {
		return c.JSON([]dto.Channel{})
	}
	e.applyThreadMessageCounts(c.UserContext(), threads)
	threadMembers, threadMemberIDs, err := e.currentUserThreadMembers(c.UserContext(), guildCtx.User.Id, threads)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	guildChannels, err := e.gc.GetGuildChannels(c.UserContext(), guildCtx.Guild.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	positions := make(map[int64]int, len(guildChannels))
	for _, guildChannel := range guildChannels {
		positions[guildChannel.ChannelId] = guildChannel.Position
	}

	resp := make([]dto.Channel, 0, len(threads))
	for i := range threads {
		if threads[i].Permissions == nil {
			if parentChannel.Permissions != nil {
				threads[i].Permissions = parentChannel.Permissions
			} else {
				threads[i].Permissions = &guildCtx.Guild.Permissions
			}
		}
		resp = append(resp, channelModelToDTOWithThreadMember(&threads[i], &guildCtx.Guild.Id, positions[threads[i].Id], nil, threadMembers[threads[i].Id], threadMemberIDs[threads[i].Id]))
	}

	return c.JSON(resp)
}

// fetchSingleChannel retrieves and validates access to a specific channel
func (e *entity) fetchSingleChannel(c *fiber.Ctx, guildCtx *guildContext, channelId int64, roles map[int64]*model.Role) error {
	guildChannel, err := e.gc.GetGuildChannel(c.UserContext(), guildCtx.Guild.Id, channelId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetChannel)
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	channel, err := e.ch.GetChannel(c.UserContext(), guildChannel.ChannelId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetChannel)
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if channel.Type == model.ChannelTypeThread && channel.Permissions == nil && channel.ParentID != nil {
		if parentChannel, err := e.ch.GetChannel(c.UserContext(), *channel.ParentID); err == nil {
			if parentChannel.Permissions != nil {
				channel.Permissions = parentChannel.Permissions
			} else {
				channel.Permissions = &guildCtx.Guild.Permissions
			}
		}
	}
	if channel.Permissions == nil {
		channel.Permissions = &guildCtx.Guild.Permissions
	}
	e.applyThreadMessageCount(c.UserContext(), &channel)

	canView, err := e.checkChannelPermissions(c, &channel, guildCtx.Guild, guildCtx.User, roles)
	if err != nil {
		return err
	}

	if !canView {
		return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
	}

	threadMember, threadMemberIDs, err := e.currentUserThreadMember(c.UserContext(), guildCtx.User.Id, &channel)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(channelModelToDTOWithThreadMember(&channel, &guildCtx.Guild.Id, guildChannel.Position, nil, threadMember, threadMemberIDs))
}

// JoinThread
//
//	@Summary	Join thread
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64				true	"Guild ID"
//	@Param		channel_id	path		int64				true	"Thread channel ID"
//	@Success	200			{object}	dto.ThreadMember	"Thread membership"
//	@failure	400			{string}	string				"Incorrect request body"
//	@failure	401			{string}	string				"Unauthorized"
//	@failure	403			{string}	string				"Forbidden"
//	@failure	500			{string}	string				"Something bad happened"
//	@Router		/guild/{guild_id}/channel/{channel_id}/thread-member/me [put]
func (e *entity) JoinThread(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}

	guildCtx, err := e.validateGuildAccess(c, guildId)
	if err != nil {
		return err
	}

	channel, _, _, canView, err := e.perm.ChannelPerm(c.UserContext(), guildCtx.Guild.Id, channelId, guildCtx.User.Id, permissions.PermServerViewChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !canView || channel == nil {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}
	if channel.Type != model.ChannelTypeThread {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetChannel)
	}

	member, err := e.tm.AddThreadMember(c.UserContext(), channelId, guildCtx.User.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to join thread")
	}

	return c.JSON(buildThreadMemberDTO(&member))
}

// LeaveThread
//
//	@Summary	Leave thread
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64	true	"Guild ID"
//	@Param		channel_id	path		int64	true	"Thread channel ID"
//	@Success	200			{string}	string	"Left"
//	@failure	400			{string}	string	"Incorrect request body"
//	@failure	401			{string}	string	"Unauthorized"
//	@failure	403			{string}	string	"Forbidden"
//	@failure	500			{string}	string	"Something bad happened"
//	@Router		/guild/{guild_id}/channel/{channel_id}/thread-member/me [delete]
func (e *entity) LeaveThread(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}

	guildCtx, err := e.validateGuildAccess(c, guildId)
	if err != nil {
		return err
	}

	channel, _, _, canView, err := e.perm.ChannelPerm(c.UserContext(), guildCtx.Guild.Id, channelId, guildCtx.User.Id, permissions.PermServerViewChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !canView || channel == nil {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}
	if channel.Type != model.ChannelTypeThread {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetChannel)
	}

	if err := e.tm.RemoveThreadMember(c.UserContext(), channelId, guildCtx.User.Id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to leave thread")
	}

	return c.SendStatus(fiber.StatusOK)
}

// Create
//
//	@Summary	Create guild
//	@Produce	json
//	@Tags		Guild
//	@Param		request	body		CreateGuildRequest	true	"Guild data"
//	@Success	200		{object}	dto.Guild			"Guild"
//	@failure	400		{string}	string				"Incorrect request body"
//	@failure	401		{string}	string				"Unauthorized"
//	@failure	500		{string}	string				"Something bad happened"
//	@Router		/guild [post]
func (e *entity) Create(c *fiber.Ctx) error {
	var req CreateGuildRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return e.createGuildWithDefaults(c, &req, user)
}

// createGuildWithDefaults creates a new guild with default channels and settings
func (e *entity) createGuildWithDefaults(c *fiber.Ctx, req *CreateGuildRequest, user *helper.JWTUser) error {
	guildId := idgen.Next()

	// Create the guild
	if err := e.g.CreateGuild(c.UserContext(), guildId, req.Name, user.Id, permissions.DefaultPermissions); err != nil {
		e.log.Error("unable to create guild", slog.String("error", err.Error()))
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateGuild)
	}

	// Set guild icon if provided
	if err := e.setGuildIconIfProvided(c, guildId, req.IconId); err != nil {
		return err
	}

	// Create default channels
	ch, err := e.createDefaultChannels(c, guildId, req.Public)
	if err != nil {
		return err
	}

	// Add creator as member
	if err := e.memb.AddMember(c.UserContext(), user.Id, guildId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Load created guild to include computed fields and icon metadata
	createdGuild, err := e.g.GetGuildById(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateGuild)
	}

	if err := e.g.SetSystemMessagesChannel(c.UserContext(), guildId, &ch); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSetSystemMessagesChannel)
	}

	return c.JSON(e.dtoGuildWithIcon(c, &createdGuild))
}

// dtoGuildWithIcon enriches guild DTO with icon metadata from Cassandra if available
func (e *entity) dtoGuildWithIcon(c *fiber.Ctx, guild *model.Guild) dto.Guild {
	dtoG := buildGuildDTO(guild)
	if guild.Icon == nil {
		return dtoG
	}

	key := fmt.Sprintf("icons:%d:%d", guild.Id, *guild.Icon)
	var cached dto.Icon
	if err := e.cache.GetJSON(c.UserContext(), key, &cached); err == nil && cached.URL != "" {
		dtoG.Icon = &cached
		return dtoG
	}

	if ic, err := e.icon.GetIcon(c.UserContext(), *guild.Icon, guild.Id); err == nil && ic.URL != nil {
		var w, h, size int64
		if ic.Width != nil {
			w = *ic.Width
		}
		if ic.Height != nil {
			h = *ic.Height
		}
		size = ic.FileSize
		var urlStr string
		if ic.URL != nil {
			urlStr = *ic.URL
		}
		ico := dto.Icon{Id: *guild.Icon, URL: urlStr, Filesize: size, Width: w, Height: h}
		dtoG.Icon = &ico
		_ = e.cache.SetJSON(c.UserContext(), key, ico)
	}
	return dtoG
}

// Delete
//
//	@Summary		Delete guild
//	@Description	Deletes a guild. Only the guild owner can delete a guild. This removes all members, all guild icons, and all guild channels.
//	@Tags			Guild
//	@Param			guild_id	path		int64	true	"Guild ID"	example(2230469276416868352)
//	@Success		200			{string}	string	"Deleted"
//	@failure		401			{string}	string	"Unauthorized"
//	@failure		403			{string}	string	"Forbidden"
//	@failure		500			{string}	string	"Internal server error"
//	@Router			/guild/{guild_id} [delete]
func (e *entity) Delete(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	// Get user and guild, ensure user is the owner
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	g, err := e.g.GetGuildById(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}
	if g.OwnerId != user.Id {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	// 1) Remove all guild channels and their messages
	guildChannels, err := e.gc.GetGuildChannels(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}
	for _, gch := range guildChannels {
		ch, chErr := e.ch.GetChannel(c.UserContext(), gch.ChannelId)
		if chErr != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
		}
		if ch.LastMessage != 0 {
			if delErr := e.msg.DeleteChannelMessages(c.UserContext(), ch.Id, ch.LastMessage); delErr != nil {
				return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUpdateChannel)
			}
		}
		if remErr := e.gc.RemoveChannel(c.UserContext(), guildId, ch.Id); remErr != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUpdateChannel)
		}
	}

	// Clean guild channels cache
	_ = e.cache.Delete(c.UserContext(), fmt.Sprintf("guild:%d:channels", guildId))

	// 2) Remove all guild icons (metadata)
	icons, err := e.icon.GetIconsByGuildId(c.UserContext(), guildId)
	if err == nil {
		for _, ic := range icons {
			_ = e.icon.RemoveIcon(c.UserContext(), ic.Id, guildId)
			_ = e.cache.Delete(c.UserContext(), fmt.Sprintf("icons:%d:%d", guildId, ic.Id))
		}
	}

	// 3) Remove all guild emojis
	emojis, err := e.emoji.DeleteGuildEmojis(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDeleteEmoji)
	}
	for _, emoji := range emojis {
		if err := e.removeEmojiObjects(c.UserContext(), emoji.Id); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDeleteEmoji)
		}
		_ = e.invalidateEmojiCache(c.UserContext(), guildId, emoji.Id)
	}

	// 4) Remove all members
	if err := e.memb.RemoveMembersByGuild(c.UserContext(), guildId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMembers)
	}

	// 5) Delete the guild itself
	if err := e.g.DeleteGuild(c.UserContext(), guildId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDeleteGuild)
	}

	return c.SendStatus(fiber.StatusOK)
}

// setGuildIconIfProvided sets guild icon if icon ID is provided and valid
func (e *entity) setGuildIconIfProvided(c *fiber.Ctx, guildId int64, iconId *int64) error {
	if iconId == nil {
		return nil
	}

	key := fmt.Sprintf("icons:%d:%d", guildId, *iconId)
	var cached dto.Icon
	if err := e.cache.GetJSON(c.UserContext(), key, &cached); err == nil && cached.URL != "" {
		if err := e.g.SetGuildIcon(c.UserContext(), guildId, cached.Id); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return nil
	}

	icon, err := e.icon.GetIcon(c.UserContext(), *iconId, guildId)
	if err != nil {
		return nil
	}
	// Cache result for reuse
	if icon.URL != nil {
		var w, h int64
		if icon.Width != nil {
			w = *icon.Width
		}
		if icon.Height != nil {
			h = *icon.Height
		}
		var urlStr string
		if icon.URL != nil {
			urlStr = *icon.URL
		}
		_ = e.cache.SetJSON(c.UserContext(), key, dto.Icon{Id: *iconId, URL: urlStr, Filesize: icon.FileSize, Width: w, Height: h})
	}

	if err := e.g.SetGuildIcon(c.UserContext(), guildId, icon.Id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return nil
}

// Update
//
//	@Summary	Update guild
//	@Produce	json
//	@Tags		Guild
//	@Param		request		body		UpdateGuildRequest	true	"Update guild data"
//	@Param		guild_id	path		int64				true	"Guild ID"	example(2230469276416868352)
//	@Success	200			{object}	dto.Guild			"Guild"
//	@failure	400			{string}	string				"Incorrect request body"
//	@failure	401			{string}	string				"Unauthorized"
//	@failure	500			{string}	string				"Something bad happened"
//	@Router		/guild/{guild_id} [patch]
func (e *entity) Update(c *fiber.Ctx) error {
	var req UpdateGuildRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return e.updateGuildWithPermissionCheck(c, guildId, user.Id, &req)
}

// SetSystemMessagesChannel
//
//	@Summary	Set system messages channel
//	@Produce	json
//	@Tags		Guild
//	@Param		request		body		SetGuildSystemMessagesChannelRequest	true	"Set system messages channel"
//	@Param		guild_id	path		int64									true	"Guild ID"	example(2230469276416868352)
//	@Success	200			{object}	dto.Guild								"Guild"
//	@failure	400			{string}	string									"Incorrect request body"
//	@failure	401			{string}	string									"Unauthorized"
//	@failure	500			{string}	string									"Something bad happened"
//	@Router		/guild/{guild_id}/systemch [patch]
func (e *entity) SetSystemMessagesChannel(c *fiber.Ctx) error {
	var req SetGuildSystemMessagesChannelRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	guild, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermAdministrator)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, ErrUnableToGetPermission)
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
	}

	if req.ChannelId != nil {
		_, err := e.gc.GetGuildChannel(c.UserContext(), guild.Id, *req.ChannelId)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetChannel)
		}
	}

	err = e.g.SetSystemMessagesChannel(c.UserContext(), guild.Id, req.ChannelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSetSystemMessagesChannel)
	}
	return c.SendStatus(fiber.StatusOK)
}

// updateGuildWithPermissionCheck validates permissions and updates guild
func (e *entity) updateGuildWithPermissionCheck(c *fiber.Ctx, guildId, userId int64, req *UpdateGuildRequest) error {
	guild, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, userId, permissions.PermServerManage)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	// Update guild
	if err := e.g.UpdateGuild(c.UserContext(), guild.Id, req.Name, req.IconId, req.Public, req.Permissions); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUpdateGuild)
	}

	// Get updated guild for response
	updatedGuild, err := e.g.GetGuildById(c.UserContext(), guild.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUpdateGuild)
	}

	// Send update event
	if err := e.sendGuildUpdateEvent(guildId, &updatedGuild); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// sendGuildUpdateEvent sends guild update message to message queue
func (e *entity) sendGuildUpdateEvent(guildId int64, guild *model.Guild) error {
	if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.UpdateGuild{
		Guild: buildGuildDTO(guild),
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
	}
	return nil
}

// CreateCategory
//
//	@Summary	Create guild channel category
//	@Produce	json
//	@Tags		Guild
//	@Param		request		body		CreateGuildChannelCategoryRequest	true	"Create category data"
//	@Param		guild_id	path		int64								true	"Guild ID"	example(2230469276416868352)
//	@Success	201			{object}	string								"Created"
//	@failure	400			{string}	string								"Incorrect request body"
//	@failure	401			{string}	string								"Unauthorized"
//	@failure	500			{string}	string								"Something bad happened"
//	@Router		/guild/{guild_id}/category [post]
func (e *entity) CreateCategory(c *fiber.Ctx) error {
	var req CreateGuildChannelCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return e.createChannelWithPermissionCheck(c, guildId, user.Id, req.Name, model.ChannelTypeGuildCategory, nil, req.Private, req.Position)
}

// createChannelWithPermissionCheck validates permissions and creates a channel
func (e *entity) createChannelWithPermissionCheck(c *fiber.Ctx, guildId, userId int64, name string, channelType model.ChannelType, parentId *int64, isPrivate bool, position int) error {
	guild, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, userId, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	channelId := idgen.Next()

	// Add channel to guild
	if err := e.gc.AddChannel(c.UserContext(), guild.Id, channelId, name, channelType, nil, isPrivate, position, nil, nil, false); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
	}

	// Send create channel event and clean cached data
	go func() {
		if err := e.sendCreateChannelEvent(guildId, guild.Id, channelId, name, channelType, nil); err != nil {
			slog.Error("unable to send create channel event", slog.String("error", err.Error()))
		}
		if err := e.cache.Delete(context.Background(), fmt.Sprintf("guild:%d:channels", guildId)); err != nil {
			slog.Error("unable to clean cached channels value", slog.String("error", err.Error()))
		}
	}()

	return c.SendStatus(fiber.StatusCreated)
}

func (e *entity) canManageThread(ctx context.Context, guildId int64, thread *model.Channel, userId int64) (bool, error) {
	if thread.Type != model.ChannelTypeThread {
		return false, nil
	}
	if thread.CreatorID != nil && *thread.CreatorID == userId {
		return true, nil
	}

	_, _, _, canManage, err := e.perm.ChannelPerm(ctx, guildId, thread.Id, userId, permissions.PermTextManageThreads)
	if err != nil {
		return false, err
	}
	return canManage, nil
}

func (e *entity) validatePatchChannelRequest(channel *model.Channel, req *PatchGuildChannelRequest) error {
	if channel.Type == model.ChannelTypeThread {
		if req.Private != nil {
			return fiber.NewError(fiber.StatusBadRequest, "threads inherit parent permissions")
		}
		if req.Name != nil {
			if err := validateThreadChannelName(*req.Name); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}
		}
		return nil
	}

	if req.Closed != nil {
		return fiber.NewError(fiber.StatusBadRequest, "only threads can be closed")
	}
	if req.Name != nil {
		if err := validateGuildChannelName(*req.Name); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}
	return nil
}

type threadCreatedMessageUpdate struct {
	channelID int64
	message   dto.Message
}

func (e *entity) buildThreadMessageAuthor(ctx context.Context, userID int64) dto.User {
	if e.user == nil || e.disc == nil {
		return dto.User{Id: userID, Name: strconv.FormatInt(userID, 10)}
	}
	user, err := e.user.GetUserById(ctx, userID)
	if err != nil {
		return dto.User{Id: userID, Name: strconv.FormatInt(userID, 10)}
	}

	discriminator, err := e.disc.GetDiscriminatorByUserId(ctx, userID)
	if err != nil {
		return dto.User{Id: userID, Name: user.Name}
	}

	author := userToDTO(user, discriminator.Discriminator)
	if user.Avatar != nil {
		if ad, err := e.getAvatarDataCached(ctx, userID, *user.Avatar); err == nil && ad != nil {
			author.Avatar = ad
		}
	}
	return author
}

func guildOptionalInt64(value int64) *int64 {
	if value == 0 {
		return nil
	}
	v := value
	return &v
}

func guildOptionalReferenceChannelID(messageChannelID, referenceChannelID, referenceID int64) *int64 {
	if referenceID == 0 {
		return nil
	}
	if referenceChannelID == 0 {
		referenceChannelID = messageChannelID
	}
	return guildOptionalInt64(referenceChannelID)
}

func (e *entity) buildGuildMessageAttachments(ctx context.Context, channelID int64, attachmentIDs []int64) []dto.Attachment {
	if len(attachmentIDs) == 0 || e.at == nil {
		return nil
	}

	attachments, err := e.at.SelectAttachmentsByChannel(ctx, channelID, attachmentIDs)
	if err != nil {
		if e.log != nil {
			e.log.Error("unable to load message attachments",
				"channel_id", channelID,
				"error", err.Error())
		}
		return nil
	}

	attachmentsByID := make(map[int64]model.Attachment, len(attachments))
	for _, attachment := range attachments {
		attachmentsByID[attachment.Id] = attachment
	}

	result := make([]dto.Attachment, 0, len(attachmentIDs))
	for _, attachmentID := range attachmentIDs {
		attachment, ok := attachmentsByID[attachmentID]
		if !ok {
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

func (e *entity) buildGuildMessageDTO(ctx context.Context, message model.Message, thread *dto.Channel) dto.Message {
	flags := model.NormalizeMessageFlags(message.Flags)
	mergedEmbeds, err := embed.ParseMergedEmbeds(message.EmbedsJSON, message.AutoEmbedsJSON, model.HasMessageFlag(flags, model.MessageFlagSuppressEmbeds))
	if err != nil && e.log != nil {
		e.log.Error("unable to parse merged message embeds",
			"channel_id", message.ChannelId,
			"message_id", message.Id,
			"error", err.Error())
	}

	return dto.Message{
		Id:                 message.Id,
		ChannelId:          message.ChannelId,
		Author:             e.buildThreadMessageAuthor(ctx, message.UserId),
		Content:            message.Content,
		Position:           guildOptionalInt64(message.Position),
		Attachments:        e.buildGuildMessageAttachments(ctx, message.ChannelId, message.Attachments),
		Embeds:             mergedEmbeds,
		Flags:              flags,
		Type:               message.Type,
		Reference:          guildOptionalInt64(message.Reference),
		ReferenceChannelId: guildOptionalReferenceChannelID(message.ChannelId, message.ReferenceChannel, message.Reference),
		ThreadId:           guildOptionalInt64(message.Thread),
		Thread:             thread,
		UpdatedAt:          message.EditedAt,
	}
}

type deletedThreadMessageRefs struct {
	parentChannelID int64
	followupMessage *model.Message
	sourceChannelID int64
	sourceMessage   *model.Message
}

func (e *entity) loadDeletedThreadMessageRefs(ctx context.Context, threadID int64) (*deletedThreadMessageRefs, error) {
	parentChannelID, followupMessageID, err := e.msg.GetThreadCreatedMessageRef(ctx, threadID)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	refs := &deletedThreadMessageRefs{parentChannelID: parentChannelID}
	followupMessage, err := e.msg.GetMessage(ctx, followupMessageID, parentChannelID)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return refs, nil
		}
		return nil, err
	}
	refs.followupMessage = &followupMessage

	if followupMessage.Reference == 0 {
		return refs, nil
	}

	sourceChannelID := parentChannelID
	if followupMessage.ReferenceChannel != 0 {
		sourceChannelID = followupMessage.ReferenceChannel
	}
	sourceMessage, err := e.msg.GetMessage(ctx, followupMessage.Reference, sourceChannelID)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return refs, nil
		}
		return nil, err
	}
	refs.sourceChannelID = sourceChannelID
	refs.sourceMessage = &sourceMessage
	return refs, nil
}

func (e *entity) sendDetachedThreadMessageUpdate(guildID int64, message dto.Message) {
	if err := e.mqt.SendChannelMessage(message.ChannelId, &mqmsg.UpdateMessage{
		GuildId: &guildID,
		Message: message,
	}); err != nil && e.log != nil {
		e.log.Error("unable to send detached thread message update",
			"channel_id", message.ChannelId,
			"message_id", message.Id,
			"error", err.Error())
	}
}

func (e *entity) detachDeletedThreadMessages(ctx context.Context, guildID, threadID int64) error {
	refs, err := e.loadDeletedThreadMessageRefs(ctx, threadID)
	if err != nil || refs == nil {
		return err
	}

	var detachErr error

	if refs.sourceMessage != nil {
		if err := e.msg.SetThread(ctx, refs.sourceMessage.Id, refs.sourceChannelID, 0); err != nil {
			detachErr = errors.Join(detachErr, fmt.Errorf("detach source message thread: %w", err))
		} else {
			refs.sourceMessage.Thread = 0
			e.sendDetachedThreadMessageUpdate(guildID, e.buildGuildMessageDTO(ctx, *refs.sourceMessage, nil))
		}
		if err := e.msg.ReleaseThreadClaim(ctx, refs.sourceChannelID, refs.sourceMessage.Id); err != nil {
			detachErr = errors.Join(detachErr, fmt.Errorf("release source message thread claim: %w", err))
		}
	}

	if refs.followupMessage != nil {
		if err := e.msg.SetThread(ctx, refs.followupMessage.Id, refs.parentChannelID, 0); err != nil {
			detachErr = errors.Join(detachErr, fmt.Errorf("detach thread-created message thread: %w", err))
		} else {
			refs.followupMessage.Thread = 0
			e.sendDetachedThreadMessageUpdate(guildID, e.buildGuildMessageDTO(ctx, *refs.followupMessage, nil))
		}
	}

	if err := e.msg.DeleteThreadCreatedMessageRef(ctx, threadID); err != nil {
		detachErr = errors.Join(detachErr, fmt.Errorf("delete thread-created message ref: %w", err))
	}

	return detachErr
}

func (e *entity) syncThreadCreatedMessage(ctx context.Context, guildID int64, thread *model.Channel, position int, memberIDs []int64) (*threadCreatedMessageUpdate, error) {
	if thread == nil || thread.Type != model.ChannelTypeThread {
		return nil, nil
	}
	e.applyThreadMessageCount(ctx, thread)

	parentChannelID, messageID, err := e.msg.GetThreadCreatedMessageRef(ctx, thread.Id)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	message, err := e.msg.GetMessage(ctx, messageID, parentChannelID)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if model.MessageType(message.Type) != model.MessageTypeThreadCreated {
		return nil, nil
	}

	if message.Content == thread.Name {
		return nil, nil
	}

	if err := e.msg.UpdateMessageContent(ctx, message.Id, parentChannelID, thread.Name); err != nil {
		return nil, err
	}

	threadDTO := channelModelToDTOWithThreadMember(thread, &guildID, position, nil, nil, memberIDs)
	return &threadCreatedMessageUpdate{
		channelID: parentChannelID,
		message: dto.Message{
			Id:                 message.Id,
			ChannelId:          parentChannelID,
			Author:             e.buildThreadMessageAuthor(ctx, message.UserId),
			Content:            thread.Name,
			Position:           guildOptionalInt64(message.Position),
			Attachments:        nil,
			Embeds:             nil,
			Flags:              model.NormalizeMessageFlags(message.Flags),
			Type:               message.Type,
			Reference:          guildOptionalInt64(message.Reference),
			ReferenceChannelId: guildOptionalReferenceChannelID(message.ChannelId, message.ReferenceChannel, message.Reference),
			ThreadId:           guildOptionalInt64(message.Thread),
			Thread:             &threadDTO,
			UpdatedAt:          nil,
		},
	}, nil
}

func (e *entity) sendThreadCreatedMessageUpdate(guildID int64, update *threadCreatedMessageUpdate) {
	if update == nil {
		return
	}

	if err := e.mqt.SendChannelMessage(update.channelID, &mqmsg.UpdateMessage{
		GuildId: &guildID,
		Message: update.message,
	}); err != nil {
		e.log.Error("unable to send thread-created message update",
			"channel_id", update.channelID,
			"message_id", update.message.Id,
			"error", err.Error())
	}
}

// sendCreateChannelEvent sends channel creation message to message queue
func (e *entity) sendCreateChannelEvent(guildId, guildModelId, channelId int64, name string, channelType model.ChannelType, parentId *int64) error {
	if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.CreateChannel{
		GuildId: &guildModelId,
		Channel: dto.Channel{
			Id:        channelId,
			Type:      channelType,
			GuildId:   &guildModelId,
			Name:      name,
			ParentId:  parentId,
			Position:  0,
			Topic:     nil,
			CreatedAt: time.Now(),
		},
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
	}
	return nil
}

// CreateChannel
//
//	@Summary	Create guild channel
//	@Produce	json
//	@Tags		Guild
//	@Param		request		body		CreateGuildChannelRequest	true	"Create channel data"
//	@Param		guild_id	path		int64						true	"Guild ID"	example(2230469276416868352)
//	@Success	201			{object}	string						"Created"
//	@failure	400			{string}	string						"Incorrect request body"
//	@failure	401			{string}	string						"Unauthorized"
//	@failure	500			{string}	string						"Something bad happened"
//	@Router		/guild/{guild_id}/channel [post]
func (e *entity) CreateChannel(c *fiber.Ctx) error {
	var req CreateGuildChannelRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return e.createChannelWithPermissionCheck(c, guildId, user.Id, req.Name, req.Type, req.ParentId, req.Private, req.Position)
}

// DeleteChannel
//
//	@Summary	Delete channel
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64	true	"Guild ID"		example(2230469276416868352)
//	@Param		channel_id	path		int64	true	"Channel ID"	example(2230469276416868352)
//	@Success	200			{object}	string	"Deleted"
//	@failure	400			{string}	string	"Incorrect request body"
//	@failure	401			{string}	string	"Unauthorized"
//	@failure	500			{string}	string	"Something bad happened"
//	@Router		/guild/{guild_id}/channel/{channel_id} [delete]
func (e *entity) DeleteChannel(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return e.deleteChannelWithPermissionCheck(c, guildId, channelId, user.Id)
}

// deleteChannelWithPermissionCheck validates permissions and deletes a channel
func (e *entity) deleteChannelWithPermissionCheck(c *fiber.Ctx, guildId, channelId, userId int64) error {
	isMember, err := e.memb.IsGuildMember(c.UserContext(), guildId, userId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMember)
	}
	if !isMember {
		return fiber.NewError(fiber.StatusForbidden, ErrNotAMember)
	}

	channel, err := e.ch.GetChannel(c.UserContext(), channelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if channel.Type == model.ChannelTypeGuildCategory {
		_, _, _, hasPermission, err := e.perm.ChannelPerm(c.UserContext(), guildId, channelId, userId, permissions.PermServerManageChannels)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if !hasPermission {
			return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
		}
	} else if channel.Type == model.ChannelTypeThread {
		canManage, err := e.canManageThread(c.UserContext(), guildId, &channel, userId)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if !canManage {
			return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
		}
	} else {
		_, _, _, hasPermission, err := e.perm.ChannelPerm(c.UserContext(), guildId, channelId, userId, permissions.PermServerManageChannels)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if !hasPermission {
			return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
		}
	}

	// Delete the channel
	if err := e.gc.RemoveChannel(c.UserContext(), guildId, channelId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Delete channel messages if any exist
	if channel.LastMessage != 0 {
		if err := e.msg.DeleteChannelMessages(c.UserContext(), channelId, channel.LastMessage); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}
	if channel.Type == model.ChannelTypeThread {
		if err := e.detachDeletedThreadMessages(c.UserContext(), guildId, channelId); err != nil {
			slog.Error("unable to detach deleted thread messages", slog.String("error", err.Error()))
		}
		if err := e.tm.RemoveThreadMembers(c.UserContext(), channelId); err != nil {
			slog.Error("unable to delete thread members", slog.String("error", err.Error()))
		}
	}

	// Send delete channel event and clean cached value
	go func() {
		if err := e.sendDeleteChannelEvent(guildId, &channel); err != nil {
			slog.Error("unable to send guild event after channel deletion", slog.String("error", err.Error()))
		}
		if err := e.cache.Delete(context.Background(), fmt.Sprintf("guild:%d:channels", guildId)); err != nil {
			slog.Error("unable to clean cached channels value", slog.String("error", err.Error()))
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

// DeleteCategory
//
//	@Summary	Delete channel category
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64	true	"Guild ID"												example(2230469276416868352)
//	@Param		category_id	path		int64	true	"Category ID (actually a channel with special type)"	example(2230469276416868352)
//	@Success	200			{object}	string	"Deleted"
//	@failure	400			{string}	string	"Incorrect request body"
//	@failure	401			{string}	string	"Unauthorized"
//	@failure	500			{string}	string	"Something bad happened"
//	@Router		/guild/{guild_id}/category/{category_id} [delete]
func (e *entity) DeleteCategory(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	categoryId, err := e.parseCategoryID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return e.deleteCategoryWithPermissionCheck(c, guildId, categoryId, user.Id)
}

// deleteCategoryWithPermissionCheck validates permissions and deletes a category
func (e *entity) deleteCategoryWithPermissionCheck(c *fiber.Ctx, guildId, categoryId, userId int64) error {
	channel, _, _, hasPermission, err := e.perm.ChannelPerm(c.UserContext(), guildId, categoryId, userId, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if !hasPermission || channel.Type != model.ChannelTypeGuildCategory {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	// Delete the category (child channels implicitly lose parent via position derivation)
	if err := e.gc.RemoveChannel(c.UserContext(), guildId, channel.Id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Send delete channel event and clean cached data
	go func() {
		if err := e.sendDeleteChannelEvent(guildId, channel); err != nil {
			slog.Error("unable to send guild event after channel deletion", slog.String("error", err.Error()))
		}
		if err := e.cache.Delete(context.Background(), fmt.Sprintf("guild:%d:channels", guildId)); err != nil {
			slog.Error("unable to clean cached channels value", slog.String("error", err.Error()))
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

// sendDeleteChannelEvent sends channel deletion message to message queue
func (e *entity) sendDeleteChannelEvent(guildId int64, channel *model.Channel) error {
	if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.DeleteChannel{
		GuildId:     &guildId,
		ChannelType: channel.Type,
		ChannelId:   channel.Id,
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
	}
	if channel.Type == model.ChannelTypeThread {
		if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.DeleteThread{
			GuildId:  &guildId,
			ThreadId: channel.Id,
		}); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
		}
	}
	return nil
}

func (e *entity) sendUpdateChannelEvent(guildId int64, channel dto.Channel) error {
	if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.UpdateChannel{
		GuildId: &guildId,
		Channel: channel,
	}); err != nil {
		return err
	}
	if channel.Type == model.ChannelTypeThread {
		if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.UpdateThread{
			GuildId: &guildId,
			Thread:  channel,
		}); err != nil {
			return err
		}
	}
	return nil
}

// PatchChannelOrder
//
//	@Summary	Change channels order
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64							true	"Guild ID"	example(2230469276416868352)
//	@Param		request		body		PatchGuildChannelOrderRequest	true	"Update channel order data"
//	@Success	200			{string}	string							"Ok"
//	@failure	400			{string}	string							"Incorrect request body"
//	@failure	404			{string}	string							"Member not found"
//	@failure	401			{string}	string							"Unauthorized"
//	@failure	406			{string}	string							"Permissions required"
//	@failure	500			{string}	string							"Something bad happened"
//	@Router		/guild/{guild_id}/channel/order [patch]
func (e *entity) PatchChannelOrder(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	var req PatchGuildChannelOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Permission check: user must be able to manage channels in this guild
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	// Ensure we only update channels that belong to this guild
	guildChannels, err := e.gc.GetGuildChannels(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	allowed := make(map[int64]struct{}, len(guildChannels))
	for _, gch := range guildChannels {
		allowed[gch.ChannelId] = struct{}{}
	}

	// Build update list
	updates := make([]model.GuildChannelUpdatePosition, 0, len(req.Channels))
	evt := make([]dto.ChannelOrder, 0, len(req.Channels))
	for _, ch := range req.Channels {
		if _, ok := allowed[ch.Id]; !ok {
			continue
		}
		updates = append(updates, model.GuildChannelUpdatePosition{
			GuildId:   guildId,
			ChannelId: ch.Id,
			Position:  ch.Position,
		})
		evt = append(evt, dto.ChannelOrder{Id: ch.Id, Position: ch.Position})
	}

	if len(updates) == 0 {
		// Nothing to update
		return c.SendStatus(fiber.StatusOK)
	}

	// Apply positions
	if err := e.gc.SetGuildChannelPosition(c.UserContext(), updates); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Notify clients about the new order and clean cached data
	go func() {
		if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.UpdateChannelList{
			GuildId:  &guildId,
			Channels: evt,
		}); err != nil {
			slog.Error("unable to send guild update event after channel reorder", slog.String("error", err.Error()))
		}
		if err := e.cache.Delete(context.Background(), fmt.Sprintf("guild:%d:channels", guildId)); err != nil {
			slog.Error("unable to clean cached channels value", slog.String("error", err.Error()))
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}

// PatchChannel
//
//	@Summary	Change channels data
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64						true	"Guild ID"		example(2230469276416868352)
//	@Param		channel_id	path		int64						true	"Channel ID"	example(2230469276416868352)
//	@Param		req			body		PatchGuildChannelRequest	true	"Request body"
//	@Success	200			{object}	dto.Channel					"Ok"
//	@failure	400			{string}	string						"Incorrect request body"
//	@failure	404			{string}	string						"Member not found"
//	@failure	401			{string}	string						"Unauthorized"
//	@failure	406			{string}	string						"Permissions required"
//	@failure	500			{string}	string						"Something bad happened"
//	@Router		/guild/{guild_id}/channel/{channel_id} [patch]
func (e *entity) PatchChannel(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}

	var req PatchGuildChannelRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Permission check: user must be able to manage channels in this guild
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	isMember, err := e.memb.IsGuildMember(c.UserContext(), guildId, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMember)
	}
	if !isMember {
		return fiber.NewError(fiber.StatusForbidden, ErrNotAMember)
	}

	guildChannel, err := e.gc.GetGuildChannel(c.UserContext(), guildId, channelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}

	channel, err := e.ch.GetChannel(c.UserContext(), guildChannel.ChannelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}
	if err := e.validatePatchChannelRequest(&channel, &req); err != nil {
		return err
	}

	if channel.Type == model.ChannelTypeThread {
		canManage, err := e.canManageThread(c.UserContext(), guildId, &channel, user.Id)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if !canManage {
			return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
		}
	} else {
		_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermServerManageChannels)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if !hasPermission {
			return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
		}
	}

	upd, err := e.ch.UpdateChannel(c.UserContext(), guildChannel.ChannelId, nil, req.Private, req.Name, req.Topic, req.Closed)
	if err != nil {
		return fiber.NewError(fiber.StatusNotModified, ErrUnableToUpdateChannel)
	}

	var threadMember *dto.ThreadMember
	var threadMemberIDs []int64
	if upd.Type == model.ChannelTypeThread {
		threadMember, threadMemberIDs, err = e.currentUserThreadMember(c.UserContext(), user.Id, &upd)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
		}
		e.applyThreadMessageCount(c.UserContext(), &upd)
	}

	var threadCreatedUpdate *threadCreatedMessageUpdate
	if upd.Type == model.ChannelTypeThread && req.Name != nil {
		threadCreatedUpdate, err = e.syncThreadCreatedMessage(c.UserContext(), guildId, &upd, guildChannel.Position, threadMemberIDs)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "unable to update thread creation message")
		}
	}

	resp := channelModelToDTOWithThreadMember(&upd, &guildId, guildChannel.Position, nil, threadMember, threadMemberIDs)

	// Notify clients about the channel update and clean cached data
	go func() {
		if err := e.sendUpdateChannelEvent(guildId, resp); err != nil {
			slog.Error("unable to send guild update event after channel update", slog.String("error", err.Error()))
		}
		e.sendThreadCreatedMessageUpdate(guildId, threadCreatedUpdate)
		if err := e.cache.Delete(context.Background(), fmt.Sprintf("guild:%d:channels", guildId)); err != nil {
			slog.Error("unable to clean cached channels value", slog.String("error", err.Error()))
		}
	}()

	return c.JSON(resp)
}

// GetMembers
//
//	@Summary	Get guild members
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64		true	"Guild ID"	example(2230469276416868352)
//	@Success	200			{array}		dto.Member	"Ok"
//	@failure	400			{string}	string		"Incorrect request body"
//	@failure	401			{string}	string		"Unauthorized"
//	@failure	500			{string}	string		"Something bad happened"
//	@Router		/guild/{guild_id}/members [get]
func (e *entity) GetMembers(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	isMember, err := e.memb.IsGuildMember(c.UserContext(), guildId, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMemberToken)
	}
	if !isMember {
		return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
	}

	members, err := e.memb.GetGuildMembers(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMembers)
	}

	var memberIds = make([]int64, len(members))
	for i, m := range members {
		memberIds[i] = m.UserId
	}

	dscs, err := e.disc.GetDiscriminatorsByUserIDs(c.UserContext(), memberIds)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetDiscriminators)
	}

	users, err := e.user.GetUsersList(c.UserContext(), memberIds)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUsers)
	}

	roles, err := e.ur.GetUsersRolesByGuild(c.UserContext(), guildId, memberIds)
	if err != nil {
		slog.Error("unable to get users roles", slog.String("error", err.Error()))
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUsersRoles)
	}

	// Build avatar data map (cached)
	avData := make(map[int64]*dto.AvatarData, len(users))
	for _, u := range users {
		if u.Avatar != nil {
			if ad, err := e.getAvatarDataCached(c.UserContext(), u.Id, *u.Avatar); err == nil && ad != nil {
				avData[u.Id] = ad
			}
		}
	}
	return c.JSON(membersToDTO(members, users, roles, dscs, avData))
}

const avatarCacheTTLSeconds = 3600 // 1 hour

func (e *entity) getAvatarDataCached(ctx context.Context, userId, avatarId int64) (*dto.AvatarData, error) {
	key := fmt.Sprintf("avatars:%d:%d", userId, avatarId)
	var ad dto.AvatarData
	if err := e.cache.GetJSON(ctx, key, &ad); err == nil && ad.URL != "" {
		return &ad, nil
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
	_ = e.cache.SetTimedJSON(ctx, key, ad, avatarCacheTTLSeconds)
	return &ad, nil
}
