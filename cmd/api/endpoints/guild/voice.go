package guild

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

// JoinVoice
//
//	@Summary		Join voice channel (get SFU signaling info)
//	@Description	Returns signaling path and a short-lived SFU token to connect to the SFU for this channel.
//	@Tags			Guild
//	@Param			guild_id	path		int64	true	"Guild ID"
//	@Param			channel_id	path		int64	true	"Channel ID"
//	@Success		200			{object}	JoinVoiceResponse
//	@failure		401			{string}	string	"Unauthorized"
//	@failure		403			{string}	string	"Forbidden"
//	@failure		503			{string}	string	"No SFU available in region"
//	@Router			/guild/{guild_id}/voice/{channel_id}/join [post]
func (e *entity) JoinVoice(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	// Validate channel is in guild, is voice, and user has Connect permission
	ch, _, _, ok, err := e.perm.ChannelPerm(c.UserContext(), guildId, channelId, user.Id, permissions.PermVoiceConnect)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !ok {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}
	if ch == nil || ch.Type != model.ChannelTypeGuildVoice {
		return fiber.NewError(fiber.StatusBadRequest, ErrNotAVoiceChannel)
	}

	// Build voice permission bitmask
	vperm, err := e.perm.GetChannelPermissions(c.UserContext(), guildId, channelId, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Try per-channel route in cache: if exists, reuse its URL.
	// Key format: voice:route:<channelId> with JSON {"id":"...","url":"..."}
	var chosen voiceRouteBinding
	if e.cache != nil {
		_ = e.cache.GetJSON(c.UserContext(), bindingKey(channelId), &chosen)
	}
	// If no binding, pick from discovery registry in cache (region -> instances) and bind to channel
	if chosen.URL == "" {
		region := e.defaultVoiceRegion
		if dbreg, err := e.ch.GetChannelVoiceRegion(c.UserContext(), channelId); err == nil && dbreg != nil && *dbreg != "" {
			region = *dbreg
		}
		// Read available instances for a region via discovery manager and pick the best load (first join)
		var pickedURL string
		var pickedID string
		if e.disco != nil {
			if list, err := e.disco.List(c.UserContext(), region); err == nil && len(list) > 0 {
				var bestLoad int64 = 1 << 60
				for _, inst := range list {
					if inst.URL == "" {
						continue
					}
					if inst.Load < bestLoad {
						bestLoad = inst.Load
						pickedURL = inst.URL
						pickedID = inst.ID
					}
				}
			}
		}
		// Discovery is mandatory: return 503 if discovery is unavailable or has no instances
		if pickedURL == "" {
			return fiber.NewError(fiber.StatusServiceUnavailable, ErrNoSFUAvailableInRegion)
		}
		chosen = voiceRouteBinding{ID: pickedID, URL: pickedURL}
		if e.cache != nil {
			// Save binding with TTL so it expires after inactivity
			_ = e.cache.SetTimedJSON(c.UserContext(), bindingKey(channelId), chosen, 60)
		}
	}

	// Issue a short-lived SFU token (typ=sfu, aud=sfu) embedding channel_id and voice perms
	now := time.Now()
	claims := struct {
		helper.Claims
		ChannelID int64  `json:"channel_id"`
		GuildID   *int64 `json:"guild_id,omitempty"`
		Perms     int64  `json:"perms"`
	}{
		Claims: helper.Claims{
			UserID:    user.Id,
			TokenType: "sfu",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "gochat",
				Audience:  []string{"sfu"},
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Minute)),
			},
		},
		ChannelID: channelId,
		GuildID:   &guildId,
		Perms:     vperm,
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(e.authSecret))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToIssueVoiceToken)
	}

	return c.JSON(JoinVoiceResponse{SFUURL: chosen.URL, SFUToken: signed})
}

func bindingKey(ch int64) string { return "voice:route:" + fmtInt64(ch) }
func fmtInt64(v int64) string    { return strconv.FormatInt(v, 10) }

// MoveMember
//
//	@Summary		Move member to voice channel
//	@Description	Move a member to another voice channel and send them connection info (SFU URL + token). Requires administrator or PermVoiceMoveMembers.
//	@Tags			Guild
//	@Param			guild_id	path		int64				true	"Guild ID"
//	@Param			request		body		MoveMemberRequest	true	"Move request"
//	@Success		200			{object}	MoveMemberResponse
//	@failure		400			{string}	string	"Bad request"
//	@failure		401			{string}	string	"Unauthorized"
//	@failure		403			{string}	string	"Forbidden"
//	@Router			/guild/{guild_id}/voice/move [post]
func (e *entity) MoveMember(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	// Permission: administrator or move members at guild scope
	_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermVoiceMoveMembers)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	var body MoveMemberRequest
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	if body.UserID == 0 || body.ChannelID == 0 || body.From == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "missing user_id, channel_id or from")
	}

	// Validate target channel is a voice channel in this guild
	gc, err := e.gc.GetGuildChannel(c.UserContext(), guildId, body.ChannelID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid channel")
	}
	ch, err := e.ch.GetChannel(c.UserContext(), gc.ChannelId)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid channel")
	}
	if ch.Type != model.ChannelTypeGuildVoice {
		return fiber.NewError(fiber.StatusBadRequest, "not a voice channel")
	}

	// Validate source (from) channel is a voice channel in this guild
	gcFrom, err := e.gc.GetGuildChannel(c.UserContext(), guildId, body.From)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid source channel")
	}
	chFrom, err := e.ch.GetChannel(c.UserContext(), gcFrom.ChannelId)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid source channel")
	}
	if chFrom.Type != model.ChannelTypeGuildVoice {
		return fiber.NewError(fiber.StatusBadRequest, "source is not a voice channel")
	}

	// Compute moved user's effective permissions in target channel
	vperm, err := e.perm.GetChannelPermissions(c.UserContext(), guildId, body.ChannelID, body.UserID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Issue a short-lived SFU token with moved=true (typ=sfu, aud=sfu)
	now := time.Now()
	claims := struct {
		helper.Claims
		ChannelID int64 `json:"channel_id"`
		Perms     int64 `json:"perms"`
		Moved     bool  `json:"moved"`
	}{
		Claims: helper.Claims{
			UserID:    body.UserID,
			TokenType: "sfu",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "gochat",
				Audience:  []string{"sfu"},
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Minute)),
			},
		},
		ChannelID: body.ChannelID,
		Perms:     vperm,
		Moved:     true,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(e.authSecret))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to issue token")
	}

	// Select SFU URL for target channel (reuse discovery + base url fallback if discovery disabled)
	var pickedURL string
	if e.disco != nil {
		region := e.defaultVoiceRegion
		if dbreg, err := e.ch.GetChannelVoiceRegion(c.UserContext(), body.ChannelID); err == nil && dbreg != nil && *dbreg != "" {
			region = *dbreg
		}
		if list, err := e.disco.List(c.UserContext(), region); err == nil && len(list) > 0 {
			var bestLoad int64 = 1 << 60
			for _, inst := range list {
				if inst.URL == "" {
					continue
				}
				if inst.Load < bestLoad {
					bestLoad = inst.Load
					pickedURL = inst.URL
				}
			}
		}
		if pickedURL == "" {
			return fiber.NewError(fiber.StatusServiceUnavailable, "no sfu available in region")
		}
	} else {
		return fiber.NewError(fiber.StatusServiceUnavailable, "no sfu discovery configured")
	}

	// Send user update with connection info
	evt := &mqmsg.VoiceMove{UserID: body.UserID, Channel: body.ChannelID, SFUURL: pickedURL, SFUToken: signed}
	if err := e.mqt.SendUserUpdate(body.UserID, evt); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to send move notice")
	}
	// Build admin URL+token to connect to the source channel to send RTCMoved
	var fromURL string
	if e.cache != nil {
		var b voiceRouteBinding
		_ = e.cache.GetJSON(c.UserContext(), bindingKey(body.From), &b)
		if b.URL != "" {
			fromURL = b.URL
		}
	}
	if fromURL == "" {
		// Resolve region for 'from' and select via discovery
		region := e.defaultVoiceRegion
		if dbreg, err := e.ch.GetChannelVoiceRegion(c.UserContext(), body.From); err == nil && dbreg != nil && *dbreg != "" {
			region = *dbreg
		}
		if e.disco != nil {
			if list, err := e.disco.List(c.UserContext(), region); err == nil && len(list) > 0 {
				var bestLoad int64 = 1 << 60
				for _, inst := range list {
					if inst.URL == "" {
						continue
					}
					if inst.Load < bestLoad {
						bestLoad = inst.Load
						fromURL = inst.URL
					}
				}
			}
		}
		if fromURL == "" {
			return fiber.NewError(fiber.StatusServiceUnavailable, "no sfu available for source channel")
		}
	}
	// Compute admin perms for 'from' channel and generate token
	adminPerms, err := e.perm.GetChannelPermissions(c.UserContext(), guildId, body.From, user.Id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	now2 := time.Now()
	adminClaims := struct {
		helper.Claims
		ChannelID int64 `json:"channel_id"`
		Perms     int64 `json:"perms"`
	}{
		Claims: helper.Claims{
			UserID:    user.Id,
			TokenType: "sfu",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "gochat",
				Audience:  []string{"sfu"},
				IssuedAt:  jwt.NewNumericDate(now2),
				ExpiresAt: jwt.NewNumericDate(now2.Add(2 * time.Minute)),
			},
		},
		ChannelID: body.From,
		Perms:     adminPerms,
	}
	atok := jwt.NewWithClaims(jwt.SigningMethodHS256, adminClaims)
	adminSigned, err := atok.SignedString([]byte(e.authSecret))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to issue admin token")
	}

	return c.JSON(MoveMemberResponse{Ok: true, FromSFUURL: fromURL, FromSFUToken: adminSigned})
}

// SetVoiceRegion
//
//	@Summary		Set channel voice region
//	@Description	Sets or clears preferred SFU region for a voice channel. Empty region clears override.
//	@Tags			Guild
//	@Param			guild_id	path		int64					true	"Guild ID"
//	@Param			channel_id	path		int64					true	"Channel ID"
//	@Param			request		body		SetVoiceRegionRequest	true	"Region payload"
//	@Success		200			{object}	SetVoiceRegionResponse
//	@failure		400			{string}	string	"Bad request"
//	@failure		401			{string}	string	"Unauthorized"
//	@failure		403			{string}	string	"Forbidden"
//	@Router			/guild/{guild_id}/voice/{channel_id}/region [patch]
func (e *entity) SetVoiceRegion(c *fiber.Ctx) error {
	guildId, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	channelId, err := e.parseChannelID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	// Require manage channels
	_, hasPermission, err := e.perm.GuildPerm(c.UserContext(), guildId, user.Id, permissions.PermServerManageChannels)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if !hasPermission {
		return fiber.NewError(fiber.StatusNotAcceptable, ErrPermissionsRequired)
	}

	// Validate channel is guild voice channel
	gc, err := e.gc.GetGuildChannel(c.UserContext(), guildId, channelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}
	ch, err := e.ch.GetChannel(c.UserContext(), gc.ChannelId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToGetChannel)
	}
	if ch.Type != model.ChannelTypeGuildVoice {
		return fiber.NewError(fiber.StatusBadRequest, "not a voice channel")
	}

	var body SetVoiceRegionRequest
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}

	reg := strings.TrimSpace(body.Region)
	if reg == "" {
		_ = e.ch.SetChannelVoiceRegion(c.UserContext(), channelId, nil)
	} else {
		// Validate region against configured list if provided
		if len(e.allowedRegions) > 0 {
			if _, ok := e.allowedRegions[reg]; !ok {
				return fiber.NewError(fiber.StatusBadRequest, "unknown region")
			}
		}
		_ = e.ch.SetChannelVoiceRegion(c.UserContext(), channelId, &reg)
		// region stored only in DB; cache not used for region reads
	}
	// Handle route rebinding when region changes:
	// If a route exists (users likely present), proactively select a new SFU for the new region,
	// store it under voice:route:{channel}, and broadcast a rebind event so clients reconnect
	// consistently. Otherwise, just clear any existing route so next join selects a new one.
	if e.cache != nil {
		var prev voiceRouteBinding
		_ = e.cache.GetJSON(c.UserContext(), bindingKey(channelId), &prev)
		if prev.URL != "" && e.disco != nil {
			// Discover a new instance for the updated region
			region := e.defaultVoiceRegion
			if dbreg, err := e.ch.GetChannelVoiceRegion(c.UserContext(), channelId); err == nil && dbreg != nil && *dbreg != "" {
				region = *dbreg
			}
			if list, err := e.disco.List(c.UserContext(), region); err == nil && len(list) > 0 {
				var chosenID, chosenURL string
				var bestLoad int64 = 1 << 60
				for _, inst := range list {
					if inst.URL == "" {
						continue
					}
					if inst.Load < bestLoad {
						bestLoad = inst.Load
						chosenID = inst.ID
						chosenURL = inst.URL
					}
				}
				if chosenURL != "" {
					_ = e.cache.SetTimedJSON(c.UserContext(), bindingKey(channelId), voiceRouteBinding{ID: chosenID, URL: chosenURL}, 60)
					// Notify clients in this channel to reconnect (call JoinVoice to fetch new route)
					_ = e.mqt.SendChannelMessage(channelId, &mqmsg.VoiceRebind{Channel: channelId})
				} else {
					// No instance available; clear binding so clients rediscover later
					_ = e.cache.Delete(c.UserContext(), bindingKey(channelId))
				}
			} else {
				// Discovery unavailable or empty; clear binding
				_ = e.cache.Delete(c.UserContext(), bindingKey(channelId))
			}
		} else {
			_ = e.cache.Delete(c.UserContext(), bindingKey(channelId))
		}
	}

	return c.JSON(SetVoiceRegionResponse{
		GuildID:   guildId,
		ChannelID: channelId,
		Region:    reg,
	})
}
