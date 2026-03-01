package guild

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/voice/discovery"
)

// pickSFU selects an SFU instance using weighted random selection.
// Weight = max(1, 1000 - load), spreading load across instances rather than
// always picking the single lowest-load one under concurrent requests.
func pickSFU(instances []discovery.Instance) (id, url string) {
	total := 0
	for _, inst := range instances {
		if inst.URL == "" {
			continue
		}
		w := 1000 - int(inst.Load)
		if w < 1 {
			w = 1
		}
		total += w
	}
	if total == 0 {
		return
	}
	r := rand.Intn(total)
	cumulative := 0
	for _, inst := range instances {
		if inst.URL == "" {
			continue
		}
		w := 1000 - int(inst.Load)
		if w < 1 {
			w = 1
		}
		cumulative += w
		if r < cumulative {
			return inst.ID, inst.URL
		}
	}
	// Fallback to last non-empty instance
	for i := len(instances) - 1; i >= 0; i-- {
		if instances[i].URL != "" {
			return instances[i].ID, instances[i].URL
		}
	}
	return
}

// sfuAdminBaseURL converts an SFU signaling WebSocket URL to an HTTP base URL
// for admin API calls (e.g. wss://host/signal → https://host).
func sfuAdminBaseURL(wsURL string) string {
	u := strings.TrimSuffix(wsURL, "/signal")
	u = strings.Replace(u, "wss://", "https://", 1)
	u = strings.Replace(u, "ws://", "http://", 1)
	return u
}

// issueAdminJWT signs a short-lived admin JWT for API → SFU close-channel calls.
func issueAdminJWT(channelID int64, authSecret string) (string, error) {
	now := time.Now()
	claims := struct {
		helper.Claims
		ChannelID int64 `json:"channel_id"`
	}{
		Claims: helper.Claims{
			UserID:    0,
			TokenType: "admin",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "gochat",
				Audience:  []string{"sfu"},
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Minute)),
			},
		},
		ChannelID: channelID,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(authSecret))
}

// notifyOldSFUClose fires an async HTTP POST to the old SFU's admin endpoint
// to close all peer connections for the given channel. Errors are logged only.
func notifyOldSFUClose(oldSFUURL string, channelID int64, authSecret string, log *slog.Logger) {
	adminToken, err := issueAdminJWT(channelID, authSecret)
	if err != nil {
		log.Error("voice region change: failed to issue admin jwt", slog.String("error", err.Error()))
		return
	}

	type closeReq struct {
		ChannelID int64 `json:"channel_id"`
	}
	body, _ := json.Marshal(closeReq{ChannelID: channelID})
	baseURL := sfuAdminBaseURL(oldSFUURL)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/admin/channel/close", bytes.NewReader(body))
	if err != nil {
		log.Error("voice region change: failed to build admin request", slog.String("error", err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("voice region change: admin close request failed", slog.String("error", err.Error()), slog.String("sfu", baseURL))
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		log.Warn("voice region change: admin close unexpected status", slog.Int("status", resp.StatusCode), slog.String("sfu", baseURL))
	}
}

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
	// Key format: voice:route:<channelId> with JSON {"id":"...","url":"...","region":"..."}
	var chosen voiceRouteBinding
	if e.cache != nil {
		_ = e.cache.GetJSON(c.UserContext(), bindingKey(channelId), &chosen)
	}

	// If no binding, pick from discovery registry and bind to channel via SetNX to prevent races.
	if chosen.URL == "" {
		region := e.defaultVoiceRegion
		if dbreg, err := e.ch.GetChannelVoiceRegion(c.UserContext(), channelId); err == nil && dbreg != nil && *dbreg != "" {
			region = *dbreg
		}
		if e.disco != nil {
			if list, err := e.disco.List(c.UserContext(), region); err == nil && len(list) > 0 {
				pickedID, pickedURL := pickSFU(list)
				if pickedURL != "" {
					chosen = voiceRouteBinding{ID: pickedID, URL: pickedURL, Region: region}
					if e.cache != nil {
						// SetNX: only write if key absent; if lost the race, re-read the winner.
						set, _ := e.cache.SetTimedJSONNX(c.UserContext(), bindingKey(channelId), chosen, 60)
						if !set {
							var winner voiceRouteBinding
							if err := e.cache.GetJSON(c.UserContext(), bindingKey(channelId), &winner); err == nil && winner.URL != "" {
								chosen = winner
							}
						}
					}
				}
			}
		}
		if chosen.URL == "" {
			return fiber.NewError(fiber.StatusServiceUnavailable, ErrNoSFUAvailableInRegion)
		}
	}

	// Issue a short-lived SFU token. Extend to 5 minutes during an active region migration.
	tokenTTL := 2 * time.Minute
	if e.cache != nil {
		if v, err := e.cache.Get(c.UserContext(), rebindMarkerKey(channelId)); err == nil && v != "" {
			tokenTTL = 5 * time.Minute
		}
	}

	now := time.Now()
	sfuClaims := struct {
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
				ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
			},
		},
		ChannelID: channelId,
		GuildID:   &guildId,
		Perms:     vperm,
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, sfuClaims)
	signed, err := tok.SignedString([]byte(e.authSecret))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToIssueVoiceToken)
	}

	return c.JSON(JoinVoiceResponse{SFUURL: chosen.URL, SFUToken: signed})
}

func bindingKey(ch int64) string      { return "voice:route:" + fmtInt64(ch) }
func rebindMarkerKey(ch int64) string { return "voice:rebind:" + fmtInt64(ch) }
func sessionHashKey(ch int64) string  { return "voice:clients:" + fmtInt64(ch) }
func fmtInt64(v int64) string         { return strconv.FormatInt(v, 10) }

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
	moveClaims := struct {
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
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, moveClaims)
	signed, err := tok.SignedString([]byte(e.authSecret))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "unable to issue token")
	}

	// Select SFU URL for target channel via weighted random selection
	var pickedURL string
	if e.disco != nil {
		region := e.defaultVoiceRegion
		if dbreg, err := e.ch.GetChannelVoiceRegion(c.UserContext(), body.ChannelID); err == nil && dbreg != nil && *dbreg != "" {
			region = *dbreg
		}
		if list, err := e.disco.List(c.UserContext(), region); err == nil && len(list) > 0 {
			_, pickedURL = pickSFU(list)
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

	// Resolve source SFU URL from cache, falling back to discovery
	var fromURL string
	if e.cache != nil {
		var b voiceRouteBinding
		_ = e.cache.GetJSON(c.UserContext(), bindingKey(body.From), &b)
		fromURL = b.URL
	}
	if fromURL == "" {
		region := e.defaultVoiceRegion
		if dbreg, err := e.ch.GetChannelVoiceRegion(c.UserContext(), body.From); err == nil && dbreg != nil && *dbreg != "" {
			region = *dbreg
		}
		if e.disco != nil {
			if list, err := e.disco.List(c.UserContext(), region); err == nil && len(list) > 0 {
				_, fromURL = pickSFU(list)
			}
		}
		if fromURL == "" {
			return fiber.NewError(fiber.StatusServiceUnavailable, "no sfu available for source channel")
		}
	}

	// Issue admin token for the caller to signal the source channel
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
//	@failure		422			{string}	string	"No SFU instances in requested region"
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

	if reg != "" {
		// I2: Validate region against allowed list
		if len(e.allowedRegions) > 0 {
			if _, ok := e.allowedRegions[reg]; !ok {
				return fiber.NewError(fiber.StatusBadRequest, "unknown region")
			}
		}
		// I2: Validate that live SFU instances exist in the requested region before writing DB
		if e.disco != nil {
			instances, err := e.disco.List(c.UserContext(), reg)
			if err != nil || len(instances) == 0 {
				return fiber.NewError(fiber.StatusUnprocessableEntity, "no SFU instances in requested region")
			}
		}
	}

	// I9: Skip DB write and rebind if the region hasn't changed
	if e.cache != nil && reg != "" {
		var prev voiceRouteBinding
		_ = e.cache.GetJSON(c.UserContext(), bindingKey(channelId), &prev)
		if prev.Region == reg {
			return c.JSON(SetVoiceRegionResponse{GuildID: guildId, ChannelID: channelId, Region: reg})
		}
	}

	// Persist region to DB
	if reg == "" {
		_ = e.ch.SetChannelVoiceRegion(c.UserContext(), channelId, nil)
	} else {
		_ = e.ch.SetChannelVoiceRegion(c.UserContext(), channelId, &reg)
	}

	// Read existing route binding and active sessions to decide whether to rebind
	if e.cache != nil && e.disco != nil {
		var prev voiceRouteBinding
		_ = e.cache.GetJSON(c.UserContext(), bindingKey(channelId), &prev)
		oldURL := prev.URL // capture before any potential overwrite

		// I1: Use session hash as authoritative activity signal (maintained by webhook handlers)
		sessions, _ := e.cache.HGetAll(c.UserContext(), sessionHashKey(channelId))

		if len(sessions) > 0 {
			// Resolve the new target region
			newRegion := e.defaultVoiceRegion
			if reg != "" {
				newRegion = reg
			} else if dbreg, err := e.ch.GetChannelVoiceRegion(c.UserContext(), channelId); err == nil && dbreg != nil && *dbreg != "" {
				newRegion = *dbreg
			}

			// Discover a new SFU instance for the updated region
			list, err := e.disco.List(c.UserContext(), newRegion)
			if err != nil || len(list) == 0 {
				// No instance available; clear binding so next join rediscovers
				_ = e.cache.Delete(c.UserContext(), bindingKey(channelId))
			} else {
				chosenID, chosenURL := pickSFU(list)
				if chosenURL == "" {
					_ = e.cache.Delete(c.UserContext(), bindingKey(channelId))
				} else {
					newBinding := voiceRouteBinding{ID: chosenID, URL: chosenURL, Region: newRegion}

					// I7: Pre-notify guild members so clients can prepare for the reconnect
					_ = e.mqt.SendGuildUpdate(guildId, &mqmsg.VoiceRegionChanging{
						ChannelId: channelId,
						Region:    newRegion,
						DelayMs:   3000,
					})

					// Capture values for the background goroutine
					cache := e.cache
					mqt := e.mqt
					authSecret := e.authSecret
					log := e.log

					// Background goroutine: sleep 3s → update cache → publish VoiceRebind → close old SFU
					go func() {
						time.Sleep(3 * time.Second)

						ctx := context.Background()
						// I9: Write new binding with region
						_ = cache.SetTimedJSON(ctx, bindingKey(channelId), newBinding, 60)
						// I6: Mark active migration so JoinVoice issues extended JWT
						_ = cache.SetTimed(ctx, rebindMarkerKey(channelId), "1", 300)
						// I4: Notify clients to reconnect with jitter to spread thundering herd
						_ = mqt.SendChannelMessage(channelId, &mqmsg.VoiceRebind{Channel: channelId, JitterMs: 3000})
						// I5: Tell old SFU to close all sessions
						if oldURL != "" {
							notifyOldSFUClose(oldURL, channelId, authSecret, log)
						}
					}()
				}
			}
		} else {
			// No active sessions; just clear the stale binding
			_ = e.cache.Delete(c.UserContext(), bindingKey(channelId))
		}
	}

	return c.JSON(SetVoiceRegionResponse{
		GuildID:   guildId,
		ChannelID: channelId,
		Region:    reg,
	})
}
