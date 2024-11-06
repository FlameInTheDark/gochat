package message

import (
	"context"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/permissions"
)

func (e *entity) checkChannelPermissions(ctx context.Context, userId, channelId int64, perms ...permissions.RolePermission) bool {
	ch, err := e.ch.GetChannel(ctx, channelId)
	if err != nil {
		e.log.Error("unable to get channel", slog.String("error", err.Error()))
		return false
	}
	g, err := e.gc.GetGuildByChannel(ctx, channelId)
	if err != nil {
		e.log.Error("unable to get guild by channel id", slog.String("error", err.Error()))
		return false
	}
	guild, err := e.g.GetGuildById(ctx, g.GuildId)
	if err != nil {
		e.log.Error("unable to get guild by id", slog.String("error", err.Error()))
		return false
	}
	if guild.OwnerId == userId {
		return true
	}
	perm, err := e.uperm.GetUserChannelPermission(ctx, channelId, userId)
	if err == nil {
		return permissions.CheckPermissions(perm.Permissions, perms...)
	} else {
		e.log.Error("unable to get user channel permissions", slog.String("error", err.Error()))
		cr, err := e.rperm.GetChannelRolePermissions(ctx, channelId)
		if err == nil {
			return permissions.CheckPermissions(ch.Permissions, perms...)
		} else {
			urs, err := e.ur.GetUserRoles(ctx, g.GuildId, userId)
			if err == nil {
				for _, r := range cr {
					for _, ur := range urs {
						if r.RoleId == ur.RoleId {
							role, err := e.role.GetRoleByID(ctx, ur.RoleId)
							if err == nil {
								return permissions.CheckPermissions(role.Permissions, perms...)
							}
						}
					}
				}
			} else {
				return permissions.CheckPermissions(ch.Permissions, perms...)
			}
		}
	}
	return false
}
