package guild

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

type permissionChecker interface {
	ChannelPerm(ctx context.Context, guildID, channelID, userID int64, perm ...permissions.RolePermission) (*model.Channel, *model.GuildChannel, *model.Guild, bool, error)
	GuildPerm(ctx context.Context, guildID, userID int64, perm ...permissions.RolePermission) (*model.Guild, bool, error)
	GetChannelPermissions(ctx context.Context, guildID, channelID, userID int64) (int64, error)
}

// KickMember
//
//	@Summary		Kick guild member
//	@Description	Removes a guild member. Allowed for guild owner, administrators, or members with PermMembershipKickMembers. Cannot target the guild owner. Members with administrator permission can only be moderated by the guild owner.
//	@Tags			Guild
//	@Param			guild_id	path		int64	true	"Guild ID"
//	@Param			user_id		path		int64	true	"User ID"
//	@Success		204			{string}	string	"No Content"
//	@failure		400			{string}	string	"Bad request"
//	@failure		404			{string}	string	"Member not found"
//	@failure		406			{string}	string	"Permissions required"
//	@Router			/guild/{guild_id}/member/{user_id}/kick [post]
func (e *entity) KickMember(c *fiber.Ctx) error {
	guildId, memberId, user, err := e.parseMemberModerationRequest(c)
	if err != nil {
		return err
	}

	if _, err := e.authorizeMemberModeration(c.UserContext(), guildId, user.Id, memberId, permissions.PermMembershipKickMembers, true); err != nil {
		return err
	}

	if err := e.memb.RemoveMember(c.UserContext(), memberId, guildId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToRemoveMember)
	}

	e.sendGuildMemberRemoved(guildId, memberId, user.Id, mqmsg.GuildMemberModerationKick, nil)
	return c.SendStatus(fiber.StatusNoContent)
}

// BanMember
//
//	@Summary		Ban guild member
//	@Description	Bans a guild member with an optional reason. Allowed for guild owner, administrators, or members with PermMembershipBanMembers. Cannot target the guild owner. Members with administrator permission can only be moderated by the guild owner.
//	@Tags			Guild
//	@Param			guild_id	path		int64				true	"Guild ID"
//	@Param			user_id		path		int64				true	"User ID"
//	@Param			request		body		BanMemberRequest	false	"Ban reason"
//	@Success		204			{string}	string				"No Content"
//	@failure		400			{string}	string				"Bad request"
//	@failure		404			{string}	string				"Member not found"
//	@failure		406			{string}	string				"Permissions required"
//	@Router			/guild/{guild_id}/member/{user_id}/ban [post]
func (e *entity) BanMember(c *fiber.Ctx) error {
	guildId, memberId, user, err := e.parseMemberModerationRequest(c)
	if err != nil {
		return err
	}

	var req BanMemberRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
		}
		if err := req.Validate(); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}

	if _, err := e.authorizeMemberModeration(c.UserContext(), guildId, user.Id, memberId, permissions.PermMembershipBanMembers, true); err != nil {
		return err
	}
	if e.ban == nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToBanMember)
	}
	if err := e.ban.BanUser(c.UserContext(), guildId, memberId, req.Reason); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToBanMember)
	}
	if err := e.memb.RemoveMember(c.UserContext(), memberId, guildId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToRemoveMember)
	}

	e.sendGuildMemberRemoved(guildId, memberId, user.Id, mqmsg.GuildMemberModerationBan, req.Reason)
	return c.SendStatus(fiber.StatusNoContent)
}

// UnbanMember
//
//	@Summary		Unban guild member
//	@Description	Removes a guild ban. Allowed for guild owner, administrators, or members with PermMembershipBanMembers. Administrators can only be moderated by the guild owner.
//	@Tags			Guild
//	@Param			guild_id	path		int64	true	"Guild ID"
//	@Param			user_id		path		int64	true	"User ID"
//	@Success		204			{string}	string	"No Content"
//	@failure		400			{string}	string	"Bad request"
//	@failure		406			{string}	string	"Permissions required"
//	@Router			/guild/{guild_id}/member/{user_id}/ban [delete]
func (e *entity) UnbanMember(c *fiber.Ctx) error {
	guildId, memberId, user, err := e.parseMemberModerationRequest(c)
	if err != nil {
		return err
	}

	if _, err := e.authorizeMemberModeration(c.UserContext(), guildId, user.Id, memberId, permissions.PermMembershipBanMembers, false); err != nil {
		return err
	}
	if e.ban == nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUnbanMember)
	}
	if err := e.ban.UnbanUser(c.UserContext(), guildId, memberId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUnbanMember)
	}

	e.sendGuildModerationEvent(guildId, memberId, user.Id, mqmsg.GuildMemberModerationUnban, nil)
	return c.SendStatus(fiber.StatusNoContent)
}

// GetBans
//
//	@Summary		Get guild bans
//	@Description	Returns banned users with optional ban reasons. Allowed for guild owner, administrators, or members with PermMembershipBanMembers.
//	@Tags			Guild
//	@Param			guild_id	path		int64			true	"Guild ID"
//	@Success		200			{array}		dto.GuildBan	"Ok"
//	@failure		400			{string}	string			"Bad request"
//	@failure		406			{string}	string			"Permissions required"
//	@Router			/guild/{guild_id}/bans [get]
func (e *entity) GetBans(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	if _, err := e.authorizeGuildPermission(c.UserContext(), guildId, user.Id, permissions.PermMembershipBanMembers); err != nil {
		return err
	}
	if e.ban == nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildBans)
	}

	bans, err := e.ban.GetGuildBans(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildBans)
	}
	if len(bans) == 0 {
		return c.JSON([]dto.GuildBan{})
	}

	userIDs := make([]int64, 0, len(bans))
	for _, ban := range bans {
		userIDs = append(userIDs, ban.UserId)
	}

	users, err := e.user.GetUsersList(c.UserContext(), userIDs)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUsers)
	}
	dscs, err := e.disc.GetDiscriminatorsByUserIDs(c.UserContext(), userIDs)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetDiscriminators)
	}

	usersByID := make(map[int64]model.User, len(users))
	for _, u := range users {
		usersByID[u.Id] = u
	}
	dscByID := make(map[int64]string, len(dscs))
	for _, d := range dscs {
		dscByID[d.UserId] = d.Discriminator
	}
	avData := make(map[int64]*dto.AvatarData, len(users))
	for _, u := range users {
		if u.Avatar != nil {
			if ad, err := e.getAvatarDataCached(c.UserContext(), u.Id, *u.Avatar); err == nil && ad != nil {
				avData[u.Id] = ad
			}
		}
	}

	result := make([]dto.GuildBan, 0, len(bans))
	for _, ban := range bans {
		userDTO := dto.User{Id: ban.UserId, Name: "Unknown User", Discriminator: dscByID[ban.UserId]}
		if u, ok := usersByID[ban.UserId]; ok {
			userDTO = userToDTO(u, dscByID[ban.UserId])
			if ad, ok := avData[ban.UserId]; ok {
				userDTO.Avatar = ad
			}
		}
		result = append(result, dto.GuildBan{User: userDTO, Reason: ban.Reason})
	}

	return c.JSON(result)
}

func (e *entity) parseMemberModerationRequest(c *fiber.Ctx) (int64, int64, *helper.JWTUser, error) {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return 0, 0, nil, err
	}
	memberId, err := e.parseUserID(c)
	if err != nil {
		return 0, 0, nil, err
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return 0, 0, nil, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	return guildId, memberId, user, nil
}

func (e *entity) authorizeGuildPermission(ctx context.Context, guildId, actorId int64, required permissions.RolePermission) (*model.Guild, error) {
	guild, err := e.g.GetGuildById(ctx, guildId)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}

	if actorId != guild.OwnerId {
		isMember, err := e.memb.IsGuildMember(ctx, guildId, actorId)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMemberToken)
		}
		if !isMember {
			return nil, fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
		}
	}

	_, allowed, err := e.perm.GuildPerm(ctx, guildId, actorId, required)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetPermission)
	}
	if !allowed {
		return nil, fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	return &guild, nil
}

func (e *entity) authorizeMemberModeration(ctx context.Context, guildId, actorId, targetId int64, required permissions.RolePermission, requireTargetMember bool) (*model.Guild, error) {
	guild, err := e.authorizeGuildPermission(ctx, guildId, actorId, required)
	if err != nil {
		return nil, err
	}

	if targetId == guild.OwnerId {
		return nil, fiber.NewError(fiber.StatusNotAcceptable, ErrCannotModerateGuildOwner)
	}
	if requireTargetMember {
		isMember, err := e.memb.IsGuildMember(ctx, guildId, targetId)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMemberToken)
		}
		if !isMember {
			return nil, fiber.NewError(fiber.StatusNotFound, ErrNotAMember)
		}
	}

	_, targetIsAdmin, err := e.perm.GuildPerm(ctx, guildId, targetId, permissions.PermAdministrator)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetPermission)
	}
	if targetIsAdmin && actorId != guild.OwnerId {
		return nil, fiber.NewError(fiber.StatusNotAcceptable, ErrOnlyOwnerCanModerateAdministrator)
	}

	return guild, nil
}

func (e *entity) sendGuildMemberRemoved(guildId, memberId, actorId int64, action mqmsg.GuildMemberModerationAction, reason *string) {
	if e.mqt == nil {
		return
	}
	logger := e.log
	if logger == nil {
		logger = slog.Default()
	}
	go func() {
		if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.RemoveGuildMember{GuildId: guildId, UserId: memberId}); err != nil {
			logger.Error("unable to send guild member remove event after moderation",
				slog.String("action", string(action)),
				slog.Int64("guild_id", guildId),
				slog.Int64("user_id", memberId),
				slog.String("error", err.Error()))
		}
		if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.GuildMemberModeration{GuildId: guildId, UserId: memberId, ActorId: actorId, Action: action, Reason: reason}); err != nil {
			logger.Error("unable to send guild member moderation event",
				slog.String("action", string(action)),
				slog.Int64("guild_id", guildId),
				slog.Int64("user_id", memberId),
				slog.String("error", err.Error()))
		}
	}()
}

func (e *entity) sendGuildModerationEvent(guildId, memberId, actorId int64, action mqmsg.GuildMemberModerationAction, reason *string) {
	if e.mqt == nil {
		return
	}
	logger := e.log
	if logger == nil {
		logger = slog.Default()
	}
	go func() {
		if err := e.mqt.SendGuildUpdate(guildId, &mqmsg.GuildMemberModeration{GuildId: guildId, UserId: memberId, ActorId: actorId, Action: action, Reason: reason}); err != nil {
			logger.Error("unable to send guild member moderation event",
				slog.String("action", string(action)),
				slog.Int64("guild_id", guildId),
				slog.Int64("user_id", memberId),
				slog.String("error", err.Error()))
		}
	}()
}

func (e *entity) isGuildUserBanned(ctx context.Context, guildId, userId int64) (bool, error) {
	if e.ban == nil {
		return false, nil
	}
	return e.ban.IsBanned(ctx, guildId, userId)
}
