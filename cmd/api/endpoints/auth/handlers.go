package auth

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

// Login
//
//	@Summary	Authentication using email and password
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
		return c.SendStatus(fiber.StatusBadRequest)
	}

	auth, err := e.auth.GetAuthenticationByEmail(c.UserContext(), login.Email)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	err = CompareHashAndPassword(auth.PasswordHash, login.Password)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	user, err := e.user.GetUserById(c.UserContext(), auth.UserId)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	if user.Blocked {
		return c.SendStatus(fiber.StatusUnauthorized)
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
		e.log.Error("unable to sign authentication token", slog.String("error", err.Error()))
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(LoginResponse{Token: t})
}

// Registration
//
//	@Summary	Registration using email address
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
		return c.SendStatus(fiber.StatusBadRequest)
	}
	token, err := helper.RandomToken(40)
	if err != nil {
		e.log.Error("unable to generate token", slog.String("error", err.Error()))
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	_, err = e.auth.GetAuthenticationByEmail(c.UserContext(), req.Email)
	if err == nil {
		return c.SendStatus(fiber.StatusFound)
	} else if !errors.Is(err, gocql.ErrNotFound) {
		e.log.Error("unable to get authentication by email", slog.String("error", err.Error()))
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	var id int64
	reg, err := e.reg.GetRegistrationByEmail(c.UserContext(), req.Email)
	if errors.Is(err, gocql.ErrNotFound) {
		id = e.id.Next()
	} else if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	} else {
		if reg.CreatedAt.Add(time.Minute).After(time.Now()) {
			return c.SendStatus(fiber.StatusTooManyRequests)
		}
		id = reg.UserId
	}
	err = e.reg.CreateRegistration(c.UserContext(), id, reg.Email, token)
	if err != nil {
		e.log.Error("unable to create registration", slog.String("error", err.Error()))
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	err = e.mailer.Send(c.UserContext(), id, req.Email, token)
	if err != nil {
		e.log.Error("unable to send email", slog.String("error", err.Error()))
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.SendStatus(fiber.StatusCreated)
}

// Confirmation
//
//	@Summary	Registration confirmation
//	@Produce	json
//	@Tags		Auth
//	@Param		request	body		ConfirmationRequest	true	"Login data"
//	@Success	201		{string}	string				"Registration completed, account created"
//	@failure	400		{string}	string				"Incorrect request body"
//	@failure	401		{string}	string				"Unauthorized"
//	@failure	500		{string}	string				"Something bad happened"
//	@Router		/auth/confitmation [post]
func (e *entity) Confirmation(c *fiber.Ctx) error {
	var req ConfirmationRequest
	err := c.BodyParser(&req)
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	if len(req.Password) < 8 {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	hash, err := HashPassword(req.Password)
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	i, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	reg, err := e.reg.GetRegistrationByUserId(c.UserContext(), i)
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	if reg.ConfirmationToken != req.Token {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	_, err = e.user.GetUserByDeterminator(c.UserContext(), req.Determinator)
	if err == nil {
		return c.SendStatus(fiber.StatusFound)
	}
	err = e.reg.RemoveRegistration(c.UserContext(), reg.UserId)
	if err != nil {
		e.log.Error("unable to remove registration", slog.String("error", err.Error()))
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	err = e.user.CreateUser(c.UserContext(), i, req.Name, req.Determinator)
	if err != nil {
		e.log.Error("unable to create user", slog.String("error", err.Error()))
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	err = e.auth.CreateAuthentication(c.UserContext(), i, reg.Email, hash)
	if err != nil {
		e.log.Error("unable to create authentication", slog.String("error", err.Error()))
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.SendStatus(fiber.StatusCreated)
}
