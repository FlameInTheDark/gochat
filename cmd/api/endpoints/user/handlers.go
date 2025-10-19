package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
)

// GetUser
//
//	@Summary	Get user
//	@Produce	json
//	@Tags		User
//	@Param		user_id	path		string		true	"User ID or 'me'"
//	@Success	200		{object}	dto.User	"User data"
//	@failure	400		{string}	string		"Bad request"
//	@failure	404		{string}	string		"User not found"
//	@failure	500		{string}	string		"Internal server error"
//	@Router		/user/{user_id} [get]
func (e *entity) GetUser(c *fiber.Ctx) error {
	// Parse user ID from path parameter
	userId, err := e.parseUserIdParam(c, "user_id")
	if err != nil {
		return err
	}

	// Fetch user data concurrently
	userDTO, err := e.fetchUserWithDiscriminatorCtx(c.UserContext(), userId)
	if err != nil {
		return err
	}

	return c.JSON(userDTO)
}

// parseUserIdParam handles parsing user ID from path parameters, supporting "me"
func (e *entity) parseUserIdParam(c *fiber.Ctx, paramName string) (int64, error) {
	idStr := c.Params(paramName)

	if idStr == "me" {
		user, err := helper.GetUser(c)
		if err != nil {
			return 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
		}
		return user.Id, nil
	}

	userId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid user ID format")
	}

	return userId, nil
}

// fetchUserWithDiscriminator fetches user data and discriminator concurrently
func (e *entity) fetchUserWithDiscriminatorCtx(ctx context.Context, userId int64) (dto.User, error) {
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

	// Fetch user data
	go func() {
		user, err := e.user.GetUserById(ctx, userId)
		userCh <- userResult{&user, err}
	}()

	// Fetch discriminator
	go func() {
		disc, err := e.disc.GetDiscriminatorByUserId(ctx, userId)
		discCh <- discResult{&disc, err}
	}()

	// Collect results
	userRes := <-userCh
	discRes := <-discCh

	// Check for errors
	if userRes.err != nil {
		return dto.User{}, helper.HttpDbError(userRes.err, ErrUnableToGetUser)
	}
	if discRes.err != nil {
		return dto.User{}, helper.HttpDbError(discRes.err, ErrUnableToGetDiscriminator)
	}

	userDTO := modelToUser(*userRes.user)
	if userRes.user.Avatar != nil {
		if ad, err := e.getAvatarDataCached(ctx, userRes.user.Id, *userRes.user.Avatar); err == nil && ad != nil {
			userDTO.Avatar = ad
		}
	}
	userDTO.Discriminator = discRes.disc.Discriminator

	return userDTO, nil
}

// ModifyUser
//
//	@Summary	Get user
//	@Produce	json
//	@Tags		User
//	@Param		request	body		ModifyUserRequest	true	"Modify user data"
//	@Success	200		{string}	string				"Ok"
//	@failure	400		{string}	string				"Incorrect ID"
//	@failure	404		{string}	string				"User not found"
//	@failure	500		{string}	string				"Something bad happened"
//	@Router		/user/me [patch]
func (e *entity) ModifyUser(c *fiber.Ctx) error {
	var req ModifyUserRequest
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	err = e.user.ModifyUser(c.UserContext(), user.Id, req.Name, req.Avatar)
	if err := helper.HttpDbError(err, ErrUnableToModifyUser); err != nil {
		return err
	}
	// Emit user update event with fresh data (best-effort). Do not use fiber.Ctx from goroutine.
	go func() {
		dtoUser, ferr := e.fetchUserWithDiscriminatorCtx(context.Background(), user.Id)
		if ferr != nil {
			slog.Error("unable to build updated user dto", slog.String("error", ferr.Error()))
			return
		}
		if err := e.mqt.SendUserUpdate(user.Id, &mqmsg.UpdateUser{User: dtoUser}); err != nil {
			slog.Error("unable to send user update event", slog.String("error", err.Error()))
		}
	}()
	return c.SendStatus(fiber.StatusOK)
}

// GetUserGuilds
//
//	@Summary	Get user guilds
//	@Produce	json
//	@Tags		User
//	@Success	200	{array}		dto.Guild	"Guilds list"
//	@failure	400	{string}	string		"Bad request"
//	@failure	404	{string}	string		"User not found"
//	@failure	500	{string}	string		"Internal server error"
//	@Router		/user/me/guilds [get]
func (e *entity) GetUserGuilds(c *fiber.Ctx) error {
	// Get authenticated user
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	// Fetch user's guilds
	guilds, err := e.fetchUserGuilds(c, user.Id)
	if err != nil {
		return err
	}

	return c.JSON(guilds)
}

// fetchUserGuilds retrieves all guilds for a user with proper error handling
func (e *entity) fetchUserGuilds(c *fiber.Ctx, userId int64) ([]dto.Guild, error) {
	// Get user's guild memberships
	memberships, err := e.member.GetUserGuilds(c.UserContext(), userId)
	if err != nil {
		return nil, helper.HttpDbError(err, ErrUnableToGetUserGuilds)
	}

	// Handle case where user has no guilds
	if len(memberships) == 0 {
		return []dto.Guild{}, nil
	}

	// Extract guild IDs
	guildIds := make([]int64, len(memberships))
	for i, membership := range memberships {
		guildIds[i] = membership.GuildId
	}

	// Fetch guild details
	guilds, err := e.guild.GetGuildsList(c.UserContext(), guildIds)
	if err != nil {
		e.log.Error("failed to fetch guild details",
			"user_id", userId,
			"guild_count", len(guildIds),
			"error", err.Error())
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUserGuilds)
	}

	// Convert to DTOs with icon metadata
	return e.guildModelToGuildMany(c, guilds), nil
}

// GetMyGuildMember
//
//	@Summary	Get user guild member
//	@Produce	json
//	@Tags		User
//	@Param		guild_id	path		int64		true	"Guild id"	example(2230469276416868352)
//	@Success	200			{object}	dto.Member	"Guild member"
//	@failure	400			{string}	string		"Bad request"
//	@failure	404			{string}	string		"Member not found"
//	@failure	500			{string}	string		"Internal server error"
//	@Router		/user/me/guilds/{guild_id}/member [get]
func (e *entity) GetMyGuildMember(c *fiber.Ctx) error {
	// Parse and validate request
	user, guildId, err := e.parseGuildMemberRequest(c)
	if err != nil {
		return err
	}

	// Fetch member data concurrently
	member, err := e.fetchGuildMemberData(c, user.Id, guildId)
	if err != nil {
		return err
	}

	return c.JSON(member)
}

// parseGuildMemberRequest handles request parsing and user authentication
func (e *entity) parseGuildMemberRequest(c *fiber.Ctx) (*helper.JWTUser, int64, error) {
	user, err := helper.GetUser(c)
	if err != nil {
		return nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	guildIdStr := c.Params("guild_id")
	guildId, err := strconv.ParseInt(guildIdStr, 10, 64)
	if err != nil {
		return nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrBadRequest)
	}

	return user, guildId, nil
}

// fetchGuildMemberData fetches all member-related data concurrently
func (e *entity) fetchGuildMemberData(c *fiber.Ctx, userId, guildId int64) (dto.Member, error) {
	// Use channels for concurrent data fetching
	type memberResult struct {
		member *model.Member
		err    error
	}
	type userResult struct {
		user *model.User
		err  error
	}
	type discResult struct {
		disc *model.Discriminator
		err  error
	}
	type rolesResult struct {
		roles []model.UserRole
		err   error
	}

	memberCh := make(chan memberResult, 1)
	userCh := make(chan userResult, 1)
	discCh := make(chan discResult, 1)
	rolesCh := make(chan rolesResult, 1)

	// Fetch member data
	go func() {
		member, err := e.member.GetMember(c.UserContext(), userId, guildId)
		memberCh <- memberResult{&member, err}
	}()

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

	// Fetch user roles
	go func() {
		roles, err := e.urole.GetUserRoles(c.UserContext(), guildId, userId)
		rolesCh <- rolesResult{roles, err}
	}()

	// Collect results
	memberRes := <-memberCh
	userRes := <-userCh
	discRes := <-discCh
	rolesRes := <-rolesCh

	// Check for errors
	if memberRes.err != nil {
		return dto.Member{}, helper.HttpDbError(memberRes.err, ErrUnableToGetMember)
	}
	if userRes.err != nil {
		return dto.Member{}, helper.HttpDbError(userRes.err, ErrUnableToGetUser)
	}
	if discRes.err != nil {
		return dto.Member{}, helper.HttpDbError(discRes.err, ErrUnableToGetDiscriminator)
	}
	if rolesRes.err != nil {
		return dto.Member{}, helper.HttpDbError(rolesRes.err, ErrUnableToGetRoles)
	}

	// Extract role IDs
	roleIds := make([]int64, len(rolesRes.roles))
	for i, role := range rolesRes.roles {
		roleIds[i] = role.RoleId
	}

	// Build user DTO with possible avatar data
	userDTO := dto.User{
		Id:            userRes.user.Id,
		Name:          userRes.user.Name,
		Discriminator: discRes.disc.Discriminator,
	}
	// Prefer member avatar over user avatar
	if memberRes.member.Avatar != nil {
		if ad, err := e.getAvatarDataCached(c.UserContext(), userRes.user.Id, *memberRes.member.Avatar); err == nil && ad != nil {
			userDTO.Avatar = ad
		}
	} else if userRes.user.Avatar != nil {
		if ad, err := e.getAvatarDataCached(c.UserContext(), userRes.user.Id, *userRes.user.Avatar); err == nil && ad != nil {
			userDTO.Avatar = ad
		}
	}

	// Build and return member DTO
	return dto.Member{
		User:     userDTO,
		Username: memberRes.member.Username,
		Avatar:   memberRes.member.Avatar,
		JoinAt:   memberRes.member.JoinAt,
		Roles:    roleIds,
	}, nil
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
		Id:          av.Id,
		URL:         *av.URL,
		ContentType: av.ContentType,
		Width:       av.Width,
		Height:      av.Height,
		Size:        av.FileSize,
	}
	_ = e.cache.SetTimedJSON(ctx, key, ad, avatarCacheTTLSeconds)
	return &ad, nil
}

// LeaveGuild
//
//	@Summary	Leave guild
//	@Produce	json
//	@Tags		User
//	@Param		guild_id	path		string	true	"Guild id"	example(2230469276416868352)
//	@Success	200			{string}	string	"ok"
//	@failure	400			{string}	string	"Bad request"
//	@failure	404			{string}	string	"Guild not found"
//	@failure	406			{string}	string	"Cannot leave own guild"
//	@failure	500			{string}	string	"Internal server error"
//	@Router		/user/me/guilds/{guild_id} [delete]
func (e *entity) LeaveGuild(c *fiber.Ctx) error {
	// Parse and validate request
	user, guildId, err := e.parseLeaveGuildRequest(c)
	if err != nil {
		return err
	}

	// Validate guild exists and user can leave it
	if err := e.validateGuildLeavePermission(c, user.Id, guildId); err != nil {
		return err
	}

	// Remove user from guild
	if err := e.member.RemoveMember(c.UserContext(), user.Id, guildId); err != nil {
		return helper.HttpDbError(err, ErrUnableToRemoveMember)
	}

	go e.mqt.SendGuildUpdate(guildId, &mqmsg.RemoveGuildMember{GuildId: guildId, UserId: user.Id})

	return c.SendStatus(fiber.StatusOK)
}

// parseLeaveGuildRequest handles request parsing and user authentication
func (e *entity) parseLeaveGuildRequest(c *fiber.Ctx) (*helper.JWTUser, int64, error) {
	user, err := helper.GetUser(c)
	if err != nil {
		return nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	guildIdStr := c.Params("guild_id")
	guildId, err := strconv.ParseInt(guildIdStr, 10, 64)
	if err != nil {
		return nil, 0, fiber.NewError(fiber.StatusBadRequest, "invalid guild ID format")
	}

	return user, guildId, nil
}

// validateGuildLeavePermission checks if user can leave the guild
func (e *entity) validateGuildLeavePermission(c *fiber.Ctx, userId, guildId int64) error {
	guild, err := e.guild.GetGuildById(c.UserContext(), guildId)
	if err != nil {
		return helper.HttpDbError(err, ErrUnableToGetGuildByID)
	}

	// Prevent guild owner from leaving their own guild
	if guild.OwnerId == userId {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrUnableToLeaveOwnServer)
	}

	return nil
}

// CreateDM
//
//	@Summary	Create DM channel
//	@Produce	json
//	@Tags		User
//	@Param		request	body		CreateDMRequest	true	"Recipient data"
//	@Success	200		{object}	dto.Channel		"Created DM channel"
//	@failure	400		{string}	string			"Bad request"
//	@failure	404		{string}	string			"User not found"
//	@failure	500		{string}	string			"Internal server error"
//	@Router		/user/me/channels [post]
func (e *entity) CreateDM(c *fiber.Ctx) error {
	// Parse and validate request
	req, user, err := e.parseDMRequest(c)
	if err != nil {
		return err
	}

	// Validate recipient exists
	recipient, err := e.validateRecipient(c, req.RecipientId)
	if err != nil {
		return err
	}

	// Check if DM channel already exists
	existingChannel, err := e.findExistingDMChannel(c, user.Id, recipient.Id)
	if err != nil {
		return err
	}
	if existingChannel != nil {
		return c.JSON(*existingChannel)
	}

	// Create new DM channel
	newChannel, err := e.createNewDMChannel(c, user.Id, req.RecipientId)
	if err != nil {
		return err
	}

	return c.JSON(newChannel)
}

// parseDMRequest handles request parsing and user authentication
func (e *entity) parseDMRequest(c *fiber.Ctx) (*CreateDMRequest, *helper.JWTUser, error) {
	var req CreateDMRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}

	if err := req.Validate(); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return &req, user, nil
}

// validateRecipient ensures the recipient user exists
func (e *entity) validateRecipient(c *fiber.Ctx, recipientId int64) (*model.User, error) {
	recipient, err := e.user.GetUserById(c.UserContext(), recipientId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fiber.NewError(fiber.StatusNotFound, "recipient user not found")
		}
		return nil, helper.HttpDbError(err, ErrUnableToGetUser)
	}
	return &recipient, nil
}

// findExistingDMChannel checks if a DM channel already exists between users
func (e *entity) findExistingDMChannel(c *fiber.Ctx, userId, recipientId int64) (*dto.Channel, error) {
	dmChannel, err := e.dm.GetDmChannel(c.UserContext(), userId, recipientId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // No existing channel
	}
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetDMChannel)
	}

	// Get the actual channel details
	channel, err := e.ch.GetChannel(c.UserContext(), dmChannel.ChannelId)
	if err != nil {
		return nil, helper.HttpDbError(err, ErrUnableToGetChannel)
	}

	// Convert to DTO and include DM participant id
	channelDTO := dto.Channel{
		Id:            channel.Id,
		Type:          channel.Type,
		GuildId:       nil,
		ParticipantId: &recipientId,
		Name:          channel.Name,
		ParentId:      channel.ParentID,
		Permissions:   channel.Permissions,
		CreatedAt:     channel.CreatedAt,
	}

	return &channelDTO, nil
}

// createNewDMChannel creates a new DM channel with proper cleanup
func (e *entity) createNewDMChannel(c *fiber.Ctx, userId, recipientId int64) (dto.Channel, error) {
	channelId := idgen.Next()

	// Create the channel
	if err := e.ch.CreateChannel(c.UserContext(), channelId, "", model.ChannelTypeDM, nil, nil, false); err != nil {
		return dto.Channel{}, helper.HttpDbError(err, ErrUnableToCreateChannel)
	}

	// Create DM channel association
	if err := e.dm.CreateDmChannel(c.UserContext(), userId, recipientId, channelId); err != nil {
		// Cleanup: delete the channel if DM association creation fails
		if cleanupErr := e.ch.DeleteChannel(c.UserContext(), channelId); cleanupErr != nil {
			e.log.Error("failed to cleanup channel after DM creation failure",
				"channel_id", channelId,
				"user_id", userId,
				"recipient_id", recipientId,
				"cleanup_error", cleanupErr.Error(),
				"original_error", err.Error())
		}
		return dto.Channel{}, helper.HttpDbError(err, ErrUnableToCreateDMChannel)
	}

	// Return the successfully created channel, include participant id
	return dto.Channel{
		Id:            channelId,
		Type:          model.ChannelTypeDM,
		GuildId:       nil,
		ParticipantId: &recipientId,
		Name:          "",
		ParentId:      nil,
		CreatedAt:     time.Now(),
	}, nil
}

// CreateGroupDM
//
//	@Summary	Create group DM channel
//	@Produce	json
//	@Tags		User
//	@Param		request	body		CreateDMManyRequest	true	"Group DM data"
//	@Success	200		{object}	dto.Channel			"Created group DM channel"
//	@failure	400		{string}	string				"Bad request"
//	@failure	403		{string}	string				"Forbidden"
//	@failure	404		{string}	string				"Not found"
//	@failure	500		{string}	string				"Internal server error"
//	@Router		/user/me/channels/group [post]
func (e *entity) CreateGroupDM(c *fiber.Ctx) error {
	// Parse and validate request
	req, user, err := e.parseGroupDMRequest(c)
	if err != nil {
		return err
	}

	// Validate recipient users exist
	if err := e.validateRecipients(c, req.RecipientsId); err != nil {
		return err
	}

	// Handle existing group DM vs new group DM
	if req.ChannelId != nil {
		return e.addUsersToExistingGroupDM(c, *req.ChannelId, user.Id, req.RecipientsId)
	}

	return e.createNewGroupDM(c, user.Id, req.RecipientsId)
}

// parseGroupDMRequest handles request parsing and user authentication
func (e *entity) parseGroupDMRequest(c *fiber.Ctx) (*CreateDMManyRequest, *helper.JWTUser, error) {
	var req CreateDMManyRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}

	if err := req.Validate(); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	return &req, user, nil
}

// validateRecipients ensures all recipient users exist and are valid
func (e *entity) validateRecipients(c *fiber.Ctx, recipientIds []int64) error {
	for _, recipientId := range recipientIds {
		if _, err := e.user.GetUserById(c.UserContext(), recipientId); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fiber.NewError(fiber.StatusNotFound, "recipient user not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUser)
		}
	}
	return nil
}

// addUsersToExistingGroupDM handles adding users to an existing group DM
func (e *entity) addUsersToExistingGroupDM(c *fiber.Ctx, channelId, userId int64, recipientIds []int64) error {
	// Verify user permissions and get channel in one operation
	channel, err := e.validateGroupDMAccess(c, channelId, userId)
	if err != nil {
		return err
	}

	// Add new participants
	if err := e.gdm.JoinGroupDmChannelMany(c.UserContext(), channelId, recipientIds); err != nil {
		return helper.HttpDbError(err, ErrUnableToJoingGroupDmChannel)
	}

	return c.JSON(e.channelToDTO(channel))
}

// validateGroupDMAccess checks if user can access the group DM and returns the channel
func (e *entity) validateGroupDMAccess(c *fiber.Ctx, channelId, userId int64) (*model.Channel, error) {
	// Get channel first to ensure it exists
	channel, err := e.ch.GetChannel(c.UserContext(), channelId)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, ErrUnableToGetChannel)
	}

	// Verify it's actually a group DM
	if channel.Type != model.ChannelTypeGroupDM {
		return nil, fiber.NewError(fiber.StatusBadRequest, "channel is not a group DM")
	}

	// Verify user is participant
	isParticipant, err := e.gdm.IsGroupDmParticipant(c.UserContext(), channelId, userId)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGroupDMChannel)
	}

	if !isParticipant {
		return nil, fiber.NewError(fiber.StatusForbidden, "not a participant in this group DM")
	}

	return &channel, nil
}

// createNewGroupDM handles creating a new group DM channel with transaction-like behavior
func (e *entity) createNewGroupDM(c *fiber.Ctx, userId int64, recipientIds []int64) error {
	channelId := idgen.Next()
	allParticipants := append([]int64{userId}, recipientIds...)

	// Create channel and add participants atomically
	channel, err := e.createGroupDMWithParticipants(c, channelId, allParticipants)
	if err != nil {
		return err
	}

	return c.JSON(channel)
}

// createGroupDMWithParticipants creates a group DM channel and adds participants with proper cleanup
func (e *entity) createGroupDMWithParticipants(c *fiber.Ctx, channelId int64, participants []int64) (dto.Channel, error) {
	// Create the channel
	if err := e.ch.CreateChannel(c.UserContext(), channelId, "", model.ChannelTypeGroupDM, nil, nil, false); err != nil {
		return dto.Channel{}, helper.HttpDbError(err, ErrUnableToCreateChannel)
	}

	// Add all participants
	if err := e.gdm.JoinGroupDmChannelMany(c.UserContext(), channelId, participants); err != nil {
		// Cleanup: delete the channel if adding participants fails
		if cleanupErr := e.ch.DeleteChannel(c.UserContext(), channelId); cleanupErr != nil {
			// Log cleanup failure but return original error
			e.log.Error("failed to cleanup channel after participant join failure",
				"channel_id", channelId,
				"cleanup_error", cleanupErr.Error(),
				"original_error", err.Error())
		}
		return dto.Channel{}, helper.HttpDbError(err, ErrUnableToJoingGroupDmChannel)
	}

	// Return the successfully created channel
	return dto.Channel{
		Id:        channelId,
		Type:      model.ChannelTypeGroupDM,
		GuildId:   nil,
		Name:      "",
		ParentId:  nil,
		CreatedAt: time.Now(),
	}, nil
}

// channelToDTO converts a model.Channel to dto.Channel for group DMs
func (e *entity) channelToDTO(channel *model.Channel) dto.Channel {
	return dto.Channel{
		Id:          channel.Id,
		Type:        channel.Type,
		GuildId:     nil, // Group DMs don't belong to guilds
		Name:        channel.Name,
		ParentId:    channel.ParentID,
		Permissions: channel.Permissions,
		CreatedAt:   channel.CreatedAt,
	}
}

// GetUserSettings
//
//	@Summary	Get current user settings (optional version gating)
//	@Produce	json
//	@Tags		User
//	@Param		version	query		int						false	"Client known version"
//	@Success	200		{object}	UserSettingsResponse	"User settings and version"
//	@Success	204		{string}	string					"No changes"
//	@failure	400		{string}	string					"Bad request"
//	@failure	500		{string}	string					"Internal server error"
//	@Router		/user/me/settings [get]
func (e *entity) GetUserSettings(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	// Parse optional version filter; default 0
	var version int64
	if v := c.Query("version"); v != "" {
		version, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseVersion)
		}
	}

	s, err := e.uset.GetUserSettings(c.UserContext(), user.Id, version)
	if err != nil {
		return helper.HttpDbError(err, ErrUnableToGetUserSettings)
	}

	memb, err := e.member.GetUserGuilds(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetMembership)
	}

	var gids = make([]int64, len(memb))
	for i, m := range memb {
		gids[i] = m.GuildId
	}

	guilds, err := e.guild.GetGuildsList(c.UserContext(), gids)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUserGuilds)
	}

	gclm, err := e.gclm.GetChannelsMessagesForGuilds(c.UserContext(), gids)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuilds)
	}

	rs, err := e.rs.GetReadStates(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetReadStates)
	}

	settings, err := modelToSettings(&s, e.guildModelToGuildMany(c, guilds), rs, gclm)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUnmarshalUserSettings)
	}
	return c.JSON(settings)
}

// SetUserSettings
//
//	@Summary	Update current user settings (replaces and bumps version)
//	@Accept		json
//	@Produce	json
//	@Tags		User
//	@Param		request	body		model.UserSettingsData	true	"User settings"
//	@Success	200		{string}	string					"ok"
//	@failure	400		{string}	string					"Bad request"
//	@failure	500		{string}	string					"Internal server error"
//	@Router		/user/me/settings [post]
func (e *entity) SetUserSettings(c *fiber.Ctx) error {
	var req model.UserSettingsData
	if err := c.BodyParser(&req); err != nil {
		slog.Error("failed to parse request body", slog.String("error", err.Error()))
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	if err := e.uset.SetUserSettings(c.UserContext(), user.Id, req); err != nil {
		return helper.HttpDbError(err, ErrUnableToSetUserSettings)
	}

	go func() {
		if err := e.mqt.SendUserUpdate(user.Id, &mqmsg.UpdateUserSettings{
			Settings: req,
		}); err != nil {
			slog.Error("unable to send update user settings event", slog.String("error", err.Error()))
		}
	}()

	return c.SendStatus(fiber.StatusOK)
}
