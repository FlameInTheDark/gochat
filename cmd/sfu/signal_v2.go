package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/pion/webrtc/v4"
	"go.opentelemetry.io/otel/codes"

	voicev2 "github.com/FlameInTheDark/gochat/cmd/sfu/signaling/v2"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/voice/dave/wire"
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
	mu sync.Mutex

	ctx    context.Context
	log    *slog.Logger
	writer *threadSafeWriter

	sessionID string
	phase     signalV2Phase
	startedAt time.Time

	userID    int64
	channelID int64
	guildID   *int64
	perms     int64
	moved     bool

	pc    *webrtc.PeerConnection
	state *peerConnectionState

	joinNotified bool
	peerAdded    bool

	rtcConnectionID string
	mediaSessionID  string

	supportsDAVE              bool
	supportsEncodedTransforms bool
	maxDAVEProtocolVersion    int
	identityKey               *voicev2.IdentityKey
	daveProtocolVersion       int
	daveEpoch                 uint64
	davePendingProtocol       int
	davePendingEpoch          uint64

	explicitClose bool
	finalized     bool
	resumeTimer   *time.Timer
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
				a.detachSignalV2Session(session)
				finalize = false
				return
			}
			if isExpectedWSReadError(err) {
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

	uid, channelID, guildID, perms, moved, err := a.authorizeJoinFields(identify.ChannelID, identify.Token)
	if err != nil {
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(voicev2.CloseCodeUnauthorized, "unauthorized")
		return nil, err
	}
	if a.sfu.IsBlocked(channelID, uid) && !moved {
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
		guildID:                   guildID,
		perms:                     perms,
		moved:                     moved,
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
	uid, channelID, _, _, _, err := a.authorizeJoinFields(resume.ChannelID, resume.Token)
	if err != nil || uid != session.userID || channelID != session.channelID {
		_ = (&threadSafeWriter{conn: conn.Conn}).SendClose(voicev2.CloseCodeUnauthorized, "unauthorized")
		return nil, err
	}

	a.bindSignalV2Session(session, conn)
	snapshot := a.dave.Snapshot(session.channelID)
	session.mu.Lock()
	session.daveProtocolVersion = snapshot.ProtocolVersion
	session.daveEpoch = snapshot.Epoch
	session.davePendingProtocol = snapshot.ProtocolVersion
	session.davePendingEpoch = snapshot.Epoch
	if session.state != nil {
		session.state.daveProtocol = snapshot.ProtocolVersion
		session.state.daveEpoch = snapshot.Epoch
	}
	session.mu.Unlock()

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
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, err.Error())
			return true
		}
		return false

	case voicev2.OpDAVEInvalidCommitWelcome:
		var invalid voicev2.InvalidCommitWelcome
		if err := json.Unmarshal(packet.D, &invalid); err != nil {
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, "invalid invalid-commit-welcome payload")
			return true
		}
		if err := a.dave.HandleInvalidCommitWelcome(session.sessionID, invalid.TransitionID); err != nil {
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, err.Error())
			return true
		}
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
	packet, err := wire.Decode(raw)
	if err != nil {
		_ = a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, err.Error())
		return true
	}
	switch packet.Opcode {
	case wire.OpcodeKeyPackage:
		if len(packet.Payloads) == 0 {
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeInvalidPayload, "missing key package payload")
			return true
		}
		if err := a.dave.HandleKeyPackage(session.sessionID, packet.Payloads[0]); err != nil {
			_ = a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, err.Error())
			return true
		}
		return false
	case wire.OpcodeCommitWelcome:
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
	session.rtcConnectionID = selectProtocol.RTCConnectionID
	session.mediaSessionID = newSignalSessionID()
	session.state.rtcConnectionID = session.rtcConnectionID
	session.state.mediaSessionID = session.mediaSessionID
	session.state.daveProtocol = session.daveProtocolVersion
	session.state.daveEpoch = session.daveEpoch

	switch session.phase {
	case signalV2PhaseAwaitSelectProtocol:
		if descType != webrtc.SDPTypeOffer {
			return a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, "initial select_protocol must carry an offer")
		}
		offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: selectProtocol.SDP}
		if err := session.pc.SetRemoteDescription(offer); err != nil {
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

		session.daveProtocolVersion = 0
		session.daveEpoch = 0
		session.davePendingProtocol = 0
		session.davePendingEpoch = 0
		session.state.daveProtocol = session.daveProtocolVersion
		session.state.daveEpoch = session.daveEpoch

		audioCodec, videoCodec := detectNegotiatedCodecs(answerToSend.SDP)
		if err := session.writer.SendVoiceGatewayPacket(voicev2.OpSessionDescription, voicev2.SessionDescription{
			Type:                answerToSend.Type.String(),
			SDP:                 answerToSend.SDP,
			RTCConnectionID:     session.rtcConnectionID,
			MediaSessionID:      session.mediaSessionID,
			AudioCodec:          audioCodec,
			VideoCodec:          videoCodec,
			DAVEProtocolVersion: session.daveProtocolVersion,
			DAVEEpoch:           session.daveEpoch,
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
		a.notifyUserJoin(session.ctx, session.userID, session.channelID, session.guildID)
		session.joinNotified = true
		session.phase = signalV2PhaseEstablished

		if err := a.dave.Connect(a.buildDAVEParticipant(session)); err != nil {
			return a.closeSignalV2Session(session, websocket.ClosePolicyViolation, err.Error())
		}
		a.sfu.RequestKeyFrame(session.channelID)

		if currentRevision := a.sfu.ChannelRevision(session.channelID); currentRevision > session.state.appliedRevision {
			a.sfu.SignalPeer(session.ctx, session.channelID, session.pc)
		}
		return nil

	case signalV2PhaseEstablished:
		remote := webrtc.SessionDescription{Type: descType, SDP: selectProtocol.SDP}
		if descType == webrtc.SDPTypeAnswer {
			if err := session.pc.SetRemoteDescription(remote); err != nil {
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
		audioCodec, videoCodec := detectNegotiatedCodecs(answerToSend.SDP)
		return session.writer.SendVoiceGatewayPacket(voicev2.OpSessionDescription, voicev2.SessionDescription{
			Type:                answerToSend.Type.String(),
			SDP:                 answerToSend.SDP,
			RTCConnectionID:     session.rtcConnectionID,
			MediaSessionID:      session.mediaSessionID,
			AudioCodec:          audioCodec,
			VideoCodec:          videoCodec,
			DAVEProtocolVersion: session.state.daveProtocol,
			DAVEEpoch:           session.state.daveEpoch,
		})

	default:
		return a.closeSignalV2Session(session, voicev2.CloseCodeWrongPhase, "unexpected select protocol")
	}
}
