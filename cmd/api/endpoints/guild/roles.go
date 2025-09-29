package guild

import (
	"database/sql"
	"errors"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/gofiber/fiber/v2"
)

// GetMemberRoles
//
//	@Summary	Get member roles
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64		true	"Guild ID"	example(2230469276416868352)
//	@Param		user_id		path		int64		true	"User ID"	example(2230469276416868352)
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

	roleIds := make([]int64, 0, len(memberRoleIds))
	for _, role := range memberRoleIds {
		roleIds = append(roleIds, role.RoleId)
	}

	roles, err := e.role.GetRolesBulk(c.UserContext(), roleIds)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}

	return c.JSON(roleModelToDTOMany(roles))
}

// GetGuildRoles
//
//	@Summary	Get guild roles
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64		true	"Guild ID"	example(2230469276416868352)
//	@Success	200			{array}		dto.Role	"Roles list"
//	@failure	400			{string}	string		"Incorrect request body"
//	@failure	401			{string}	string		"Unauthorized"
//	@failure	500			{string}	string		"Something bad happened"
//	@Router		/guild/{guild_id}/roles [get]
func (e *entity) GetGuildRoles(c *fiber.Ctx) error {
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

	roles, err := e.role.GetGuildRoles(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}
	return c.JSON(roleModelToDTOMany(roles))
}

// CreateGuildRole
//
//	@Summary	Create guild role
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64					true	"Guild ID"	example(2230469276416868352)
//	@Param		req			body		CreateGuildRoleRequest	true	"Role data"
//	@Success	201			{object}	dto.Role				"Role"
//	@failure	400			{string}	string					"Incorrect request body"
//	@failure	401			{string}	string					"Unauthorized"
//	@failure	406			{string}	string					"Permissions required"
//	@failure	500			{string}	string					"Something bad happened"
//	@Router		/guild/{guild_id}/roles [post]
func (e *entity) CreateGuildRole(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	var req CreateGuildRoleRequest
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

	_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermServerManageRoles)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	roleId := idgen.Next()
	if err := e.role.CreateRole(c.UserContext(), roleId, guildId, req.Name, req.Color, req.Permissions); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}

	created := dto.Role{Id: roleId, GuildId: guildId, Name: req.Name, Color: req.Color, Permissions: req.Permissions}
	go e.mqt.SendGuildUpdate(guildId, &mqmsg.CreateGuildRole{Role: created})
	return c.Status(fiber.StatusCreated).JSON(created)
}

// PatchGuildRole
//
//	@Summary	Update guild role
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64					true	"Guild ID"	example(2230469276416868352)
//	@Param		role_id		path		int64					true	"Role ID"	example(2230469276416868352)
//	@Param		req			body		PatchGuildRoleRequest	true	"Role changes"
//	@Success	200			{object}	dto.Role				"Role"
//	@failure	400			{string}	string					"Incorrect request body"
//	@failure	401			{string}	string					"Unauthorized"
//	@failure	404			{string}	string					"Role not found"
//	@failure	406			{string}	string					"Permissions required"
//	@failure	500			{string}	string					"Something bad happened"
//	@Router		/guild/{guild_id}/roles/{role_id} [patch]
func (e *entity) PatchGuildRole(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	roleId, err := e.parseRoleID(c)
	if err != nil {
		return err
	}

	var req PatchGuildRoleRequest
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

	_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermServerManageRoles)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	r, err := e.role.GetRoleByID(c.UserContext(), roleId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}
	if r.GuildId != guildId {
		return fiber.NewError(fiber.StatusBadRequest, ErrRoleNotInGuild)
	}

	if req.Name != nil {
		if err := e.role.SetRoleName(c.UserContext(), roleId, *req.Name); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
		}
	}
	if req.Color != nil {
		if err := e.role.SetRoleColor(c.UserContext(), roleId, *req.Color); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
		}
	}
	if req.Permissions != nil {
		if err := e.role.SetRolePermissions(c.UserContext(), roleId, *req.Permissions); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
		}
	}

	// Return updated role
	ur, err := e.role.GetRoleByID(c.UserContext(), roleId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}
	role := roleModelToDTO(ur)
	go e.mqt.SendGuildUpdate(guildId, &mqmsg.UpdateGuildRole{GuildId: guildId, Role: role})

	return c.JSON(role)
}

// DeleteGuildRole
//
//	@Summary	Delete guild role
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64	true	"Guild ID"	example(2230469276416868352)
//	@Param		role_id		path		int64	true	"Role ID"	example(2230469276416868352)
//	@Success	200			{string}	string	"Deleted"
//	@failure	400			{string}	string	"Incorrect request body"
//	@failure	401			{string}	string	"Unauthorized"
//	@failure	404			{string}	string	"Role not found"
//	@failure	406			{string}	string	"Permissions required"
//	@failure	500			{string}	string	"Something bad happened"
//	@Router		/guild/{guild_id}/roles/{role_id} [delete]
func (e *entity) DeleteGuildRole(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	roleId, err := e.parseRoleID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermServerManageRoles)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	r, err := e.role.GetRoleByID(c.UserContext(), roleId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}
	if r.GuildId != guildId {
		return fiber.NewError(fiber.StatusBadRequest, ErrRoleNotInGuild)
	}

	if err := e.ur.RemoveRoleAssignments(c.UserContext(), guildId, roleId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}

	if err := e.role.RemoveRole(c.UserContext(), roleId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}
	return c.SendStatus(fiber.StatusOK)
}

// AddMemberRole
//
//	@Summary	Assign role to member
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64	true	"Guild ID"	example(2230469276416868352)
//	@Param		user_id		path		int64	true	"User ID"	example(2230469276416868352)
//	@Param		role_id		path		int64	true	"Role ID"	example(2230469276416868352)
//	@Success	200			{string}	string	"Ok"
//	@failure	400			{string}	string	"Bad request"
//	@failure	401			{string}	string	"Unauthorized"
//	@failure	404			{string}	string	"Member not found"
//	@failure	406			{string}	string	"Permissions required"
//	@failure	500			{string}	string	"Something bad happened"
//	@Router		/guild/{guild_id}/member/{user_id}/roles/{role_id} [put]
func (e *entity) AddMemberRole(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	memberId, err := e.parseUserID(c)
	if err != nil {
		return err
	}

	roleId, err := e.parseRoleID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermServerManageRoles)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	isGuildMember, err := e.memb.IsGuildMember(c.UserContext(), guildId, memberId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMemberToken)
	}
	if !isGuildMember {
		return fiber.NewError(fiber.StatusNotFound, ErrNotAMember)
	}

	r, err := e.role.GetRoleByID(c.UserContext(), roleId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}
	if r.GuildId != guildId {
		return fiber.NewError(fiber.StatusBadRequest, ErrRoleNotInGuild)
	}

	if err := e.ur.AddUserRole(c.UserContext(), guildId, memberId, roleId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSetUserRole)
	}

	go e.mqt.SendGuildUpdate(guildId, &mqmsg.AddGuildMemberRole{GuildId: guildId, RoleId: roleId, UserId: memberId})

	return c.SendStatus(fiber.StatusOK)
}

// RemoveMemberRole
//
//	@Summary	Remove role from member
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64	true	"Guild ID"	example(2230469276416868352)
//	@Param		user_id		path		int64	true	"User ID"	example(2230469276416868352)
//	@Param		role_id		path		int64	true	"Role ID"	example(2230469276416868352)
//	@Success	200			{string}	string	"Ok"
//	@failure	400			{string}	string	"Bad request"
//	@failure	401			{string}	string	"Unauthorized"
//	@failure	404			{string}	string	"Member not found"
//	@failure	406			{string}	string	"Permissions required"
//	@failure	500			{string}	string	"Something bad happened"
//	@Router		/guild/{guild_id}/member/{user_id}/roles/{role_id} [delete]
func (e *entity) RemoveMemberRole(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	memberId, err := e.parseUserID(c)
	if err != nil {
		return err
	}

	roleId, err := e.parseRoleID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermServerManageRoles)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	isGuildMember, err := e.memb.IsGuildMember(c.UserContext(), guildId, memberId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMemberToken)
	}
	if !isGuildMember {
		return fiber.NewError(fiber.StatusNotFound, ErrNotAMember)
	}

	r, err := e.role.GetRoleByID(c.UserContext(), roleId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRoles)
	}
	if r.GuildId != guildId {
		return fiber.NewError(fiber.StatusBadRequest, ErrRoleNotInGuild)
	}

	if err := e.ur.RemoveUserRole(c.UserContext(), guildId, memberId, roleId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToRemoveUserRole)
	}

	go e.mqt.SendGuildUpdate(guildId, &mqmsg.RemoveGuildMemberRole{GuildId: guildId, RoleId: roleId, UserId: memberId})

	return c.SendStatus(fiber.StatusOK)
}

// GetChannelRolePermissions
//
//	@Summary	List channel role permissions
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64					true	"Guild ID"		example(2230469276416868352)
//	@Param		channel_id	path		int64					true	"Channel ID"	example(2230469276416868352)
//	@Success	200			{array}		ChannelRolePermission	"List of role permissions"
//	@failure	400			{string}	string					"Incorrect request"
//	@failure	401			{string}	string					"Unauthorized"
//	@failure	406			{string}	string					"Permissions required"
//	@failure	500			{string}	string					"Something bad happened"
//	@Router		/guild/{guild_id}/channel/{channel_id}/roles [get]
func (e *entity) GetChannelRolePermissions(c *fiber.Ctx) error {
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

	// Only membership is required to view
	isMember, err := e.memb.IsGuildMember(c.UserContext(), guildId, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMemberToken)
	}
	if !isMember {
		return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
	}

	// Ensure channel belongs to guild (exists in mapping)
	if _, err := e.gc.GetGuildChannel(c.UserContext(), guildId, channelId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}

	perms, err := e.rperm.GetChannelRolePermissions(c.UserContext(), channelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannelRolePerms)
	}
	out := make([]ChannelRolePermission, 0, len(perms))
	for _, p := range perms {
		out = append(out, ChannelRolePermission{RoleId: p.RoleId, Accept: p.Accept, Deny: p.Deny})
	}
	return c.JSON(out)
}

// GetChannelRolePermission
//
//	@Summary	Get channel role permission
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64					true	"Guild ID"		example(2230469276416868352)
//	@Param		channel_id	path		int64					true	"Channel ID"	example(2230469276416868352)
//	@Param		role_id		path		int64					true	"Role ID"		example(2230469276416868352)
//	@Success	200			{object}	ChannelRolePermission	"Role permission"
//	@failure	400			{string}	string					"Incorrect request"
//	@failure	401			{string}	string					"Unauthorized"
//	@failure	404			{string}	string					"Not found"
//	@failure	406			{string}	string					"Permissions required"
//	@failure	500			{string}	string					"Something bad happened"
//	@Router		/guild/{guild_id}/channel/{channel_id}/roles/{role_id} [get]
func (e *entity) GetChannelRolePermission(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}
	roleId, err := e.parseRoleID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	// Only membership is required to view
	isMember, err := e.memb.IsGuildMember(c.UserContext(), guildId, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMemberToken)
	}
	if !isMember {
		return fiber.NewError(fiber.StatusUnauthorized, ErrPermissionsRequired)
	}

	if _, err := e.gc.GetGuildChannel(c.UserContext(), guildId, channelId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}

	p, err := e.rperm.GetChannelRolePermission(c.UserContext(), channelId, roleId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetChannelRolePerms)
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannelRolePerms)
	}

	return c.JSON(ChannelRolePermission{RoleId: p.RoleId, Accept: p.Accept, Deny: p.Deny})
}

// SetChannelRolePermission
//
//	@Summary	Set channel role permission (create or replace)
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64							true	"Guild ID"		example(2230469276416868352)
//	@Param		channel_id	path		int64							true	"Channel ID"	example(2230469276416868352)
//	@Param		role_id		path		int64							true	"Role ID"		example(2230469276416868352)
//	@Param		req			body		ChannelRolePermissionRequest	true	"Permission mask"
//	@Success	200			{string}	string							"Ok"
//	@failure	400			{string}	string							"Incorrect request"
//	@failure	401			{string}	string							"Unauthorized"
//	@failure	404			{string}	string							"Role or channel not found"
//	@failure	406			{string}	string							"Permissions required"
//	@failure	500			{string}	string							"Something bad happened"
//	@Router		/guild/{guild_id}/channel/{channel_id}/roles/{role_id} [put]
func (e *entity) SetChannelRolePermission(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}
	roleId, err := e.parseRoleID(c)
	if err != nil {
		return err
	}

	var req ChannelRolePermissionRequest
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
	_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	// Ensure channel belongs to guild
	if _, err := e.gc.GetGuildChannel(c.UserContext(), guildId, channelId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}
	// Ensure role belongs to guild
	r, err := e.role.GetRoleByID(c.UserContext(), roleId)
	if err != nil || r.GuildId != guildId {
		return fiber.NewError(fiber.StatusNotFound, ErrRoleNotInGuild)
	}

	// Upsert behavior: update if exists, else insert
	if _, err := e.rperm.GetChannelRolePermission(c.UserContext(), channelId, roleId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := e.rperm.SetChannelRolePermission(c.UserContext(), channelId, roleId, req.Accept, req.Deny); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSetChannelRolePerm)
			}
			return c.SendStatus(fiber.StatusOK)
		}
		// Other DB error
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannelRolePerms)
	}

	// Sanitize permissions override to prevent setting restricted permissions
	req.Accept = permissions.SanitizeChannelOverrides(req.Accept)
	req.Deny = permissions.SanitizeChannelOverrides(req.Deny)

	if err := e.rperm.UpdateChannelRolePermission(c.UserContext(), channelId, roleId, req.Accept, req.Deny); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUpdateChannelRole)
	}
	return c.SendStatus(fiber.StatusOK)
}

// UpdateChannelRolePermission
//
//	@Summary	Update channel role permission
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64							true	"Guild ID"		example(2230469276416868352)
//	@Param		channel_id	path		int64							true	"Channel ID"	example(2230469276416868352)
//	@Param		role_id		path		int64							true	"Role ID"		example(2230469276416868352)
//	@Param		req			body		ChannelRolePermissionRequest	true	"Permission mask"
//	@Success	200			{string}	string							"Ok"
//	@failure	400			{string}	string							"Incorrect request"
//	@failure	401			{string}	string							"Unauthorized"
//	@failure	404			{string}	string							"Role or channel not found"
//	@failure	406			{string}	string							"Permissions required"
//	@failure	500			{string}	string							"Something bad happened"
//	@Router		/guild/{guild_id}/channel/{channel_id}/roles/{role_id} [patch]
func (e *entity) UpdateChannelRolePermission(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}
	roleId, err := e.parseRoleID(c)
	if err != nil {
		return err
	}

	var req ChannelRolePermissionRequest
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
	_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	if _, err := e.gc.GetGuildChannel(c.UserContext(), guildId, channelId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}
	r, err := e.role.GetRoleByID(c.UserContext(), roleId)
	if err != nil || r.GuildId != guildId {
		return fiber.NewError(fiber.StatusNotFound, ErrRoleNotInGuild)
	}

	req.Accept = permissions.SanitizeChannelOverrides(req.Accept)
	req.Deny = permissions.SanitizeChannelOverrides(req.Deny)

	if err := e.rperm.UpdateChannelRolePermission(c.UserContext(), channelId, roleId, req.Accept, req.Deny); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUpdateChannelRole)
	}
	return c.SendStatus(fiber.StatusOK)
}

// RemoveChannelRolePermission
//
//	@Summary	Remove channel role permission
//	@Produce	json
//	@Tags		Guild Roles
//	@Param		guild_id	path		int64	true	"Guild ID"		example(2230469276416868352)
//	@Param		channel_id	path		int64	true	"Channel ID"	example(2230469276416868352)
//	@Param		role_id		path		int64	true	"Role ID"		example(2230469276416868352)
//	@Success	200			{string}	string	"Ok"
//	@failure	400			{string}	string	"Incorrect request"
//	@failure	401			{string}	string	"Unauthorized"
//	@failure	404			{string}	string	"Role or channel not found"
//	@failure	406			{string}	string	"Permissions required"
//	@failure	500			{string}	string	"Something bad happened"
//	@Router		/guild/{guild_id}/channel/{channel_id}/roles/{role_id} [delete]
func (e *entity) RemoveChannelRolePermission(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}
	roleId, err := e.parseRoleID(c)
	if err != nil {
		return err
	}
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
	if _, err := e.gc.GetGuildChannel(c.UserContext(), guildId, channelId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}
	// Ensure role belongs to guild
	r, err := e.role.GetRoleByID(c.UserContext(), roleId)
	if err != nil || r.GuildId != guildId {
		return fiber.NewError(fiber.StatusNotFound, ErrRoleNotInGuild)
	}
	if err := e.rperm.RemoveChannelRolePermission(c.UserContext(), channelId, roleId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToRemoveChannelRole)
	}
	return c.SendStatus(fiber.StatusOK)
}
