package rolecheck

import (
	"context"
	"errors"

	"github.com/gocql/gocql"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

// getUserRoleIDs retrieves user role IDs for a given guild and user
// Returns a slice of role IDs and an error if any
func (e *Entity) getUserRoleIDs(ctx context.Context, guildID, userID int64) ([]int64, error) {
	userRoles, err := e.ur.GetUserRoles(ctx, guildID, userID)
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
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
	// Administrator permission is always checked
	perm = append(perm, permissions.PermAdministrator)

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

	// Get channel information
	channel, err := e.ch.GetChannel(ctx, gc.ChannelId)
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
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, nil, nil, false, err
	}

	// Private channel access flag
	var allowPrivate = !channel.Private

	// Get channel role permissions
	channelRolePerms, err := e.chrp.GetChannelRolePermissions(ctx, channelID)
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
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
	userChannelPerm, err := e.chup.GetUserChannelPermission(ctx, channelID, userID)
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, nil, nil, false, err
	}

	// Apply user-specific permissions if found
	if !errors.Is(err, gocql.ErrNotFound) { // User has specific permissions
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

// GuildPerm checks if a user has the specified permissions for a guild
// Returns guild, permission status, and error
func (e *Entity) GuildPerm(ctx context.Context, guildID, userID int64, perm ...permissions.RolePermission) (*model.Guild, bool, error) {
	// Administrator permission is always checked
	perm = append(perm, permissions.PermAdministrator)

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
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
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
