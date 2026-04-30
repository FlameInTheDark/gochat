package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	daveserver "github.com/FlameInTheDark/go-dave/server"
	"github.com/gofiber/contrib/websocket"
	"github.com/pion/webrtc/v4"
	"go.opentelemetry.io/otel/codes"

	voicev2 "github.com/FlameInTheDark/gochat/cmd/stream/signaling/v2"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/observability"
	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
)

const (
	signalProtocolVersion1 = 1
	signalProtocolVersion2 = 2
)

type signalV2Phase int

const (
	signalV2PhaseAwaitHandshake signalV2Phase = iota
	signalV2PhaseAwaitSelectProtocol
	signalV2PhaseEstablished
)

type signalV2Session struct {
	ctx                       context.Context
	sessionID                 string
	guildID                   *int64
	log                       *slog.Logger
	writer                    *threadSafeWriter
	pc                        *webrtc.PeerConnection
	state                     *peerConnectionState
	rtcConnectionID           string
	mediaSessionID            string
	identityKey               *voicev2.IdentityKey
	resumeTimer               *time.Timer
	startedAt                 time.Time
	daveEpoch                 uint64
	davePendingEpoch          uint64
	userID                    int64
	channelID                 int64
	voiceChannelID            int64
	perms                     int64
	role                      string
	sourceType                string
	audioMode                 string
	mu                        sync.Mutex
	maxDAVEProtocolVersion    int
	daveProtocolVersion       int
	davePendingProtocol       int
	phase                     signalV2Phase
	supportsDAVE              bool
	supportsEncodedTransforms bool
	explicitClose             bool
	finalized                 bool
	joinNotified              bool
	peerAdded                 bool
}

func (s *signalV2Session) setSessionDescriptionIDs(rtcConnectionID, mediaSessionID string) {
	s.mu.Lock()
	s.rtcConnectionID = rtcConnectionID
	s.mediaSessionID = mediaSessionID
	state := s.state
	daveProtocol := s.daveProtocolVersion
	daveEpoch := s.daveEpoch
	s.mu.Unlock()

	if state != nil {
		state.setSessionDescriptionMetadata(rtcConnectionID, mediaSessionID, daveProtocol, daveEpoch)
	}
}

func (s *signalV2Session) setDAVEState(protocol int, epoch uint64) {
	s.mu.Lock()
	s.daveProtocolVersion = protocol
	s.daveEpoch = epoch
	state := s.state
	s.mu.Unlock()

	if state != nil {
		state.setDAVEState(protocol, epoch)
	}
}

func (s *signalV2Session) setDAVEStateAndPending(protocol int, epoch uint64) {
	s.mu.Lock()
	s.daveProtocolVersion = protocol
	s.daveEpoch = epoch
	s.davePendingProtocol = protocol
	s.davePendingEpoch = epoch
	state := s.state
	s.mu.Unlock()

	if state != nil {
		state.setDAVEState(protocol, epoch)
	}
}

func (s *signalV2Session) setDAVEPending(protocol int, epoch uint64) {
	s.mu.Lock()
	s.davePendingProtocol = protocol
	s.davePendingEpoch = epoch
	s.mu.Unlock()
}

func (s *signalV2Session) setDAVEPendingProtocol(protocol int) {
	s.mu.Lock()
	s.davePendingProtocol = protocol
	if protocol == 0 {
		s.davePendingEpoch = 0
	}
	s.mu.Unlock()
}

func (s *signalV2Session) applyPendingDAVEState() {
	s.mu.Lock()
	s.daveProtocolVersion = s.davePendingProtocol
	s.daveEpoch = s.davePendingEpoch
	protocol := s.daveProtocolVersion
	epoch := s.daveEpoch
	state := s.state
	s.mu.Unlock()

	if state != nil {
		state.setDAVEState(protocol, epoch)
	}
}

func (s *signalV2Session) sessionDescriptionSnapshot() (rtcConnectionID, mediaSessionID string, daveProtocol int, daveEpoch uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.rtcConnectionID, s.mediaSessionID, s.daveProtocolVersion, s.daveEpoch
}

func parseSignalProtocolVersion(raw string) (int, bool) {
	switch strings.TrimSpace(raw) {
	case "", "1":
		return signalProtocolVersion1, true
	case "2":
		return signalProtocolVersion2, true
	default:
		return 0, false
	}
}

func (a *App) handleSignalWSV2(c *websocket.Conn) {
	requestCtx := context.Background()
	if raw := c.Locals("request_context"); raw != nil {
		if current, ok := raw.(context.Context); ok && current != nil {
			requestCtx = observability.BackgroundFromContext(current)
		}
	}
	signalCtx, signalSpan := observability.Tracer("gochat/sfu").Start(requestCtx, "sfu.signal.v2")
	log := helper.WithContext(a.log, signalCtx)
	defer signalSpan.End()
	defer func() { _ = c.Close() }()

	var (
		session        *signalV2Session
		heartbeatReset chan struct{}
		heartbeatDone  chan struct{}
		finalize       bool
	)
	defer func() {
		if heartbeatDone != nil {
			close(heartbeatDone)
		}
		if finalize && session != nil {
			a.finalizeSignalV2Session(session)
		}
	}()

	for {
		mt, raw, err := c.ReadMessage()
		if err != nil {
			if session != nil && session.phase == signalV2PhaseEstablished && !session.explicitClose && shouldDetachSignalV2Session(err) {
				session.log.Info(
					"detaching stream signal session after websocket read close",
					slog.String("error", err.Error()),
					slog.String("session_id", session.sessionID),
					slog.Int64("stream_id", session.channelID),
					slog.Int64("user", session.userID),
				)
				a.detachSignalV2Session(session)
				finalize = false
				return
			}
			if isExpectedWSReadError(err) {
				if session != nil {
					session.log.Info(
						"stream signal websocket closed",
						slog.String("error", err.Error()),
						slog.String("session_id", session.sessionID),
						slog.Int64("stream_id", session.channelID),
						slog.Int64("user", session.userID),
					)
				} else {
					log.Info("stream signal websocket closed before session established", slog.String("error", err.Error()))
				}
				return
			}
			signalSpan.RecordError(err)
			signalSpan.SetStatus(codes.Error, err.Error())
			log.Warn("failed to read v2 signal message", slog.String("error", err.Error()))
			return
		}

		switch mt {
		case websocket.TextMessage:
			var packet voicev2.IncomingPacket
			if err := json.Unmarshal(raw, &packet); err != nil {
				if session != nil {
					_ = a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, "invalid payload")
					return
				}
				_ = (&threadSafeWriter{conn: c.Conn}).SendClose(voicev2.CloseCodeInvalidPayload, "invalid payload")
				return
			}
			if session == nil {
				created, closeNow := a.handleSignalV2Handshake(c, signalCtx, log, &packet, &session, &heartbeatReset, &heartbeatDone)
				if closeNow {
					finalize = session != nil
					return
				}
				if created {
					finalize = true
				}
				continue
			}
			if a.handleSignalV2TextPacket(session, &packet, heartbeatReset) {
				return
			}

		case websocket.BinaryMessage:
			if session == nil || session.phase != signalV2PhaseEstablished {
				if session != nil {
					_ = a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, "binary messages are not allowed before session is established")
				} else {
					_ = (&threadSafeWriter{conn: c.Conn}).SendClose(voicev2.CloseCodeWrongPhase, "binary messages are not allowed before session is established")
				}
				return
			}
			if a.handleSignalV2BinaryPacket(session, raw) {
				return
			}

		default:
			if session != nil {
				_ = a.closeSignalV2Session(session, voicev2.CloseCodeUnsupportedMedium, "unsupported websocket message type")
			} else {
				_ = (&threadSafeWriter{conn: c.Conn}).SendClose(voicev2.CloseCodeUnsupportedMedium, "unsupported websocket message type")
			}
			return
		}
	}
}

func (a *App) handleSignalV2Handshake(
	conn *websocket.Conn,
	signalCtx context.Context,
	log *slog.Logger,
	packet *voicev2.IncomingPacket,
	session **signalV2Session,
	heartbeatReset *chan struct{},
	heartbeatDone *chan struct{},
) (created bool, closeNow bool) {
	switch packet.Op {
	case voicev2.OpIdentify:
		next, err := a.handleSignalV2Identify(conn, signalCtx, log, packet)
		if err != nil || next == nil {
			return false, true
		}
		*session = next
		created = true
		*heartbeatReset = make(chan struct{}, 1)
		*heartbeatDone = make(chan struct{})
		go a.runSignalV2Heartbeat(next, *heartbeatReset, *heartbeatDone)
		return true, false

	case voicev2.OpResume:
		resumed, err := a.handleSignalV2Resume(conn, packet)
		if err != nil || resumed == nil {
			return false, true
		}
		*session = resumed
		created = true
		*heartbeatReset = make(chan struct{}, 1)
		*heartbeatDone = make(chan struct{})
		go a.runSignalV2Heartbeat(resumed, *heartbeatReset, *heartbeatDone)
		return true, false

	default:
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(voicev2.CloseCodeWrongPhase, "identify or resume expected")
		return false, true
	}
}

func (a *App) handleSignalV2Identify(conn *websocket.Conn, signalCtx context.Context, log *slog.Logger, packet *voicev2.IncomingPacket) (*signalV2Session, error) {
	var identify voicev2.Identify
	if err := json.Unmarshal(packet.D, &identify); err != nil || identify.ChannelID == 0 || identify.Token == "" {
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(voicev2.CloseCodeInvalidPayload, "invalid identify payload")
		return nil, err
	}
	normalizeSignalV2IdentifyDAVE(&identify)

	uid, channelID, voiceChannelID, guildID, perms, role, sourceType, audioMode, err := a.authorizeJoinFields(identify.ChannelID, identify.Token)
	if err != nil {
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(voicev2.CloseCodeUnauthorized, "unauthorized")
		return nil, err
	}
	if a.sfu.IsBlocked(channelID, uid) {
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(voicev2.CloseCodeUnauthorized, "blocked")
		return nil, errors.New("blocked")
	}

	supportsDAVE := a.cfg.DAVEEnabled && identify.SupportsEncodedTransforms && identify.MaxDAVEProtocolVersion > 0
	if !supportsDAVE && a.cfg.DAVERequiredDefault {
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(voicev2.CloseCodeDAVERequired, "dave is required")
		return nil, errors.New("dave required")
	}

	pc, err := a.webrtcAPI.NewPeerConnection(a.iceConfig)
	if err != nil {
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(websocket.ClosePolicyViolation, "unable to create peer connection")
		return nil, err
	}
	if err := a.setupTransceivers(pc); err != nil {
		_ = pc.Close()
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(websocket.ClosePolicyViolation, "unable to configure peer connection")
		return nil, err
	}

	sessionCtx := helper.ContextWithUserID(signalCtx, uid)
	writer := &threadSafeWriter{conn: conn.Conn}
	session := &signalV2Session{
		ctx:                       sessionCtx,
		log:                       helper.WithContext(a.log, sessionCtx),
		writer:                    writer,
		sessionID:                 newSignalSessionID(),
		phase:                     signalV2PhaseAwaitSelectProtocol,
		startedAt:                 time.Now(),
		userID:                    uid,
		channelID:                 channelID,
		voiceChannelID:            voiceChannelID,
		guildID:                   guildID,
		perms:                     perms,
		role:                      role,
		sourceType:                sourceType,
		audioMode:                 audioMode,
		pc:                        pc,
		supportsDAVE:              supportsDAVE,
		supportsEncodedTransforms: identify.SupportsEncodedTransforms,
		maxDAVEProtocolVersion:    identify.MaxDAVEProtocolVersion,
		identityKey:               identify.IdentityKey,
	}
	session.state = &peerConnectionState{
		peerConnection: pc,
		websocket:      writer,
		userID:         uid,
		perms:          perms,
		signalVersion:  signalProtocolVersion2,
	}
	a.registerPeerCallbacks(session.ctx, pc, writer, session.state, uid, channelID, perms)
	a.registerSignalV2Session(session)

	if err := writer.SendVoiceGatewayPacket(voicev2.OpHello, voicev2.Hello{
		V:                 voicev2.ProtocolVersion,
		HeartbeatInterval: a.cfg.SignalHeartbeatIntervalMS,
		SessionID:         session.sessionID,
	}); err != nil {
		return nil, err
	}
	if err := writer.SendVoiceGatewayPacket(voicev2.OpReady, a.signalV2ReadyPayload(perms)); err != nil {
		return nil, err
	}

	return session, nil
}

func normalizeSignalV2IdentifyDAVE(identify *voicev2.Identify) {
	if identify == nil {
		return
	}

	if identify.MaxDAVEProtocolVersion > voicev2.MaxDAVEProtocol {
		identify.MaxDAVEProtocolVersion = voicev2.MaxDAVEProtocol
	}
	if identify.MaxDAVEProtocolVersion > 0 {
		identify.SupportsEncodedTransforms = true
	}
	if identify.DAVESupported && !identify.SupportsEncodedTransforms {
		identify.SupportsEncodedTransforms = true
	}
	if identify.MaxDAVEProtocolVersion == 0 && identify.SupportsEncodedTransforms {
		identify.MaxDAVEProtocolVersion = voicev2.MaxDAVEProtocol
	}
}

func (a *App) handleSignalV2Resume(conn *websocket.Conn, packet *voicev2.IncomingPacket) (*signalV2Session, error) {
	var resume voicev2.Resume
	if err := json.Unmarshal(packet.D, &resume); err != nil || resume.SessionID == "" || resume.ChannelID == 0 || resume.Token == "" {
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(voicev2.CloseCodeInvalidPayload, "invalid resume payload")
		return nil, err
	}

	session := a.getSignalV2Session(resume.SessionID)
	if session == nil {
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(voicev2.CloseCodeSessionExpired, "session expired")
		return nil, errors.New("session expired")
	}
	uid, channelID, _, _, _, _, _, _, err := a.authorizeJoinFields(resume.ChannelID, resume.Token)
	if err != nil || uid != session.userID || channelID != session.channelID {
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(voicev2.CloseCodeUnauthorized, "unauthorized")
		return nil, err
	}

	a.bindSignalV2Session(session, conn)
	snapshot := a.dave.Snapshot(session.channelID)
	session.setDAVEStateAndPending(snapshot.ProtocolVersion, snapshot.Epoch)

	if err := session.writer.SendVoiceGatewayPacket(voicev2.OpHello, voicev2.Hello{
		V:                 voicev2.ProtocolVersion,
		HeartbeatInterval: a.cfg.SignalHeartbeatIntervalMS,
		SessionID:         session.sessionID,
	}); err != nil {
		return nil, err
	}
	if err := session.writer.SendVoiceGatewayPacket(voicev2.OpResumed, voicev2.Resumed{
		SessionID:           session.sessionID,
		DAVEProtocolVersion: snapshot.ProtocolVersion,
		DAVEEpoch:           snapshot.Epoch,
	}); err != nil {
		return nil, err
	}

	return session, nil
}

func (a *App) runSignalV2Heartbeat(session *signalV2Session, reset <-chan struct{}, done <-chan struct{}) {
	timeout := time.Duration(a.cfg.SignalHeartbeatIntervalMS)*time.Millisecond + a.signalHeartbeatGrace
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-done:
			return
		case <-reset:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(timeout)
		case <-timer.C:
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeHeartbeatTimeout, "heartbeat timeout")
			return
		}
	}
}

func (a *App) handleSignalV2TextPacket(session *signalV2Session, packet *voicev2.IncomingPacket, heartbeatReset chan<- struct{}) bool {
	switch packet.Op {
	case voicev2.OpHeartbeat:
		select {
		case heartbeatReset <- struct{}{}:
		default:
		}
		var heartbeat voicev2.Heartbeat
		_ = json.Unmarshal(packet.D, &heartbeat)
		return session.writer.SendVoiceGatewayPacket(voicev2.OpHeartbeatACK, voicev2.HeartbeatACK{T: heartbeat.T}) != nil

	case voicev2.OpSelectProtocol:
		return a.handleSignalV2SelectProtocol(session, json.RawMessage(packet.D)) != nil

	case voicev2.OpSpeaking:
		if session.phase != signalV2PhaseEstablished {
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, "speaking is not allowed before session is established")
			return true
		}
		var speaking voicev2.Speaking
		_ = json.Unmarshal(packet.D, &speaking)
		value := 0
		if speaking.Speaking != 0 {
			value = 1
		}
		a.sfu.BroadcastSpeaking(session.ctx, session.channelID, session.userID, value)
		return false

	case voicev2.OpDAVETransitionReady:
		var ready voicev2.TransitionReady
		if err := json.Unmarshal(packet.D, &ready); err != nil {
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, "invalid transition ready payload")
			return true
		}
		if err := a.dave.HandleTransitionReady(session.sessionID, ready.TransitionID); err != nil {
			if strings.EqualFold(err.Error(), "unexpected transition ready") {
				session.log.Warn(
					"ignoring unexpected dave transition ready",
					slog.Uint64("transition_id", uint64(ready.TransitionID)),
					slog.String("session_id", session.sessionID),
					slog.Int64("stream_id", session.channelID),
					slog.Int64("user", session.userID),
				)
				return false
			}
			session.log.Warn(
				"dave transition ready rejected",
				slog.Uint64("transition_id", uint64(ready.TransitionID)),
				slog.String("reason", err.Error()),
				slog.String("session_id", session.sessionID),
				slog.Int64("stream_id", session.channelID),
				slog.Int64("user", session.userID),
			)
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, err.Error())
			return true
		}
		session.log.Info(
			"dave transition ready accepted",
			slog.Uint64("transition_id", uint64(ready.TransitionID)),
			slog.String("session_id", session.sessionID),
			slog.Int64("stream_id", session.channelID),
			slog.Int64("user", session.userID),
		)
		return false

	case voicev2.OpDAVEInvalidCommitWelcome:
		var invalid voicev2.InvalidCommitWelcome
		if err := json.Unmarshal(packet.D, &invalid); err != nil {
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, "invalid invalid-commit-welcome payload")
			return true
		}
		if err := a.dave.HandleInvalidCommitWelcome(session.sessionID, invalid.TransitionID); err != nil {
			session.log.Warn(
				"dave invalid commit welcome rejected",
				slog.Uint64("transition_id", uint64(invalid.TransitionID)),
				slog.String("reason", err.Error()),
				slog.String("session_id", session.sessionID),
				slog.Int64("stream_id", session.channelID),
				slog.Int64("user", session.userID),
			)
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, err.Error())
			return true
		}
		session.log.Warn(
			"dave invalid commit welcome accepted",
			slog.Uint64("transition_id", uint64(invalid.TransitionID)),
			slog.String("session_id", session.sessionID),
			slog.Int64("stream_id", session.channelID),
			slog.Int64("user", session.userID),
		)
		return false

	case voicev2.OpResume, voicev2.OpIdentify:
		_ = a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, "handshake already completed")
		return true

	default:
		_ = a.closeSignalV2Session(session, voicev2.CloseCodeUnknownOpcode, "unsupported opcode")
		return true
	}
}

func (a *App) handleSignalV2BinaryPacket(session *signalV2Session, raw []byte) bool {
	packet, err := daveserver.DecodeBinaryMessage(raw)
	if err != nil {
		_ = a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, err.Error())
		return true
	}
	switch packet.Opcode {
	case daveserver.OpcodeKeyPackage:
		if len(packet.Payloads) == 0 {
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, "missing key package payload")
			return true
		}
		if err := a.dave.HandleKeyPackage(session.sessionID, packet.Payloads[0]); err != nil {
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, err.Error())
			return true
		}
		return false
	case daveserver.OpcodeCommitWelcome:
		if err := a.dave.HandleCommitWelcome(session.sessionID, packet.Commit, packet.Welcome); err != nil {
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, err.Error())
			return true
		}
		return false
	default:
		_ = a.closeSignalV2Session(session, voicev2.CloseCodeUnknownOpcode, "unsupported binary opcode")
		return true
	}
}

func (a *App) handleSignalV2SelectProtocol(session *signalV2Session, raw json.RawMessage) error {
	var selectProtocol voicev2.SelectProtocol
	if err := json.Unmarshal(raw, &selectProtocol); err != nil || selectProtocol.Protocol == "" || selectProtocol.SDP == "" || selectProtocol.RTCConnectionID == "" {
		return a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, "invalid select protocol payload")
	}
	if !strings.EqualFold(selectProtocol.Protocol, "webrtc") {
		return a.closeSignalV2Session(session, voicev2.CloseCodeUnsupportedMedium, "only webrtc protocol is supported")
	}

	descType := parseVoiceGatewaySDPType(selectProtocol.Type, webrtc.SDPTypeOffer)
	session.setSessionDescriptionIDs(selectProtocol.RTCConnectionID, newSignalSessionID())

	switch session.phase {
	case signalV2PhaseAwaitSelectProtocol:
		if descType != webrtc.SDPTypeOffer {
			return a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, "initial select_protocol must carry an offer")
		}
		offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: selectProtocol.SDP}
		if err := session.pc.SetRemoteDescription(offer); err != nil {
			session.log.Warn("failed to apply v2 offer",
				slog.String("error", err.Error()),
				slog.String("signaling_state", session.pc.SignalingState().String()),
				slog.Int("phase", int(session.phase)),
				slog.String("rtc_connection_id", selectProtocol.RTCConnectionID),
			)
			return a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, "unable to apply offer")
		}
		a.telemetry.Offer(session.ctx, "inbound", signalPeerAttrs(session.channelID, session.userID, session.guildID)...)

		revision := uint64(0)
		if ch := a.sfu.GetChannel(session.channelID); ch != nil {
			revision = ch.preparePeerInitialSync(session.state)
		}

		answer, err := session.pc.CreateAnswer(nil)
		if err != nil {
			return a.closeSignalV2Session(session, websocket.ClosePolicyViolation, "unable to create answer")
		}
		if err := session.pc.SetLocalDescription(answer); err != nil {
			return a.closeSignalV2Session(session, websocket.ClosePolicyViolation, "unable to set local description")
		}
		answerToSend := waitForGatheredLocalDescription(session.pc, answer)
		if a.sfu.maxAudioBitrateBps > 0 {
			answerToSend.SDP = limitAudioBitrateInSDP(answerToSend.SDP, a.sfu.maxAudioBitrateBps)
		}
		if a.sfu.maxVideoBitrateBps > 0 {
			answerToSend.SDP = limitVideoBitrateInSDP(answerToSend.SDP, a.sfu.maxVideoBitrateBps)
		}

		session.setDAVEStateAndPending(0, 0)
		rtcConnectionID, mediaSessionID, daveProtocol, daveEpoch := session.sessionDescriptionSnapshot()

		audioCodec, videoCodec := detectNegotiatedCodecs(answerToSend.SDP)
		if err := session.writer.SendVoiceGatewayPacket(voicev2.OpSessionDescription, voicev2.SessionDescription{
			Type:                answerToSend.Type.String(),
			SDP:                 answerToSend.SDP,
			RTCConnectionID:     rtcConnectionID,
			MediaSessionID:      mediaSessionID,
			AudioCodec:          audioCodec,
			VideoCodec:          videoCodec,
			DAVEProtocolVersion: daveProtocol,
			DAVEEpoch:           daveEpoch,
		}); err != nil {
			return err
		}

		session.state.negotiated = true
		session.state.offeredRevision = revision
		session.state.appliedRevision = revision
		a.sfu.AddPeer(session.ctx, session.channelID, session.state)
		session.peerAdded = true
		a.totalPeers.Add(1)
		a.telemetry.Join(session.ctx, signalPeerAttrs(session.channelID, session.userID, session.guildID)...)
		if session.role == streammeta.RolePublisher {
			a.notifyUserJoin(session.ctx, session.userID, session.channelID, session.voiceChannelID, session.guildID, session.sourceType, session.audioMode, session.sessionID)
			session.joinNotified = true
		}
		session.phase = signalV2PhaseEstablished

		if err := a.dave.Connect(a.buildDAVEParticipant(session)); err != nil {
			return a.closeSignalV2Session(session, websocket.ClosePolicyViolation, err.Error())
		}
		a.sfu.RequestKeyFrameBurst(session.channelID)

		if currentRevision := a.sfu.ChannelRevision(session.channelID); currentRevision > session.state.appliedRevision {
			a.sfu.SignalPeer(session.ctx, session.channelID, session.pc)
		}
		return nil

	case signalV2PhaseEstablished:
		remote := webrtc.SessionDescription{Type: descType, SDP: selectProtocol.SDP}
		if descType == webrtc.SDPTypeAnswer {
			if err := session.pc.SetRemoteDescription(remote); err != nil {
				session.log.Warn("failed to apply v2 answer",
					slog.String("error", err.Error()),
					slog.String("signaling_state", session.pc.SignalingState().String()),
					slog.Int("phase", int(session.phase)),
					slog.String("rtc_connection_id", selectProtocol.RTCConnectionID),
				)
				return a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, "unable to apply answer")
			}
			a.sfu.ApplyAnswer(session.ctx, session.channelID, session.pc)
			a.telemetry.Answer(session.ctx, "inbound", signalPeerAttrs(session.channelID, session.userID, session.guildID)...)
			return nil
		}
		if descType != webrtc.SDPTypeOffer {
			return a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, "unsupported description type")
		}
		if err := session.pc.SetRemoteDescription(remote); err != nil {
			session.log.Warn("failed to apply v2 offer",
				slog.String("error", err.Error()),
				slog.String("signaling_state", session.pc.SignalingState().String()),
				slog.Int("phase", int(session.phase)),
				slog.String("rtc_connection_id", selectProtocol.RTCConnectionID),
			)
			return a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, "unable to apply offer")
		}
		answer, err := session.pc.CreateAnswer(nil)
		if err != nil {
			return a.closeSignalV2Session(session, websocket.ClosePolicyViolation, "unable to create answer")
		}
		if err := session.pc.SetLocalDescription(answer); err != nil {
			return a.closeSignalV2Session(session, websocket.ClosePolicyViolation, "unable to set local description")
		}
		answerToSend := waitForGatheredLocalDescription(session.pc, answer)
		if a.sfu.maxAudioBitrateBps > 0 {
			answerToSend.SDP = limitAudioBitrateInSDP(answerToSend.SDP, a.sfu.maxAudioBitrateBps)
		}
		if a.sfu.maxVideoBitrateBps > 0 {
			answerToSend.SDP = limitVideoBitrateInSDP(answerToSend.SDP, a.sfu.maxVideoBitrateBps)
		}
		rtcConnectionID, mediaSessionID, daveProtocol, daveEpoch := session.sessionDescriptionSnapshot()
		audioCodec, videoCodec := detectNegotiatedCodecs(answerToSend.SDP)
		return session.writer.SendVoiceGatewayPacket(voicev2.OpSessionDescription, voicev2.SessionDescription{
			Type:                answerToSend.Type.String(),
			SDP:                 answerToSend.SDP,
			RTCConnectionID:     rtcConnectionID,
			MediaSessionID:      mediaSessionID,
			AudioCodec:          audioCodec,
			VideoCodec:          videoCodec,
			DAVEProtocolVersion: daveProtocol,
			DAVEEpoch:           daveEpoch,
		})

	default:
		return a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, "unexpected select protocol")
	}
}
