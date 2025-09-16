package guild

import (
	"database/sql"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
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

// validateGuildAccess validates user access to guild and returns guild context
func (e *entity) validateGuildAccess(c *fiber.Ctx, guildId int64) (*guildContext, error) {
	user, err := helper.GetUser(c)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	member, err := e.memb.GetMember(c.UserContext(), user.Id, guildId)
	if err != nil {
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

	roles, err := e.role.GetRolesBulk(c.UserContext(), roleIds)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	roleMap := make(map[int64]*model.Role)
	for i, role := range roles {
		roleMap[role.Id] = &roles[i]
	}

	return roleMap, nil
}

// checkChannelPermissions validates if user can view a private channel
func (e *entity) checkChannelPermissions(c *fiber.Ctx, channel *model.Channel, guild *model.Guild, user *helper.JWTUser, roles map[int64]*model.Role) (bool, error) {
	if !channel.Private || guild.OwnerId == user.Id {
		return true, nil
	}

	// Check user-specific channel permissions
	userPerm, err := e.uperm.GetUserChannelPermission(c.UserContext(), channel.Id, user.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if errors.Is(err, sql.ErrNoRows) {
		// Check role-based permissions
		rolePerms, err := e.rperm.GetChannelRolePermissions(c.UserContext(), channel.Id)
		if err != nil {
			return false, fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		var combinedRole int64
		for _, rp := range rolePerms {
			if role, exists := roles[rp.RoleId]; exists {
				combinedRole = permissions.AddRoles(combinedRole, role.Permissions)
			}
		}

		return permissions.CheckPermissions(combinedRole, permissions.PermServerViewChannels), nil
	}

	// Calculate final permissions
	basePerms := *channel.Permissions
	if channel.Permissions == nil {
		basePerms = guild.Permissions
	}
	finalPerms := permissions.SubtractRoles(permissions.AddRoles(basePerms, userPerm.Accept), userPerm.Deny)

	return permissions.CheckPermissions(finalPerms, permissions.PermServerViewChannels), nil
}

// createDefaultChannels creates default text category and general channel for new guild
func (e *entity) createDefaultChannels(c *fiber.Ctx, guildId int64, isPublic bool) error {
	categoryId := idgen.Next()
	channelId := idgen.Next()

	// Create text category
	if err := e.ch.CreateChannel(c.UserContext(), categoryId, "text", model.ChannelTypeGuildCategory, nil, nil, isPublic); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Create general channel
	if err := e.ch.CreateChannel(c.UserContext(), channelId, "general", model.ChannelTypeGuild, &categoryId, nil, isPublic); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Add channels to guild
	if err := e.gc.AddChannel(c.UserContext(), guildId, categoryId, 0); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if err := e.gc.AddChannel(c.UserContext(), guildId, channelId, 0); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return nil
}

// Get
//
//	@Summary	Get guild
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64		true	"Guild id"
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

	return c.JSON(buildGuildDTO(guildCtx.Guild))
}

// GetChannels
//
//	@Summary	Get guild channels
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64		true	"Guild id"
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

	roles, err := e.getUserRoles(c, guildCtx.Guild.Id, guildCtx.User.Id)
	if err != nil {
		return err
	}

	return e.fetchAndFilterChannels(c, guildCtx, roles)
}

// fetchAndFilterChannels retrieves guild channels and filters based on permissions
func (e *entity) fetchAndFilterChannels(c *fiber.Ctx, guildCtx *guildContext, roles map[int64]*model.Role) error {
	guildChannels, err := e.gc.GetGuildChannels(c.UserContext(), guildCtx.Guild.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var channelIds []int64
	channelMap := make(map[int64]*model.GuildChannel)
	for i, gch := range guildChannels {
		channelMap[gch.ChannelId] = &guildChannels[i]
		channelIds = append(channelIds, gch.ChannelId)
	}

	channels, err := e.ch.GetChannelsBulk(c.UserContext(), channelIds)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var visibleChannels []dto.Channel
	for _, ch := range channels {
		if guildChannel, exists := channelMap[ch.Id]; exists {
			// Set default permissions if not set
			if ch.Permissions == nil {
				ch.Permissions = &guildCtx.Guild.Permissions
			}

			// Check if user can view this channel
			canView, err := e.checkChannelPermissions(c, &ch, guildCtx.Guild, guildCtx.User, roles)
			if err != nil {
				return err
			}

			if canView {
				visibleChannels = append(visibleChannels, channelModelToDTO(&ch, &guildCtx.Guild.Id, guildChannel.Position))
			}
		}
	}

	return c.JSON(visibleChannels)
}

// GetChannel
//
//	@Summary	Get guild channel
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64		true	"Guild id"
//	@Param		channel_id	path		int64		true	"Channel id"
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

// fetchSingleChannel retrieves and validates access to a specific channel
func (e *entity) fetchSingleChannel(c *fiber.Ctx, guildCtx *guildContext, channelId int64, roles map[int64]*model.Role) error {
	guildChannel, err := e.gc.GetGuildChannel(c.UserContext(), guildCtx.Guild.Id, channelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	channel, err := e.ch.GetChannel(c.UserContext(), guildChannel.ChannelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Set default permissions if not set
	if channel.Permissions == nil {
		channel.Permissions = &guildCtx.Guild.Permissions
	}

	// Check if user can view this channel
	canView, err := e.checkChannelPermissions(c, &channel, guildCtx.Guild, guildCtx.User, roles)
	if err != nil {
		return err
	}

	if !canView {
		return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
	}

	return c.JSON(channelModelToDTO(&channel, &guildCtx.Guild.Id, guildChannel.Position))
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
	if err := e.createDefaultChannels(c, guildId, req.Public); err != nil {
		return err
	}

	// Add creator as member
	if err := e.memb.AddMember(c.UserContext(), user.Id, guildId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(dto.Guild{
		Id:          guildId,
		Name:        req.Name,
		Icon:        req.IconId,
		Owner:       user.Id,
		Public:      req.Public,
		Permissions: permissions.DefaultPermissions,
	})
}

// setGuildIconIfProvided sets guild icon if icon ID is provided and valid
func (e *entity) setGuildIconIfProvided(c *fiber.Ctx, guildId int64, iconId *int64) error {
	if iconId == nil {
		return nil
	}

	icon, err := e.icon.GetIcon(c.UserContext(), *iconId)
	if err != nil {
		// Icon not found or error - continue without setting icon
		return nil
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
//	@Param		guild_id	path		int64				true	"Guild ID"
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
//	@Param		guild_id	path		int64								true	"Guild ID"
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

	return e.createChannelWithPermissionCheck(c, guildId, user.Id, req.Name, model.ChannelTypeGuildCategory, nil, req.Private)
}

// createChannelWithPermissionCheck validates permissions and creates a channel
func (e *entity) createChannelWithPermissionCheck(c *fiber.Ctx, guildId, userId int64, name string, channelType model.ChannelType, parentId *int64, isPrivate bool) error {
	guild, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, userId, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	channelId := idgen.Next()

	// Create the channel
	if err := e.ch.CreateChannel(c.UserContext(), channelId, name, channelType, parentId, nil, isPrivate); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
	}

	// Add channel to guild
	if err := e.gc.AddChannel(c.UserContext(), guild.Id, channelId, 0); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
	}

	// Send create channel event
	if err := e.sendCreateChannelEvent(guildId, guild.Id, channelId, name, channelType, parentId); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
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
//	@Param		guild_id	path		int64						true	"Guild ID"
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

	return e.createChannelWithPermissionCheck(c, guildId, user.Id, req.Name, req.Type, req.ParentId, req.Private)
}

// DeleteChannel
//
//	@Summary	Delete channel
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64	true	"Guild ID"
//	@Param		channel_id	path		int64	true	"Channel ID"
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
	channel, _, _, hasPermission, err := e.perm.ChannelPerm(c.UserContext(), guildId, channelId, userId, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if !hasPermission || channel.Type == model.ChannelTypeGuildCategory {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	// Delete the channel
	if err := e.ch.DeleteChannel(c.UserContext(), channelId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Delete channel messages if any exist
	if channel.LastMessage != 0 {
		if err := e.msg.DeleteChannelMessages(c.UserContext(), channelId, channel.LastMessage); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	// Send delete channel event
	if err := e.sendDeleteChannelEvent(guildId, channel); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteCategory
//
//	@Summary	Delete channel category
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64	true	"Guild ID"
//	@Param		category_id	path		int64	true	"Category ID (actually a channel with special type)"
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
	channel, _, guild, hasPermission, err := e.perm.ChannelPerm(c.UserContext(), guildId, categoryId, userId, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if !hasPermission || channel.Type != model.ChannelTypeGuildCategory {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	// Handle child channels - move them out of the category
	if err := e.handleCategoryChildChannels(c, guild.Id, categoryId); err != nil {
		return err
	}

	// Delete the category
	if err := e.ch.DeleteChannel(c.UserContext(), channel.Id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Send delete channel event
	if err := e.sendDeleteChannelEvent(guildId, channel); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// handleCategoryChildChannels moves child channels out of a category before deletion
func (e *entity) handleCategoryChildChannels(c *fiber.Ctx, guildId, categoryId int64) error {
	guildChannels, err := e.gc.GetGuildChannels(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var channelIds []int64
	for _, gch := range guildChannels {
		channelIds = append(channelIds, gch.ChannelId)
	}

	channels, err := e.ch.GetChannelsBulk(c.UserContext(), channelIds)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var childChannelIds []int64
	for _, ch := range channels {
		if ch.ParentID != nil && *ch.ParentID == categoryId {
			childChannelIds = append(childChannelIds, ch.Id)
		}
	}

	if len(childChannelIds) > 0 {
		// Remove parent from child channels
		if err := e.ch.SetChannelParentBulk(c.UserContext(), childChannelIds, nil); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		// Reset positions for child channels
		if err := e.gc.ResetGuildChannelPositionBulk(c.UserContext(), childChannelIds, guildId); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	return nil
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
	return nil
}

// GetMemberRoles
//
//	@Summary	Get member roles
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64		true	"Guild ID"
//	@Param		user_id		path		int64		true	"User ID"
//	@Success	200			{array}		dto.Role	"List of user roles"
//	@failure	400			{string}	string		"Incorrect request body"
//	@failure	404			{string}	string		"Member not found"
//	@failure	401			{string}	string		"Unauthorized"
//	@failure	500			{string}	string		"Something bad happened"
//	@Router		/guild/{guild_id}/member/{user_id}/roles [get]
func (e *entity) GetMemberRoles(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	memberId, err := e.parseUserID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	isUserMember, err := e.memb.IsGuildMember(c.UserContext(), guildId, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUserToken)
	}
	if !isUserMember {
		return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
	}

	isGuildMember, err := e.memb.IsGuildMember(c.UserContext(), guildId, memberId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMemberToken)
	}
	if !isGuildMember {
		return fiber.NewError(fiber.StatusNotFound, ErrNotAMember)
	}

	memberRoleIds, err := e.ur.GetUserRoles(c.UserContext(), guildId, memberId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}

	roleIds := make([]int64, len(memberRoleIds))
	for _, role := range memberRoleIds {
		roleIds = append(roleIds, role.RoleId)
	}

	roles, err := e.role.GetRolesBulk(c.UserContext(), roleIds)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}

	return c.JSON(roleModelToDTOMany(roles))
}

// PatchChannelOrder
//
//	@Summary	Change channels order
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64							true	"Guild ID"
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

	// Notify clients about the new order
	if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.UpdateChannelList{
		GuildId:  &guildId,
		Channels: evt,
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
	}

	return c.SendStatus(fiber.StatusOK)
}

// PatchChannel
//
//	@Summary	Change channels data
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64						true	"Guild ID"
//	@Param		channel_id	path		int64						true	"Channel ID"
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

	_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	if req.ParentId != nil && *req.ParentId == channelId {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToSetParentAsSelf)
	}

	guildChannel, err := e.gc.GetGuildChannel(c.UserContext(), guildId, channelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}

	ch, err := e.ch.GetChannel(c.UserContext(), guildChannel.ChannelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}

	if req.ParentId != nil && ch.Type == model.ChannelTypeGuildCategory {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToSetParentForCategory)
	}

	upd, err := e.ch.UpdateChannel(c.UserContext(), guildChannel.ChannelId, req.ParentId, req.Private, req.Name, req.Topic)
	if err != nil {
		return fiber.NewError(fiber.StatusNotModified, ErrUnableToUpdateChannel)
	}

	resp := channelModelToDTO(&upd, &guildId, guildChannel.Position)

	// Notify clients about the channel update
	if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.UpdateChannel{
		GuildId: &guildId,
		Channel: resp,
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
	}

	return c.JSON(resp)
}
