package applicationcommand

import (
	"time"

	appcmd "github.com/FlameInTheDark/gochat/internal/applicationcommands"
	"github.com/FlameInTheDark/gochat/internal/botauth"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/gofiber/fiber/v2"
)

func (e *Entity) ListGlobalCommands(c *fiber.Ctx) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	commands, err := e.appcmd.ListBotCommands(c.UserContext(), applicationID, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to list application commands")
	}
	return c.JSON(commands)
}

func (e *Entity) ListGuildCommands(c *fiber.Ctx) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	guildID, err := guildIDParam(c)
	if err != nil {
		return err
	}
	commands, err := e.appcmd.ListBotCommands(c.UserContext(), applicationID, guildID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to list application commands")
	}
	return c.JSON(commands)
}

func (e *Entity) CreateGlobalCommand(c *fiber.Ctx) error {
	return e.createCommand(c, nil)
}

func (e *Entity) CreateGuildCommand(c *fiber.Ctx) error {
	guildID, err := guildIDParam(c)
	if err != nil {
		return err
	}
	return e.createCommand(c, guildID)
}

func (e *Entity) createCommand(c *fiber.Ctx, guildID *int64) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	var req appcmd.ApplicationCommand
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid application command body")
	}
	cmd, err := prepareCommand(applicationID, guildID, req, 0)
	if err != nil {
		return err
	}
	if err := e.appcmd.CreateCommand(c.UserContext(), cmd); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	created, err := e.appcmd.GetCommandForBot(c.UserContext(), applicationID, cmd.ID)
	if err != nil {
		return commandLookupError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (e *Entity) BulkOverwriteGlobalCommands(c *fiber.Ctx) error {
	return e.bulkOverwriteCommands(c, nil)
}

func (e *Entity) BulkOverwriteGuildCommands(c *fiber.Ctx) error {
	guildID, err := guildIDParam(c)
	if err != nil {
		return err
	}
	return e.bulkOverwriteCommands(c, guildID)
}

func (e *Entity) bulkOverwriteCommands(c *fiber.Ctx, guildID *int64) error {
	_, applicationID, err := e.requireApplication(c)
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
		cmd, err := prepareCommand(applicationID, guildID, raw, 0)
		if err != nil {
			return err
		}
		commands = append(commands, cmd)
	}
	if err := e.appcmd.BulkOverwrite(c.UserContext(), applicationID, guildID, commands); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	out, err := e.appcmd.ListBotCommands(c.UserContext(), applicationID, guildID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to list application commands")
	}
	return c.JSON(out)
}

func (e *Entity) GetCommand(c *fiber.Ctx) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	commandID, err := parseInt64Param(c, "command_id")
	if err != nil {
		return err
	}
	command, err := e.appcmd.GetCommandForBot(c.UserContext(), applicationID, commandID)
	if err != nil {
		return commandLookupError(err)
	}
	return c.JSON(command)
}

func (e *Entity) EditCommand(c *fiber.Ctx) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	commandID, err := parseInt64Param(c, "command_id")
	if err != nil {
		return err
	}
	existing, err := e.appcmd.GetCommandForBot(c.UserContext(), applicationID, commandID)
	if err != nil {
		return commandLookupError(err)
	}
	var req appcmd.ApplicationCommand
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid application command body")
	}
	guildID := existing.GuildID
	cmd, err := prepareCommand(applicationID, guildID, req, commandID)
	if err != nil {
		return err
	}
	if err := e.appcmd.ReplaceCommand(c.UserContext(), cmd); err != nil {
		return commandLookupError(err)
	}
	updated, err := e.appcmd.GetCommandForBot(c.UserContext(), applicationID, commandID)
	if err != nil {
		return commandLookupError(err)
	}
	return c.JSON(updated)
}

func (e *Entity) DeleteCommand(c *fiber.Ctx) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	commandID, err := parseInt64Param(c, "command_id")
	if err != nil {
		return err
	}
	if err := e.appcmd.DeleteCommand(c.UserContext(), applicationID, commandID); err != nil {
		return commandLookupError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (e *Entity) RespondInteraction(c *fiber.Ctx) error {
	principal, ok := botauthFromCtx(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "missing bot principal")
	}
	token := c.Params("interaction_token")
	record, err := e.loadInteraction(c, principal.BotUserID, token)
	if err != nil {
		return err
	}
	if err := requireInteractionID(c, record); err != nil {
		return err
	}
	if time.Now().After(record.CreatedAt.Add(appcmd.InteractionDeadline)) {
		return fiber.NewError(fiber.StatusNotFound, "unknown interaction")
	}
	var response appcmd.InteractionResponse
	if err := c.BodyParser(&response); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid interaction response")
	}
	switch response.Type {
	case appcmd.ResponseTypePong:
		if err := e.appcmd.AckInteraction(c.UserContext(), record.ID, appcmd.AckStateResponded, nil); err != nil {
			return duplicateAckError()
		}
		return c.SendStatus(fiber.StatusNoContent)
	case appcmd.ResponseTypeDeferredChannelMessageSource:
		if response.Data != nil && appcmd.HasResponseFlag(response.Data, appcmd.MessageFlagEphemeral) {
			_ = e.setEphemeralResponse(c.UserContext(), record.ID, &appcmd.InteractionResponseData{Flags: response.Data.Flags | appcmd.MessageFlagLoading})
		}
		if err := e.appcmd.AckInteraction(c.UserContext(), record.ID, appcmd.AckStateDeferred, nil); err != nil {
			return duplicateAckError()
		}
		return c.SendStatus(fiber.StatusNoContent)
	case appcmd.ResponseTypeChannelMessageWithSource:
		var messageID *int64
		if appcmd.HasResponseFlag(response.Data, appcmd.MessageFlagEphemeral) {
			if err := e.setEphemeralResponse(c.UserContext(), record.ID, response.Data); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "unable to store ephemeral response")
			}
		} else {
			msg, err := e.createPublicInteractionMessage(c, record, response.Data)
			if err != nil {
				return err
			}
			id := msg.Id
			messageID = &id
		}
		if err := e.appcmd.AckInteraction(c.UserContext(), record.ID, appcmd.AckStateResponded, messageID); err != nil {
			return duplicateAckError()
		}
		return c.SendStatus(fiber.StatusNoContent)
	case appcmd.ResponseTypeAutocompleteResult:
		if response.Data == nil {
			response.Data = &appcmd.InteractionResponseData{}
		}
		if len(response.Data.Choices) > 25 {
			return fiber.NewError(fiber.StatusBadRequest, "autocomplete responses cannot contain more than 25 choices")
		}
		if err := e.setAutocompleteResponse(c.UserContext(), record.ID, response.Data.Choices); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "unable to store autocomplete response")
		}
		if err := e.appcmd.AckInteraction(c.UserContext(), record.ID, appcmd.AckStateResponded, nil); err != nil {
			return duplicateAckError()
		}
		return c.SendStatus(fiber.StatusNoContent)
	case appcmd.ResponseTypeModal:
		if err := e.appcmd.AckInteraction(c.UserContext(), record.ID, appcmd.AckStateResponded, nil); err != nil {
			return duplicateAckError()
		}
		return c.JSON(response)
	default:
		return fiber.NewError(fiber.StatusBadRequest, "unsupported interaction response type")
	}
}

func (e *Entity) GetOriginalInteractionResponse(c *fiber.Ctx) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	record, err := e.loadInteraction(c, applicationID, c.Params("interaction_token"))
	if err != nil {
		return err
	}
	if data, ok, err := e.getEphemeralResponse(c.UserContext(), record.ID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to load ephemeral response")
	} else if ok {
		return c.JSON(data)
	}
	if record.InitialResponseID == nil {
		return fiber.NewError(fiber.StatusNotFound, "original interaction response not found")
	}
	out, err := e.getMessageResponse(c, record, *record.InitialResponseID)
	if err != nil {
		return err
	}
	return c.JSON(out)
}

func (e *Entity) EditOriginalInteractionResponse(c *fiber.Ctx) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	record, err := e.loadInteraction(c, applicationID, c.Params("interaction_token"))
	if err != nil {
		return err
	}
	var data appcmd.InteractionResponseData
	if err := c.BodyParser(&data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid interaction response body")
	}
	if appcmd.HasResponseFlag(&data, appcmd.MessageFlagEphemeral) {
		if err := e.setEphemeralResponse(c.UserContext(), record.ID, &data); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "unable to store ephemeral response")
		}
		return c.JSON(data)
	}
	if record.InitialResponseID == nil {
		if _, ok, err := e.getEphemeralResponse(c.UserContext(), record.ID); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "unable to load ephemeral response")
		} else if ok {
			if err := e.setEphemeralResponse(c.UserContext(), record.ID, &data); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "unable to store ephemeral response")
			}
			return c.JSON(data)
		}
		out, err := e.createPublicInteractionMessage(c, record, &data)
		if err != nil {
			return err
		}
		if err := e.appcmd.SetInitialResponse(c.UserContext(), record.ID, out.Id); err != nil {
			_ = e.deleteMessageResponse(c, record, out.Id)
			return duplicateAckError()
		}
		return c.JSON(out)
	}
	out, err := e.editMessageResponse(c, record, *record.InitialResponseID, &data)
	if err != nil {
		return err
	}
	return c.JSON(out)
}

func (e *Entity) DeleteOriginalInteractionResponse(c *fiber.Ctx) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	record, err := e.loadInteraction(c, applicationID, c.Params("interaction_token"))
	if err != nil {
		return err
	}
	e.deleteEphemeralResponse(c.UserContext(), record.ID)
	if record.InitialResponseID != nil {
		if err := e.deleteMessageResponse(c, record, *record.InitialResponseID); err != nil {
			return err
		}
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (e *Entity) CreateFollowupMessage(c *fiber.Ctx) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	record, err := e.loadInteraction(c, applicationID, c.Params("interaction_token"))
	if err != nil {
		return err
	}
	var data appcmd.InteractionResponseData
	if err := c.BodyParser(&data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid followup body")
	}
	if appcmd.HasResponseFlag(&data, appcmd.MessageFlagEphemeral) {
		if err := e.setEphemeralResponse(c.UserContext(), idgen.Next(), &data); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "unable to store ephemeral followup")
		}
		return c.Status(fiber.StatusCreated).JSON(data)
	}
	out, err := e.createPublicInteractionMessage(c, record, &data)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(out)
}

func (e *Entity) EditFollowupMessage(c *fiber.Ctx) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	record, err := e.loadInteraction(c, applicationID, c.Params("interaction_token"))
	if err != nil {
		return err
	}
	messageID, err := parseInt64Param(c, "message_id")
	if err != nil {
		return err
	}
	var data appcmd.InteractionResponseData
	if err := c.BodyParser(&data); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid followup body")
	}
	out, err := e.editMessageResponse(c, record, messageID, &data)
	if err != nil {
		return err
	}
	return c.JSON(out)
}

func (e *Entity) DeleteFollowupMessage(c *fiber.Ctx) error {
	_, applicationID, err := e.requireApplication(c)
	if err != nil {
		return err
	}
	record, err := e.loadInteraction(c, applicationID, c.Params("interaction_token"))
	if err != nil {
		return err
	}
	messageID, err := parseInt64Param(c, "message_id")
	if err != nil {
		return err
	}
	if err := e.deleteMessageResponse(c, record, messageID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func botauthFromCtx(c *fiber.Ctx) (*botauth.Principal, bool) {
	principal, ok := botauth.FromFiber(c)
	return principal, ok
}
