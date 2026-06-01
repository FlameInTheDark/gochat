package developer

import (
	"database/sql"
	"errors"
	"strconv"

	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/gofiber/fiber/v2"
)

func (e *entity) ListBotCommands(c *fiber.Ctx) error {
	_, botID, err := e.requireOwnedBot(c)
	if err != nil {
		return err
	}
	guildID, err := commandGuildID(c)
	if err != nil {
		return err
	}
	commands, err := e.cmd.ListBotCommands(c.UserContext(), botID, guildID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to list application commands")
	}
	return c.JSON(commands)
}

func (e *entity) CreateBotCommand(c *fiber.Ctx) error {
	_, botID, err := e.requireOwnedBot(c)
	if err != nil {
		return err
	}
	guildID, err := commandGuildID(c)
	if err != nil {
		return err
	}
	var req appcmd.ApplicationCommand
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid application command body")
	}
	cmd, err := prepareDeveloperCommand(botID, guildID, req, 0)
	if err != nil {
		return err
	}
	if err := e.cmd.CreateCommand(c.UserContext(), cmd); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	created, err := e.cmd.GetCommandForBot(c.UserContext(), botID, cmd.ID)
	if err != nil {
		return commandError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (e *entity) BulkOverwriteBotCommands(c *fiber.Ctx) error {
	_, botID, err := e.requireOwnedBot(c)
	if err != nil {
		return err
	}
	guildID, err := commandGuildID(c)
	if err != nil {
		return err
	}
	var req []appcmd.ApplicationCommand
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid application command body")
	}
	if len(req) > 200 {
		return fiber.NewError(fiber.StatusBadRequest, "bulk overwrite cannot contain more than 200 commands")
	}
	commands := make([]appcmd.ApplicationCommand, 0, len(req))
	for _, raw := range req {
		cmd, err := prepareDeveloperCommand(botID, guildID, raw, 0)
		if err != nil {
			return err
		}
		commands = append(commands, cmd)
	}
	if err := e.cmd.BulkOverwrite(c.UserContext(), botID, guildID, commands); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	out, err := e.cmd.ListBotCommands(c.UserContext(), botID, guildID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to list application commands")
	}
	return c.JSON(out)
}

func (e *entity) GetBotCommand(c *fiber.Ctx) error {
	_, botID, err := e.requireOwnedBot(c)
	if err != nil {
		return err
	}
	commandID, err := parseIntParam(c, "command_id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errUnableToParseID)
	}
	command, err := e.cmd.GetCommandForBot(c.UserContext(), botID, commandID)
	if err != nil {
		return commandError(err)
	}
	return c.JSON(command)
}

func (e *entity) UpdateBotCommand(c *fiber.Ctx) error {
	_, botID, err := e.requireOwnedBot(c)
	if err != nil {
		return err
	}
	commandID, err := parseIntParam(c, "command_id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errUnableToParseID)
	}
	existing, err := e.cmd.GetCommandForBot(c.UserContext(), botID, commandID)
	if err != nil {
		return commandError(err)
	}
	var req appcmd.ApplicationCommand
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid application command body")
	}
	cmd, err := prepareDeveloperCommand(botID, existing.GuildID, req, commandID)
	if err != nil {
		return err
	}
	if err := e.cmd.ReplaceCommand(c.UserContext(), cmd); err != nil {
		return commandError(err)
	}
	updated, err := e.cmd.GetCommandForBot(c.UserContext(), botID, commandID)
	if err != nil {
		return commandError(err)
	}
	return c.JSON(updated)
}

func (e *entity) DeleteBotCommand(c *fiber.Ctx) error {
	_, botID, err := e.requireOwnedBot(c)
	if err != nil {
		return err
	}
	commandID, err := parseIntParam(c, "command_id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errUnableToParseID)
	}
	if err := e.cmd.DeleteCommand(c.UserContext(), botID, commandID); err != nil {
		return commandError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (e *entity) requireOwnedBot(c *fiber.Ctx) (int64, int64, error) {
	owner, err := helper.GetUser(c)
	if err != nil {
		return 0, 0, fiber.NewError(fiber.StatusUnauthorized, errUnableToGetUser)
	}
	botID, err := parseIntParam(c, "bot_id")
	if err != nil {
		return 0, 0, fiber.NewError(fiber.StatusBadRequest, errUnableToParseID)
	}
	if _, err = e.bot.GetBotForOwner(c.UserContext(), owner.Id, botID); err != nil {
		return 0, 0, fiber.NewError(fiber.StatusNotFound, errUnableToGetBot)
	}
	return owner.Id, botID, nil
}

func commandGuildID(c *fiber.Ctx) (*int64, error) {
	raw := c.Query("guild_id")
	if raw == "" {
		return nil, nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid guild_id")
	}
	return &id, nil
}

func prepareDeveloperCommand(botID int64, guildID *int64, cmd appcmd.ApplicationCommand, existingID int64) (appcmd.ApplicationCommand, error) {
	cmd = appcmd.NormalizeCommand(cmd)
	cmd.ApplicationID = botID
	cmd.GuildID = guildID
	if existingID != 0 {
		cmd.ID = existingID
	} else if cmd.ID == 0 {
		cmd.ID = idgen.Next()
	}
	cmd.Version = idgen.Next()
	if err := appcmd.ValidateCommand(cmd); err != nil {
		return appcmd.ApplicationCommand{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return cmd, nil
}

func commandError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fiber.NewError(fiber.StatusNotFound, "application command not found")
	}
	return fiber.NewError(fiber.StatusInternalServerError, "unable to get application command")
}
