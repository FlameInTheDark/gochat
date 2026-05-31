package developer

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/FlameInTheDark/gochat/internal/botauth"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
)

const (
	errBadRequest       = "incorrect request"
	errUnableToGetUser  = "unable to get user"
	errUnableToGetBot   = "unable to get bot"
	errUnableToParseID  = "unable to parse id"
	errUnableToSaveBot  = "unable to save bot"
	errUnableToGetToken = "unable to get bot token"
	errBotTagInvalid    = "tags must be 2-32 characters and contain only letters, numbers, hyphens, and underscores"
	errTooManyBotTags   = "bot can have at most 10 tags"
)

var botDiscoveryTagRegex = regexp.MustCompile(`^[a-z0-9_-]{2,32}$`)

const maxBotDiscoveryTags = 10

type CreateBotRequest struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Public             bool     `json:"public"`
	DefaultPermissions int64    `json:"default_permissions"`
	Tags               []string `json:"tags"`
}

type UpdateBotRequest struct {
	Name               *string  `json:"name,omitempty"`
	Bio                *string  `json:"bio,omitempty"`
	Avatar             *int64   `json:"avatar,omitempty"`
	Banner             *int64   `json:"banner,omitempty"`
	BannerColor        *int     `json:"banner_color,omitempty"`
	PanelColor         *int     `json:"panel_color,omitempty"`
	Description        *string  `json:"description,omitempty"`
	Public             *bool    `json:"public,omitempty"`
	DefaultPermissions *int64   `json:"default_permissions,omitempty"`
	Disabled           *bool    `json:"disabled,omitempty"`
	Tags               []string `json:"tags,omitempty"`
}

type CreateTokenRequest struct {
	Name string `json:"name"`
}

type CreateGrantRequest struct {
	RequestedPermissions int64 `json:"requested_permissions"`
	ExpiresInSeconds     int64 `json:"expires_in_seconds"`
	MaxUses              int   `json:"max_uses"`
}

type BotResponse struct {
	BotUserId          int64    `json:"bot_user_id"`
	OwnerUserId        int64    `json:"owner_user_id"`
	User               dto.User `json:"user"`
	Description        string   `json:"description"`
	Tags               []string `json:"tags"`
	Public             bool     `json:"public"`
	DefaultPermissions int64    `json:"default_permissions"`
	Disabled           bool     `json:"disabled"`
	CreatedAt          int64    `json:"created_at"`
	UpdatedAt          int64    `json:"updated_at"`
}

type TokenCreateResponse struct {
	Token     string         `json:"token"`
	TokenData model.BotToken `json:"token_data"`
}

type GrantCreateResponse struct {
	Token string                `json:"token"`
	Grant model.BotInstallGrant `json:"grant"`
}

type BotAuthorizationPreview struct {
	Bot                  BotResponse `json:"bot"`
	RequestedPermissions int64       `json:"requested_permissions"`
	GrantId              *int64      `json:"grant_id,omitempty"`
}

func parseIntParam(c *fiber.Ctx, name string) (int64, error) {
	return strconv.ParseInt(c.Params(name), 10, 64)
}

func validateBotName(name string) error {
	name = strings.TrimSpace(name)
	if len([]rune(name)) < 2 || len([]rune(name)) > 32 {
		return fiber.NewError(fiber.StatusBadRequest, "bot name must be between 2 and 32 characters")
	}
	return nil
}

func normalizeBotDiscoveryTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func validateBotDiscoveryTags(tags []string) error {
	return validation.Validate(tags,
		validation.Length(0, maxBotDiscoveryTags).Error(errTooManyBotTags),
		validation.Each(validation.Match(botDiscoveryTagRegex).Error(errBotTagInvalid)),
	)
}

func (e *entity) botResponse(ctx *fiber.Ctx, bot model.Bot) (BotResponse, error) {
	user, err := e.user.GetUserById(ctx.UserContext(), bot.BotUserId)
	if err != nil {
		return BotResponse{}, err
	}
	tagsByBot, err := e.bot.GetTagsByBots(ctx.UserContext(), []int64{bot.BotUserId})
	if err != nil {
		return BotResponse{}, err
	}
	tags := tagsByBot[bot.BotUserId]
	if tags == nil {
		tags = []string{}
	}
	var disc string
	if d, err := e.disc.GetDiscriminatorByUserId(ctx.UserContext(), bot.BotUserId); err == nil {
		disc = d.Discriminator
	}
	avatar := e.avatarData(ctx, user.Id, user.Avatar)
	banner := e.bannerData(ctx, user.Id, user.Banner)
	return BotResponse{
		BotUserId:   bot.BotUserId,
		OwnerUserId: bot.OwnerUserId,
		User: dto.User{
			Id:            user.Id,
			Name:          user.Name,
			Discriminator: disc,
			Bio:           user.Bio,
			BannerColor:   user.BannerColor,
			PanelColor:    user.PanelColor,
			Avatar:        avatar,
			Banner:        banner,
			IsBot:         user.IsBot(),
		},
		Description:        bot.Description,
		Tags:               tags,
		Public:             bot.Public,
		DefaultPermissions: bot.DefaultPermissions,
		Disabled:           bot.Disabled,
		CreatedAt:          bot.CreatedAt.UnixMilli(),
		UpdatedAt:          bot.UpdatedAt.UnixMilli(),
	}, nil
}

func normalizeGrantRequest(req *CreateGrantRequest, defaultPermissions int64) {
	if req.ExpiresInSeconds <= 0 {
		req.ExpiresInSeconds = int64((24 * time.Hour).Seconds())
	}
	if req.MaxUses < 0 {
		req.MaxUses = 0
	} else if req.MaxUses == 0 {
		req.MaxUses = 1
	}
	if req.RequestedPermissions == 0 {
		req.RequestedPermissions = defaultPermissions
	}
}

func grantHasUsesRemaining(grant model.BotInstallGrant) bool {
	return grant.MaxUses == 0 || grant.Uses < grant.MaxUses
}

func (e *entity) botAuthorizationPreview(ctx *fiber.Ctx) (BotAuthorizationPreview, error) {
	var (
		bot         model.Bot
		grantID     *int64
		permissions int64
		err         error
	)

	if token := strings.TrimSpace(ctx.Query("grant_token")); token != "" {
		grant, err := e.bot.GetGrantByHash(ctx.UserContext(), botauth.HashToken(token))
		if err != nil {
			return BotAuthorizationPreview{}, fiber.NewError(fiber.StatusUnauthorized, "invalid bot install grant")
		}
		if grant.RevokedAt != nil || time.Now().After(grant.ExpiresAt) || !grantHasUsesRemaining(grant) {
			return BotAuthorizationPreview{}, fiber.NewError(fiber.StatusUnauthorized, "bot install grant is expired")
		}
		bot, err = e.bot.GetBot(ctx.UserContext(), grant.BotUserId)
		if err != nil {
			return BotAuthorizationPreview{}, fiber.NewError(fiber.StatusNotFound, errUnableToGetBot)
		}
		permissions = grant.RequestedPermissions
		grantID = &grant.Id
	} else {
		botID, parseErr := strconv.ParseInt(ctx.Query("bot_user_id"), 10, 64)
		if parseErr != nil || botID <= 0 {
			return BotAuthorizationPreview{}, fiber.NewError(fiber.StatusBadRequest, "bot_user_id is required")
		}
		bot, err = e.bot.GetBot(ctx.UserContext(), botID)
		if err != nil {
			return BotAuthorizationPreview{}, fiber.NewError(fiber.StatusNotFound, errUnableToGetBot)
		}
		if !bot.Public {
			return BotAuthorizationPreview{}, fiber.NewError(fiber.StatusForbidden, "bot is not public")
		}
		permissions = bot.DefaultPermissions
		if raw := strings.TrimSpace(ctx.Query("permissions")); raw != "" {
			permissions, err = strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return BotAuthorizationPreview{}, fiber.NewError(fiber.StatusBadRequest, "invalid permissions")
			}
		}
		if permissions&^bot.DefaultPermissions != 0 {
			return BotAuthorizationPreview{}, fiber.NewError(fiber.StatusForbidden, "requested permissions exceed bot defaults")
		}
	}

	if bot.Disabled {
		return BotAuthorizationPreview{}, fiber.NewError(fiber.StatusForbidden, "bot is disabled")
	}
	resp, err := e.botResponse(ctx, bot)
	if err != nil {
		return BotAuthorizationPreview{}, err
	}
	return BotAuthorizationPreview{Bot: resp, RequestedPermissions: permissions, GrantId: grantID}, nil
}
