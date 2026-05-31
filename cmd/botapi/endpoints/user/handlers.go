package user

import (
	"github.com/FlameInTheDark/gochat/internal/botauth"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/gofiber/fiber/v2"
)

// Me
//
//	@Summary	Get current bot account
//	@Produce	json
//	@Tags		Bot User
//	@Security	BotToken
//	@Success	200	{object}	MeResponse
//	@Failure	401	{string}	string	"Unauthorized"
//	@Router		/bot/api/v1/user/me [get]
func (e *Entity) Me(c *fiber.Ctx) error {
	principal, ok := botauth.FromFiber(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "missing bot principal")
	}
	return c.JSON(MeResponse{
		BotUserId:   principal.BotUserID,
		OwnerUserId: principal.OwnerUserID,
		User: dto.User{
			Id:          principal.User.Id,
			Name:        principal.User.Name,
			Bio:         principal.User.Bio,
			BannerColor: principal.User.BannerColor,
			PanelColor:  principal.User.PanelColor,
			IsBot:       principal.User.IsBot(),
		},
		Description:        principal.Bot.Description,
		Public:             principal.Bot.Public,
		DefaultPermissions: principal.Bot.DefaultPermissions,
	})
}
