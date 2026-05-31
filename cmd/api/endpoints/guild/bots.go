package guild

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/FlameInTheDark/gochat/internal/botauth"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/gofiber/fiber/v2"
)

type InstallBotRequest struct {
	BotUserId          int64  `json:"bot_user_id"`
	GrantToken         string `json:"grant_token"`
	GrantedPermissions int64  `json:"granted_permissions"`
}

func (r *InstallBotRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		BotUserId          jsonInt64 `json:"bot_user_id"`
		GrantToken         string    `json:"grant_token"`
		GrantedPermissions int64     `json:"granted_permissions"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.BotUserId = int64(raw.BotUserId)
	r.GrantToken = raw.GrantToken
	r.GrantedPermissions = raw.GrantedPermissions
	return nil
}

func botInstallGrantHasUsesRemaining(grant model.BotInstallGrant) bool {
	return grant.MaxUses == 0 || grant.Uses < grant.MaxUses
}

type jsonInt64 int64

func (v *jsonInt64) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		*v = 0
		return nil
	}
	if strings.HasPrefix(raw, `"`) {
		unquoted, err := strconv.Unquote(raw)
		if err != nil {
			return err
		}
		raw = strings.TrimSpace(unquoted)
		if raw == "" {
			*v = 0
			return nil
		}
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}
	*v = jsonInt64(parsed)
	return nil
}

type InstalledBotResponse struct {
	BotUserId          int64      `json:"bot_user_id"`
	User               dto.User   `json:"user"`
	Roles              []int64    `json:"roles"`
	GrantedPermissions int64      `json:"granted_permissions"`
	InstallerUserId    int64      `json:"installer_user_id"`
	GrantId            *int64     `json:"grant_id,omitempty"`
	CreatedAt          int64      `json:"created_at"`
	Member             dto.Member `json:"member"`
}

// ListBotAuthorizationGuilds
//
//	@Summary	List guilds available for bot authorization
//	@Produce	json
//	@Tags		Guild Bots
//	@Security	BearerAuth
//	@Success	200	{array}		dto.Guild
//	@Failure	401	{string}	string	"Unauthorized"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/guild/bots/authorize-guilds [get]
func (e *entity) ListBotAuthorizationGuilds(c *fiber.Ctx) error {
	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, ErrUnableToGetUserToken)
	}
	memberships, err := e.memb.GetUserGuilds(c.UserContext(), user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuilds)
	}
	if len(memberships) == 0 {
		return c.JSON([]dto.Guild{})
	}
	guildIDs := make([]int64, 0, len(memberships))
	for _, membership := range memberships {
		guildIDs = append(guildIDs, membership.GuildId)
	}
	guilds, err := e.g.GetGuildsList(c.UserContext(), guildIDs)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuilds)
	}

	out := make([]dto.Guild, 0, len(guilds))
	for _, guild := range guilds {
		if guild.OwnerId == user.Id {
			out = append(out, e.dtoGuildWithIcon(c, &guild))
			continue
		}
		_, ok, err := e.perm.GuildPerm(c.UserContext(), guild.Id, user.Id, permissions.PermAdministrator)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetPermission)
		}
		if ok {
			out = append(out, e.dtoGuildWithIcon(c, &guild))
		}
	}
	return c.JSON(out)
}

// ListGuildBots
//
//	@Summary	List installed guild bots
//	@Produce	json
//	@Tags		Guild Bots
//	@Security	BearerAuth
//	@Param		guild_id	path		int64	true	"Guild id"
//	@Success	200			{array}		InstalledBotResponse
//	@Failure	401			{string}	string	"Unauthorized"
//	@Failure	403			{string}	string	"Forbidden"
//	@Failure	500			{string}	string	"Internal server error"
//	@Router		/guild/{guild_id}/bots [get]
func (e *entity) ListGuildBots(c *fiber.Ctx) error {
	guildID, actorID, err := e.requireBotInstallAdmin(c)
	if err != nil {
		return err
	}
	_ = actorID
	installs, err := e.bot.ListGuildBots(c.UserContext(), guildID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to list guild bots")
	}
	out := make([]InstalledBotResponse, 0, len(installs))
	for _, install := range installs {
		item, err := e.installedBotResponse(c, install)
		if err == nil {
			out = append(out, item)
		}
	}
	return c.JSON(out)
}

// InstallBot
//
//	@Summary	Install a bot into a guild
//	@Accept		json
//	@Produce	json
//	@Tags		Guild Bots
//	@Security	BearerAuth
//	@Param		guild_id	path		int64				true	"Guild id"
//	@Param		request		body		InstallBotRequest	true	"Install request"
//	@Success	201			{object}	InstalledBotResponse
//	@Failure	400			{string}	string	"Bad request"
//	@Failure	401			{string}	string	"Unauthorized"
//	@Failure	403			{string}	string	"Forbidden"
//	@Failure	404			{string}	string	"Not found"
//	@Failure	500			{string}	string	"Internal server error"
//	@Router		/guild/{guild_id}/bots [post]
func (e *entity) InstallBot(c *fiber.Ctx) error {
	guildID, actorID, err := e.requireBotInstallAdmin(c)
	if err != nil {
		return err
	}
	var req InstallBotRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	var bot model.Bot
	var grantID *int64
	granted := req.GrantedPermissions
	if req.GrantToken != "" {
		grant, err := e.bot.GetGrantByHash(c.UserContext(), botauth.HashToken(req.GrantToken))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid bot install grant")
		}
		if grant.RevokedAt != nil || time.Now().After(grant.ExpiresAt) || !botInstallGrantHasUsesRemaining(grant) {
			return fiber.NewError(fiber.StatusUnauthorized, "bot install grant is expired")
		}
		bot, err = e.bot.GetBot(c.UserContext(), grant.BotUserId)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "bot not found")
		}
		granted = grant.RequestedPermissions
		grantID = &grant.Id
	} else {
		if req.BotUserId == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "bot_user_id is required")
		}
		bot, err = e.bot.GetBot(c.UserContext(), req.BotUserId)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "bot not found")
		}
		if !bot.Public {
			return fiber.NewError(fiber.StatusForbidden, "bot is not public")
		}
		if granted == 0 {
			granted = bot.DefaultPermissions
		}
		if granted&^bot.DefaultPermissions != 0 {
			return fiber.NewError(fiber.StatusForbidden, "requested permissions exceed bot defaults")
		}
	}
	if bot.Disabled {
		return fiber.NewError(fiber.StatusForbidden, "bot is disabled")
	}

	isMember, err := e.memb.IsGuildMember(c.UserContext(), guildID, bot.BotUserId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMember)
	}
	if !isMember {
		if err := e.memb.AddMember(c.UserContext(), bot.BotUserId, guildID); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildMember)
		}
	}
	if err := e.bot.UpsertBotGuild(c.UserContext(), model.BotGuild{
		BotUserId:          bot.BotUserId,
		GuildId:            guildID,
		GrantedPermissions: granted,
		InstallerUserId:    actorID,
		GrantId:            grantID,
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to install bot")
	}
	if grantID != nil {
		_ = e.bot.UseGrant(c.UserContext(), *grantID)
	}
	e.publishBotSearchUpsert(c.UserContext(), bot.BotUserId, e.log)
	install, err := e.bot.GetBotGuild(c.UserContext(), bot.BotUserId, guildID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get installed bot")
	}
	resp, err := e.installedBotResponse(c, install)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to get installed bot")
	}
	e.publishBotMemberAdded(c, guildID, resp.Member)
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// RemoveBot
//
//	@Summary	Remove a bot from a guild
//	@Produce	json
//	@Tags		Guild Bots
//	@Security	BearerAuth
//	@Param		guild_id	path	int64	true	"Guild id"
//	@Param		bot_id		path	int64	true	"Bot user id"
//	@Success	204
//	@Failure	400	{string}	string	"Bad request"
//	@Failure	401	{string}	string	"Unauthorized"
//	@Failure	403	{string}	string	"Forbidden"
//	@Failure	500	{string}	string	"Internal server error"
//	@Router		/guild/{guild_id}/bots/{bot_id} [delete]
func (e *entity) RemoveBot(c *fiber.Ctx) error {
	guildID, _, err := e.requireBotInstallAdmin(c)
	if err != nil {
		return err
	}
	botID, err := strconv.ParseInt(c.Params("bot_id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetGuildByID)
	}
	if err := e.bot.DeleteBotGuild(c.UserContext(), botID, guildID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to remove bot")
	}
	e.publishBotSearchUpsert(c.UserContext(), botID, e.log)
	if isMember, err := e.memb.IsGuildMember(c.UserContext(), guildID, botID); err == nil && isMember {
		_ = e.memb.RemoveMember(c.UserContext(), botID, guildID)
	}
	asyncCtx := observability.BackgroundFromContext(c.UserContext())
	go func() {
		_ = mq.SendGuildUpdate(asyncCtx, e.mqt, guildID, &mqmsg.RemoveGuildMember{GuildId: guildID, UserId: botID})
	}()
	return c.SendStatus(fiber.StatusNoContent)
}

func (e *entity) requireBotInstallAdmin(c *fiber.Ctx) (guildID, actorID int64, err error) {
	user, err := helper.GetUser(c)
	if err != nil {
		return 0, 0, fiber.NewError(fiber.StatusUnauthorized, ErrUnableToGetUserToken)
	}
	guildID, err = strconv.ParseInt(c.Params("guild_id"), 10, 64)
	if err != nil {
		return 0, 0, fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetGuildByID)
	}
	guild, err := e.g.GetGuildById(c.UserContext(), guildID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, fiber.NewError(fiber.StatusNotFound, ErrUnableToGetGuildByID)
		}
		return 0, 0, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetGuildByID)
	}
	if guild.OwnerId == user.Id {
		return guildID, user.Id, nil
	}
	_, ok, err := e.perm.GuildPerm(c.UserContext(), guildID, user.Id, permissions.PermAdministrator)
	if err != nil {
		return 0, 0, fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetPermission)
	}
	if !ok {
		return 0, 0, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}
	return guildID, user.Id, nil
}

func (e *entity) installedBotResponse(c *fiber.Ctx, install model.BotGuild) (InstalledBotResponse, error) {
	u, err := e.user.GetUserById(c.UserContext(), install.BotUserId)
	if err != nil {
		return InstalledBotResponse{}, err
	}
	dsc := ""
	if d, err := e.disc.GetDiscriminatorByUserId(c.UserContext(), install.BotUserId); err == nil {
		dsc = d.Discriminator
	}
	member, err := e.memb.GetMember(c.UserContext(), install.BotUserId, install.GuildId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return InstalledBotResponse{}, err
	}
	roles, _ := e.ur.GetUserRoles(c.UserContext(), install.GuildId, install.BotUserId)
	roleIDs := make([]int64, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.RoleId)
	}
	var avatarData *dto.AvatarData
	if member.Avatar != nil {
		if ad, err := e.getAvatarDataCached(c.UserContext(), u.Id, *member.Avatar); err == nil && ad != nil {
			avatarData = ad
		}
	}
	if avatarData == nil && u.Avatar != nil {
		if ad, err := e.getAvatarDataCached(c.UserContext(), u.Id, *u.Avatar); err == nil && ad != nil {
			avatarData = ad
		}
	}
	memberDTO := memberToDTO(member, u, dsc, avatarData, roleIDs)
	if u.Banner == nil {
		memberDTO.User.Banner = &dto.BannerData{Exists: false}
	} else if e.bn != nil {
		if bd, err := e.getBannerDataCached(c.UserContext(), u.Id, *u.Banner); err == nil && bd != nil {
			memberDTO.User.Banner = bd
		}
	}
	userDTO := memberDTO.User
	return InstalledBotResponse{
		BotUserId:          install.BotUserId,
		User:               userDTO,
		Roles:              roleIDs,
		GrantedPermissions: install.GrantedPermissions,
		InstallerUserId:    install.InstallerUserId,
		GrantId:            install.GrantId,
		CreatedAt:          install.CreatedAt.UnixMilli(),
		Member:             memberDTO,
	}, nil
}

func (e *entity) publishBotMemberAdded(c *fiber.Ctx, guildID int64, member dto.Member) {
	asyncCtx := observability.BackgroundFromContext(c.UserContext())
	go func() {
		_ = mq.SendGuildUpdate(asyncCtx, e.mqt, guildID, &mqmsg.AddGuildMember{
			GuildId: guildID,
			UserId:  member.User.Id,
			Member:  member,
		})
	}()
}
