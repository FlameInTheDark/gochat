package rolecheck

import (
	"context"
	"errors"

	"github.com/gocql/gocql"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

func (e *Entity) ChannelPerm(ctx context.Context, guildID, channelID, userID int64, perm ...permissions.RolePermission) (*model.Channel, *model.GuildChannel, *model.Guild, bool, error) {
	perm = append(perm, permissions.PermAdministrator)
	guild, err := e.g.GetGuildById(ctx, guildID)
	if err != nil {
		return nil, nil, nil, false, err
	}
	if userID == guild.OwnerId {
		return nil, nil, nil, true, nil
	}
	gc, err := e.gc.GetGuildChannel(ctx, guildID, channelID)
	if err != nil {
		return nil, nil, nil, false, err
	}
	channel, err := e.ch.GetChannel(ctx, gc.ChannelId)
	if err != nil {
		return nil, nil, nil, false, err
	}
	var permAll int64
	if channel.Permissions != nil {
		permAll = *channel.Permissions
	} else {
		permAll = guild.Permissions
	}
	ur, err := e.ur.GetUserRoles(ctx, guildID, userID)
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, nil, nil, false, err
	}
	var rids = make([]int64, len(ur))
	for i, r := range ur {
		rids[i] = r.RoleId
	}

	rs, err := e.role.GetRolesBulk(ctx, rids)
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, nil, nil, false, err
	}
	var allowPrivate = !channel.Private

	chrp, err := e.chrp.GetChannelRolePermissions(ctx, channelID)
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, nil, nil, false, err
	}
	var chrpm = make(map[int64]*model.ChannelRolesPermission)
	for i, p := range chrp {
		chrpm[p.RoleId] = &chrp[i]
	}
	for _, r := range rs {
		if _, ok := chrpm[r.Id]; ok {
			allowPrivate = true
			r.Permissions = permissions.AddRoles(r.Permissions, chrpm[r.Id].Accept)
			r.Permissions = permissions.SubtractRoles(r.Permissions, chrpm[r.Id].Deny)
		}
		permAll = permissions.AddRoles(permAll, r.Permissions)
	}

	chup, err := e.chup.GetUserChannelPermission(ctx, channelID, userID)
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, nil, nil, false, err
	}
	if !errors.Is(err, gocql.ErrNotFound) {
		allowPrivate = true
		permAll = permissions.AddRoles(permAll, chup.Accept)
		permAll = permissions.SubtractRoles(permAll, chup.Deny)
	}
	if !allowPrivate {
		return nil, nil, nil, false, nil
	}
	if permissions.CheckPermissions(permAll, perm...) {
		return &channel, &gc, &guild, true, nil
	}
	return nil, nil, nil, false, nil
}

func (e *Entity) GuildPerm(ctx context.Context, guildID, userID int64, perm ...permissions.RolePermission) (*model.Guild, bool, error) {
	perm = append(perm, permissions.PermAdministrator)
	guild, err := e.g.GetGuildById(ctx, guildID)
	if err != nil {
		return nil, false, err
	}
	if userID == guild.OwnerId {
		return nil, true, nil
	}
	permAll := guild.Permissions
	ur, err := e.ur.GetUserRoles(ctx, guildID, userID)
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, false, err
	}
	var rids = make([]int64, len(ur))
	for i, r := range ur {
		rids[i] = r.RoleId
	}
	rs, err := e.role.GetRolesBulk(ctx, rids)
	if err != nil && !errors.Is(err, gocql.ErrNotFound) {
		return nil, false, err
	}

	for _, r := range rs {
		permAll = permissions.Allowing(permAll, r.Permissions)
	}
	if permissions.CheckPermissions(permAll, perm...) {
		return &guild, true, nil
	}
	return nil, false, nil
}
