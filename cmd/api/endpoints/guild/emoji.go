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

// CreateEmoji
//
//	@Summary	Create guild emoji metadata
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64				true	"Guild ID"
//	@Param		request		body		CreateEmojiRequest	true	"Emoji metadata"
//	@Success	200			{object}	dto.EmojiUpload		"Emoji upload metadata"
//	@failure	400			{string}	string				"Bad request"
//	@failure	401			{string}	string				"Unauthorized"
//	@failure	403			{string}	string				"Forbidden"
//	@failure	409			{string}	string				"Conflict"
//	@failure	500			{string}	string				"Internal server error"
//	@Router		/guild/{guild_id}/emojis [post]
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

	uploadMeta, err := e.reserveEmojiUpload(c.UserContext(), guildId, user.Id, req)
	if err != nil {
		return err
	}
	_ = e.invalidateEmojiCache(c.UserContext(), guildId, uploadMeta.Id)
	return c.JSON(uploadMeta)
}

func (e *entity) reserveEmojiUpload(ctx context.Context, guildId, userId int64, req CreateEmojiRequest) (dto.EmojiUpload, error) {
	_ = e.emoji.PruneExpired(ctx, guildId)

	record := model.GuildEmoji{
		GuildId:          guildId,
		Id:               idgen.Next(),
		Name:             req.Name,
		NameNormalized:   emojiutil.NormalizeName(req.Name),
		CreatorId:        userId,
		DeclaredFileSize: req.FileSize,
		UploadExpiresAt:  time.Now().UTC().Add(time.Duration(e.attachTTL) * time.Second),
	}

	reused, err := e.emoji.ReusePendingPlaceholder(ctx, record)
	switch {
	case err == nil:
		return dto.EmojiUpload{Id: reused.Id, GuildId: reused.GuildId, Name: reused.Name}, nil
	case errors.Is(err, emojirepo.ErrEmojiNameTaken):
		return dto.EmojiUpload{}, fiber.NewError(fiber.StatusConflict, ErrEmojiNameTaken)
	case err != nil && !errors.Is(err, emojirepo.ErrEmojiNotFound):
		return dto.EmojiUpload{}, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateEmoji)
	}

	count, err := e.emoji.CountActiveGuildEmojis(ctx, guildId)
	if err != nil {
		return dto.EmojiUpload{}, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateEmoji)
	}
	if count >= emojiutil.MaxActivePerGuild {
		return dto.EmojiUpload{}, fiber.NewError(fiber.StatusConflict, ErrEmojiActiveLimitExceeded)
	}

	if err := e.emoji.CreatePlaceholder(ctx, record); err != nil {
		if errors.Is(err, emojirepo.ErrEmojiNameTaken) {
			reused, reuseErr := e.emoji.ReusePendingPlaceholder(ctx, record)
			switch {
			case reuseErr == nil:
				return dto.EmojiUpload{Id: reused.Id, GuildId: reused.GuildId, Name: reused.Name}, nil
			case errors.Is(reuseErr, emojirepo.ErrEmojiNameTaken):
				return dto.EmojiUpload{}, fiber.NewError(fiber.StatusConflict, ErrEmojiNameTaken)
			default:
				return dto.EmojiUpload{}, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateEmoji)
			}
		}
		return dto.EmojiUpload{}, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToCreateEmoji)
	}

	return dto.EmojiUpload{Id: record.Id, GuildId: record.GuildId, Name: record.Name}, nil
}

// ListEmojis
//
//	@Summary	List guild emojis
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64			true	"Guild ID"
//	@Success	200			{array}		dto.GuildEmoji	"Guild emojis"
//	@failure	400			{string}	string			"Bad request"
//	@failure	401			{string}	string			"Unauthorized"
//	@failure	403			{string}	string			"Forbidden"
//	@failure	500			{string}	string			"Internal server error"
//	@Router		/guild/{guild_id}/emojis [get]
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

// UpdateEmoji
//
//	@Summary	Update guild emoji
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64				true	"Guild ID"
//	@Param		emoji_id	path		int64				true	"Emoji ID"
//	@Param		request		body		UpdateEmojiRequest	true	"Emoji update data"
//	@Success	200			{object}	dto.GuildEmoji		"Updated emoji"
//	@failure	400			{string}	string				"Bad request"
//	@failure	401			{string}	string				"Unauthorized"
//	@failure	403			{string}	string				"Forbidden"
//	@failure	404			{string}	string				"Not found"
//	@failure	409			{string}	string				"Conflict"
//	@failure	500			{string}	string				"Internal server error"
//	@Router		/guild/{guild_id}/emojis/{emoji_id} [patch]
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

// DeleteEmoji
//
//	@Summary	Delete guild emoji
//	@Produce	json
//	@Tags		Guild
//	@Param		guild_id	path		int64	true	"Guild ID"
//	@Param		emoji_id	path		int64	true	"Emoji ID"
//	@Success	200			{string}	string	"OK"
//	@failure	400			{string}	string	"Bad request"
//	@failure	401			{string}	string	"Unauthorized"
//	@failure	403			{string}	string	"Forbidden"
//	@failure	404			{string}	string	"Not found"
//	@failure	500			{string}	string	"Internal server error"
//	@Router		/guild/{guild_id}/emojis/{emoji_id} [delete]
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
