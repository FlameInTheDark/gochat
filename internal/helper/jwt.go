package helper

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/savsgio/gotils/strings"
)

type JWTUser struct {
	Id int64
}

func GetUser(c *fiber.Ctx) (*JWTUser, error) {
	user, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return nil, fmt.Errorf("could not find user in context")
	}
	return GetUserFromToken(user)
}

func GetUserFromToken(token *jwt.Token) (*JWTUser, error) {
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("could not get claims")
	}

	return &JWTUser{
		Id: claims.UserID,
	}, nil
}

type Claims struct {
	UserID    int64  `json:"user_id"`
	TokenType string `json:"typ"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func generateJTI() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func IssueTokens(userID int64, secret string) (access, refresh string, err error) {
	now := time.Now()

	accessTok := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:    userID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "gochat",
			Audience:  []string{"api"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	})
	if access, err = accessTok.SignedString([]byte(secret)); err != nil {
		return
	}

	refreshTok := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "gochat",
			Audience:  []string{"refresh"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * 24 * time.Hour)),
			ID:        generateJTI(),
		},
	})
	refresh, err = refreshTok.SignedString([]byte(secret))
	return
}

func RequireTokenType(expect string, audience ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tok, _ := c.Locals("user").(*jwt.Token)
		if tok == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "unable to get token to check type")
		}
		cl, _ := tok.Claims.(*Claims)
		if cl == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "unable to get claims to check type")
		}

		if cl.TokenType != expect {
			return fiber.NewError(fiber.StatusUnauthorized, "wrong token type")
		}

		for _, v := range audience {
			if !strings.Include(cl.Audience, v) {
				return fiber.NewError(fiber.StatusUnauthorized, "wrong audience")
			}
		}
		return c.Next()
	}
}
