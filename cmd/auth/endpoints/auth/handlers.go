package auth

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mailer"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/gofiber/fiber/v2"
)

// Login
//
//	@Summary		Authentication
//	@Description	Validates email and password. Returns tokens immediately when two-factor auth is disabled, or a login challenge when a second factor is required.
//	@Accept			json
//	@Produce		json
//	@Tags			Auth
//	@Param			request	body		LoginRequest			true	"Login data"
//	@Success		200		{object}	LoginResponse			"Authenticated without second factor"
//	@Success		202		{object}	LoginChallengeResponse	"Second-factor challenge created"
//	@failure		400		{string}	string					"Incorrect request body"
//	@failure		401		{string}	string					"Unauthorized"
//	@failure		429		{string}	string					"Too many authentication attempts"
//	@failure		500		{string}	string					"Something bad happened"
//	@Router			/auth/login [post]
func (e *entity) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := e.parseAndValidate(c, "login", &req); err != nil {
		return err
	}
	if err := e.loginRateLimit(c.UserContext(), req.Email, c.IP()); err != nil {
		return err
	}
	auth, err := e.auth.GetAuthenticationByEmail(c.UserContext(), req.Email)
	if err := helper.HttpDbError(err, ErrUnableToGetAuthenticationByEmail); err != nil {
		return err
	}

	err = CompareHashAndPassword(auth.PasswordHash, req.Password)
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

	factor, _, err := e.loadActiveFactorBundle(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetActiveFactor)
	}
	if factor != nil {
		challenge, err := e.createLoginChallenge(c.UserContext(), auth, *factor)
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusAccepted).JSON(challenge)
	}

	t, rt, err := helper.IssueTokens(user.Id, auth.SessionVersion, e.secret)
	if err != nil {
		return err
	}

	return c.JSON(LoginResponse{Token: t, RefreshToken: rt})
}

// RefreshToken
//
//	@Summary	Refresh authentication token
//	@Produce	json
//	@Tags		Auth
//	@Param		Authorization	header		string	true	"Refresh token instead of auth"
//	@Success	200				{object}	RefreshTokenResponse
//	@failure	400				{string}	string	"Incorrect request body"
//	@failure	401				{string}	string	"Unauthorized"
//	@failure	500				{string}	string	"Something bad happened"
//	@Router		/auth/refresh [get]
func (e *entity) RefreshToken(c *fiber.Ctx) error {
	tu, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetUserFromToken)
	}

	u, err := e.user.GetUserById(c.UserContext(), tu.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, ErrUnableToGetUserById)
	}
	if u.Blocked {
		return fiber.NewError(fiber.StatusUnauthorized, ErrUserIsBanned)
	}

	auth, err := e.auth.GetAuthenticationByUserId(c.UserContext(), u.Id)
	if err := helper.HttpDbError(err, ErrUnableToGetAuthenticationByUserId); err != nil {
		return err
	}

	t, rt, err := helper.IssueTokens(u.Id, auth.SessionVersion, e.secret)
	if err != nil {
		return err
	}

	return c.JSON(RefreshTokenResponse{Token: t, RefreshToken: rt})
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
	log := observability.LoggerFromFiber(c, e.log)

	var req RegisterRequest
	if err := e.parseAndValidate(c, "registration", &req); err != nil {
		return err
	}
	_, err := e.auth.GetAuthenticationByEmail(c.UserContext(), req.Email)
	if err == nil {
		return c.SendStatus(fiber.StatusFound)
	} else if !errors.Is(err, sql.ErrNoRows) {
		log.Error("unable to get authentication by email", slog.String("error", err.Error()))
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetAuthenticationByEmail)
	}

	var id int64
	var token string
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
		token = reg.ConfirmationToken
	}

	if token == "" {
		newToken, tErr := helper.RandomToken(40)
		if tErr != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGenerateToken)
		}
		token = newToken
	}

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		err = e.reg.CreateRegistration(c.UserContext(), id, req.Email, token)
		if err := helper.HttpDbError(err, ErrUnableToCreateRegistration); err != nil {
			return err
		}
	}

	err = e.mailer.Send(c.UserContext(), id, req.Email, token, mailer.EmailTypeRegistration)
	if err != nil {
		log.Error("send email error", slog.String("error", err.Error()))
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
//	@Router		/auth/confirmation [post]
func (e *entity) Confirmation(c *fiber.Ctx) error {
	log := observability.LoggerFromFiber(c, e.log)

	var req ConfirmationRequest
	if err := e.parseAndValidate(c, "confirmation", &req); err != nil {
		return err
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
		log.Error("unable to create authentication", slog.String("error", err.Error()))
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.SendStatus(fiber.StatusCreated)
}

// PasswordRecovery
//
//	@Summary	Password Recovery
//	@Produce	json
//	@Tags		Auth
//	@Param		request	body		PasswordRecoveryRequest	true	"Email for password recovery"
//	@Success	202		{string}	string					"Recovery email sent"
//	@failure	400		{string}	string					"Incorrect request body"
//	@failure	404		{string}	string					"Email not found"
//	@failure	409		{string}	string					"Recovery email already sent"
//	@failure	410		{string}	string					"User is banned"
//	@failure	429		{string}	string					"Try again later"
//	@failure	500		{string}	string					"Something bad happened"
//	@Router		/auth/recovery [post]
func (e *entity) PasswordRecovery(c *fiber.Ctx) error {
	log := observability.LoggerFromFiber(c, e.log)

	var req PasswordRecoveryRequest
	if err := e.parseAndValidate(c, "recovery", &req); err != nil {
		return err
	}

	// Check if the email exists in the authentication table
	auth, err := e.auth.GetAuthenticationByEmail(c.UserContext(), req.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return fiber.NewError(fiber.StatusNotFound, ErrEmailNotFound)
	}
	if err != nil {
		log.Error("unable to get authentication by email", slog.String("error", err.Error()))
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetAuthenticationByEmail)
	}

	user, err := e.user.GetUserById(c.UserContext(), auth.UserId)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, ErrUnableToGetUserById)
	}
	if user.Blocked {
		return fiber.NewError(fiber.StatusGone, ErrUserIsBanned)
	}

	// Generate a token for password reset
	token, err := helper.RandomToken(40)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGenerateToken)
	}

	// Check if there's already a recovery record for this user
	rec, err := e.auth.GetRecoveryByUserId(c.UserContext(), auth.UserId)
	if err == nil {
		// If recovery exists and was created less than a minute ago, return error
		if rec.CreatedAt.Add(time.Minute).After(time.Now()) {
			return fiber.NewError(fiber.StatusConflict, ErrRecoveryEmailAlreadySent)
		}
		// Remove the existing recovery
		err = e.auth.RemoveRecovery(c.UserContext(), auth.UserId)
		if err := helper.HttpDbError(err, ErrUnableToRemoveRegistration); err != nil {
			return err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetRecoveryByUserId)
	}

	expiration := time.Now().Add(time.Hour * 24)

	// Create a new recovery record with the reset token
	err = e.auth.CreateRecovery(c.UserContext(), auth.UserId, token, expiration)
	if err := helper.HttpDbError(err, ErrUnableToCreateRegistration); err != nil {
		return err
	}

	// Send the password reset email
	err = e.mailer.Send(c.UserContext(), auth.UserId, req.Email, token, mailer.EmailTypePasswordReset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendEmail)
	}

	return c.SendStatus(fiber.StatusAccepted)
}

// PasswordReset
//
//	@Summary	Password Reset
//	@Produce	json
//	@Tags		Auth
//	@Param		request	body		PasswordResetRequest	true	"Password reset data"
//	@Success	200		{string}	string					"Password reset successful"
//	@failure	400		{string}	string					"Incorrect request body"
//	@failure	401		{string}	string					"Unauthorized"
//	@failure	500		{string}	string					"Something bad happened"
//	@Router		/auth/reset [post]
func (e *entity) PasswordReset(c *fiber.Ctx) error {
	log := observability.LoggerFromFiber(c, e.log)

	var req PasswordResetRequest
	if err := e.parseAndValidate(c, "reset", &req); err != nil {
		return err
	}

	// Get the recovery record for the user
	rec, err := e.auth.GetRecoveryByUserId(c.UserContext(), req.Id)
	if err := helper.HttpDbError(err, ErrUnableToGetRegistrationById); err != nil {
		return err
	}
	if time.Now().After(rec.ExpiresAt) {
		if err := e.auth.RemoveRecovery(c.UserContext(), req.Id); err != nil {
			log.Warn("unable to remove expired recovery", slog.String("error", err.Error()))
		}
		return fiber.NewError(fiber.StatusUnauthorized, ErrRecoveryExpired)
	}

	// Verify the token
	if rec.Token != req.Token {
		return fiber.NewError(fiber.StatusUnauthorized, ErrTokenIsIncorrect)
	}

	// Hash the new password
	hash, err := HashPassword(req.Password)
	if err := helper.HttpDbError(err, ErrUnableToGetPasswordHash); err != nil {
		return err
	}

	tx, err := e.db.BeginTxx(c.UserContext(), nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSetPasswordHash)
	}
	defer func() { _ = tx.Rollback() }()

	// Update the password hash in the authentication table
	err = e.auth.SetPasswordHashTx(c.UserContext(), tx, req.Id, hash)
	if err != nil {
		log.Error("unable to set password hash", slog.String("error", err.Error()))
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSetPasswordHash)
	}
	version, err := e.auth.BumpSessionVersionTx(c.UserContext(), tx, req.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToBumpSessionVersion)
	}

	// Remove the registration record
	err = e.auth.RemoveRecoveryTx(c.UserContext(), tx, req.Id)
	if err := helper.HttpDbError(err, ErrUnableToRemoveRegistration); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSetPasswordHash)
	}
	if err := e.completeSessionMutation(c.UserContext(), req.Id, version); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
	}

	return c.SendStatus(fiber.StatusOK)
}
