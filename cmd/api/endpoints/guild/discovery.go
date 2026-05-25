package guild

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/gofiber/fiber/v2"
)

// GetGuildDiscovery
//
//	@Summary		Get guild discovery settings
//	@Description	Owner-only read for public discovery state, discovery description, tags, and hydrated display data persisted in YugabyteDB YSQL.
//	@Produce		json
//	@Tags			Guild
//	@Param			guild_id	path		int64								true	"Guild id"	example(2230469276416868352)
//	@Success		200			{object}	dto.GuildDiscoveryUpdateResponse	"Guild discovery settings"
//	@failure		400			{string}	string								"Bad request"
//	@failure		401			{string}	string								"Unauthorized"
//	@failure		403			{string}	string								"Forbidden"
//	@failure		404			{string}	string								"Guild not found"
//	@failure		500			{string}	string								"Internal server error"
//	@Router			/guild/{guild_id}/discovery [get]
func (e *entity) GetGuildDiscovery(c *fiber.Ctx) error {
	guildID, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	guild, err := e.g.GetGuildById(c.UserContext(), guildID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetGuildByID)
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}
	if guild.OwnerId != user.Id {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	dtoGuild, err := e.discoveryDTO(c.UserContext(), guildID, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}

	return c.JSON(dto.GuildDiscoveryUpdateResponse{Guild: dtoGuild})
}

// UpdateGuildDiscovery
//
//	@Summary		Update guild discovery settings
//	@Description	Owner-only update for public discovery state, discovery description, and normalized tags. The API persists this data in YugabyteDB YSQL and notifies the search service to refresh its OpenSearch document.
//	@Accept			json
//	@Produce		json
//	@Tags			Guild
//	@Param			guild_id	path		int64								true	"Guild id"	example(2230469276416868352)
//	@Param			request		body		GuildDiscoveryUpdateRequest			true	"Guild discovery settings"
//	@Success		200			{object}	dto.GuildDiscoveryUpdateResponse	"Updated guild discovery settings"
//	@Success		202			{object}	dto.GuildDiscoveryUpdateResponse	"Updated settings, indexing notification is pending"
//	@failure		400			{string}	string								"Bad request"
//	@failure		401			{string}	string								"Unauthorized"
//	@failure		403			{string}	string								"Forbidden"
//	@failure		404			{string}	string								"Guild not found"
//	@failure		500			{string}	string								"Internal server error"
//	@Router			/guild/{guild_id}/discovery [patch]
func (e *entity) UpdateGuildDiscovery(c *fiber.Ctx) error {
	guildID, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	guild, err := e.g.GetGuildById(c.UserContext(), guildID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, ErrUnableToGetGuildByID)
		}
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}
	if guild.OwnerId != user.Id {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	var req GuildDiscoveryUpdateRequest
	if err = c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	req.Description = strings.TrimSpace(req.Description)
	req.Tags = normalizeGuildDiscoveryTags(req.Tags)
	if err = req.Validate(); err != nil {
		return badRequestValidationError(err)
	}

	if err = e.gd.SetGuildDiscovery(c.UserContext(), guildID, req.Public, req.Description, req.Tags); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToUpdateGuild)
	}
	e.deleteGuildCache(c.UserContext(), guildID)

	dtoGuild, err := e.discoveryDTO(c.UserContext(), guildID, req.Tags)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}

	indexingPending := false
	if req.Public {
		indexingPending = !e.notifyGuildDiscoveryUpsert(c.UserContext(), guildID, e.log)
	} else {
		indexingPending = !e.notifyGuildDiscoveryDelete(c.UserContext(), guildID, e.log)
	}

	status := fiber.StatusOK
	if indexingPending {
		status = fiber.StatusAccepted
	}
	return c.Status(status).JSON(dto.GuildDiscoveryUpdateResponse{Guild: dtoGuild, IndexingPending: indexingPending})
}

func (e *entity) discoveryDTO(ctx context.Context, guildID int64, fallbackTags []string) (dto.GuildDiscovery, error) {
	guild, err := e.g.GetGuildById(ctx, guildID)
	if err != nil {
		return dto.GuildDiscovery{}, err
	}

	tagsByGuild, err := e.gd.GetTagsByGuilds(ctx, []int64{guildID})
	if err != nil {
		return dto.GuildDiscovery{}, err
	}
	statsByGuild, err := e.gd.GetStatsByGuilds(ctx, []int64{guildID})
	if err != nil {
		return dto.GuildDiscovery{}, err
	}

	tags := tagsByGuild[guildID]
	if tags == nil {
		tags = fallbackTags
	}
	if tags == nil {
		tags = []string{}
	}
	result := dto.GuildDiscovery{
		Id:           guild.Id,
		Name:         guild.Name,
		Description:  guild.Description,
		MembersCount: statsByGuild[guildID].MembersCount,
		Tags:         tags,
	}
	if guild.Icon != nil {
		if iconDTO, err := e.discoveryIconDTO(ctx, guild.Id, *guild.Icon); err == nil && iconDTO != nil {
			result.Icon = iconDTO
		}
	}
	return result, nil
}

func (e *entity) discoveryIconDTO(ctx context.Context, guildID, iconID int64) (*dto.Icon, error) {
	ic, err := e.icon.GetIcon(ctx, iconID, guildID)
	if err != nil {
		return nil, err
	}
	if ic.URL == nil || *ic.URL == "" {
		return nil, nil
	}
	var width, height int64
	if ic.Width != nil {
		width = *ic.Width
	}
	if ic.Height != nil {
		height = *ic.Height
	}
	return &dto.Icon{
		Id:       ic.Id,
		URL:      *ic.URL,
		Filesize: ic.FileSize,
		Width:    width,
		Height:   height,
	}, nil
}

func (e *entity) notifyGuildDiscoveryUpsert(ctx context.Context, guildID int64, logger *slog.Logger) bool {
	if e.smq == nil {
		return false
	}
	if err := e.smq.UpsertGuild(ctx, dto.GuildIndexMessage{GuildId: guildID}); err != nil {
		if logger != nil {
			logger.Error("unable to publish guild search upsert", slog.Int64("guild_id", guildID), slog.String("error", err.Error()))
		}
		return false
	}
	return true
}

func (e *entity) notifyGuildDiscoveryDelete(ctx context.Context, guildID int64, logger *slog.Logger) bool {
	if e.smq == nil {
		return false
	}
	if err := e.smq.DeleteGuild(ctx, dto.GuildIndexDeleteMessage{GuildId: guildID}); err != nil {
		if logger != nil {
			logger.Error("unable to publish guild search delete", slog.Int64("guild_id", guildID), slog.String("error", err.Error()))
		}
		return false
	}
	return true
}
