package auth

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
)

// Login
//
//	@Summary	Authentication
//	@Produce	json
//	@Tags		Auth
//	@Param		request	body		LoginRequest	true	"Login data"
//	@Success	200		{object}	LoginResponse
//	@failure	400		{string}	string	"Incorrect request body"
//	@failure	401		{string}	string	"Unauthorized"
//	@failure	500		{string}	string	"Something bad happened"
//	@Router		/auth/login [post]
func (e *entity) Login(c *fiber.Ctx) error {
	var login LoginRequest
	err := c.BodyParser(&login)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	auth, err := e.auth.GetAuthenticationByEmail(c.UserContext(), login.Email)
	if err := helper.HttpDbError(err, ErrUnableToGetAuthenticationByEmail); err != nil {
		return err
	}

	err = CompareHashAndPassword(auth.PasswordHash, login.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, ErrUnableToCompareHash)
	}

	user, err := e.user.GetUserById(c.UserContext(), auth.UserId)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, ErrUnableToGetUserById)
	}
	if user.Blocked {
		return fiber.NewError(fiber.StatusUnauthorized, ErrUserIsBanned)
	}

	// Create the Claims
	claims := jwt.MapClaims{
		"name": user.Name,
		"id":   user.Id,
		"exp":  time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(e.secret))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSignAuthenticationToken)
	}

	return c.JSON(LoginResponse{Token: t})
}

// Registration
//
//	@Summary	Registration
//	@Produce	json
//	@Tags		Auth
//	@Param		request	body		RegisterRequest	true	"Login data"
//	@Success	201		{string}	string			"Registration email sent"
//	@failure	302		{string}	string			"User already exist"
//	@failure	400		{string}	string			"Incorrect request body"
//	@failure	429		{string}	string			"Try again later"
//	@failure	500		{string}	string			"Something bad happened"
//	@Router		/auth/registration [post]
func (e *entity) Registration(c *fiber.Ctx) error {
	var req RegisterRequest
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	token, err := helper.RandomToken(40)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGenerateToken)
	}
	_, err = e.auth.GetAuthenticationByEmail(c.UserContext(), req.Email)
	if err == nil {
		return c.SendStatus(fiber.StatusFound)
	} else if !errors.Is(err, sql.ErrNoRows) {
		e.log.Error("unable to get authentication by email", slog.String("error", err.Error()))
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetAuthenticationByEmail)
	}
	var id int64
	reg, err := e.reg.GetRegistrationByEmail(c.UserContext(), req.Email)
	if errors.Is(err, sql.ErrNoRows) {
		id = idgen.Next()
	} else if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, ErrUnableToGetRegistrationByEmail)
	} else {
		if reg.CreatedAt.Add(time.Minute).After(time.Now()) {
			return fiber.NewError(fiber.StatusTooManyRequests, ErrEmailAlreadySent)
		}
		id = reg.UserId
	}
	err = e.reg.CreateRegistration(c.UserContext(), id, req.Email, token)
	if err := helper.HttpDbError(err, ErrUnableToCreateRegistration); err != nil {
		return err
	}
	err = e.mailer.Send(c.UserContext(), id, req.Email, token)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendEmail)
	}
	return c.SendStatus(fiber.StatusCreated)
}

// Confirmation
//
//	@Summary	Confirmation
//	@Produce	json
//	@Tags		Auth
//	@Param		request	body		ConfirmationRequest	true	"Login data"
//	@Success	201		{string}	string				"Registration completed, account created"
//	@failure	400		{string}	string				"Incorrect request body"
//	@failure	401		{string}	string				"Unauthorized"
//	@failure	409		{string}	string				"Discriminator is not unique"
//	@failure	500		{string}	string				"Something bad happened"
//	@Router		/auth/confitmation [post]
func (e *entity) Confirmation(c *fiber.Ctx) error {
	var req ConfirmationRequest
	err := c.BodyParser(&req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	if len(req.Password) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, ErrPasswordIsTooShort)
	}
	hash, err := HashPassword(req.Password)
	if err := helper.HttpDbError(err, ErrUnableToGetPasswordHash); err != nil {
		return err
	}
	reg, err := e.reg.GetRegistrationByUserId(c.UserContext(), req.Id)
	if err := helper.HttpDbError(err, ErrUnableToGetRegistrationById); err != nil {
		return err
	}
	if reg.ConfirmationToken != req.Token {
		return fiber.NewError(fiber.StatusUnauthorized, ErrTokenIsIncorrect)
	}
	_, err = e.disc.GetUserIdByDiscriminator(c.UserContext(), req.Discriminator)
	if err == nil {
		return fiber.NewError(fiber.StatusConflict, ErrDiscriminatorIsNotUnique)
	}
	err = e.reg.RemoveRegistration(c.UserContext(), reg.UserId)
	if err := helper.HttpDbError(err, ErrUnableToRemoveRegistration); err != nil {
		return err
	}
	err = e.user.CreateUser(c.UserContext(), req.Id, req.Name)
	if err := helper.HttpDbError(err, ErrUnableToCreateUser); err != nil {
		return err
	}
	err = e.disc.CreateDiscriminator(c.UserContext(), req.Id, req.Discriminator)
	if err := helper.HttpDbError(err, ErrUnableToCreateDiscriminator); err != nil {
		return err
	}
	err = e.auth.CreateAuthentication(c.UserContext(), req.Id, reg.Email, hash)
	if err != nil {
		e.log.Error("unable to create authentication", slog.String("error", err.Error()))
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.SendStatus(fiber.StatusCreated)
}
