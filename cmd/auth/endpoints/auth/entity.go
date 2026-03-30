package auth

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"

	"github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/authentication"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/authfactor"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/registration"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mailer"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "auth"

func (e *entity) Init(router fiber.Router) {
	router.Post("/login", e.Login)
	router.Post("/login/2fa/totp", e.LoginTOTP)
	router.Post("/login/2fa/recovery-code", e.LoginRecoveryCode)
	router.Post("/login/2fa/email/start", e.LoginEmailRecoveryStart)
	router.Post("/login/2fa/email/verify", e.LoginEmailRecoveryVerify)
	router.Post("/registration", e.Registration)
	router.Post("/confirmation", e.Confirmation)
	router.Post("/recovery", e.PasswordRecovery)
	router.Post("/reset", e.PasswordReset)
	router.Group("/refresh", e.refreshMiddleware).Get("", e.RefreshToken)

	passwordGroup := router.Group("/password", e.accessMiddleware)
	passwordGroup.Post("/change", e.ChangePassword)

	twoFactorGroup := router.Group("/2fa", e.accessMiddleware)
	twoFactorGroup.Get("", e.GetTwoFactorStatus)
	twoFactorGroup.Post("/totp/setup", e.StartTOTPSetup)
	twoFactorGroup.Post("/totp/confirm", e.ConfirmTOTPSetup)
	twoFactorGroup.Post("/recovery-codes/regenerate", e.RegenerateRecoveryCodes)
	twoFactorGroup.Delete("", e.DisableTwoFactor)
}

type entity struct {
	name    string
	appName string
	secret  string
	db      *sqlx.DB

	// Services
	log            *slog.Logger
	cache          cache.Cache
	mqt            mq.SendTransporter
	secretBox      *helper.SecretBox
	sessionChecker *helper.SessionVersionChecker

	// DB entities
	auth              authentication.Authentication
	factor            authfactor.AuthFactor
	user              user.User
	reg               registration.Registration
	mailer            *mailer.Mailer
	disc              discriminator.Discriminator
	accessMiddleware  fiber.Handler
	refreshMiddleware fiber.Handler
}

func (e *entity) Name() string {
	return e.name
}

func New(pg *pgdb.DB, cache cache.Cache, m *mailer.Mailer, transporter mq.SendTransporter, appName, secret string, secretBox *helper.SecretBox, sessionChecker *helper.SessionVersionChecker, log *slog.Logger, accessMiddleware, refreshMiddleware fiber.Handler) server.Entity {
	return &entity{
		name:              entityName,
		appName:           appName,
		secret:            secret,
		db:                pg.Conn(),
		log:               log,
		cache:             cache,
		mqt:               transporter,
		secretBox:         secretBox,
		sessionChecker:    sessionChecker,
		auth:              authentication.New(pg.Conn()),
		factor:            authfactor.New(pg.Conn()),
		user:              user.New(pg.Conn()),
		reg:               registration.New(pg.Conn()),
		disc:              discriminator.New(pg.Conn()),
		mailer:            m,
		accessMiddleware:  accessMiddleware,
		refreshMiddleware: refreshMiddleware,
	}
}
