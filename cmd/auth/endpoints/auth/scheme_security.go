package auth

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	ErrUnableToGetAuthenticationByUserId = "unable to get authentication by user id"
	ErrUnableToGetActiveFactor           = "unable to get active factor"
	ErrUnableToGetTOTPFactor             = "unable to get TOTP factor"
	ErrUnableToListRecoveryCodes         = "unable to list recovery codes"
	ErrUnableToReplaceRecoveryCodes      = "unable to replace recovery codes"
	ErrUnableToCreateFactor              = "unable to create factor"
	ErrUnableToDeleteFactor              = "unable to delete factor"
	ErrUnableToPersistSetup              = "unable to persist TOTP setup"
	ErrUnableToLoadSetup                 = "unable to load TOTP setup"
	ErrUnableToDeleteSetup               = "unable to delete TOTP setup"
	ErrUnableToCreateLoginChallenge      = "unable to create login challenge"
	ErrUnableToLoadLoginChallenge        = "unable to load login challenge"
	ErrUnableToDeleteLoginChallenge      = "unable to delete login challenge"
	ErrUnableToValidateSession           = "unable to validate session"
	ErrUnableToEncryptSecret             = "unable to encrypt secret"
	ErrUnableToDecryptSecret             = "unable to decrypt secret"
	ErrUnableToGenerateRecoveryCodes     = "unable to generate recovery codes"
	ErrUnableToGenerateEmailCode         = "unable to generate email code"
	ErrUnableToSendRecoveryCodeEmail     = "unable to send recovery code email"
	ErrUnableToConsumeRecoveryCode       = "unable to consume recovery code"
	ErrUnableToBumpSessionVersion        = "unable to bump session version"
	ErrTwoFactorAlreadyEnabled           = "two-factor auth is already enabled"
	ErrTwoFactorNotEnabled               = "two-factor auth is not enabled"
	ErrSetupExpired                      = "setup is expired or invalid"
	ErrLoginChallengeExpired             = "login challenge is expired or invalid"
	ErrInvalidTwoFactorCode              = "two-factor code is invalid"
	ErrInvalidRecoveryCode               = "recovery code is invalid"
	ErrTooManyAuthAttempts               = "too many authentication attempts"
	ErrTooManyRecoveryEmails             = "too many recovery emails"

	ErrCurrentPasswordRequired = "current password is required"
	ErrNewPasswordRequired     = "new password is required"
	ErrCodeRequired            = "code is required"
	ErrCodeTypeInvalid         = "code type must be totp or recovery_code"
	ErrChallengeIDRequired     = "challenge_id is required"
	ErrSetupIDRequired         = "setup_id is required"
	ErrOTPInvalidFormat        = "OTP must be exactly 6 digits"
)

const (
	codeTypeTOTP          = "totp"
	codeTypeRecoveryCode  = "recovery_code"
	codeTypeEmailRecovery = "email_recovery"

	authFactorTypeTOTP     = "totp"
	authFactorStatusActive = "active"
)

var validSecurityCodeTypes = validation.In(codeTypeTOTP, codeTypeRecoveryCode)

type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" example:"OldP@ssw0rd123"`
	NewPassword     string `json:"new_password" example:"NewP@ssw0rd123"`
	CodeType        string `json:"code_type,omitempty" example:"totp"`
	Code            string `json:"code,omitempty" example:"123456"`
}

func (r PasswordChangeRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.CurrentPassword, validation.Required.Error(ErrCurrentPasswordRequired)),
		validation.Field(&r.NewPassword,
			validation.Required.Error(ErrNewPasswordRequired),
			validation.RuneLength(8, 0).Error(ErrPasswordIsTooShort),
			validation.RuneLength(0, 50).Error(ErrPasswordTooLong),
		),
		validation.Field(&r.CodeType,
			validation.When(r.CodeType != "" || r.Code != "",
				validation.Required.Error(ErrCodeTypeInvalid),
				validSecurityCodeTypes.Error(ErrCodeTypeInvalid),
			),
		),
		validation.Field(&r.Code,
			validation.When(r.CodeType != "" || r.Code != "",
				validation.Required.Error(ErrCodeRequired),
			),
		),
	)
}

type TwoFactorStatusResponse struct {
	Enabled                bool    `json:"enabled" example:"true"`
	FactorType             *string `json:"factor_type,omitempty" example:"totp"`
	RecoveryCodesRemaining int     `json:"recovery_codes_remaining" example:"10"`
}

type TOTPSetupRequest struct {
	CurrentPassword string `json:"current_password" example:"VerYstR0NgP@66WoR6"`
}

func (r TOTPSetupRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.CurrentPassword, validation.Required.Error(ErrCurrentPasswordRequired)),
	)
}

type TOTPSetupResponse struct {
	SetupID     string    `json:"setup_id" example:"b7af1b0bf9f0d78d913cb1fa746785a86f98b1ad"`
	OtpauthURI  string    `json:"otpauth_uri" example:"otpauth://totp/GoChat:user@example.com?algorithm=SHA1&digits=6&issuer=GoChat&period=30&secret=JBSWY3DPEHPK3PXP"`
	ManualKey   string    `json:"manual_key" example:"JBSWY3DPEHPK3PXP"`
	Issuer      string    `json:"issuer" example:"GoChat"`
	AccountName string    `json:"account_name" example:"user@example.com"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type TOTPConfirmRequest struct {
	SetupID string `json:"setup_id" example:"b7af1b0bf9f0d78d913cb1fa746785a86f98b1ad"`
	Code    string `json:"code" example:"123456"`
}

func (r TOTPConfirmRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.SetupID, validation.Required.Error(ErrSetupIDRequired)),
		validation.Field(&r.Code,
			validation.Required.Error(ErrCodeRequired),
			validation.Match(otpRegex).Error(ErrOTPInvalidFormat),
		),
	)
}

type RecoveryCodesResponse struct {
	Token         string   `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"`
	RefreshToken  string   `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"`
	RecoveryCodes []string `json:"recovery_codes" example:"AB12CD34EF56,GH78JK90LM12"`
}

type RecoveryCodesRegenerateRequest struct {
	CurrentPassword string `json:"current_password" example:"VerYstR0NgP@66WoR6"`
	CodeType        string `json:"code_type" example:"totp"`
	Code            string `json:"code" example:"123456"`
}

func (r RecoveryCodesRegenerateRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.CurrentPassword, validation.Required.Error(ErrCurrentPasswordRequired)),
		validation.Field(&r.CodeType,
			validation.Required.Error(ErrCodeTypeInvalid),
			validSecurityCodeTypes.Error(ErrCodeTypeInvalid),
		),
		validation.Field(&r.Code, validation.Required.Error(ErrCodeRequired)),
	)
}

type DisableTwoFactorRequest struct {
	CurrentPassword string `json:"current_password" example:"VerYstR0NgP@66WoR6"`
	CodeType        string `json:"code_type" example:"totp"`
	Code            string `json:"code" example:"123456"`
}

func (r DisableTwoFactorRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.CurrentPassword, validation.Required.Error(ErrCurrentPasswordRequired)),
		validation.Field(&r.CodeType,
			validation.Required.Error(ErrCodeTypeInvalid),
			validSecurityCodeTypes.Error(ErrCodeTypeInvalid),
		),
		validation.Field(&r.Code, validation.Required.Error(ErrCodeRequired)),
	)
}

type LoginChallengeResponse struct {
	ChallengeID string    `json:"challenge_id" example:"e2ec6cc32bcc5b37d2d3b99d6f450c086f8b54c9"`
	Methods     []string  `json:"methods" example:"totp,recovery_code,email_recovery"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type LoginTOTPRequest struct {
	ChallengeID string `json:"challenge_id" example:"e2ec6cc32bcc5b37d2d3b99d6f450c086f8b54c9"`
	Code        string `json:"code" example:"123456"`
}

func (r LoginTOTPRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ChallengeID, validation.Required.Error(ErrChallengeIDRequired)),
		validation.Field(&r.Code,
			validation.Required.Error(ErrCodeRequired),
			validation.Match(otpRegex).Error(ErrOTPInvalidFormat),
		),
	)
}

type LoginRecoveryCodeRequest struct {
	ChallengeID string `json:"challenge_id" example:"e2ec6cc32bcc5b37d2d3b99d6f450c086f8b54c9"`
	Code        string `json:"code" example:"AB12CD34EF56"`
}

func (r LoginRecoveryCodeRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ChallengeID, validation.Required.Error(ErrChallengeIDRequired)),
		validation.Field(&r.Code, validation.Required.Error(ErrCodeRequired)),
	)
}

type LoginEmailStartRequest struct {
	ChallengeID string `json:"challenge_id" example:"e2ec6cc32bcc5b37d2d3b99d6f450c086f8b54c9"`
}

func (r LoginEmailStartRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ChallengeID, validation.Required.Error(ErrChallengeIDRequired)),
	)
}

type LoginEmailVerifyRequest struct {
	ChallengeID string `json:"challenge_id" example:"e2ec6cc32bcc5b37d2d3b99d6f450c086f8b54c9"`
	Code        string `json:"code" example:"123456"`
}

func (r LoginEmailVerifyRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ChallengeID, validation.Required.Error(ErrChallengeIDRequired)),
		validation.Field(&r.Code,
			validation.Required.Error(ErrCodeRequired),
			validation.Match(otpRegex).Error(ErrOTPInvalidFormat),
		),
	)
}
