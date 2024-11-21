package guild

import (
	"errors"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

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
	guildId := c.Params("guild_id")
	id, err := strconv.ParseInt(guildId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	m, err := e.memb.GetMember(c.UserContext(), u.Id, id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMember)
	}
	guild, err := e.g.GetGuildById(c.UserContext(), m.GuildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(dto.Guild{
		Id:     guild.Id,
		Name:   guild.Name,
		Icon:   guild.Icon,
		Owner:  guild.OwnerId,
		Public: guild.Public,
	})
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
	guildId := c.Params("guild_id")
	id, err := strconv.ParseInt(guildId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	m, err := e.memb.GetMember(c.UserContext(), u.Id, id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMember)
	}
	guild, err := e.g.GetGuildById(c.UserContext(), m.GuildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	urs, err := e.ur.GetUserRoles(c.UserContext(), guild.Id, u.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	var rids []int64
	for _, ur := range urs {
		rids = append(rids, ur.RoleId)
	}
	roles, err := e.role.GetRolesBulk(c.UserContext(), rids)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	var rm = make(map[int64]*model.Role)
	for i, ur := range roles {
		rm[ur.Id] = &roles[i]
	}
	gch, err := e.gc.GetGuildChannels(c.UserContext(), guild.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	var chids []int64
	var chm = make(map[int64]*model.GuildChannel)
	for i, ch := range gch {
		chm[ch.ChannelId] = &gch[i]
		chids = append(chids, ch.ChannelId)
	}

	chs, err := e.ch.GetChannelsBulk(c.UserContext(), chids)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var channels []dto.Channel
	for _, ch := range chs {
		if gc, ok := chm[ch.Id]; ok {
			if ch.Permissions == nil {
				ch.Permissions = &guild.Permissions
			}
			if ch.Private && guild.OwnerId != u.Id {
				cr, err := e.uperm.GetUserChannelPermission(c.UserContext(), ch.Id, u.Id)
				if err != nil {
					return fiber.NewError(fiber.StatusInternalServerError, err.Error())
				} else if errors.Is(err, gocql.ErrNotFound) {
					crs, err := e.rperm.GetChannelRolePermissions(c.UserContext(), ch.Id)
					if err != nil {
						return fiber.NewError(fiber.StatusInternalServerError, err.Error())
					}
					var role int64
					for _, r := range crs {
						if ur, ok := rm[r.RoleId]; ok {
							role = permissions.AddRoles(role, ur.Permissions)
						}
					}
					if !permissions.CheckPermissions(role, permissions.PermServerViewChannels) {
						continue
					}
				}
				if !permissions.CheckPermissions(permissions.SubtractRoles(permissions.AddRoles(*ch.Permissions, cr.Accept), cr.Deny), permissions.PermServerViewChannels) {
					continue
				}
			}
			channels = append(channels, channelModelToDTO(&ch, &id, gc.Position))
		}
	}
	return c.JSON(channels)
}

// Create
//
//	@Summary	Create guild
//	@Produce	json
//	@Tags		Guild
//	@Param		request	body		UpdateGuildRequest	true	"Guild data"
//	@Success	200		{object}	dto.Guild			"Guild"
//	@failure	400		{string}	string				"Incorrect request body"
//	@failure	401		{string}	string				"Unauthorized"
//	@failure	500		{string}	string				"Something bad happened"
//	@Router		/guild [post]
func (e *entity) Create(c *fiber.Ctx) error {
	var req CreateGuildRequest
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	gid := idgen.Next()
	err = e.g.CreateGuild(c.UserContext(), gid, req.Name, u.Id, permissions.DefaultPermissions)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateGuild)
	}

	if req.IconId != nil {
		icon, err := e.icon.GetIcon(c.UserContext(), *req.IconId)
		if err == nil {
			err = e.g.SetGuildIcon(c.UserContext(), gid, icon.Id)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateGuild)
			}
		}
	}

	chcatid := idgen.Next()
	chid := idgen.Next()
	err = e.ch.CreateChannel(c.UserContext(), chcatid, "text", model.ChannelTypeGuildCategory, nil, nil, req.Public)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateGuild)
	}
	err = e.ch.CreateChannel(c.UserContext(), chid, "general", model.ChannelTypeGuild, &chcatid, nil, req.Public)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateGuild)
	}
	err = e.gc.AddChannel(c.UserContext(), gid, chcatid, 0)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateGuild)
	}
	err = e.gc.AddChannel(c.UserContext(), gid, chid, 0)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateGuild)
	}
	err = e.memb.AddMember(c.UserContext(), u.Id, gid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateGuild)
	}

	return c.JSON(dto.Guild{
		Id:     gid,
		Name:   req.Name,
		Icon:   req.IconId,
		Owner:  u.Id,
		Public: req.Public,
	})
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
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	guildId := c.Params("guild_id")
	id, err := strconv.ParseInt(guildId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	g, ok, err := e.perm.GuildPerm(c.UserContext(), id, u.Id, permissions.PermServerManage)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if ok {
		err := e.g.UpdateGuild(c.UserContext(), id, req.Name, req.IconId, req.Public)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUpdateGuild)
		}
		err = e.mqt.SendChannelMessage(id, &mqmsg.UpdateGuild{
			Guild: dto.Guild{
				Id:     g.Id,
				Name:   "",
				Icon:   nil,
				Owner:  0,
				Public: false,
			},
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
		}
		return c.SendStatus(fiber.StatusOK)
	}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
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
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	guildId := c.Params("guild_id")
	id, err := strconv.ParseInt(guildId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	g, ok, err := e.perm.GuildPerm(c.UserContext(), id, u.Id, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if ok {
		chid := idgen.Next()
		err := e.ch.CreateChannel(c.UserContext(), chid, req.Name, model.ChannelTypeGuildCategory, nil, nil, req.Private)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
		}

		err = e.gc.AddChannel(c.UserContext(), g.Id, chid, 0)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
		}

		err = e.mqt.SendGuildUpdate(id, &mqmsg.CreateChannel{
			GuildId: &g.Id,
			Channel: dto.Channel{
				Id:        chid,
				Type:      model.ChannelTypeGuildCategory,
				GuildId:   &g.Id,
				Name:      req.Name,
				ParentId:  nil,
				Position:  0,
				Topic:     nil,
				CreatedAt: time.Now(),
			},
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
		}
		return c.SendStatus(fiber.StatusCreated)
	}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
}

// CreateChannel
//
//	@Summary	Add guild channel category
//	@Produce	json
//	@Tags		Guild
//	@Param		request		body		CreateGuildChannelCategoryRequest	true	"Create category data"
//	@Param		guild_id	path		int64								true	"Guild ID"
//	@Success	201			{object}	string								"Created"
//	@failure	400			{string}	string								"Incorrect request body"
//	@failure	401			{string}	string								"Unauthorized"
//	@failure	500			{string}	string								"Something bad happened"
//	@Router		/guild/{guild_id}/channel [post]
func (e *entity) CreateChannel(c *fiber.Ctx) error {
	var req CreateGuildChannelRequest
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	guildId := c.Params("guild_id")
	id, err := strconv.ParseInt(guildId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	g, ok, err := e.perm.GuildPerm(c.UserContext(), id, u.Id, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if ok {
		chid := idgen.Next()
		err := e.ch.CreateChannel(c.UserContext(), chid, req.Name, req.Type, req.ParentId, nil, req.Private)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
		}

		err = e.gc.AddChannel(c.UserContext(), g.Id, chid, 0)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
		}

		err = e.mqt.SendGuildUpdate(id, &mqmsg.CreateChannel{
			GuildId: &g.Id,
			Channel: dto.Channel{
				Id:        chid,
				Type:      req.Type,
				GuildId:   &g.Id,
				Name:      req.Name,
				ParentId:  req.ParentId,
				Position:  0,
				Topic:     nil,
				CreatedAt: time.Now(),
			},
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
		}
		return c.SendStatus(fiber.StatusCreated)
	}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
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
	guild := c.Params("guild_id")
	gid, err := strconv.ParseInt(guild, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}
	channelId := c.Params("channel_id")
	chid, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	ch, _, _, ok, err := e.perm.ChannelPerm(c.UserContext(), gid, chid, u.Id, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if ok && ch.Type != model.ChannelTypeGuildCategory {
		err = e.ch.DeleteChannel(c.UserContext(), chid)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		err = e.msg.DeleteChannelMessages(c.UserContext(), chid)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		err = e.mqt.SendGuildUpdate(gid, &mqmsg.DeleteChannel{
			GuildId:     &gid,
			ChannelType: ch.Type,
			ChannelId:   ch.Id,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
		}
		return c.SendStatus(fiber.StatusOK)
	}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
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
	guild := c.Params("guild_id")
	gid, err := strconv.ParseInt(guild, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectGuildID)
	}
	channelId := c.Params("category_id")
	chid, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrIncorrectChannelID)
	}
	u, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	ch, _, g, ok, err := e.perm.ChannelPerm(c.UserContext(), gid, chid, u.Id, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if ok && ch.Type == model.ChannelTypeGuildCategory {
		gch, err := e.gc.GetGuildChannels(c.UserContext(), g.Id)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		var chids []int64
		for _, ch := range gch {
			chids = append(chids, ch.ChannelId)
		}
		chans, err := e.ch.GetChannelsBulk(c.UserContext(), chids)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		chids = []int64{}
		for _, ch := range chans {
			if ch.ParentID != nil && *ch.ParentID == chid {
				chids = append(chids, ch.Id)
			}
		}
		if len(chids) > 0 {
			err = e.ch.SetChannelParentBulk(c.UserContext(), chids, nil)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
			err = e.gc.ResetGuildChannelPositionBulk(c.UserContext(), chids, g.Id)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
		}
		err = e.ch.DeleteChannel(c.UserContext(), ch.Id)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		err = e.mqt.SendGuildUpdate(g.Id, &mqmsg.DeleteChannel{
			GuildId:     &gid,
			ChannelType: ch.Type,
			ChannelId:   ch.Id,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateChannelGroup)
		}
		return c.SendStatus(fiber.StatusOK)
	}
	return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
}
