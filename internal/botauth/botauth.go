package botauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	botrepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/bot"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	Scheme      = "Bot"
	tokenPrefix = "gcb_"
)

var (
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthScheme = errors.New("bot runtime authentication requires Authorization: Bot <token>")
	ErrInvalidToken      = errors.New("invalid bot token")
	ErrDisabledBot       = errors.New("bot is disabled")
)

type Principal struct {
	BotUserID   int64
	OwnerUserID int64
	TokenID     int64
	TokenPrefix string
	Bot         model.Bot
	User        model.User
}

func GenerateToken() (plain, prefix, hash string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", "", err
	}
	plain = tokenPrefix + base64.RawURLEncoding.EncodeToString(raw)
	if len(plain) < 16 {
		return "", "", "", fmt.Errorf("generated token too short")
	}
	return plain, plain[:16], HashToken(plain), nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func ParseAuthorization(header string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", ErrMissingAuthHeader
	}
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, Scheme) || strings.TrimSpace(token) == "" {
		return "", ErrInvalidAuthScheme
	}
	return strings.TrimSpace(token), nil
}

func Authenticate(ctx context.Context, bots botrepo.Bot, users user.User, authorization string) (*Principal, error) {
	token, err := ParseAuthorization(authorization)
	if err != nil {
		return nil, err
	}
	rec, err := bots.GetTokenByHash(ctx, HashToken(token))
	if err != nil {
		return nil, ErrInvalidToken
	}
	bot, err := bots.GetBot(ctx, rec.BotUserId)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if bot.Disabled {
		return nil, ErrDisabledBot
	}
	u, err := users.GetUserById(ctx, bot.BotUserId)
	if err != nil || !u.IsBot() {
		return nil, ErrInvalidToken
	}
	_ = bots.TouchToken(ctx, rec.Id)
	return &Principal{
		BotUserID:   bot.BotUserId,
		OwnerUserID: bot.OwnerUserId,
		TokenID:     rec.Id,
		TokenPrefix: rec.TokenPrefix,
		Bot:         bot,
		User:        u,
	}, nil
}

func Middleware(bots botrepo.Bot, users user.User) fiber.Handler {
	return func(c *fiber.Ctx) error {
		principal, err := Authenticate(c.UserContext(), bots, users, c.Get("Authorization"))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}
		c.Locals("bot_principal", principal)
		c.Locals("user", &jwt.Token{Claims: &helper.Claims{
			UserID:    principal.BotUserID,
			TokenType: "bot",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:   "gochat",
				Audience: []string{"bot"},
			},
		}})
		return c.Next()
	}
}

func FromFiber(c *fiber.Ctx) (*Principal, bool) {
	principal, ok := c.Locals("bot_principal").(*Principal)
	return principal, ok
}
