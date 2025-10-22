package rolecheck

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

// getUserRoleIDs retrieves user role IDs for a given guild and user
// Returns a slice of role IDs and an error if any
func (e *Entity) getUserRoleIDs(ctx context.Context, guildID, userID int64) ([]int64, error) {
	userRoles, err := e.ur.GetUserRoles(ctx, guildID, userID)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]int64, len(userRoles))
	for i, role := range userRoles {
		roleIDs[i] = role.RoleId
	}

	return roleIDs, nil
}

// ChannelPerm checks if a user has the specified permissions for a channel
// Returns channel, guild channel, guild, permission status, and error
func (e *Entity) ChannelPerm(ctx context.Context, guildID, channelID, userID int64, perm ...permissions.RolePermission) (*model.Channel, *model.GuildChannel, *model.Guild, bool, error) {
	// Get channel information
	channel, err := e.ch.GetChannel(ctx, channelID)
	if err != nil {
		return nil, nil, nil, false, err
	}

	// Handle different channel types
	switch channel.Type {
	case model.ChannelTypeDM:
		// For DM channels, check if the user is a participant
		isParticipant, err := e.dm.IsDmChannelParticipant(ctx, channelID, userID)
		if err != nil {
			return nil, nil, nil, false, err
		}
		if isParticipant {
			return &channel, nil, nil, true, nil
		}
		return nil, nil, nil, false, nil

	case model.ChannelTypeGroupDM:
		// For group DM channels, check if the user is a participant
		isParticipant, err := e.gdm.IsGroupDmParticipant(ctx, channelID, userID)
		if err != nil {
			return nil, nil, nil, false, err
		}
		if isParticipant {
			return &channel, nil, nil, true, nil
		}
		return nil, nil, nil, false, nil

	case model.ChannelTypeThread:
		// For thread channels, inherit parent channel permissions
		if channel.ParentID == nil {
			return nil, nil, nil, false, fmt.Errorf("thread channel has no parent")
		}
		// Recursively check permissions on the parent channel
		return e.ChannelPerm(ctx, guildID, *channel.ParentID, userID, perm...)

	default:
		// For guild channels, check specific permissions
		// Get guild information
		guild, err := e.g.GetGuildById(ctx, guildID)
		if err != nil {
			return nil, nil, nil, false, err
		}

		// Get guild channel information
		gc, err := e.gc.GetGuildChannel(ctx, guildID, channelID)
		if err != nil {
			return nil, nil, nil, false, err
		}

		// Guild owner has all permissions
		if userID == guild.OwnerId {
			return &channel, &gc, &guild, true, nil
		}

		// Determine base permissions
		var permAll int64
		if channel.Permissions != nil {
			permAll = *channel.Permissions
		} else {
			permAll = guild.Permissions
		}

		// Get user role IDs
		roleIDs, err := e.getUserRoleIDs(ctx, guildID, userID)
		if err != nil {
			return nil, nil, nil, false, err
		}

		// Get role information
		roles, err := e.role.GetRolesBulk(ctx, roleIDs)
		if err != nil {
			return nil, nil, nil, false, err
		}

		// Private channel access flag
		var allowPrivate = !channel.Private

		// Get channel role permissions
		channelRolePerms, err := e.chrp.GetChannelRolePermissions(ctx, channelID)
		if err != nil {
			return nil, nil, nil, false, err
		}

		// Create a map for faster lookup
		channelRolePermsMap := make(map[int64]*model.ChannelRolesPermission)
		for i, p := range channelRolePerms {
			channelRolePermsMap[p.RoleId] = &channelRolePerms[i]
		}

		// Process role permissions
		for _, role := range roles {
			if channelRolePerm, ok := channelRolePermsMap[role.Id]; ok {
				allowPrivate = true
				role.Permissions = permissions.AddRoles(role.Permissions, channelRolePerm.Accept)
				role.Permissions = permissions.SubtractRoles(role.Permissions, channelRolePerm.Deny)
			}
			permAll = permissions.AddRoles(permAll, role.Permissions)
		}

		// Get user-specific channel permissions
		userChannelPerm, ucpErr := e.chup.GetUserChannelPermission(ctx, channelID, userID)
		if ucpErr != nil && !errors.Is(ucpErr, sql.ErrNoRows) {
			return nil, nil, nil, false, ucpErr
		}

		// Apply user-specific permissions if found
		if ucpErr == nil { // User has specific permissions
			allowPrivate = true
			permAll = permissions.AddRoles(permAll, userChannelPerm.Accept)
			permAll = permissions.SubtractRoles(permAll, userChannelPerm.Deny)
		}

		// Check if user can access private channel
		if !allowPrivate {
			return nil, nil, nil, false, nil
		}

		// Check if user has all required permissions
		if permissions.CheckPermissions(permAll, perm...) {
			return &channel, &gc, &guild, true, nil
		}

		return nil, nil, nil, false, nil
	}
}

// ChannelEffectivePermissions computes the combined permission bitmask for a user in a given channel.
// It mirrors ChannelPerm logic but returns the effective permission flags instead of checking specific bits.
func (e *Entity) GetChannelPermissions(ctx context.Context, guildID, channelID, userID int64) (int64, error) {
	// Get channel information
	channel, err := e.ch.GetChannel(ctx, channelID)
	if err != nil {
		return 0, err
	}

	// Handle DM/group DM/thread specially: use access semantics similar to ChannelPerm
	switch channel.Type {
	case model.ChannelTypeDM:
		ok, err := e.dm.IsDmChannelParticipant(ctx, channelID, userID)
		if err != nil {
			return 0, err
		}
		if ok {
			return permissions.CreatePermissions(permissions.PermServerViewChannels), nil
		}
		return 0, nil
	case model.ChannelTypeGroupDM:
		ok, err := e.gdm.IsGroupDmParticipant(ctx, channelID, userID)
		if err != nil {
			return 0, err
		}
		if ok {
			return permissions.CreatePermissions(permissions.PermServerViewChannels), nil
		}
		return 0, nil
	case model.ChannelTypeThread:
		if channel.ParentID == nil {
			return 0, fmt.Errorf("thread channel has no parent")
		}
		return e.GetChannelPermissions(ctx, guildID, *channel.ParentID, userID)
	}

	// Guild + guild channel info
	guild, err := e.g.GetGuildById(ctx, guildID)
	if err != nil {
		return 0, err
	}
	if _, err := e.gc.GetGuildChannel(ctx, guildID, channelID); err != nil {
		return 0, err
	}

	// Guild owner has all permissions
	if userID == guild.OwnerId {
		return int64(^uint64(0) >> 1), nil
	}

	// Base perms from channel or guild
	var permAll int64
	if channel.Permissions != nil {
		permAll = *channel.Permissions
	} else {
		permAll = guild.Permissions
	}

	// Roles
	roleIDs, err := e.getUserRoleIDs(ctx, guildID, userID)
	if err != nil {
		return 0, err
	}
	roles, err := e.role.GetRolesBulk(ctx, roleIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	// Channel role overrides
	channelRolePerms, err := e.chrp.GetChannelRolePermissions(ctx, channelID)
	if err != nil {
		return 0, err
	}
	channelRolePermsMap := make(map[int64]*model.ChannelRolesPermission)
	for i, p := range channelRolePerms {
		channelRolePermsMap[p.RoleId] = &channelRolePerms[i]
	}

	// Apply role permissions
	for _, r := range roles {
		if crp, ok := channelRolePermsMap[r.Id]; ok {
			r.Permissions = permissions.AddRoles(r.Permissions, crp.Accept)
			r.Permissions = permissions.SubtractRoles(r.Permissions, crp.Deny)
		}
		permAll = permissions.AddRoles(permAll, r.Permissions)
	}

	// User-specific overrides
	if ucp, err := e.chup.GetUserChannelPermission(ctx, channelID, userID); err == nil {
		permAll = permissions.AddRoles(permAll, ucp.Accept)
		permAll = permissions.SubtractRoles(permAll, ucp.Deny)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	// If channel is private and user has no role overrides, access may still be restricted in higher layers.
	return permAll, nil
}

// GuildPerm checks if a user has the specified permissions for a guild
// Returns guild, permission status, and error
func (e *Entity) GuildPerm(ctx context.Context, guildID, userID int64, perm ...permissions.RolePermission) (*model.Guild, bool, error) {
	// Get guild information
	guild, err := e.g.GetGuildById(ctx, guildID)
	if err != nil {
		return nil, false, err
	}

	// Guild owner has all permissions
	if userID == guild.OwnerId {
		return &guild, true, nil
	}

	// Start with guild base permissions
	permAll := guild.Permissions

	// Get user role IDs
	roleIDs, err := e.getUserRoleIDs(ctx, guildID, userID)
	if err != nil {
		return nil, false, err
	}

	// Get role information
	roles, err := e.role.GetRolesBulk(ctx, roleIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	// Process role permissions
	for _, role := range roles {
		permAll = permissions.AddRoles(permAll, role.Permissions)
	}

	// Check if user has all required permissions
	if permissions.CheckPermissions(permAll, perm...) {
		return &guild, true, nil
	}
	return nil, false, nil
}
