package search

import (
	"context"
	"strconv"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/guildsearch"
	"github.com/gofiber/fiber/v2"
)

const maxGuildSearchLimit = 16

// SearchGuilds
//
//	@Summary		Search public guilds
//	@Description	Searches only public guilds in OpenSearch, then hydrates ordered results from PostgreSQL/Citus. Limit defaults to 16 and is capped at 16.
//	@Produce		json
//	@Tags			Search
//	@Param			q		query		string								false	"Search text for guild name, description, and tags"
//	@Param			tags	query		string								false	"Comma-separated normalized tags"	example(go,voice)
//	@Param			sort	query		string								false	"Sort mode"							Enums(best_match,popularity,alphabetical)	default(best_match)
//	@Param			page	query		int									false	"Zero-based page number"			minimum(0)									default(0)
//	@Param			limit	query		int									false	"Results per page, capped at 16"	minimum(1)									maximum(16)	default(16)
//	@Success		200		{object}	dto.GuildDiscoverySearchResponse	"Public guild search results"
//	@failure		400		{string}	string								"Bad request"
//	@failure		401		{string}	string								"Unauthorized"
//	@failure		500		{string}	string								"Internal server error"
//	@Router			/search/guilds [get]
func (e *entity) SearchGuilds(c *fiber.Ctx) error {
	page, limit := parseGuildSearchPagination(c)
	tags := normalizeGuildDiscoveryTags(strings.Split(c.Query("tags"), ","))
	sortMode := guildsearch.SortMode(c.Query("sort", string(guildsearch.SortBestMatch)))
	switch sortMode {
	case guildsearch.SortBestMatch, guildsearch.SortPopularity, guildsearch.SortAlphabetical:
	default:
		return fiber.NewError(fiber.StatusBadRequest, ErrSortInvalid)
	}

	res, err := e.guildSearch.SearchGuilds(c.UserContext(), guildsearch.SearchRequest{
		Query: c.Query("q"),
		Tags:  tags,
		Sort:  sortMode,
		From:  page * limit,
		Size:  limit,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFindMessages)
	}

	guilds, err := e.hydrateGuildDiscovery(c.UserContext(), res.Ids)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}

	return c.JSON(dto.GuildDiscoverySearchResponse{
		Guilds: guilds,
		Pages:  (res.Total + limit - 1) / limit,
	})
}

// SearchGuildTags
//
//	@Summary		Autocomplete public guild tags
//	@Description	Returns tag suggestions from tags attached to public guilds only. Limit defaults to 16 and is capped at 16.
//	@Produce		json
//	@Tags			Search
//	@Param			q		query		string	false	"Tag prefix"
//	@Param			limit	query		int		false	"Maximum tags to return, capped at 16"	minimum(1)	maximum(16)	default(16)
//	@Success		200		{array}		string	"Tag suggestions"
//	@failure		401		{string}	string	"Unauthorized"
//	@failure		500		{string}	string	"Internal server error"
//	@Router			/search/guild-tags [get]
func (e *entity) SearchGuildTags(c *fiber.Ctx) error {
	limit := parsePositiveInt(c.Query("limit"), maxGuildSearchLimit)
	if limit > maxGuildSearchLimit {
		limit = maxGuildSearchLimit
	}
	tags, err := e.guildSearch.SearchTags(c.UserContext(), strings.ToLower(strings.TrimSpace(c.Query("q"))), limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFindMessages)
	}
	return c.JSON(tags)
}

func (e *entity) hydrateGuildDiscovery(ctx context.Context, ids []int64) ([]dto.GuildDiscovery, error) {
	if len(ids) == 0 {
		return []dto.GuildDiscovery{}, nil
	}
	guilds, err := e.g.GetGuildsList(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]model.Guild, len(guilds))
	publicIDs := make([]int64, 0, len(guilds))
	for _, guild := range guilds {
		if !guild.Public {
			continue
		}
		byID[guild.Id] = guild
		publicIDs = append(publicIDs, guild.Id)
	}

	tags, err := e.gd.GetTagsByGuilds(ctx, publicIDs)
	if err != nil {
		return nil, err
	}
	stats, err := e.gd.GetStatsByGuilds(ctx, publicIDs)
	if err != nil {
		return nil, err
	}

	out := make([]dto.GuildDiscovery, 0, len(publicIDs))
	for _, id := range ids {
		guild, ok := byID[id]
		if !ok {
			continue
		}
		result := dto.GuildDiscovery{
			Id:           guild.Id,
			Name:         guild.Name,
			Description:  guild.Description,
			MembersCount: stats[guild.Id].MembersCount,
			Tags:         tags[guild.Id],
		}
		if result.Tags == nil {
			result.Tags = []string{}
		}
		if guild.Icon != nil {
			if iconDTO, err := e.guildIconDTO(ctx, guild.Id, *guild.Icon); err == nil && iconDTO != nil {
				result.Icon = iconDTO
			}
		}
		out = append(out, result)
	}
	return out, nil
}

func (e *entity) guildIconDTO(ctx context.Context, guildID, iconID int64) (*dto.Icon, error) {
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

func parseGuildSearchPagination(c *fiber.Ctx) (int, int) {
	page := parsePositiveInt(c.Query("page"), 0)
	limit := parsePositiveInt(c.Query("limit"), maxGuildSearchLimit)
	if limit <= 0 || limit > maxGuildSearchLimit {
		limit = maxGuildSearchLimit
	}
	return page, limit
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return fallback
	}
	return value
}
