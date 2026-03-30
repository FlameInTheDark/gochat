package auth

import (
	"errors"
	"time"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// GetTwoFactorStatus
//
//	@Summary		Get two-factor auth status
//	@Description	Returns whether TOTP two-factor auth is enabled for the authenticated user and how many unused recovery codes remain.
//	@Produce		json
//	@Tags			Auth
//	@Security		BearerAuth
//	@Success		200	{object}	TwoFactorStatusResponse	"Current two-factor auth status"
//	@failure		401	{string}	string					"Unauthorized"
//	@failure		500	{string}	string					"Something bad happened"
//	@Router			/auth/2fa [get]
func (e *entity) GetTwoFactorStatus(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserFromToken)
	}

	factor, _, err := e.loadActiveFactorBundle(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetActiveFactor)
	}
	if factor == nil {
		return c.JSON(TwoFactorStatusResponse{Enabled: false, RecoveryCodesRemaining: 0})
	}

	count, err := e.factor.CountUnusedRecoveryCodes(c.UserContext(), user.Id, factor.FactorID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToListRecoveryCodes)
	}
	factorType := factor.FactorType
	return c.JSON(TwoFactorStatusResponse{
		Enabled:                true,
		FactorType:             &factorType,
		RecoveryCodesRemaining: count,
	})
}

// StartTOTPSetup
//
//	@Summary		Start TOTP setup
//	@Description	Verifies the current password and returns a short-lived TOTP provisioning payload. The client should render the QR code locally from the returned otpauth URI.
//	@Accept			json
//	@Produce		json
//	@Tags			Auth
//	@Security		BearerAuth
//	@Param			request	body		TOTPSetupRequest	true	"Current password for TOTP setup"
//	@Success		200		{object}	TOTPSetupResponse	"TOTP setup payload"
//	@failure		400		{string}	string				"Incorrect request body"
//	@failure		401		{string}	string				"Unauthorized"
//	@failure		409		{string}	string				"Two-factor auth is already enabled"
//	@failure		500		{string}	string				"Something bad happened"
//	@Router			/auth/2fa/totp/setup [post]
func (e *entity) StartTOTPSetup(c *fiber.Ctx) error {
	var req TOTPSetupRequest
	if err := e.parseAndValidate(c, "totp_setup_start", &req); err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserFromToken)
	}
	authRec, err := e.auth.GetAuthenticationByUserId(c.UserContext(), user.Id)
	if err := helper.HttpDbError(err, ErrUnableToGetAuthenticationByUserId); err != nil {
		return err
	}
	if err := e.validateCurrentPassword(authRec, req.CurrentPassword); err != nil {
		return err
	}

	factor, _, err := e.loadActiveFactorBundle(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetActiveFactor)
	}
	if factor != nil {
		return fiber.NewError(fiber.StatusConflict, ErrTwoFactorAlreadyEnabled)
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      e.appName,
		AccountName: authRec.Email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGenerateToken)
	}

	secret, err := e.secretBox.Encrypt([]byte(key.Secret()))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToEncryptSecret)
	}
	setupID, err := helper.RandomToken(40)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGenerateToken)
	}
	expiresAt := time.Now().Add(pendingTOTPSetupTTL)
	if err := e.savePendingTOTPSetup(c.UserContext(), pendingTOTPSetup{
		SetupID:     setupID,
		UserID:      user.Id,
		Email:       authRec.Email,
		Issuer:      e.appName,
		AccountName: authRec.Email,
		Secret:      secret,
		ExpiresAt:   expiresAt,
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToPersistSetup)
	}

	return c.JSON(TOTPSetupResponse{
		SetupID:     setupID,
		OtpauthURI:  key.URL(),
		ManualKey:   key.Secret(),
		Issuer:      e.appName,
		AccountName: authRec.Email,
		ExpiresAt:   expiresAt,
	})
}

// ConfirmTOTPSetup
//
//	@Summary		Confirm TOTP setup
//	@Description	Validates the authenticator code for a pending TOTP setup, enables two-factor auth, rotates the session version, and returns new tokens plus recovery codes.
//	@Accept			json
//	@Produce		json
//	@Tags			Auth
//	@Security		BearerAuth
//	@Param			request	body		TOTPConfirmRequest		true	"Pending setup ID and authenticator code"
//	@Success		200		{object}	RecoveryCodesResponse	"TOTP enabled and recovery codes issued"
//	@failure		400		{string}	string					"Incorrect request body"
//	@failure		401		{string}	string					"Unauthorized"
//	@failure		409		{string}	string					"Two-factor auth is already enabled"
//	@failure		500		{string}	string					"Something bad happened"
//	@Router			/auth/2fa/totp/confirm [post]
func (e *entity) ConfirmTOTPSetup(c *fiber.Ctx) error {
	var req TOTPConfirmRequest
	if err := e.parseAndValidate(c, "totp_setup_confirm", &req); err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserFromToken)
	}
	setup, err := e.loadPendingTOTPSetup(c.UserContext(), req.SetupID)
	if err != nil || setup.UserID != user.Id || time.Now().After(setup.ExpiresAt) {
		return fiber.NewError(fiber.StatusUnauthorized, ErrSetupExpired)
	}

	factor, _, err := e.loadActiveFactorBundle(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetActiveFactor)
	}
	if factor != nil {
		return fiber.NewError(fiber.StatusConflict, ErrTwoFactorAlreadyEnabled)
	}

	secret, err := e.secretBox.Decrypt(setup.Secret)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDecryptSecret)
	}
	valid, err := totp.ValidateCustom(req.Code, string(secret), time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
	}
	if !valid {
		return fiber.NewError(fiber.StatusUnauthorized, ErrInvalidTwoFactorCode)
	}

	tx, err := e.db.BeginTxx(c.UserContext(), nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateFactor)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	factorID := idgenNext()
	recoveryCodes, storedCodes, err := e.generateRecoveryCodes()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGenerateRecoveryCodes)
	}
	for i := range storedCodes {
		storedCodes[i].UserID = user.Id
		storedCodes[i].FactorID = factorID
	}
	err = e.factor.CreateTOTPFactorTx(c.UserContext(), tx, model.AuthFactor{
		UserID:      user.Id,
		FactorID:    factorID,
		FactorType:  authFactorTypeTOTP,
		DisplayName: "Authenticator app",
		Status:      authFactorStatusActive,
		CreatedAt:   now,
		VerifiedAt:  &now,
		LastUsedAt:  &now,
	}, model.AuthTOTPFactor{
		UserID:           user.Id,
		FactorID:         factorID,
		SecretCiphertext: setup.Secret.Ciphertext,
		SecretNonce:      setup.Secret.Nonce,
		Algorithm:        "SHA1",
		Digits:           6,
		PeriodSeconds:    30,
		CreatedAt:        now,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateFactor)
	}
	if err := e.factor.ReplaceRecoveryCodesTx(c.UserContext(), tx, user.Id, factorID, storedCodes); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToReplaceRecoveryCodes)
	}
	version, err := e.auth.BumpSessionVersionTx(c.UserContext(), tx, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToBumpSessionVersion)
	}
	if err := tx.Commit(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateFactor)
	}

	_ = e.deletePendingTOTPSetup(c.UserContext(), req.SetupID)
	if err := e.completeSessionMutation(c.UserContext(), user.Id, version); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
	}

	token, refresh, err := helper.IssueTokens(user.Id, version, e.secret)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSignAuthenticationToken)
	}
	return c.JSON(RecoveryCodesResponse{Token: token, RefreshToken: refresh, RecoveryCodes: recoveryCodes})
}

// ChangePassword
//
//	@Summary		Change password
//	@Description	Changes the current password. When two-factor auth is enabled, the request must also include either a TOTP code or a recovery code. Successful changes rotate the session version and return fresh tokens.
//	@Accept			json
//	@Produce		json
//	@Tags			Auth
//	@Security		BearerAuth
//	@Param			request	body		PasswordChangeRequest	true	"Password change request"
//	@Success		200		{object}	LoginResponse			"Password changed and new tokens issued"
//	@failure		400		{string}	string					"Incorrect request body"
//	@failure		401		{string}	string					"Unauthorized"
//	@failure		500		{string}	string					"Something bad happened"
//	@Router			/auth/password/change [post]
func (e *entity) ChangePassword(c *fiber.Ctx) error {
	var req PasswordChangeRequest
	if err := e.parseAndValidate(c, "password_change", &req); err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserFromToken)
	}
	authRec, err := e.auth.GetAuthenticationByUserId(c.UserContext(), user.Id)
	if err := helper.HttpDbError(err, ErrUnableToGetAuthenticationByUserId); err != nil {
		return err
	}
	if err := e.validateCurrentPassword(authRec, req.CurrentPassword); err != nil {
		return err
	}

	factor, _, err := e.loadActiveFactorBundle(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetActiveFactor)
	}
	if factor != nil && (req.CodeType == "" || req.Code == "") {
		return fiber.NewError(fiber.StatusBadRequest, ErrCodeRequired)
	}

	hash, err := HashPassword(req.NewPassword)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetPasswordHash)
	}
	tx, err := e.db.BeginTxx(c.UserContext(), nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSetPasswordHash)
	}
	defer func() { _ = tx.Rollback() }()

	if factor != nil {
		if err := e.verifySecurityCode(c.UserContext(), tx, user.Id, *factor, req.CodeType, req.Code); err != nil {
			return err
		}
	}
	if err := e.auth.SetPasswordHashTx(c.UserContext(), tx, user.Id, hash); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSetPasswordHash)
	}
	version, err := e.auth.BumpSessionVersionTx(c.UserContext(), tx, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToBumpSessionVersion)
	}
	if err := tx.Commit(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSetPasswordHash)
	}
	if err := e.completeSessionMutation(c.UserContext(), user.Id, version); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
	}

	token, refresh, err := helper.IssueTokens(user.Id, version, e.secret)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSignAuthenticationToken)
	}
	return c.JSON(LoginResponse{Token: token, RefreshToken: refresh})
}

// RegenerateRecoveryCodes
//
//	@Summary		Regenerate recovery codes
//	@Description	Requires the current password and an active second factor verification, replaces all recovery codes, rotates the session version, and returns the new codes once.
//	@Accept			json
//	@Produce		json
//	@Tags			Auth
//	@Security		BearerAuth
//	@Param			request	body		RecoveryCodesRegenerateRequest	true	"Recovery code regeneration request"
//	@Success		200		{object}	RecoveryCodesResponse			"New recovery codes issued"
//	@failure		400		{string}	string							"Incorrect request body"
//	@failure		401		{string}	string							"Unauthorized"
//	@failure		409		{string}	string							"Two-factor auth is not enabled"
//	@failure		500		{string}	string							"Something bad happened"
//	@Router			/auth/2fa/recovery-codes/regenerate [post]
func (e *entity) RegenerateRecoveryCodes(c *fiber.Ctx) error {
	var req RecoveryCodesRegenerateRequest
	if err := e.parseAndValidate(c, "recovery_codes_regenerate", &req); err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserFromToken)
	}
	authRec, err := e.auth.GetAuthenticationByUserId(c.UserContext(), user.Id)
	if err := helper.HttpDbError(err, ErrUnableToGetAuthenticationByUserId); err != nil {
		return err
	}
	if err := e.validateCurrentPassword(authRec, req.CurrentPassword); err != nil {
		return err
	}

	factor, _, err := e.loadActiveFactorBundle(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetActiveFactor)
	}
	if factor == nil {
		return fiber.NewError(fiber.StatusConflict, ErrTwoFactorNotEnabled)
	}

	tx, err := e.db.BeginTxx(c.UserContext(), nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToReplaceRecoveryCodes)
	}
	defer func() { _ = tx.Rollback() }()

	if err := e.verifySecurityCode(c.UserContext(), tx, user.Id, *factor, req.CodeType, req.Code); err != nil {
		return err
	}
	recoveryCodes, storedCodes, err := e.generateRecoveryCodes()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGenerateRecoveryCodes)
	}
	for i := range storedCodes {
		storedCodes[i].UserID = user.Id
		storedCodes[i].FactorID = factor.FactorID
	}
	if err := e.factor.ReplaceRecoveryCodesTx(c.UserContext(), tx, user.Id, factor.FactorID, storedCodes); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToReplaceRecoveryCodes)
	}
	version, err := e.auth.BumpSessionVersionTx(c.UserContext(), tx, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToBumpSessionVersion)
	}
	if err := tx.Commit(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToReplaceRecoveryCodes)
	}
	if err := e.completeSessionMutation(c.UserContext(), user.Id, version); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
	}

	token, refresh, err := helper.IssueTokens(user.Id, version, e.secret)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSignAuthenticationToken)
	}
	return c.JSON(RecoveryCodesResponse{Token: token, RefreshToken: refresh, RecoveryCodes: recoveryCodes})
}

// DisableTwoFactor
//
//	@Summary		Disable two-factor auth
//	@Description	Requires the current password and a valid TOTP or recovery code, disables the active factor, deletes recovery codes, rotates the session version, and returns fresh tokens.
//	@Accept			json
//	@Produce		json
//	@Tags			Auth
//	@Security		BearerAuth
//	@Param			request	body		DisableTwoFactorRequest	true	"Disable two-factor auth request"
//	@Success		200		{object}	LoginResponse			"Two-factor auth disabled and new tokens issued"
//	@failure		400		{string}	string					"Incorrect request body"
//	@failure		401		{string}	string					"Unauthorized"
//	@failure		409		{string}	string					"Two-factor auth is not enabled"
//	@failure		500		{string}	string					"Something bad happened"
//	@Router			/auth/2fa [delete]
func (e *entity) DisableTwoFactor(c *fiber.Ctx) error {
	var req DisableTwoFactorRequest
	if err := e.parseAndValidate(c, "disable_2fa", &req); err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserFromToken)
	}
	authRec, err := e.auth.GetAuthenticationByUserId(c.UserContext(), user.Id)
	if err := helper.HttpDbError(err, ErrUnableToGetAuthenticationByUserId); err != nil {
		return err
	}
	if err := e.validateCurrentPassword(authRec, req.CurrentPassword); err != nil {
		return err
	}

	factor, _, err := e.loadActiveFactorBundle(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetActiveFactor)
	}
	if factor == nil {
		return fiber.NewError(fiber.StatusConflict, ErrTwoFactorNotEnabled)
	}

	tx, err := e.db.BeginTxx(c.UserContext(), nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDeleteFactor)
	}
	defer func() { _ = tx.Rollback() }()

	if err := e.verifySecurityCode(c.UserContext(), tx, user.Id, *factor, req.CodeType, req.Code); err != nil {
		return err
	}
	if err := e.factor.DeleteFactorsTx(c.UserContext(), tx, user.Id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDeleteFactor)
	}
	version, err := e.auth.BumpSessionVersionTx(c.UserContext(), tx, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToBumpSessionVersion)
	}
	if err := tx.Commit(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDeleteFactor)
	}
	if err := e.completeSessionMutation(c.UserContext(), user.Id, version); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
	}

	token, refresh, err := helper.IssueTokens(user.Id, version, e.secret)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSignAuthenticationToken)
	}
	return c.JSON(LoginResponse{Token: token, RefreshToken: refresh})
}

// LoginTOTP
//
//	@Summary		Complete login with TOTP
//	@Description	Consumes a pending login challenge and verifies the authenticator code. On success the challenge is removed and access and refresh tokens are returned.
//	@Accept			json
//	@Produce		json
//	@Tags			Auth
//	@Param			request	body		LoginTOTPRequest	true	"Login challenge ID and TOTP code"
//	@Success		200		{object}	LoginResponse		"Authenticated after TOTP verification"
//	@failure		400		{string}	string				"Incorrect request body"
//	@failure		401		{string}	string				"Unauthorized"
//	@failure		429		{string}	string				"Too many authentication attempts"
//	@failure		500		{string}	string				"Something bad happened"
//	@Router			/auth/login/2fa/totp [post]
func (e *entity) LoginTOTP(c *fiber.Ctx) error {
	var req LoginTOTPRequest
	if err := e.parseAndValidate(c, "login_totp", &req); err != nil {
		return err
	}
	if err := e.challengeRateLimit(c.UserContext(), req.ChallengeID, c.IP()); err != nil {
		return err
	}

	challenge, err := e.loadLoginChallenge(c.UserContext(), req.ChallengeID)
	if err != nil || time.Now().After(challenge.ExpiresAt) {
		return fiber.NewError(fiber.StatusUnauthorized, ErrLoginChallengeExpired)
	}
	factor, _, err := e.loadActiveFactorBundle(c.UserContext(), challenge.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetActiveFactor)
	}
	if factor == nil || factor.FactorID != challenge.FactorID {
		return fiber.NewError(fiber.StatusUnauthorized, ErrLoginChallengeExpired)
	}

	tx, err := e.db.BeginTxx(c.UserContext(), nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
	}
	defer func() { _ = tx.Rollback() }()

	if err := e.verifySecurityCode(c.UserContext(), tx, challenge.UserID, *factor, codeTypeTOTP, req.Code); err != nil {
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) && fiberErr.Code == fiber.StatusUnauthorized {
			_ = e.incrementChallengeAttempts(c.UserContext(), challenge)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
	}

	authRec, err := e.auth.GetAuthenticationByUserId(c.UserContext(), challenge.UserID)
	if err := helper.HttpDbError(err, ErrUnableToGetAuthenticationByUserId); err != nil {
		return err
	}
	_ = e.deleteLoginChallenge(c.UserContext(), challenge.ChallengeID)
	resp, err := e.issueTokensForAuthentication(authRec)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSignAuthenticationToken)
	}
	return c.JSON(resp)
}

// LoginRecoveryCode
//
//	@Summary		Complete login with recovery code
//	@Description	Consumes a pending login challenge and verifies a recovery code. Recovery codes are single use and are invalid after successful consumption.
//	@Accept			json
//	@Produce		json
//	@Tags			Auth
//	@Param			request	body		LoginRecoveryCodeRequest	true	"Login challenge ID and recovery code"
//	@Success		200		{object}	LoginResponse				"Authenticated after recovery code verification"
//	@failure		400		{string}	string						"Incorrect request body"
//	@failure		401		{string}	string						"Unauthorized"
//	@failure		429		{string}	string						"Too many authentication attempts"
//	@failure		500		{string}	string						"Something bad happened"
//	@Router			/auth/login/2fa/recovery-code [post]
func (e *entity) LoginRecoveryCode(c *fiber.Ctx) error {
	var req LoginRecoveryCodeRequest
	if err := e.parseAndValidate(c, "login_recovery_code", &req); err != nil {
		return err
	}
	if err := e.challengeRateLimit(c.UserContext(), req.ChallengeID, c.IP()); err != nil {
		return err
	}

	challenge, err := e.loadLoginChallenge(c.UserContext(), req.ChallengeID)
	if err != nil || time.Now().After(challenge.ExpiresAt) {
		return fiber.NewError(fiber.StatusUnauthorized, ErrLoginChallengeExpired)
	}
	factor, _, err := e.loadActiveFactorBundle(c.UserContext(), challenge.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetActiveFactor)
	}
	if factor == nil || factor.FactorID != challenge.FactorID {
		return fiber.NewError(fiber.StatusUnauthorized, ErrLoginChallengeExpired)
	}

	tx, err := e.db.BeginTxx(c.UserContext(), nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
	}
	defer func() { _ = tx.Rollback() }()

	if err := e.verifySecurityCode(c.UserContext(), tx, challenge.UserID, *factor, codeTypeRecoveryCode, req.Code); err != nil {
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) && fiberErr.Code == fiber.StatusUnauthorized {
			_ = e.incrementChallengeAttempts(c.UserContext(), challenge)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToValidateSession)
	}

	authRec, err := e.auth.GetAuthenticationByUserId(c.UserContext(), challenge.UserID)
	if err := helper.HttpDbError(err, ErrUnableToGetAuthenticationByUserId); err != nil {
		return err
	}
	_ = e.deleteLoginChallenge(c.UserContext(), challenge.ChallengeID)
	resp, err := e.issueTokensForAuthentication(authRec)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSignAuthenticationToken)
	}
	return c.JSON(resp)
}

// LoginEmailRecoveryStart
//
//	@Summary		Send email recovery code for login
//	@Description	Sends a short-lived email recovery code for a pending login challenge. This flow is intended only for completing login when the authenticator device is unavailable.
//	@Accept			json
//	@Produce		json
//	@Tags			Auth
//	@Param			request	body		LoginEmailStartRequest	true	"Login challenge ID"
//	@Success		202		{string}	string					"Recovery code email sent"
//	@failure		400		{string}	string					"Incorrect request body"
//	@failure		401		{string}	string					"Unauthorized"
//	@failure		429		{string}	string					"Too many authentication attempts"
//	@failure		500		{string}	string					"Something bad happened"
//	@Router			/auth/login/2fa/email/start [post]
func (e *entity) LoginEmailRecoveryStart(c *fiber.Ctx) error {
	var req LoginEmailStartRequest
	if err := e.parseAndValidate(c, "login_email_recovery_start", &req); err != nil {
		return err
	}
	if err := e.challengeRateLimit(c.UserContext(), req.ChallengeID, c.IP()); err != nil {
		return err
	}

	challenge, err := e.loadLoginChallenge(c.UserContext(), req.ChallengeID)
	if err != nil || time.Now().After(challenge.ExpiresAt) {
		return fiber.NewError(fiber.StatusUnauthorized, ErrLoginChallengeExpired)
	}
	if err := e.emailRecoveryRateLimit(c.UserContext(), challenge.UserID, c.IP()); err != nil {
		return err
	}
	if challenge.EmailRecoverySentAt != nil && time.Since(*challenge.EmailRecoverySentAt) < emailRecoveryResendAfter {
		return fiber.NewError(fiber.StatusTooManyRequests, ErrTooManyRecoveryEmails)
	}

	code, err := randomDigits(6)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGenerateEmailCode)
	}
	hash, err := HashPassword(code)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGenerateEmailCode)
	}
	if err := e.sendEmailRecoveryCode(c.UserContext(), challenge.Email, code); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSendRecoveryCodeEmail)
	}
	now := time.Now()
	expiresAt := now.Add(emailRecoveryCodeTTL)
	challenge.EmailRecoveryCodeHash = hash
	challenge.EmailRecoverySentAt = &now
	challenge.EmailRecoveryExpiresAt = &expiresAt
	if err := e.saveLoginChallenge(c.UserContext(), challenge); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateLoginChallenge)
	}
	return c.SendStatus(fiber.StatusAccepted)
}

// LoginEmailRecoveryVerify
//
//	@Summary		Complete login with email recovery code
//	@Description	Verifies the short-lived email recovery code for a pending login challenge and returns access and refresh tokens.
//	@Accept			json
//	@Produce		json
//	@Tags			Auth
//	@Param			request	body		LoginEmailVerifyRequest	true	"Login challenge ID and email recovery code"
//	@Success		200		{object}	LoginResponse			"Authenticated after email recovery verification"
//	@failure		400		{string}	string					"Incorrect request body"
//	@failure		401		{string}	string					"Unauthorized"
//	@failure		429		{string}	string					"Too many authentication attempts"
//	@failure		500		{string}	string					"Something bad happened"
//	@Router			/auth/login/2fa/email/verify [post]
func (e *entity) LoginEmailRecoveryVerify(c *fiber.Ctx) error {
	var req LoginEmailVerifyRequest
	if err := e.parseAndValidate(c, "login_email_recovery_verify", &req); err != nil {
		return err
	}
	if err := e.challengeRateLimit(c.UserContext(), req.ChallengeID, c.IP()); err != nil {
		return err
	}

	challenge, err := e.loadLoginChallenge(c.UserContext(), req.ChallengeID)
	if err != nil || time.Now().After(challenge.ExpiresAt) {
		return fiber.NewError(fiber.StatusUnauthorized, ErrLoginChallengeExpired)
	}
	if challenge.EmailRecoveryCodeHash == "" || challenge.EmailRecoveryExpiresAt == nil || time.Now().After(*challenge.EmailRecoveryExpiresAt) {
		return fiber.NewError(fiber.StatusUnauthorized, ErrInvalidTwoFactorCode)
	}
	if err := CompareHashAndPassword(challenge.EmailRecoveryCodeHash, req.Code); err != nil {
		_ = e.incrementChallengeAttempts(c.UserContext(), challenge)
		return fiber.NewError(fiber.StatusUnauthorized, ErrInvalidTwoFactorCode)
	}

	authRec, err := e.auth.GetAuthenticationByUserId(c.UserContext(), challenge.UserID)
	if err := helper.HttpDbError(err, ErrUnableToGetAuthenticationByUserId); err != nil {
		return err
	}
	_ = e.deleteLoginChallenge(c.UserContext(), challenge.ChallengeID)
	resp, err := e.issueTokensForAuthentication(authRec)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToSignAuthenticationToken)
	}
	return c.JSON(resp)
}
