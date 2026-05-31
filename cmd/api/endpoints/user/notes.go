package user

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

const maxUserNoteLength = 500

type UpsertUserNoteRequest struct {
	Note string `json:"note" example:"Met during the release party"`
}

func (r UpsertUserNoteRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Note,
			validation.RuneLength(0, maxUserNoteLength).Error(ErrUserNoteTooLong),
		),
	)
}

// UpsertUserNote
//
//	@Summary	Save a private note for another user
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		user_id	path		int64					true	"Target user ID"
//	@Param		request	body		UpsertUserNoteRequest	true	"Private note"
//	@Success	200		{string}	string					"OK"
//	@Router		/user/me/notes/{user_id} [put]
func (e *entity) UpsertUserNote(c *fiber.Ctx) error {
	targetId, err := strconv.ParseInt(c.Params("user_id"), 10, 64)
	if err != nil || targetId <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseID)
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	if targetId == user.Id {
		return fiber.NewError(fiber.StatusBadRequest, ErrCannotNoteSelf)
	}

	var req UpsertUserNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseRequestBody)
	}
	if err := req.Validate(); err != nil {
		return badRequestValidationError(err)
	}

	note := strings.TrimSpace(req.Note)
	if note == "" {
		if err := e.notes.DeleteNote(c.UserContext(), user.Id, targetId); err != nil {
			return helper.HttpDbError(err, ErrUnableToSaveUserNote)
		}
		return c.SendStatus(fiber.StatusOK)
	}

	if _, err := e.user.GetUserById(c.UserContext(), targetId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetUser)
		}
		return helper.HttpDbError(err, ErrUnableToGetUser)
	}

	if err := e.notes.UpsertNote(c.UserContext(), user.Id, targetId, note); err != nil {
		return helper.HttpDbError(err, ErrUnableToSaveUserNote)
	}
	return c.SendStatus(fiber.StatusOK)
}

// DeleteUserNote
//
//	@Summary	Delete a private note for another user
//	@Tags		User
//	@Param		user_id	path		int64	true	"Target user ID"
//	@Success	200		{string}	string	"OK"
//	@Router		/user/me/notes/{user_id} [delete]
func (e *entity) DeleteUserNote(c *fiber.Ctx) error {
	targetId, err := strconv.ParseInt(c.Params("user_id"), 10, 64)
	if err != nil || targetId <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseID)
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	if err := e.notes.DeleteNote(c.UserContext(), user.Id, targetId); err != nil {
		return helper.HttpDbError(err, ErrUnableToSaveUserNote)
	}
	return c.SendStatus(fiber.StatusOK)
}
