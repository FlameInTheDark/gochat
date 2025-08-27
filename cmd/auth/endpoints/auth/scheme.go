package auth

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	ErrUnableToCreateDiscriminator      = "unable to create discriminator"
	ErrUnableToCreateUser               = "unable to create user"
	ErrUnableToGetUserById              = "unable to get user by ID"
	ErrUnableToRemoveRegistration       = "unable to remove registration"
	ErrDiscriminatorIsNotUnique         = "discriminator is not unique"
	ErrTokenIsIncorrect                 = "token is incorrect"
	ErrUnableToGetRegistrationById      = "unable to get registration by id"
	ErrUnableToCreateRegistration       = "unable to create registration"
	ErrUnableToGetRegistrationByEmail   = "unable to get registration by email"
	ErrUnableToGetPasswordHash          = "unable to get password hash"
	ErrPasswordIsTooShort               = "password is too short"
	ErrUnableToParseBody                = "unable to parse body"
	ErrUnableToSendEmail                = "unable to send email"
	ErrEmailAlreadySent                 = "email already sent"
	ErrUnableToGenerateToken            = "unable to generate token"
	ErrUnableToGetAuthenticationByEmail = "unable to get authentication by email"
	ErrUnableToSignAuthenticationToken  = "unable to sign authentication token"
	ErrUserIsBanned                     = "user is banned"
	ErrUnableToCompareHash              = "unable to compare hash"
	ErrEmailNotFound                    = "email not found"
	ErrUnableToSetPasswordHash          = "unable to set password hash"
	ErrRecoveryEmailAlreadySent         = "recovery email already sent"

	// Validation error messages
	ErrNameRequired               = "name is required"
	ErrNameTooShort               = "name must be at least 4 characters"
	ErrNameTooLong                = "name must be less than 20 characters"
	ErrDiscriminatorRequired      = "discriminator is required"
	ErrDiscriminatorTooShort      = "discriminator must be at least 4 characters"
	ErrDiscriminatorTooLong       = "discriminator must be less than 20 characters"
	ErrDiscriminatorInvalidFormat = "discriminator can only contain lowercase letters, numbers, underscore, hyphen, and dot"
	ErrTokenRequired              = "token is required"
	ErrTokenInvalidLength         = "token must be exactly 40 characters"
	ErrIdRequired                 = "id is required"
	ErrIdInvalid                  = "id must be a positive number"
	ErrPasswordTooLong            = "password must be less than 50 characters"
	ErrEmailRequired              = "email is required"
	ErrEmailInvalidFormat         = "email format is invalid"
	ErrPasswordRequired           = "password is required"
)

var (
	emailRegex         = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	discriminatorRegex = regexp.MustCompile(`^[a-z0-9._-]+$`)
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email,
			validation.Required.Error(ErrEmailRequired),
			validation.Match(emailRegex).Error(ErrEmailInvalidFormat),
		),
		validation.Field(&r.Password,
			validation.Required.Error(ErrPasswordRequired),
		),
	)
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Email string `json:"email"`
}

func (r RegisterRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email,
			validation.Required.Error(ErrEmailRequired),
			validation.Match(emailRegex).Error(ErrEmailInvalidFormat),
		),
	)
}

type ConfirmationRequest struct {
	Id            int64  `json:"id"`
	Token         string `json:"token"`
	Name          string `json:"name"`
	Discriminator string `json:"discriminator"`
	Password      string `json:"password"`
}

func (r ConfirmationRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Password,
			validation.Required,
			validation.RuneLength(8, 0).Error(ErrPasswordIsTooShort),
			validation.RuneLength(0, 50).Error(ErrPasswordTooLong),
		),
		validation.Field(&r.Name,
			validation.Required.Error(ErrNameRequired),
			validation.RuneLength(4, 0).Error(ErrNameTooShort),
			validation.RuneLength(0, 20).Error(ErrNameTooLong),
		),
		validation.Field(&r.Token,
			validation.Required.Error(ErrTokenRequired),
			validation.Length(40, 40).Error(ErrTokenInvalidLength),
		),
		validation.Field(&r.Discriminator,
			validation.Required.Error(ErrDiscriminatorRequired),
			validation.RuneLength(4, 0).Error(ErrDiscriminatorTooShort),
			validation.RuneLength(0, 20).Error(ErrDiscriminatorTooLong),
			validation.Match(discriminatorRegex).Error(ErrDiscriminatorInvalidFormat),
		),
		validation.Field(&r.Id,
			validation.Required.Error(ErrIdRequired),
			validation.Min(int64(1)).Error(ErrIdInvalid),
		),
	)
}

type PasswordRecoveryRequest struct {
	Email string `json:"email"`
}

func (r PasswordRecoveryRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Email,
			validation.Required.Error(ErrEmailRequired),
			validation.Match(emailRegex).Error(ErrEmailInvalidFormat),
		),
	)
}

type PasswordResetRequest struct {
	Id       int64  `json:"id"`
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (r PasswordResetRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Password,
			validation.Required,
			validation.RuneLength(8, 0).Error(ErrPasswordIsTooShort),
			validation.RuneLength(0, 50).Error(ErrPasswordTooLong),
		),
		validation.Field(&r.Token,
			validation.Required.Error(ErrTokenRequired),
			validation.Length(40, 40).Error(ErrTokenInvalidLength),
		),
		validation.Field(&r.Id,
			validation.Required.Error(ErrIdRequired),
			validation.Min(int64(1)).Error(ErrIdInvalid),
		),
	)
}
