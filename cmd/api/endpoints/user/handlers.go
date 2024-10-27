package user

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
)

// GetUser
//
//	@Summary	Get user
//	@Produce	json
//	@Tags		User
//	@Param		user_id	path		string		true	"User ID or 'me'"
//	@Success	200		{object}	dto.User	"User data"
//	@failure	400		{string}	string		"Incorrect ID"
//	@failure	404		{string}	string		"User not found"
//	@failure	500		{string}	string		"Something bad happened"
//	@Router		/user/{user_id} [get]
func (e *entity) GetUser(c *fiber.Ctx) error {
	id := c.Params("user_id")
	var userId int64
	if id == "me" {
		user, err := helper.GetUser(c)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		userId = user.Id
	} else {
		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		userId = i
	}

	user, err := e.user.GetUserById(c.UserContext(), userId)
	if err := helper.HttpDbError(err, ErrUnableToGetUser); err != nil {
		return err
	}

	disc, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), userId)
	if err := helper.HttpDbError(err, ErrUnableToGetDiscriminator); err != nil {
		return err
	}

	u := modelToUser(user)
	u.Discriminator = disc.Discriminator

	return c.JSON(u)
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
//	@Router		/me [patch]
func (e *entity) ModifyUser(c *fiber.Ctx) error {
	var req ModifyUserRequest
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	err = e.user.ModifyUser(c.UserContext(), user.Id, req.Name, req.Avatar)
	if err := helper.HttpDbError(err, ErrUnableToModifyUser); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

// GetUserGuilds
//
//	@Summary	Get user guilds
//	@Produce	json
//	@Tags		User
//	@Success	200	{array}		dto.Guild	"Guilds list"
//	@failure	400	{string}	string		"Incorrect ID"
//	@failure	404	{string}	string		"User not found"
//	@failure	500	{string}	string		"Something bad happened"
//	@Router		/user/me/guilds [get]
func (e *entity) GetUserGuilds(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	ms, err := e.member.GetUserGuilds(c.UserContext(), user.Id)
	if err := helper.HttpDbError(err, ErrUnableToGetUserGuilds); err != nil {
		return err
	}
	var ids = make([]int64, len(ms))
	for i, m := range ms {
		ids[i] = m.GuildId
	}
	var guilds []model.Guild
	gs, err := e.guild.GetGuildsList(c.UserContext(), ids)
	if errors.Is(err, gocql.ErrNotFound) {
		return c.JSON(guilds)
	} else if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetGuilds)
	}
	return c.JSON(guildModelToGuildMany(gs, user.Id))
}

// GetMyGuildMember
//
//	@Summary	Get user guild member
//	@Produce	json
//	@Tags		User
//	@Param		guild_id	path		string		true	"Guild id"
//	@Success	200			{object}	dto.Member	"Guild member"
//	@failure	400			{string}	string		"Incorrect ID"
//	@failure	404			{string}	string		"User not found"
//	@failure	500			{string}	string		"Something bad happened"
//	@Router		/user/me/guilds/{guild_id}/member [get]
func (e *entity) GetMyGuildMember(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	id := c.Params("guild_id")
	guildId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrBadRequest)
	}
	m, err := e.member.GetMember(c.UserContext(), user.Id, guildId)
	if err := helper.HttpDbError(err, ErrUnableToGetMember); err != nil {
		return err
	}
	u, err := e.user.GetUserById(c.UserContext(), user.Id)
	if err := helper.HttpDbError(err, ErrUnableToGetUser); err != nil {
		return err
	}
	disc, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), user.Id)
	if err := helper.HttpDbError(err, ErrUnableToGetDiscriminator); err != nil {
		return err
	}
	r, err := e.urole.GetUserRoles(c.UserContext(), guildId, user.Id)
	if err := helper.HttpDbError(err, ErrUnableToGetRoles); err != nil {
		return err
	}
	ids := make([]int64, len(r))
	for i, r := range r {
		ids[i] = r.RoleId
	}
	return c.JSON(dto.Member{
		User: dto.User{
			Id:            u.Id,
			Name:          u.Name,
			Discriminator: disc.Discriminator,
			Avatar:        u.Avatar,
		},
		Username: m.Username,
		Avatar:   m.Avatar,
		JoinAt:   m.JoinAt,
		Roles:    ids,
	})
}

// LeaveGuild
//
//	@Summary	Leave guild
//	@Produce	json
//	@Tags		User
//	@Param		guild_id	path		string	true	"Guild id"
//	@Success	200			{string}	string	"ok"
//	@failure	400			{string}	string	"Incorrect ID"
//	@failure	404			{string}	string	"User not found"
//	@failure	406			{string}	string	"Unable to leave your guild"
//	@failure	500			{string}	string	"Something bad happened"
//	@Router		/user/me/guilds/{guild_id} [delete]
func (e *entity) LeaveGuild(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUser)
	}
	id := c.Params("guild_id")
	guildId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseID)
	}
	g, err := e.guild.GetGuildById(c.UserContext(), guildId)
	if err := helper.HttpDbError(err, ErrUnableToGetGuildByID); err != nil {
		return err
	}
	if g.OwnerId == user.Id {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrUnableToLeaveOwnServer)
	}
	err = e.member.RemoveMember(c.UserContext(), user.Id, guildId)
	if err := helper.HttpDbError(err, ErrUnableToRemoveMember); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}

// CreateDM
//
//	@Summary	Create DM channel
//	@Produce	json
//	@Tags		User
//	@Param		request	body		CreateDMRequest	true	"Recipient data"
//	@Success	200		{string}	string			"ok"
//	@failure	400		{string}	string			"Incorrect ID"
//	@failure	404		{string}	string			"User not found"
//	@failure	500		{string}	string			"Something bad happened"
//	@Router		/user/me/channels [post]
func (e *entity) CreateDM(c *fiber.Ctx) error {
	var req CreateDMRequest
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	rec, err := e.user.GetUserById(c.UserContext(), req.RecipientId)
	if err := helper.HttpDbError(err, ErrUnableToGetUser); err != nil {
		return err
	}
	rc, err := e.dm.GetDmChannel(c.UserContext(), user.Id, rec.Id)
	if errors.Is(err, gocql.ErrNotFound) {
		chId := idgen.Next()
		err = e.ch.CreateChannel(c.UserContext(), chId, "", model.ChannelTypeDM, nil, 0)
		if err := helper.HttpDbError(err, ErrUnableToCreateChannel); err != nil {
			return err
		}
		err = e.dm.CreateDmChannel(c.UserContext(), user.Id, req.RecipientId, chId)
		if err := helper.HttpDbError(err, ErrUnableToCreateDMChannel); err != nil {
			return err
		}
		return c.JSON(dto.Channel{
			Id:          chId,
			Type:        model.ChannelTypeDM,
			GuildId:     nil,
			Name:        "",
			ParentId:    nil,
			Permissions: 0,
			CreatedAt:   time.Now(),
		})
	} else if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetDMChannel)
	}
	ch, err := e.ch.GetChannel(c.UserContext(), rc.ChannelId)
	if err := helper.HttpDbError(err, ErrUnableToGetChannel); err != nil {
		return err
	}
	return c.JSON(dto.Channel{
		Id:          ch.Id,
		Type:        ch.Type,
		GuildId:     ch.GuildId,
		Name:        ch.Name,
		ParentId:    ch.ParentID,
		Permissions: ch.Permissions,
		CreatedAt:   ch.CreatedAt,
	})
}

// CreateDM
//
//	@Summary	Create group DM channel
//	@Produce	json
//	@Tags		User
//	@Param		request	body		CreateDMManyRequest	true	"Group DM data"
//	@Success	200		{string}	string				"ok"
//	@failure	400		{string}	string				"Incorrect ID"
//	@failure	404		{string}	string				"Not found"
//	@failure	500		{string}	string				"Something bad happened"
//	@Router		/user/me/channels/group [post]
func (e *entity) CreateGroupDM(c *fiber.Ctx) error {
	var req CreateDMManyRequest
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	var ch dto.Channel
	if req.ChannelId != nil {
		is, err := e.gdm.IsGroupDmParticipant(c.UserContext(), *req.ChannelId, user.Id)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGroupDMChannel)
		}
		if is {
			uch, err := e.ch.GetChannel(c.UserContext(), *req.ChannelId)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
			}
			ch = dto.Channel{
				Id:          uch.Id,
				Type:        uch.Type,
				GuildId:     uch.GuildId,
				Name:        uch.Name,
				ParentId:    uch.ParentID,
				Permissions: uch.Permissions,
				CreatedAt:   uch.CreatedAt,
			}
		}
	}
	id := idgen.Next()
	err = e.ch.CreateChannel(c.UserContext(), id, "", model.ChannelTypeGroupDM, nil, 0)
	if err := helper.HttpDbError(err, ErrUnableToCreateChannel); err != nil {
		return err
	}
	ch = dto.Channel{
		Id:          id,
		Type:        model.ChannelTypeGroupDM,
		GuildId:     nil,
		Name:        "",
		ParentId:    nil,
		Permissions: 0,
		CreatedAt:   time.Now(),
	}
	err = e.gdm.JoinGroupDmChannelMany(c.UserContext(), id, req.RecipientsId)
	if err := helper.HttpDbError(err, ErrUnableToJoingGroupDmChannel); err != nil {
		return err
	}
	return c.JSON(ch)
}
