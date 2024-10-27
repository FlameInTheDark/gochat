package auth

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
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Email string `json:"email"`
}

type ConfirmationRequest struct {
	Id            int64  `json:"id"`
	Token         string `json:"token"`
	Name          string `json:"name"`
	Discriminator string `json:"discriminator"`
	Password      string `json:"password"`
}
