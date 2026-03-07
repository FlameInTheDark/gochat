package guild

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	emojirepo "github.com/FlameInTheDark/gochat/internal/database/pgentities/emoji"
	"github.com/FlameInTheDark/gochat/internal/dto"
	emojiutil "github.com/FlameInTheDark/gochat/internal/emoji"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/upload"
)

func (e *entity) CreateEmoji(c *fiber.Ctx) error {
	var req CreateEmojiRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	if err := e.ensureGuildEmojiPermission(c.UserContext(), guildId, user.Id, permissions.PermCreateExpressions); err != nil {
		return err
	}
	_ = e.emoji.PruneExpired(c.UserContext(), guildId)
	count, err := e.emoji.CountActiveGuildEmojis(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateEmoji)
	}
	if count >= emojiutil.MaxActivePerGuild {
		return fiber.NewError(fiber.StatusConflict, ErrEmojiActiveLimitExceeded)
	}

	id := idgen.Next()
	record := model.GuildEmoji{
		GuildId:          guildId,
		Id:               id,
		Name:             req.Name,
		NameNormalized:   emojiutil.NormalizeName(req.Name),
		CreatorId:        user.Id,
		DeclaredFileSize: req.FileSize,
		UploadExpiresAt:  time.Now().UTC().Add(time.Duration(e.attachTTL) * time.Second),
	}
	if err := e.emoji.CreatePlaceholder(c.UserContext(), record); err != nil {
		if errors.Is(err, emojirepo.ErrEmojiNameTaken) {
			return fiber.NewError(fiber.StatusConflict, ErrEmojiNameTaken)
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateEmoji)
	}
	_ = e.invalidateEmojiCache(c.UserContext(), guildId, id)
	return c.JSON(dto.EmojiUpload{Id: id, GuildId: guildId, Name: req.Name})
}

func (e *entity) ListEmojis(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	if _, err = e.validateGuildAccess(c, guildId); err != nil {
		return err
	}

	emojis, err := e.getCachedGuildEmojis(c.UserContext(), guildId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetEmojis)
	}
	return c.JSON(emojis)
}

func (e *entity) UpdateEmoji(c *fiber.Ctx) error {
	var req UpdateEmojiRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	emojiId, err := e.parseEmojiID(c)
	if err != nil {
		return err
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	if err := e.ensureGuildEmojiPermission(c.UserContext(), guildId, user.Id, permissions.PermManageExpressions); err != nil {
		return err
	}

	updated, err := e.emoji.Rename(c.UserContext(), guildId, emojiId, req.Name, emojiutil.NormalizeName(req.Name))
	if err != nil {
		switch {
		case errors.Is(err, emojirepo.ErrEmojiNotFound):
			return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetEmojis)
		case errors.Is(err, emojirepo.ErrEmojiNameTaken):
			return fiber.NewError(fiber.StatusConflict, ErrEmojiNameTaken)
		default:
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUpdateEmoji)
		}
	}

	_ = e.invalidateEmojiCache(c.UserContext(), guildId, emojiId)
	go e.publishEmojiUpdate(guildId, guildEmojiToDTO(updated))
	return c.JSON(guildEmojiToDTO(updated))
}

func (e *entity) DeleteEmoji(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	emojiId, err := e.parseEmojiID(c)
	if err != nil {
		return err
	}
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}
	if err := e.ensureGuildEmojiPermission(c.UserContext(), guildId, user.Id, permissions.PermManageExpressions); err != nil {
		return err
	}

	emoji, err := e.emoji.GetGuildEmoji(c.UserContext(), guildId, emojiId)
	if err != nil {
		if errors.Is(err, emojirepo.ErrEmojiNotFound) {
			return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetEmojis)
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDeleteEmoji)
	}
	if err := e.removeEmojiObjects(c.UserContext(), emoji.Id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDeleteEmoji)
	}
	if _, err = e.emoji.Delete(c.UserContext(), guildId, emojiId); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToDeleteEmoji)
	}
	_ = e.invalidateEmojiCache(c.UserContext(), guildId, emojiId)
	go e.publishEmojiDelete(guildId, emojiId)
	return c.SendStatus(fiber.StatusOK)
}

func (e *entity) ensureGuildEmojiPermission(ctx context.Context, guildId, userId int64, perm permissions.RolePermission) error {
	_, ok, err := e.perm.GuildPerm(ctx, guildId, userId, perm)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetGuildByID)
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetPermission)
	}
	if !ok {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}
	return nil
}

func (e *entity) getCachedGuildEmojis(ctx context.Context, guildId int64) ([]dto.GuildEmoji, error) {
	var cached []dto.GuildEmoji
	if e.cache != nil {
		if err := e.cache.GetJSON(ctx, emojiutil.GuildCacheKey(guildId), &cached); err == nil {
			if cached == nil {
				return []dto.GuildEmoji{}, nil
			}
			return cached, nil
		}
	}

	rows, err := e.emoji.ListReadyGuildEmojis(ctx, guildId)
	if err != nil {
		return nil, err
	}
	cached = guildEmojisToDTO(rows)
	if cached == nil {
		cached = []dto.GuildEmoji{}
	}
	if e.cache != nil {
		_ = e.cache.SetTimedJSON(ctx, emojiutil.GuildCacheKey(guildId), cached, emojiutil.GuildCacheTTLSeconds)
	}
	return cached, nil
}

func (e *entity) invalidateEmojiCache(ctx context.Context, guildId, emojiId int64) error {
	if e.cache == nil {
		return nil
	}
	if err := e.cache.Delete(ctx, emojiutil.GuildCacheKey(guildId)); err != nil {
		return err
	}
	return e.cache.Delete(ctx, emojiutil.LookupCacheKey(emojiId))
}

func (e *entity) removeEmojiObjects(ctx context.Context, emojiId int64) error {
	if e.storage == nil {
		return fmt.Errorf(ErrEmojiStorageUnavailable)
	}
	for _, key := range []string{upload.EmojiMasterKey(emojiId), upload.EmojiSizedKey(emojiId, 96), upload.EmojiSizedKey(emojiId, 44)} {
		if err := e.storage.RemoveAttachment(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (e *entity) publishEmojiUpdate(guildId int64, emoji dto.GuildEmoji) {
	_ = e.mqt.SendGuildUpdate(guildId, &mqmsg.UpdateGuildEmoji{Emoji: emoji})
}

func (e *entity) publishEmojiCreate(guildId int64, emoji dto.GuildEmoji) {
	_ = e.mqt.SendGuildUpdate(guildId, &mqmsg.CreateGuildEmoji{Emoji: emoji})
}

func (e *entity) publishEmojiDelete(guildId, emojiId int64) {
	_ = e.mqt.SendGuildUpdate(guildId, &mqmsg.DeleteGuildEmoji{GuildId: guildId, EmojiId: emojiId})
}
