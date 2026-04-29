package guild

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/idgen"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/presence"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
)

const streamTokenTTL = time.Minute

func (e *entity) parseStreamID(c *fiber.Ctx) (int64, error) {
	streamID, err := strconv.ParseInt(c.Params("stream_id"), 10, 64)
	if err != nil {
		return 0, fiber.NewError(http.StatusBadRequest, ErrUnableToGetStream)
	}
	return streamID, nil
}

func (e *entity) loadStreamMetadata(ctx context.Context, streamID int64) (streammeta.Metadata, bool, error) {
	if e == nil || e.cache == nil || streamID == 0 {
		return streammeta.Metadata{}, false, nil
	}

	var meta streammeta.Metadata
	if err := e.cache.GetJSON(ctx, streammeta.MetaKey(streamID), &meta); err != nil || meta.ID == 0 {
		return streammeta.Metadata{}, false, nil
	}
	return meta, true, nil
}

func (e *entity) loadUserStream(ctx context.Context, userID int64) (streammeta.Metadata, bool, error) {
	if e == nil || e.cache == nil || userID == 0 {
		return streammeta.Metadata{}, false, nil
	}

	var meta streammeta.Metadata
	if err := e.cache.GetJSON(ctx, streammeta.UserKey(userID), &meta); err != nil || meta.ID == 0 {
		return streammeta.Metadata{}, false, nil
	}
	return meta, true, nil
}

func (e *entity) isVoiceMember(ctx context.Context, channelID, userID int64) (bool, error) {
	if e == nil || e.cache == nil {
		return false, nil
	}
	value, err := e.cache.HGet(ctx, sessionHashKey(channelID), fmtInt64(userID))
	if err != nil {
		return false, err
	}
	return value != "", nil
}

func hasPermissionMask(mask int64, required permissions.RolePermission) bool {
	if mask&int64(permissions.PermAdministrator) != 0 {
		return true
	}
	return mask&int64(required) != 0
}

func buildVoiceStreamSummary(meta streammeta.Metadata) VoiceStreamSummary {
	return VoiceStreamSummary{
		ID:          meta.ID,
		OwnerUserID: meta.OwnerUserID,
		ChannelID:   meta.ChannelID,
		SourceType:  meta.SourceType,
		AudioMode:   meta.AudioMode,
		StartedAt:   meta.StartedAt,
	}
}

func (e *entity) issueStreamToken(userID int64, meta streammeta.Metadata, role string, binding streammeta.RouteBinding) (string, error) {
	if binding.ID == "" {
		return "", fmt.Errorf("stream route id is required")
	}

	now := time.Now()
	claims := streammeta.Claims{
		Claims: helper.Claims{
			UserID:    userID,
			TokenType: "stream",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "gochat",
				Audience:  []string{"stream"},
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(streamTokenTTL)),
			},
		},
		StreamID:    meta.ID,
		ChannelID:   meta.ChannelID,
		GuildID:     meta.GuildID,
		OwnerUserID: meta.OwnerUserID,
		RouteID:     binding.ID,
		Role:        role,
		SourceType:  meta.SourceType,
		AudioMode:   meta.AudioMode,
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(e.authSecret))
}

func (e *entity) publishStreamStop(ctx context.Context, meta streammeta.Metadata, reason string) {
	_ = mq.SendGuildUpdate(ctx, e.mqt, meta.GuildID, &mqmsg.GuildMemberStopStream{
		GuildId:   meta.GuildID,
		ChannelId: meta.ChannelID,
		UserId:    meta.OwnerUserID,
		StreamId:  meta.ID,
		Reason:    reason,
	})
	if e.pstore != nil {
		_ = e.pstore.ClearActiveStream(ctx, meta.OwnerUserID)
		_, _ = presence.Refresh(ctx, e.pstore, e.natsConn, meta.OwnerUserID, streamStateTTLSeconds)
	}
}

func (e *entity) clearStreamState(ctx context.Context, meta streammeta.Metadata, reason string) error {
	if e.cache == nil || meta.ID == 0 {
		return nil
	}

	activeRaw, _ := e.cache.HGet(ctx, streammeta.ChannelKey(meta.ChannelID), fmtInt64(meta.ID))
	_ = e.cache.Delete(ctx, streammeta.MetaKey(meta.ID))
	_ = e.cache.Delete(ctx, streammeta.RouteKey(meta.ID))
	_ = e.cache.Delete(ctx, streammeta.UserKey(meta.OwnerUserID))
	_ = e.cache.Delete(ctx, streammeta.RebindKey(meta.ID))
	_ = e.cache.HDel(ctx, streammeta.ChannelKey(meta.ChannelID), fmtInt64(meta.ID))

	if activeRaw != "" {
		e.publishStreamStop(ctx, meta, reason)
	}
	return nil
}

func (e *entity) validateVoiceStreamChannel(ctx context.Context, guildID, channelID, userID int64, requireVideo, requireMembership bool) (*model.Channel, int64, error) {
	ch, _, _, ok, err := e.perm.ChannelPerm(ctx, guildID, channelID, userID, permissions.PermVoiceConnect)
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		return nil, 0, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}
	if ch == nil || ch.Type != model.ChannelTypeGuildVoice {
		return nil, 0, fiber.NewError(fiber.StatusBadRequest, ErrNotAVoiceChannel)
	}

	permsMask, err := e.perm.GetChannelPermissions(ctx, guildID, channelID, userID)
	if err != nil {
		return nil, 0, err
	}
	if requireVideo && !hasPermissionMask(permsMask, permissions.PermVoiceVideo) {
		return nil, 0, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	if requireMembership {
		member, err := e.isVoiceMember(ctx, channelID, userID)
		if err != nil {
			return nil, 0, err
		}
		if !member {
			return nil, 0, fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
		}
	}

	return ch, permsMask, nil
}

// ListStreams
//
//	@Summary		List active voice-channel streams
//	@Description	Returns the currently active streams for a voice channel.
//	@Tags			Guild
//	@Param			guild_id	path	int64	true	"Guild ID"
//	@Param			channel_id	path	int64	true	"Channel ID"
//	@Success		200			{array}	VoiceStreamSummary
//	@Router			/guild/{guild_id}/voice/{channel_id}/streams [get]
func (e *entity) ListStreams(c *fiber.Ctx) error {
	guildID, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	channelID, err := e.parseChannelID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	if _, _, err := e.validateVoiceStreamChannel(c.UserContext(), guildID, channelID, user.Id, false, false); err != nil {
		return err
	}

	streams, err := e.listChannelStreams(c.UserContext(), channelID)
	if err != nil {
		return e.voiceInternalError(c, ErrUnableToGetStream, err)
	}

	summaries := make([]VoiceStreamSummary, len(streams))
	for i, stream := range streams {
		summaries[i] = buildVoiceStreamSummary(stream)
	}
	return c.JSON(summaries)
}

// StartStream
//
//	@Summary		Start a voice-channel stream
//	@Description	Creates or resumes the caller's screen-share stream for the voice channel.
//	@Tags			Guild
//	@Param			guild_id	path		int64						true	"Guild ID"
//	@Param			channel_id	path		int64						true	"Channel ID"
//	@Param			request		body		CreateVoiceStreamRequest	true	"Stream start payload"
//	@Success		200			{object}	CreateVoiceStreamResponse
//	@Router			/guild/{guild_id}/voice/{channel_id}/streams [post]
func (e *entity) StartStream(c *fiber.Ctx) error {
	guildID, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	channelID, err := e.parseChannelID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	if _, _, err := e.validateVoiceStreamChannel(c.UserContext(), guildID, channelID, user.Id, true, true); err != nil {
		return err
	}

	var req CreateVoiceStreamRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToParseBody)
	}
	if err := req.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if existing, ok, _ := e.loadUserStream(c.UserContext(), user.Id); ok && existing.ID != 0 {
		if existing.ChannelID != channelID || existing.GuildID != guildID {
			return fiber.NewError(fiber.StatusConflict, ErrAlreadyStreamingInAnotherChannel)
		}

		binding, err := e.streamBindingForJoin(c.UserContext(), existing.ID, existing.ChannelID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, "stream discovery unavailable")
		}
		if binding.URL == "" {
			return fiber.NewError(fiber.StatusServiceUnavailable, ErrNoStreamServiceAvailableInRegion)
		}

		existing.RouteID = binding.ID
		existing.RouteURL = binding.URL
		existing.Region = binding.Region
		_ = e.cache.SetTimedJSON(c.UserContext(), streammeta.MetaKey(existing.ID), existing, streamStateTTLSeconds)
		_ = e.cache.SetTimedJSON(c.UserContext(), streammeta.UserKey(user.Id), existing, streamStateTTLSeconds)

		token, err := e.issueStreamToken(user.Id, existing, streammeta.RolePublisher, binding)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToIssueStreamToken)
		}
		return c.JSON(CreateVoiceStreamResponse{
			StreamID:    existing.ID,
			StreamURL:   binding.URL,
			StreamToken: token,
			Stream:      buildVoiceStreamSummary(existing),
		})
	}

	streamID := idgen.Next()
	binding, err := e.selectStreamBinding(c.UserContext(), streamID, e.effectiveStreamRegion(c.UserContext(), channelID), false)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "stream discovery unavailable")
	}
	if binding.URL == "" {
		return fiber.NewError(fiber.StatusServiceUnavailable, ErrNoStreamServiceAvailableInRegion)
	}

	meta := streammeta.Metadata{
		ActiveStream: streammeta.ActiveStream{
			ID:         streamID,
			ChannelID:  channelID,
			SourceType: req.SourceType,
			AudioMode:  req.AudioMode,
			StartedAt:  time.Now().Unix(),
		},
		GuildID:     guildID,
		OwnerUserID: user.Id,
		Region:      binding.Region,
		RouteID:     binding.ID,
		RouteURL:    binding.URL,
	}
	if e.cache != nil {
		_ = e.cache.SetTimedJSON(c.UserContext(), streammeta.MetaKey(streamID), meta, streamStateTTLSeconds)
		_ = e.cache.SetTimedJSON(c.UserContext(), streammeta.RouteKey(streamID), binding, streamRouteInitialTTLSeconds)
		_ = e.cache.SetTimedJSON(c.UserContext(), streammeta.UserKey(user.Id), meta, streamStateTTLSeconds)
	}

	token, err := e.issueStreamToken(user.Id, meta, streammeta.RolePublisher, binding)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToIssueStreamToken)
	}

	return c.JSON(CreateVoiceStreamResponse{
		StreamID:    streamID,
		StreamURL:   binding.URL,
		StreamToken: token,
		Stream:      buildVoiceStreamSummary(meta),
	})
}

// JoinStream
//
//	@Summary		Join a voice-channel stream as viewer
//	@Description	Returns receive-only stream signaling info for an active voice-channel stream.
//	@Tags			Guild
//	@Param			guild_id	path		int64	true	"Guild ID"
//	@Param			channel_id	path		int64	true	"Channel ID"
//	@Param			stream_id	path		int64	true	"Stream ID"
//	@Success		200			{object}	JoinVoiceStreamResponse
//	@Router			/guild/{guild_id}/voice/{channel_id}/streams/{stream_id}/join [post]
func (e *entity) JoinStream(c *fiber.Ctx) error {
	guildID, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	channelID, err := e.parseChannelID(c)
	if err != nil {
		return err
	}
	streamID, err := e.parseStreamID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	if _, _, err := e.validateVoiceStreamChannel(c.UserContext(), guildID, channelID, user.Id, false, true); err != nil {
		return err
	}

	meta, ok, err := e.loadStreamMetadata(c.UserContext(), streamID)
	if err != nil {
		return e.voiceInternalError(c, ErrUnableToGetStream, err)
	}
	if !ok || meta.ID == 0 || meta.ChannelID != channelID || meta.GuildID != guildID {
		return fiber.NewError(fiber.StatusNotFound, ErrStreamNotFound)
	}
	if e.cache == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, ErrUnableToGetStream)
	}
	if activeRaw, err := e.cache.HGet(c.UserContext(), streammeta.ChannelKey(channelID), fmtInt64(streamID)); err != nil || activeRaw == "" {
		return fiber.NewError(fiber.StatusNotFound, ErrStreamNotFound)
	}

	binding, err := e.streamBindingForJoin(c.UserContext(), streamID, channelID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "stream discovery unavailable")
	}
	if binding.URL == "" {
		return fiber.NewError(fiber.StatusServiceUnavailable, ErrNoStreamServiceAvailableInRegion)
	}

	token, err := e.issueStreamToken(user.Id, meta, streammeta.RoleViewer, binding)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, ErrUnableToIssueStreamToken)
	}

	return c.JSON(JoinVoiceStreamResponse{
		StreamID:    meta.ID,
		StreamURL:   binding.URL,
		StreamToken: token,
	})
}

// StopStream
//
//	@Summary		Stop a voice-channel stream
//	@Description	Stops the caller's active stream in the voice channel.
//	@Tags			Guild
//	@Param			guild_id	path	int64	true	"Guild ID"
//	@Param			channel_id	path	int64	true	"Channel ID"
//	@Param			stream_id	path	int64	true	"Stream ID"
//	@Success		204
//	@Router			/guild/{guild_id}/voice/{channel_id}/streams/{stream_id} [delete]
func (e *entity) StopStream(c *fiber.Ctx) error {
	guildID, err := e.parseGuildID(c)
	if err != nil {
		return err
	}
	channelID, err := e.parseChannelID(c)
	if err != nil {
		return err
	}
	streamID, err := e.parseStreamID(c)
	if err != nil {
		return err
	}

	user, err := helper.GetUser(c)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, ErrUnableToGetUserToken)
	}

	meta, ok, err := e.loadStreamMetadata(c.UserContext(), streamID)
	if err != nil {
		return e.voiceInternalError(c, ErrUnableToGetStream, err)
	}
	if !ok || meta.ID == 0 || meta.ChannelID != channelID || meta.GuildID != guildID {
		// Stop is intentionally idempotent so reconnect/recovery cleanup can
		// safely retry after the stream has already been removed elsewhere.
		return c.SendStatus(fiber.StatusNoContent)
	}
	if meta.OwnerUserID != user.Id {
		return fiber.NewError(fiber.StatusForbidden, ErrPermissionsRequired)
	}

	if err := e.clearStreamState(c.UserContext(), meta, "owner_stop"); err != nil {
		return e.voiceInternalError(c, ErrUnableToGetStream, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
