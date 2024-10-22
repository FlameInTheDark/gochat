package user

import (
	"errors"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/gocql/gocql"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

// GetUser
//
//	@Summary	Get user
//	@Produce	json
//	@Tags		User
//	@Param		user_id	path		string			true	"user id or @me"
//	@Success	200		{object}	UserResponse	"User data"
//	@failure	400		{string}	string			"Incorrect ID"
//	@failure	404		{string}	string			"User not found"
//	@Router		/user/{user_id} [get]
func (e *entity) GetUser(c *fiber.Ctx) error {
	id := c.Params("user_id")
	var userId int64
	if id == "@me" {
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
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.JSON(modelToUser(user))
}

// GetUser
//
//	@Summary	Get user guilds
//	@Produce	json
//	@Tags		User
//	@Param		user_id	path		string			true	"user id or @me"
//	@Success	200		{array}		UserGuild		"Guilds list"
//	@failure	400		{string}	string			"Incorrect ID"
//	@failure	404		{string}	string			"User not found"
//	@Router		/user/@me/guilds [get]
func (e *entity) GetUserGuilds(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	ms, err := e.member.GetUserGuilds(c.UserContext(), user.Id)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
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
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.JSON(guildModelToGuildMany(gs, user.Id))
}
