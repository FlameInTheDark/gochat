package developer

import (
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/botauth"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/gofiber/fiber/v2"
)

// CreateBot
//
//	@Summary	Create a bot
//	@Accept		json
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		request	body		CreateBotRequest	true	"Bot configuration"
//	@Success	201		{object}	BotResponse
//	@Failure	400		{string}	string	"Bad request"
//	@Failure	401		{string}	string	"Unauthorized"
//	@Failure	500		{string}	string	"Internal server error"
//	@Router		/developer/bots [post]
func (e *entity) CreateBot(c *fiber.Ctx) error {
	owner, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, errUnableToGetUser)
	}
	var req CreateBotRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errBadRequest)
	}
	if err := validateBotName(req.Name); err != nil {
		return err
	}
	botID := idgen.Next()
	if err := e.user.CreateUserWithFlags(c.UserContext(), botID, req.Name, model.UserFlagBot); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToSaveBot)
	}
	if err := e.disc.CreateDiscriminator(c.UserContext(), botID, fmt.Sprintf("bot-%d", botID)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToSaveBot)
	}
	bot := model.Bot{
		BotUserId:          botID,
		OwnerUserId:        owner.Id,
		Description:        req.Description,
		Public:             req.Public,
		DefaultPermissions: req.DefaultPermissions,
	}
	if err := e.bot.CreateBot(c.UserContext(), bot); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToSaveBot)
	}
	bot, _ = e.bot.GetBot(c.UserContext(), botID)
	out, err := e.botResponse(c, bot)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetBot)
	}
	return c.Status(fiber.StatusCreated).JSON(out)
}

// ListBots
//
//	@Summary	List owned bots
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Success	200	{array}		BotResponse
//	@Failure	401	{string}	string	"Unauthorized"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/developer/bots [get]
func (e *entity) ListBots(c *fiber.Ctx) error {
	owner, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, errUnableToGetUser)
	}
	bots, err := e.bot.ListOwnerBots(c.UserContext(), owner.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetBot)
	}
	out := make([]BotResponse, 0, len(bots))
	for _, b := range bots {
		dto, err := e.botResponse(c, b)
		if err == nil {
			out = append(out, dto)
		}
	}
	return c.JSON(out)
}

// SearchPublicBots
//
//	@Summary	Search public bots
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		query	query		string	false	"Search query"
//	@Param		limit	query		int		false	"Page size"
//	@Param		offset	query		int		false	"Offset"
//	@Success	200		{array}		BotResponse
//	@Failure	500		{string}	string	"Internal server error"
//	@Router		/developer/bots/public [get]
func (e *entity) SearchPublicBots(c *fiber.Ctx) error {
	limit := uint64(c.QueryInt("limit", 25))
	if limit == 0 || limit > 100 {
		limit = 25
	}
	offset := uint64(c.QueryInt("offset", 0))
	bots, err := e.bot.SearchPublicBots(c.UserContext(), c.Query("query"), limit, offset)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetBot)
	}
	out := make([]BotResponse, 0, len(bots))
	for _, b := range bots {
		dto, err := e.botResponse(c, b)
		if err == nil {
			out = append(out, dto)
		}
	}
	return c.JSON(out)
}

// GetBot
//
//	@Summary	Get an owned bot
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		bot_id	path		int64	true	"Bot user id"
//	@Success	200		{object}	BotResponse
//	@Failure	400		{string}	string	"Bad request"
//	@Failure	401		{string}	string	"Unauthorized"
//	@Failure	404		{string}	string	"Not found"
//	@Failure	500		{string}	string	"Internal server error"
//	@Router		/developer/bots/{bot_id} [get]
func (e *entity) GetBot(c *fiber.Ctx) error {
	owner, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, errUnableToGetUser)
	}
	botID, err := parseIntParam(c, "bot_id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errUnableToParseID)
	}
	bot, err := e.bot.GetBotForOwner(c.UserContext(), owner.Id, botID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, errUnableToGetBot)
	}
	out, err := e.botResponse(c, bot)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetBot)
	}
	return c.JSON(out)
}

// UpdateBot
//
//	@Summary	Update an owned bot
//	@Accept		json
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		bot_id	path		int64				true	"Bot user id"
//	@Param		request	body		UpdateBotRequest	true	"Bot update"
//	@Success	200		{object}	BotResponse
//	@Failure	400		{string}	string	"Bad request"
//	@Failure	401		{string}	string	"Unauthorized"
//	@Failure	404		{string}	string	"Not found"
//	@Failure	500		{string}	string	"Internal server error"
//	@Router		/developer/bots/{bot_id} [patch]
func (e *entity) UpdateBot(c *fiber.Ctx) error {
	owner, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, errUnableToGetUser)
	}
	botID, err := parseIntParam(c, "bot_id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errUnableToParseID)
	}
	if _, err = e.bot.GetBotForOwner(c.UserContext(), owner.Id, botID); err != nil {
		return fiber.NewError(fiber.StatusNotFound, errUnableToGetBot)
	}
	var req UpdateBotRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errBadRequest)
	}
	if req.Name != nil {
		if err := validateBotName(*req.Name); err != nil {
			return err
		}
	}
	if req.Name != nil || req.Avatar != nil || req.Bio != nil || req.BannerColor != nil || req.PanelColor != nil {
		if err := e.user.ModifyUser(c.UserContext(), botID, req.Name, req.Avatar, req.Bio, req.BannerColor, req.PanelColor); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, errUnableToSaveBot)
		}
	}
	if req.Banner != nil {
		if err := e.user.SetUserBanner(c.UserContext(), botID, *req.Banner); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, errUnableToSaveBot)
		}
	}
	if err := e.bot.UpdateBot(c.UserContext(), botID, req.Description, req.Public, req.DefaultPermissions, req.Disabled); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToSaveBot)
	}
	bot, err := e.bot.GetBot(c.UserContext(), botID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetBot)
	}
	out, err := e.botResponse(c, bot)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetBot)
	}
	return c.JSON(out)
}

// DeleteBot
//
//	@Summary	Delete an owned bot
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		bot_id	path	int64	true	"Bot user id"
//	@Success	204
//	@Failure	400	{string}	string	"Bad request"
//	@Failure	401	{string}	string	"Unauthorized"
//	@Failure	404	{string}	string	"Not found"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/developer/bots/{bot_id} [delete]
func (e *entity) DeleteBot(c *fiber.Ctx) error {
	owner, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, errUnableToGetUser)
	}
	botID, err := parseIntParam(c, "bot_id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errUnableToParseID)
	}
	if _, err = e.bot.GetBotForOwner(c.UserContext(), owner.Id, botID); err != nil {
		return fiber.NewError(fiber.StatusNotFound, errUnableToGetBot)
	}
	if err := e.bot.DeleteBot(c.UserContext(), botID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToSaveBot)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// PreviewBotAuthorization
//
//	@Summary	Preview bot authorization
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		grant_token	query		string	false	"Install grant token"
//	@Param		bot_user_id	query		int64	false	"Public bot user id"
//	@Param		permissions	query		int64	false	"Requested permission bitmask"
//	@Success	200			{object}	BotAuthorizationPreview
//	@Failure	400			{string}	string	"Bad request"
//	@Failure	401			{string}	string	"Unauthorized"
//	@Failure	403			{string}	string	"Forbidden"
//	@Failure	404			{string}	string	"Not found"
//	@Router		/developer/bots/authorize/preview [get]
func (e *entity) PreviewBotAuthorization(c *fiber.Ctx) error {
	preview, err := e.botAuthorizationPreview(c)
	if err != nil {
		return err
	}
	return c.JSON(preview)
}

// CreateToken
//
//	@Summary	Create a bot runtime token
//	@Accept		json
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		bot_id	path		int64				true	"Bot user id"
//	@Param		request	body		CreateTokenRequest	true	"Token request"
//	@Success	201		{object}	TokenCreateResponse
//	@Failure	400		{string}	string	"Bad request"
//	@Failure	401		{string}	string	"Unauthorized"
//	@Failure	404		{string}	string	"Not found"
//	@Failure	500		{string}	string	"Internal server error"
//	@Router		/developer/bots/{bot_id}/tokens [post]
func (e *entity) CreateToken(c *fiber.Ctx) error {
	bot, err := e.requireOwnerBot(c)
	if err != nil {
		return err
	}
	var req CreateTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errBadRequest)
	}
	if req.Name == "" {
		req.Name = "Runtime token"
	}
	plain, prefix, hash, err := botauth.GenerateToken()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetToken)
	}
	token := model.BotToken{
		Id:          idgen.Next(),
		BotUserId:   bot.BotUserId,
		Name:        req.Name,
		TokenHash:   hash,
		TokenPrefix: prefix,
	}
	if err := e.bot.CreateToken(c.UserContext(), token); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetToken)
	}
	tokens, _ := e.bot.ListTokens(c.UserContext(), bot.BotUserId)
	for _, t := range tokens {
		if t.Id == token.Id {
			token = t
			break
		}
	}
	return c.Status(fiber.StatusCreated).JSON(TokenCreateResponse{Token: plain, TokenData: token})
}

// ListTokens
//
//	@Summary	List bot runtime tokens
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		bot_id	path		int64	true	"Bot user id"
//	@Success	200		{array}		model.BotToken
//	@Failure	401		{string}	string	"Unauthorized"
//	@Failure	404		{string}	string	"Not found"
//	@Failure	500		{string}	string	"Internal server error"
//	@Router		/developer/bots/{bot_id}/tokens [get]
func (e *entity) ListTokens(c *fiber.Ctx) error {
	bot, err := e.requireOwnerBot(c)
	if err != nil {
		return err
	}
	tokens, err := e.bot.ListTokens(c.UserContext(), bot.BotUserId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetToken)
	}
	return c.JSON(tokens)
}

// RevokeToken
//
//	@Summary	Revoke a bot runtime token
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		bot_id		path	int64	true	"Bot user id"
//	@Param		token_id	path	int64	true	"Token id"
//	@Success	204
//	@Failure	400	{string}	string	"Bad request"
//	@Failure	401	{string}	string	"Unauthorized"
//	@Failure	404	{string}	string	"Not found"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/developer/bots/{bot_id}/tokens/{token_id} [delete]
func (e *entity) RevokeToken(c *fiber.Ctx) error {
	bot, err := e.requireOwnerBot(c)
	if err != nil {
		return err
	}
	tokenID, err := parseIntParam(c, "token_id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errUnableToParseID)
	}
	if err := e.bot.RevokeToken(c.UserContext(), bot.BotUserId, tokenID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetToken)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// CreateGrant
//
//	@Summary	Create a bot install grant
//	@Accept		json
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		bot_id	path		int64				true	"Bot user id"
//	@Param		request	body		CreateGrantRequest	true	"Grant request"
//	@Success	201		{object}	GrantCreateResponse
//	@Failure	400		{string}	string	"Bad request"
//	@Failure	401		{string}	string	"Unauthorized"
//	@Failure	404		{string}	string	"Not found"
//	@Failure	500		{string}	string	"Internal server error"
//	@Router		/developer/bots/{bot_id}/grants [post]
func (e *entity) CreateGrant(c *fiber.Ctx) error {
	bot, err := e.requireOwnerBot(c)
	if err != nil {
		return err
	}
	var req CreateGrantRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errBadRequest)
	}
	normalizeGrantRequest(&req, bot.DefaultPermissions)
	plain, prefix, hash, err := botauth.GenerateToken()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetToken)
	}
	grant := model.BotInstallGrant{
		Id:                   idgen.Next(),
		BotUserId:            bot.BotUserId,
		OwnerUserId:          bot.OwnerUserId,
		TokenHash:            hash,
		TokenPrefix:          prefix,
		RequestedPermissions: req.RequestedPermissions,
		ExpiresAt:            time.Now().Add(time.Duration(req.ExpiresInSeconds) * time.Second),
		MaxUses:              req.MaxUses,
	}
	if err := e.bot.CreateGrant(c.UserContext(), grant); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetToken)
	}
	grants, _ := e.bot.ListGrants(c.UserContext(), bot.BotUserId)
	for _, g := range grants {
		if g.Id == grant.Id {
			grant = g
			break
		}
	}
	return c.Status(fiber.StatusCreated).JSON(GrantCreateResponse{Token: plain, Grant: grant})
}

// ListGrants
//
//	@Summary	List bot install grants
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		bot_id	path		int64	true	"Bot user id"
//	@Success	200		{array}		model.BotInstallGrant
//	@Failure	401		{string}	string	"Unauthorized"
//	@Failure	404		{string}	string	"Not found"
//	@Failure	500		{string}	string	"Internal server error"
//	@Router		/developer/bots/{bot_id}/grants [get]
func (e *entity) ListGrants(c *fiber.Ctx) error {
	bot, err := e.requireOwnerBot(c)
	if err != nil {
		return err
	}
	grants, err := e.bot.ListGrants(c.UserContext(), bot.BotUserId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetToken)
	}
	return c.JSON(grants)
}

// RevokeGrant
//
//	@Summary	Revoke a bot install grant
//	@Produce	json
//	@Tags		Developer Bots
//	@Security	BearerAuth
//	@Param		bot_id		path	int64	true	"Bot user id"
//	@Param		grant_id	path	int64	true	"Grant id"
//	@Success	204
//	@Failure	400	{string}	string	"Bad request"
//	@Failure	401	{string}	string	"Unauthorized"
//	@Failure	404	{string}	string	"Not found"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/developer/bots/{bot_id}/grants/{grant_id} [delete]
func (e *entity) RevokeGrant(c *fiber.Ctx) error {
	bot, err := e.requireOwnerBot(c)
	if err != nil {
		return err
	}
	grantID, err := parseIntParam(c, "grant_id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errUnableToParseID)
	}
	if err := e.bot.RevokeGrant(c.UserContext(), bot.BotUserId, grantID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, errUnableToGetToken)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (e *entity) requireOwnerBot(c *fiber.Ctx) (model.Bot, error) {
	owner, err := helper.GetUser(c)
	if err != nil {
		return model.Bot{}, fiber.NewError(fiber.StatusUnauthorized, errUnableToGetUser)
	}
	botID, err := parseIntParam(c, "bot_id")
	if err != nil {
		return model.Bot{}, fiber.NewError(fiber.StatusBadRequest, errUnableToParseID)
	}
	bot, err := e.bot.GetBotForOwner(c.UserContext(), owner.Id, botID)
	if err != nil {
		return model.Bot{}, fiber.NewError(fiber.StatusNotFound, errUnableToGetBot)
	}
	return bot, nil
}
