package helper

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var ErrInvalidSessionVersion = errors.New("invalid session version")

type SessionVersionRepository interface {
	GetSessionVersion(ctx context.Context, userID int64) (int64, error)
}

type SessionVersionCache interface {
	GetInt64(ctx context.Context, key string) (int64, error)
	SetInt64(ctx context.Context, key string, val int64) error
}

type SessionVersionChecker struct {
	repo  SessionVersionRepository
	cache SessionVersionCache
}

func NewSessionVersionChecker(repo SessionVersionRepository, cache SessionVersionCache) *SessionVersionChecker {
	return &SessionVersionChecker{repo: repo, cache: cache}
}

func SessionVersionCacheKey(userID int64) string {
	return fmt.Sprintf("auth:session_version:%d", userID)
}

func (c *SessionVersionChecker) Current(ctx context.Context, userID int64) (int64, error) {
	if c == nil || c.repo == nil {
		return 0, fmt.Errorf("session version checker is not configured")
	}

	if c.cache != nil {
		version, err := c.cache.GetInt64(ctx, SessionVersionCacheKey(userID))
		if err == nil {
			return version, nil
		}
		if !errors.Is(err, redis.Nil) {
			return 0, err
		}
	}

	version, err := c.repo.GetSessionVersion(ctx, userID)
	if err != nil {
		return 0, err
	}
	if c.cache != nil {
		_ = c.cache.SetInt64(ctx, SessionVersionCacheKey(userID), version)
	}
	return version, nil
}

func (c *SessionVersionChecker) Sync(ctx context.Context, userID, version int64) error {
	if c == nil || c.cache == nil {
		return nil
	}
	return c.cache.SetInt64(ctx, SessionVersionCacheKey(userID), version)
}

func (c *SessionVersionChecker) ValidateClaims(ctx context.Context, claims *Claims) error {
	if c == nil || claims == nil {
		return nil
	}
	current, err := c.Current(ctx, claims.UserID)
	if err != nil {
		return err
	}
	if current != claims.SessionVersion {
		return ErrInvalidSessionVersion
	}
	return nil
}

func RequireSessionVersion(checker *SessionVersionChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if checker == nil {
			return c.Next()
		}
		tok, _ := c.Locals("user").(*jwt.Token)
		if tok == nil {
			return c.Next()
		}
		claims, _ := tok.Claims.(*Claims)
		if claims == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "unable to get claims")
		}
		if err := checker.ValidateClaims(c.UserContext(), claims); err != nil {
			if errors.Is(err, ErrInvalidSessionVersion) {
				return fiber.NewError(fiber.StatusUnauthorized, "session expired")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "unable to validate session")
		}
		return c.Next()
	}
}
