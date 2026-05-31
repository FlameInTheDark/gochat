package search

import (
	"context"
	"strings"

	"github.com/FlameInTheDark/gochat/internal/botsearch"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/gofiber/fiber/v2"
)

const maxBotSearchLimit = 16

// SearchBots
//
//	@Summary		Search public bots
//	@Description	Searches only public enabled bots in OpenSearch, then hydrates ordered results from YugabyteDB YSQL. Limit defaults to 16 and is capped at 16.
//	@Produce		json
//	@Tags			Search
//	@Param			q		query		string							false	"Search text for bot name, description, bio, and tags"
//	@Param			tags	query		string							false	"Comma-separated normalized tags"	example(moderation,music)
//	@Param			sort	query		string							false	"Sort mode"							Enums(best_match,popularity,alphabetical)	default(best_match)
//	@Param			page	query		int								false	"Zero-based page number"			minimum(0)									default(0)
//	@Param			limit	query		int								false	"Results per page, capped at 16"	minimum(1)									maximum(16)	default(16)
//	@Success		200		{object}	dto.BotDiscoverySearchResponse	"Public bot search results"
//	@failure		400		{string}	string							"Bad request"
//	@failure		401		{string}	string							"Unauthorized"
//	@failure		500		{string}	string							"Internal server error"
//	@Router			/search/bots [get]
func (e *entity) SearchBots(c *fiber.Ctx) error {
	page, limit := parseBotSearchPagination(c)
	tags := normalizeGuildDiscoveryTags(strings.Split(c.Query("tags"), ","))
	sortMode := botsearch.SortMode(c.Query("sort", string(botsearch.SortBestMatch)))
	switch sortMode {
	case botsearch.SortBestMatch, botsearch.SortPopularity, botsearch.SortAlphabetical:
	default:
		return fiber.NewError(fiber.StatusBadRequest, ErrSortInvalid)
	}

	res, err := e.botSearch.SearchBots(c.UserContext(), botsearch.SearchRequest{
		Query: c.Query("q"),
		Tags:  tags,
		Sort:  sortMode,
		From:  page * limit,
		Size:  limit,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFindMessages)
	}

	bots, err := e.hydrateBotDiscovery(c.UserContext(), res.Ids)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetBots)
	}

	return c.JSON(dto.BotDiscoverySearchResponse{
		Bots:  bots,
		Pages: (res.Total + limit - 1) / limit,
	})
}

// SearchBotTags
//
//	@Summary		Autocomplete public bot tags
//	@Description	Returns tag suggestions from tags attached to public enabled bots only. Limit defaults to 16 and is capped at 16.
//	@Produce		json
//	@Tags			Search
//	@Param			q		query		string	false	"Tag prefix"
//	@Param			limit	query		int		false	"Maximum tags to return, capped at 16"	minimum(1)	maximum(16)	default(16)
//	@Success		200		{array}		string	"Tag suggestions"
//	@failure		401		{string}	string	"Unauthorized"
//	@failure		500		{string}	string	"Internal server error"
//	@Router			/search/bot-tags [get]
func (e *entity) SearchBotTags(c *fiber.Ctx) error {
	limit := parsePositiveInt(c.Query("limit"), maxBotSearchLimit)
	if limit > maxBotSearchLimit {
		limit = maxBotSearchLimit
	}
	tags, err := e.botSearch.SearchTags(c.UserContext(), strings.ToLower(strings.TrimSpace(c.Query("q"))), limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToFindMessages)
	}
	return c.JSON(tags)
}

func (e *entity) hydrateBotDiscovery(ctx context.Context, ids []int64) ([]dto.BotDiscovery, error) {
	if len(ids) == 0 {
		return []dto.BotDiscovery{}, nil
	}
	bots, err := e.bot.GetBotsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]model.Bot, len(bots))
	publicIDs := make([]int64, 0, len(bots))
	for _, bot := range bots {
		if !bot.Public || bot.Disabled {
			continue
		}
		byID[bot.BotUserId] = bot
		publicIDs = append(publicIDs, bot.BotUserId)
	}

	users, err := e.user.GetUsersList(ctx, publicIDs)
	if err != nil {
		return nil, err
	}
	usersByID := make(map[int64]model.User, len(users))
	for _, user := range users {
		usersByID[user.Id] = user
	}

	discs, err := e.disc.GetDiscriminatorsByUserIDs(ctx, publicIDs)
	if err != nil {
		return nil, err
	}
	discsByID := make(map[int64]string, len(discs))
	for _, disc := range discs {
		discsByID[disc.UserId] = disc.Discriminator
	}

	tags, err := e.bot.GetTagsByBots(ctx, publicIDs)
	if err != nil {
		return nil, err
	}
	counts, err := e.bot.GetInstallCounts(ctx, publicIDs)
	if err != nil {
		return nil, err
	}

	out := make([]dto.BotDiscovery, 0, len(publicIDs))
	for _, id := range ids {
		bot, ok := byID[id]
		if !ok {
			continue
		}
		user, ok := usersByID[id]
		if !ok {
			continue
		}
		botTags := tags[id]
		if botTags == nil {
			botTags = []string{}
		}
		out = append(out, dto.BotDiscovery{
			BotUserId:          bot.BotUserId,
			User:               e.botDiscoveryUserDTO(ctx, user, discsByID[id]),
			Description:        bot.Description,
			Tags:               botTags,
			DefaultPermissions: bot.DefaultPermissions,
			InstallsCount:      counts[id],
			CreatedAt:          bot.CreatedAt.UnixMilli(),
			UpdatedAt:          bot.UpdatedAt.UnixMilli(),
		})
	}
	return out, nil
}

func (e *entity) botDiscoveryUserDTO(ctx context.Context, user model.User, discriminator string) dto.User {
	result := dto.User{
		Id:            user.Id,
		Name:          user.Name,
		Discriminator: discriminator,
		Bio:           user.Bio,
		BannerColor:   user.BannerColor,
		PanelColor:    user.PanelColor,
		IsBot:         user.IsBot(),
	}
	if user.Avatar != nil {
		if avatar, err := e.av.GetAvatar(ctx, *user.Avatar, user.Id); err == nil && avatar.Done && avatar.URL != nil && *avatar.URL != "" {
			result.Avatar = &dto.AvatarData{
				Id:          avatar.Id,
				URL:         *avatar.URL,
				ContentType: avatar.ContentType,
				Width:       avatar.Width,
				Height:      avatar.Height,
				Size:        avatar.FileSize,
			}
		}
	}
	if user.Banner == nil {
		result.Banner = &dto.BannerData{Exists: false}
	} else if banner, err := e.bn.GetBanner(ctx, *user.Banner, user.Id); err == nil && banner.Done && banner.URL != nil && *banner.URL != "" {
		result.Banner = &dto.BannerData{
			Exists:      true,
			Id:          banner.Id,
			URL:         *banner.URL,
			ContentType: banner.ContentType,
			Width:       banner.Width,
			Height:      banner.Height,
			Size:        banner.FileSize,
		}
	} else {
		result.Banner = &dto.BannerData{Exists: false}
	}
	return result
}

func parseBotSearchPagination(c *fiber.Ctx) (int, int) {
	page := parsePositiveInt(c.Query("page"), 0)
	limit := parsePositiveInt(c.Query("limit"), maxBotSearchLimit)
	if limit <= 0 || limit > maxBotSearchLimit {
		limit = maxBotSearchLimit
	}
	return page, limit
}

const errUnableToGetBots = "unable to get bots"
