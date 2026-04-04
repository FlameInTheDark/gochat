package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	daveserver "github.com/FlameInTheDark/go-dave/server"
	"github.com/gofiber/contrib/websocket"
	"github.com/pion/webrtc/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

func (a *App) handleSignalWSV1(c *websocket.Conn) {
	requestCtx := context.Background()
	if raw := c.Locals("request_context"); raw != nil {
		if current, ok := raw.(context.Context); ok && current != nil {
			requestCtx = observability.BackgroundFromContext(current)
		}
	}
	signalCtx, signalSpan := observability.Tracer("gochat/sfu").Start(requestCtx, "sfu.signal")
	log := helper.WithContext(a.log, signalCtx)
	defer signalSpan.End()
	defer func() { _ = c.Close() }()

	// Phase 1: Handshake - read join envelope and authorize.
	joinEnv, err := a.readJoinEnvelope(c)
	if err != nil {
		signalSpan.RecordError(err)
		signalSpan.SetStatus(codes.Error, err.Error())
		log.Warn("invalid join envelope", slog.String("error", err.Error()))
		_ = (&threadSafeWriter{conn: c.Conn}).SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: ErrorResponse{Error: "invalid message"}})
		return
	}
	uid, channelID, guildID, perms, _, err := a.authorizeJoin(joinEnv)
	if err != nil {
		signalSpan.RecordError(err)
		signalSpan.SetStatus(codes.Error, err.Error())
		log.Warn("join unauthorized", slog.String("error", err.Error()))
		_ = (&threadSafeWriter{conn: c.Conn}).SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: ErrorResponse{Error: err.Error()}})
		return
	}
	sessionCtx := helper.ContextWithUserID(signalCtx, uid)
	peerAttrs := []attribute.KeyValue{
		attribute.Int64("voice.channel_id", channelID),
		attribute.Int64("user.id", uid),
	}
	if guildID != nil {
		peerAttrs = append(peerAttrs, attribute.Int64("guild.id", *guildID))
	}
	a.telemetry.Join(sessionCtx, peerAttrs...)
	sessionStarted := time.Now()
	log = helper.WithContext(a.log, sessionCtx)

	if a.sfu.IsBlocked(channelID, uid) {
		log.Warn("blocked user tried to join", slog.Int64("user", uid), slog.Int64("channel", channelID))
		_ = (&threadSafeWriter{conn: c.Conn}).SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: ErrorResponse{Error: "blocked"}})
		return
	}
	if a.cfg.DAVERequiredDefault {
		log.Warn("legacy signal protocol rejected because dave is required", slog.Int64("user", uid), slog.Int64("channel", channelID))
		_ = (&threadSafeWriter{conn: c.Conn}).SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: ErrorResponse{Error: "dave is required; use signal v2"}})
		return
	}

	// Phase 2: Setup - create PeerConnection and register it.
	a.notifyUserJoin(sessionCtx, uid, channelID, guildID)
	defer a.notifyUserLeave(observability.BackgroundFromContext(sessionCtx), uid, channelID, guildID)

	pc, err := a.webrtcAPI.NewPeerConnection(a.iceConfig)
	if err != nil {
		signalSpan.RecordError(err)
		signalSpan.SetStatus(codes.Error, err.Error())
		log.Error("failed to create peer connection", slog.String("error", err.Error()))
		return
	}
	defer func() { _ = pc.Close() }()

	writer := &threadSafeWriter{conn: c.Conn}
	legacySessionID := newSignalSessionID()
	state := &peerConnectionState{peerConnection: pc, websocket: writer, userID: uid, perms: perms, signalVersion: signalProtocolVersion1}

	if err := a.setupTransceivers(pc); err != nil {
		log.Error("failed to setup transceivers", slog.String("error", err.Error()))
		return
	}

	a.registerPeerCallbacks(sessionCtx, pc, writer, state, uid, channelID, perms)

	if err := writer.SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: JoinAck{Ok: true}}); err != nil {
		log.Warn("failed to send join ack", slog.String("error", err.Error()))
		return
	}

	log.Info("client joined", slog.Int64("user", uid), slog.Int64("channel", channelID))

	a.sfu.AddPeer(sessionCtx, channelID, state)
	a.totalPeers.Add(1)
	if err := a.dave.Connect(daveserver.Participant{
		SessionID:     legacySessionID,
		UserID:        uid,
		ChannelID:     channelID,
		SignalVersion: signalProtocolVersion1,
	}); err != nil {
		log.Warn("failed to update dave coordinator for legacy peer", slog.String("error", err.Error()))
	}
	defer func() {
		writer.Close()
		_ = a.dave.Disconnect(legacySessionID)
		a.sfu.RemovePeer(sessionCtx, channelID, pc)
		a.totalPeers.Add(-1)
		a.telemetry.Leave(sessionCtx, sessionStarted, peerAttrs...)
		log.Info("client left", slog.Int64("user", uid), slog.Int64("channel", channelID))
	}()

	a.sfu.SignalPeer(sessionCtx, channelID, pc)

	// Phase 3: Message loop
	a.messageLoop(sessionCtx, c, pc, writer, uid, perms, channelID)
}

func (a *App) messageLoop(
	ctx context.Context,
	c *websocket.Conn,
	pc *webrtc.PeerConnection,
	writer *threadSafeWriter,
	uid, perms, channelID int64,
) {
	for {
		_, raw, err := c.ReadMessage()
		if err != nil {
			a.log.Warn("failed to read message", slog.String("error", err.Error()))
			return
		}

		// Try simple event-based format first.
		var msg websocketMessage
		if err := json.Unmarshal(raw, &msg); err == nil && msg.Event != "" {
			if a.handleSimpleMessage(ctx, msg, pc, writer, uid, channelID) {
				return
			}
			continue
		}

		// Fall back to legacy envelope format.
		var env envelope
		if err := json.Unmarshal(raw, &env); err == nil && env.OP != 0 {
			if a.handleLegacyEnvelope(ctx, env, pc, writer, uid, perms, channelID) {
				return
			}
			continue
		}

		a.log.Warn("unrecognized message format", slog.Int64("user", uid), slog.Int64("channel", channelID))
	}
}

// handleSimpleMessage processes simple event-based WebSocket messages.
// Returns true if the connection should be closed.
func (a *App) handleSimpleMessage(
	ctx context.Context,
	msg websocketMessage,
	pc *webrtc.PeerConnection,
	writer *threadSafeWriter,
	uid, channelID int64,
) bool {
	attrs := []attribute.KeyValue{
		attribute.Int64("voice.channel_id", channelID),
		attribute.Int64("user.id", uid),
	}
	switch msg.Event {
	case "candidate":
		var cand webrtc.ICECandidateInit
		if err := json.Unmarshal([]byte(msg.Data), &cand); err != nil {
			a.log.Warn("failed to parse candidate", slog.String("error", err.Error()))
			return true
		}
		if err := pc.AddICECandidate(cand); err != nil {
			a.log.Warn("failed to add candidate", slog.String("error", err.Error()))
			return true
		}
		a.telemetry.Candidate(ctx, "inbound", attrs...)

	case "answer":
		var answer webrtc.SessionDescription
		if err := json.Unmarshal([]byte(msg.Data), &answer); err != nil {
			a.log.Warn("failed to parse answer", slog.String("error", err.Error()))
			return true
		}
		if answer.Type == webrtc.SDPType(0) {
			answer.Type = webrtc.SDPTypeAnswer
		}
		if err := pc.SetRemoteDescription(answer); err != nil {
			a.log.Warn("failed to set remote description", slog.String("error", err.Error()))
			return true
		}
		a.sfu.ApplyAnswer(ctx, channelID, pc)
		a.telemetry.Answer(ctx, "inbound", attrs...)

	case "negotiate":
		a.log.Info("client requested renegotiation", slog.Int64("channel", channelID), slog.Int64("user", uid))
		a.telemetry.Renegotiation(ctx, attrs...)
		a.sfu.SignalPeer(ctx, channelID, pc)

	case "speaking":
		speaking := parseSpeakingData(msg.Data)
		a.log.Debug("speaking event", slog.Int64("user", uid), slog.Int64("channel", channelID), slog.Int("speaking", speaking))
		a.sfu.BroadcastSpeaking(ctx, channelID, uid, speaking)

	default:
		a.log.Warn("unknown message", slog.String("event", msg.Event))
	}
	return false
}

func (a *App) handleLegacyEnvelope(ctx context.Context, env envelope, pc *webrtc.PeerConnection, writer *threadSafeWriter, uid int64, perms int64, channelID int64) bool {
	switch env.OP {
	case int(mqmsg.OPCodeHeartBeat):
		var hb heartbeatData
		_ = json.Unmarshal(env.D, &hb)
		_ = writer.SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeHeartBeat), D: HeartbeatReply{Pong: true, ServerTS: time.Now().UnixMilli(), Nonce: hb.Nonce, TS: hb.TS}})
		return false

	case int(mqmsg.OPCodeRTC):
		return a.handleLegacyRTCEvent(ctx, env, pc, writer, uid, perms, channelID)
	}
	return false
}

func (a *App) handleLegacyRTCEvent(ctx context.Context, env envelope, pc *webrtc.PeerConnection, writer *threadSafeWriter, uid int64, perms int64, channelID int64) bool {
	attrs := []attribute.KeyValue{
		attribute.Int64("voice.channel_id", channelID),
		attribute.Int64("user.id", uid),
	}
	switch env.T {
	case int(mqmsg.EventTypeRTCAnswer):
		var ans rtcAnswer
		if err := json.Unmarshal(env.D, &ans); err != nil || ans.SDP == "" {
			return false
		}
		descType := parseLegacySDPType(ans.Type)
		desc := webrtc.SessionDescription{Type: descType, SDP: ans.SDP}
		if err := pc.SetRemoteDescription(desc); err != nil {
			a.log.Warn("failed to apply legacy answer", slog.String("error", err.Error()))
		} else {
			a.sfu.ApplyAnswer(ctx, channelID, pc)
		}
		a.telemetry.Answer(ctx, "inbound", attrs...)

	case int(mqmsg.EventTypeRTCCandidate):
		var cand rtcCandidate
		if err := json.Unmarshal(env.D, &cand); err != nil || cand.Candidate == "" {
			return false
		}
		if err := pc.AddICECandidate(webrtc.ICECandidateInit{Candidate: cand.Candidate, SDPMid: cand.SDPMid, SDPMLineIndex: cand.SDPMLineIndex}); err != nil {
			a.log.Warn("failed to add legacy candidate", slog.String("error", err.Error()))
		}
		a.telemetry.Candidate(ctx, "inbound", attrs...)

	case int(mqmsg.EventTypeRTCLeave):
		return true

	case int(mqmsg.EventTypeRTCServerMuteUser):
		if !hasPerm(perms, permissions.PermVoiceMuteMembers) {
			a.log.Warn("mute denied: insufficient permissions", slog.Int64("user", uid))
			return false
		}
		var data muteUserData
		if err := json.Unmarshal(env.D, &data); err != nil {
			return false
		}
		a.sfu.ServerMuteUser(ctx, channelID, data.User, data.Muted)

	case int(mqmsg.EventTypeRTCServerDeafenUser):
		if !hasPerm(perms, permissions.PermVoiceDeafenMembers) {
			a.log.Warn("deafen denied: insufficient permissions", slog.Int64("user", uid))
			return false
		}
		var data deafenUserData
		if err := json.Unmarshal(env.D, &data); err != nil {
			return false
		}
		a.sfu.ServerDeafenUser(ctx, channelID, data.User, data.Deafened)

	case int(mqmsg.EventTypeRTCServerKickUser):
		if !hasPerm(perms, permissions.PermVoiceMoveMembers) {
			a.log.Warn("kick denied: insufficient permissions", slog.Int64("user", uid))
			return false
		}
		var data kickUserData
		if err := json.Unmarshal(env.D, &data); err != nil {
			return false
		}
		a.sfu.KickUser(ctx, channelID, data.User)

	case int(mqmsg.EventTypeRTCServerBlockUser):
		if !hasPerm(perms, permissions.PermVoiceMoveMembers) {
			a.log.Warn("block denied: insufficient permissions", slog.Int64("user", uid))
			return false
		}
		var data blockEvent
		if err := json.Unmarshal(env.D, &data); err != nil {
			return false
		}
		a.sfu.BlockUser(ctx, channelID, data.UserId, data.Block)
	}
	return false
}

func (a *App) readJoinEnvelope(c *websocket.Conn) (rtcJoinEnvelope, error) {
	var env rtcJoinEnvelope
	if c.Conn != nil {
		_ = c.SetReadDeadline(time.Now().Add(joinHandshakeTimeout))
	}
	if err := c.ReadJSON(&env); err != nil {
		return rtcJoinEnvelope{}, err
	}
	if c.Conn != nil {
		_ = c.SetReadDeadline(time.Time{})
	}
	return env, nil
}

func (a *App) authorizeJoin(env rtcJoinEnvelope) (int64, int64, *int64, int64, bool, error) {
	if env.OP != int(mqmsg.OPCodeRTC) || env.T != int(mqmsg.EventTypeRTCJoin) || env.D.Token == "" || env.D.Channel == 0 {
		return 0, 0, nil, 0, false, fmt.Errorf("expected join")
	}
	uid, tokChannel, tokGuild, perms, moved, err := a.validateJoinToken(env.D.Token)
	if err != nil {
		return 0, 0, nil, 0, false, fmt.Errorf("unauthorized")
	}
	if tokChannel != 0 && tokChannel != int64(env.D.Channel) {
		return 0, 0, nil, 0, false, fmt.Errorf("unauthorized")
	}
	return uid, int64(env.D.Channel), tokGuild, perms, moved, nil
}
