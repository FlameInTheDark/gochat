package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	recm "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	slogfiber "github.com/samber/slog-fiber"

	"github.com/FlameInTheDark/gochat/cmd/sfu/config"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

// App is the top-level SFU application.
type App struct {
	app  *fiber.App
	cfg  *config.Config
	log  *slog.Logger
	shut *shutter.Shut
	sfu  *SFU

	iceConfig webrtc.Configuration
	webrtcAPI *webrtc.API // Custom API with restricted MediaEngine

	instID      string
	totalPeers  atomic.Int64
	discoverLog sync.Once
}

// websocketMessage is the simple event-based message format used over WebSocket.
type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

// NewApp creates a fully configured SFU application.
func NewApp(shut *shutter.Shut, logger *slog.Logger) *App {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("unable to load config", slog.String("error", err.Error()))
		panic(err)
	}

	iceCfg := buildICEConfig(cfg.STUNServers)
	api := buildWebRTCAPI(logger)

	fiberApp := fiber.New(fiber.Config{DisableStartupMessage: true})
	lm := slogfiber.NewWithFilters(logger, slogfiber.IgnorePath("/metrics"))
	fiberApp.Use(lm)
	fiberApp.Use(recm.New())

	// Prometheus metrics
	h := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})
	fiberApp.Get("/metrics", adaptor.HTTPHandler(h))

	// Compute max audio bitrate in bps (0 means disabled)
	var maxAudioBps uint64
	if cfg.MaxAudioBitrateKbps > 0 {
		maxAudioBps = uint64(cfg.MaxAudioBitrateKbps) * 1000
	}
	// Clamp margin to [0,100]
	marginPct := cfg.AudioBitrateMarginPercent
	if marginPct < 0 {
		marginPct = 0
	} else if marginPct > 100 {
		marginPct = 100
	}
	sfu := NewSFU(cfg.WebhookURL, cfg.WebhookToken, logger, maxAudioBps, cfg.EnforceAudioBitrate, marginPct)

	a := &App{
		app:       fiberApp,
		cfg:       cfg,
		log:       logger,
		shut:      shut,
		sfu:       sfu,
		instID:    cfg.ServiceID,
		iceConfig: iceCfg,
		webrtcAPI: api,
	}

	fiberApp.Get("/signal", websocket.New(a.handleSignalWS, websocket.Config{}))
	fiberApp.Post("/admin/channel/close", a.handleAdminCloseChannel)
	go sfu.RunKeyFrameTicker()

	return a
}

// Start begins listening and blocks until SIGINT/SIGTERM.
func (a *App) Start() {
	a.log.Info("SFU starting", slog.String("addr", a.cfg.ServerAddress))
	go func() {
		if err := a.app.Listen(a.cfg.ServerAddress); err != nil {
			a.log.Error("failed to start sfu", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	go a.discoveryHeartbeat()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

// Close gracefully stops the application and underlying SFU.
func (a *App) Close() error {
	a.sfu.Close()
	return a.app.Shutdown()
}

// ---------------------------------------------------------------------------
// WebRTC setup helpers
// ---------------------------------------------------------------------------

// buildICEConfig creates a webrtc.Configuration from a list of STUN server URLs.
func buildICEConfig(stunServers []string) webrtc.Configuration {
	iceCfg := webrtc.Configuration{}
	for _, raw := range stunServers {
		url := strings.TrimSpace(raw)
		if url == "" {
			continue
		}
		iceCfg.ICEServers = append(iceCfg.ICEServers, webrtc.ICEServer{URLs: []string{url}})
	}
	if len(iceCfg.ICEServers) == 0 {
		iceCfg.ICEServers = []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}}
	}
	return iceCfg
}

// buildWebRTCAPI creates a webrtc.API with a restricted MediaEngine (Opus + VP8 + VP9 only)
// and TWCC header extensions registered for bandwidth estimation.
// Uses minimal interceptors to avoid crashes in the RTCP receiver report interceptor.
func buildWebRTCAPI(logger *slog.Logger) *webrtc.API {
	me := &webrtc.MediaEngine{}

	// Audio: Opus only (48kHz, 2ch)
	if err := me.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeOpus,
			ClockRate:   48000,
			Channels:    2,
			SDPFmtpLine: "minptime=10;useinbandfec=1",
		},
		PayloadType: 111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		logger.Error("failed to register Opus codec", slog.String("error", err.Error()))
	}

	// Video: VP8 (widely supported, low complexity)
	if err := me.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:  webrtc.MimeTypeVP8,
			ClockRate: 90000,
		},
		PayloadType: 96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		logger.Error("failed to register VP8 codec", slog.String("error", err.Error()))
	}

	// Video: VP9 (better quality at same bitrate, optional)
	if err := me.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:  webrtc.MimeTypeVP9,
			ClockRate: 90000,
		},
		PayloadType: 98,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		logger.Error("failed to register VP9 codec", slog.String("error", err.Error()))
	}

	// Register TWCC header extension for bandwidth estimation
	twccURI := "http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01"
	if err := me.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: twccURI}, webrtc.RTPCodecTypeVideo); err != nil {
		logger.Warn("failed to register TWCC extension for video", slog.String("error", err.Error()))
	}
	if err := me.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: twccURI}, webrtc.RTPCodecTypeAudio); err != nil {
		logger.Warn("failed to register TWCC extension for audio", slog.String("error", err.Error()))
	}

	return webrtc.NewAPI(webrtc.WithMediaEngine(me))
}

// ---------------------------------------------------------------------------
// Webhook notification helpers
// ---------------------------------------------------------------------------

// notifyUserJoin sends an async webhook notification for a user joining voice.
func (a *App) notifyUserJoin(uid, channelID int64, guildID *int64) {
	go func() {
		resp, err := a.sfu.httpClient.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("X-Webhook-Token", a.cfg.WebhookToken).
			SetBody(UserJoinNotify{UserId: uid, ChannelId: channelID, GuildId: guildID}).
			Post(a.cfg.WebhookURL + "/api/v1/webhook/sfu/voice/join")
		if err != nil {
			a.log.Error("user join notify failed", slog.String("error", err.Error()))
		} else if resp.StatusCode() != 200 {
			a.log.Warn("user join notify unexpected status", slog.Int("status", resp.StatusCode()))
		}
	}()
}

// notifyUserLeave sends a synchronous webhook notification for a user leaving voice.
// Called in a defer, so it runs before the WebSocket is torn down.
func (a *App) notifyUserLeave(uid, channelID int64, guildID *int64) {
	resp, err := a.sfu.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Webhook-Token", a.cfg.WebhookToken).
		SetBody(UserLeaveNotify{UserId: uid, ChannelId: channelID, GuildId: guildID}).
		Post(a.cfg.WebhookURL + "/api/v1/webhook/sfu/voice/leave")
	if err != nil {
		a.log.Error("user leave notify failed", slog.String("error", err.Error()))
	} else if resp.StatusCode() != 200 {
		a.log.Warn("user leave notify unexpected status", slog.Int("status", resp.StatusCode()))
	}
}

// ---------------------------------------------------------------------------
// WebSocket signal handler
// ---------------------------------------------------------------------------

func (a *App) handleSignalWS(c *websocket.Conn) {
	defer func() { _ = c.Close() }()

	// Phase 1: Handshake вЂ” read join envelope and authorize.
	joinEnv, err := a.readJoinEnvelope(c)
	if err != nil {
		a.log.Warn("invalid join envelope", slog.String("error", err.Error()))
		_ = (&threadSafeWriter{conn: c.Conn}).SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: ErrorResponse{Error: "invalid message"}})
		return
	}
	uid, channelID, guildID, perms, _, err := a.authorizeJoin(joinEnv)
	if err != nil {
		a.log.Warn("join unauthorized", slog.String("error", err.Error()))
		_ = (&threadSafeWriter{conn: c.Conn}).SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: ErrorResponse{Error: err.Error()}})
		return
	}

	if a.sfu.IsBlocked(channelID, uid) {
		a.log.Warn("blocked user tried to join", slog.Int64("user", uid), slog.Int64("channel", channelID))
		_ = (&threadSafeWriter{conn: c.Conn}).SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: ErrorResponse{Error: "blocked"}})
		return
	}

	// Phase 2: Setup вЂ” create PeerConnection and register it.
	a.notifyUserJoin(uid, channelID, guildID)
	defer a.notifyUserLeave(uid, channelID, guildID)

	pc, err := a.webrtcAPI.NewPeerConnection(a.iceConfig)
	if err != nil {
		a.log.Error("failed to create peer connection", slog.String("error", err.Error()))
		return
	}
	defer func() { _ = pc.Close() }()

	writer := &threadSafeWriter{conn: c.Conn}
	state := &peerConnectionState{peerConnection: pc, websocket: writer, userID: uid, perms: perms}

	if err := a.setupTransceivers(pc); err != nil {
		a.log.Error("failed to setup transceivers", slog.String("error", err.Error()))
		return
	}

	a.registerPeerCallbacks(pc, writer, state, uid, channelID, perms)

	if err := writer.SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: JoinAck{Ok: true}}); err != nil {
		a.log.Warn("failed to send join ack", slog.String("error", err.Error()))
		return
	}

	a.log.Info("client joined", slog.Int64("user", uid), slog.Int64("channel", channelID))

	a.sfu.AddPeer(channelID, state)
	a.totalPeers.Add(1)
	defer func() {
		writer.Close()
		a.sfu.RemovePeer(channelID, pc)
		a.totalPeers.Add(-1)
		a.log.Info("client left", slog.Int64("user", uid), slog.Int64("channel", channelID))
	}()

	a.sfu.SignalChannel(channelID)

	// Phase 3: Message loop
	a.messageLoop(c, pc, writer, uid, perms, channelID)
}

// setupTransceivers adds audio and video sendrecv transceivers to the peer connection.
func (a *App) setupTransceivers(pc *webrtc.PeerConnection) error {
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := pc.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendrecv,
		}); err != nil {
			return fmt.Errorf("add transceiver %s: %w", typ, err)
		}
	}
	return nil
}

// registerPeerCallbacks sets up OnICECandidate, OnConnectionStateChange, and OnTrack.
func (a *App) registerPeerCallbacks(
	pc *webrtc.PeerConnection,
	writer *threadSafeWriter,
	state *peerConnectionState,
	uid, channelID, perms int64,
) {
	pc.OnICECandidate(func(i *webrtc.ICECandidate) {
		if err := writer.SendRTCCandidate(i); err != nil {
			a.log.Warn("failed to send candidate", slog.String("error", err.Error()))
		}
	})

	pc.OnConnectionStateChange(func(connState webrtc.PeerConnectionState) {
		a.log.Info("connection state change", slog.String("state", connState.String()), slog.Int64("user", uid))
		switch connState {
		case webrtc.PeerConnectionStateFailed:
			if err := pc.Close(); err != nil {
				a.log.Warn("failed to close peer connection", slog.String("error", err.Error()))
			}
		case webrtc.PeerConnectionStateClosed:
			a.sfu.SignalChannel(channelID)
		}
	})

	pc.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		a.handleInboundTrack(pc, state, t, uid, channelID, perms)
	})
}

// handleInboundTrack processes a single inbound track, forwarding RTP packets to
// a local track while enforcing permissions and bitrate limits.
func (a *App) handleInboundTrack(
	pc *webrtc.PeerConnection,
	state *peerConnectionState,
	t *webrtc.TrackRemote,
	uid, channelID, perms int64,
) {
	// Recover from panics in the track read loop. Pion's internal buffers
	// can trigger a fault when the PeerConnection is closed during a read.
	defer func() {
		if r := recover(); r != nil {
			a.log.Error("recovered panic in OnTrack goroutine",
				slog.Any("panic", r),
				slog.Int64("user", uid),
				slog.Int64("channel", channelID),
				slog.String("track", t.ID()),
			)
		}
	}()

	a.log.Info("inbound track", slog.String("kind", t.Kind().String()), slog.String("id", t.ID()))

	// Permission enforcement
	if t.Kind() == webrtc.RTPCodecTypeAudio && !hasPerm(perms, permissions.PermVoiceSpeak) {
		a.log.Warn("rejecting audio track: no PermVoiceSpeak", slog.Int64("user", uid))
		return
	}
	if t.Kind() == webrtc.RTPCodecTypeVideo && !hasPerm(perms, permissions.PermVoiceVideo) {
		a.log.Warn("rejecting video track: no PermVoiceVideo", slog.Int64("user", uid))
		return
	}

	// Server mute check
	if state.serverMuted {
		a.log.Info("rejecting track: user is server-muted", slog.Int64("user", uid))
		return
	}

	trackLocal := a.sfu.AddTrack(channelID, uid, t)
	if trackLocal == nil {
		a.log.Warn("failed to create forwarding track", slog.Int64("user", uid), slog.Int64("channel", channelID), slog.String("track", t.ID()), slog.String("kind", t.Kind().String()))
		return
	}
	defer a.sfu.RemoveTrack(channelID, trackLocal)

	a.forwardRTP(pc, t, trackLocal, uid, channelID)
}

// forwardRTP reads RTP packets from the remote track and writes them to the local track.
// Handles audio bitrate enforcement when configured.
func (a *App) forwardRTP(
	pc *webrtc.PeerConnection,
	remote *webrtc.TrackRemote,
	local *webrtc.TrackLocalStaticRTP,
	uid, channelID int64,
) {
	// Recover from panics in pion's interceptor chain. The RTCP receiver report
	// interceptor can crash with a nil pointer dereference when the PeerConnection
	// is closed during RTP processing (race condition in pion/interceptor v0.1.41).
	defer func() {
		if r := recover(); r != nil {
			a.log.Error("recovered panic in forwardRTP",
				slog.Any("panic", r),
				slog.Int64("user", uid),
				slog.Int64("channel", channelID),
				slog.String("track", remote.ID()),
			)
		}
	}()

	buf := make([]byte, 1500)
	rtpPacket := &rtp.Packet{}

	// Bitrate enforcement for audio if configured
	enforce := a.sfu.enforceAudioBitrate &&
		a.sfu.maxAudioBitrateBps > 0 &&
		remote.Kind() == webrtc.RTPCodecTypeAudio

	limitWithMargin := float64(a.sfu.maxAudioBitrateBps)
	if a.sfu.audioBitrateMarginPct > 0 {
		limitWithMargin *= 1.0 + float64(a.sfu.audioBitrateMarginPct)/100.0
	}

	var (
		windowStart   = time.Now()
		bytesInWindow int64
		overCount     int
		firstPacket   = true
	)

	for {
		// Bail out early if the PeerConnection is no longer active.
		if pcState := pc.ConnectionState(); pcState == webrtc.PeerConnectionStateClosed || pcState == webrtc.PeerConnectionStateFailed {
			return
		}

		n, _, err := remote.Read(buf)
		if err != nil {
			// EOF / closed are expected on normal disconnect вЂ” debug only.
			if pc.ConnectionState() == webrtc.PeerConnectionStateClosed ||
				pc.ConnectionState() == webrtc.PeerConnectionStateFailed {
				a.log.Debug("rtp read stopped (peer closed)",
					slog.Int64("user", uid), slog.Int64("channel", channelID),
					slog.String("track", remote.ID()), slog.String("kind", remote.Kind().String()))
			} else {
				a.log.Warn("rtp read error",
					slog.Int64("user", uid), slog.Int64("channel", channelID),
					slog.String("track", remote.ID()), slog.String("kind", remote.Kind().String()),
					slog.String("error", err.Error()))
			}
			return
		}

		if firstPacket {
			firstPacket = false
			a.log.Debug("first rtp packet received",
				slog.Int64("user", uid), slog.Int64("channel", channelID),
				slog.String("track", remote.ID()), slog.String("kind", remote.Kind().String()),
				slog.String("codec", remote.Codec().MimeType))
		}

		if err = rtpPacket.Unmarshal(buf[:n]); err != nil {
			a.log.Warn("failed to unmarshal rtp packet",
				slog.Int64("user", uid), slog.Int64("channel", channelID),
				slog.String("error", err.Error()))
			return
		}

		if enforce {
			bytesInWindow += int64(n)
			elapsed := time.Since(windowStart)
			if elapsed >= time.Second {
				bps := (float64(bytesInWindow) * 8.0) / elapsed.Seconds()
				a.log.Debug("audio bitrate window",
					slog.Int64("user", uid), slog.Int64("channel", channelID),
					slog.Float64("bps", bps), slog.Float64("limit_bps", limitWithMargin))
				if bps > limitWithMargin {
					overCount++
				} else {
					overCount = 0
				}
				windowStart = time.Now()
				bytesInWindow = 0
				// Allow brief spikes, disconnect on sustained exceed (2+ consecutive windows)
				if overCount >= 2 {
					a.log.Warn("disconnecting peer due to audio bitrate limit exceed",
						slog.Int64("user", uid),
						slog.Float64("bps", bps),
						slog.Float64("limit_bps", limitWithMargin),
						slog.Int64("channel", channelID))
					_ = pc.Close()
					return
				}
			}
		}

		if err = local.WriteRTP(rtpPacket); err != nil {
			a.log.Warn("failed to write rtp to local track",
				slog.Int64("user", uid), slog.Int64("channel", channelID),
				slog.String("track", remote.ID()), slog.String("error", err.Error()))
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Message loop
// ---------------------------------------------------------------------------

func (a *App) messageLoop(
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

		// Try simple event-based format first
		var msg websocketMessage
		if err := json.Unmarshal(raw, &msg); err == nil && msg.Event != "" {
			if a.handleSimpleMessage(msg, pc, writer, uid, channelID) {
				return
			}
			continue
		}

		// Fall back to legacy envelope format
		var env envelope
		if err := json.Unmarshal(raw, &env); err == nil && env.OP != 0 {
			if a.handleLegacyEnvelope(env, pc, writer, uid, perms, channelID) {
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
	msg websocketMessage,
	pc *webrtc.PeerConnection,
	writer *threadSafeWriter,
	uid, channelID int64,
) bool {
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

	case "negotiate":
		a.log.Info("client requested renegotiation", slog.Int64("channel", channelID), slog.Int64("user", uid))
		a.sfu.SignalChannel(channelID)

	case "speaking":
		speaking := parseSpeakingData(msg.Data)
		a.log.Debug("speaking event", slog.Int64("user", uid), slog.Int64("channel", channelID), slog.Int("speaking", speaking))
		a.sfu.BroadcastSpeaking(channelID, uid, speaking)

	default:
		a.log.Warn("unknown message", slog.String("event", msg.Event))
	}
	return false
}

// parseSpeakingData extracts a speaking indicator (0 or 1) from various payload formats.
func parseSpeakingData(data string) int {
	switch data {
	case "1", "\"1\"":
		return 1
	case "0", "\"0\"", "":
		return 0
	}
	var aux struct {
		Speaking int `json:"speaking"`
	}
	_ = json.Unmarshal([]byte(data), &aux)
	if aux.Speaking != 0 {
		return 1
	}
	return 0
}

// ---------------------------------------------------------------------------
// Legacy envelope handler
// ---------------------------------------------------------------------------

func (a *App) handleLegacyEnvelope(env envelope, pc *webrtc.PeerConnection, writer *threadSafeWriter, uid int64, perms int64, channelID int64) bool {
	switch env.OP {
	case int(mqmsg.OPCodeHeartBeat):
		var hb heartbeatData
		_ = json.Unmarshal(env.D, &hb)
		_ = writer.SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeHeartBeat), D: HeartbeatReply{Pong: true, ServerTS: time.Now().UnixMilli(), Nonce: hb.Nonce, TS: hb.TS}})
		return false

	case int(mqmsg.OPCodeRTC):
		return a.handleLegacyRTCEvent(env, pc, writer, uid, perms, channelID)
	}
	return false
}

func (a *App) handleLegacyRTCEvent(env envelope, pc *webrtc.PeerConnection, writer *threadSafeWriter, uid int64, perms int64, channelID int64) bool {
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
		}

	case int(mqmsg.EventTypeRTCCandidate):
		var cand rtcCandidate
		if err := json.Unmarshal(env.D, &cand); err != nil || cand.Candidate == "" {
			return false
		}
		if err := pc.AddICECandidate(webrtc.ICECandidateInit{Candidate: cand.Candidate, SDPMid: cand.SDPMid, SDPMLineIndex: cand.SDPMLineIndex}); err != nil {
			a.log.Warn("failed to add legacy candidate", slog.String("error", err.Error()))
		}

	case int(mqmsg.EventTypeRTCLeave):
		return true

	// в”Ђв”Ђ Server control events в”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђв”Ђ
	case int(mqmsg.EventTypeRTCServerMuteUser):
		if !hasPerm(perms, permissions.PermVoiceMuteMembers) {
			a.log.Warn("mute denied: insufficient permissions", slog.Int64("user", uid))
			return false
		}
		var data muteUserData
		if err := json.Unmarshal(env.D, &data); err != nil {
			return false
		}
		a.sfu.ServerMuteUser(channelID, data.User, data.Muted)

	case int(mqmsg.EventTypeRTCServerDeafenUser):
		if !hasPerm(perms, permissions.PermVoiceDeafenMembers) {
			a.log.Warn("deafen denied: insufficient permissions", slog.Int64("user", uid))
			return false
		}
		var data deafenUserData
		if err := json.Unmarshal(env.D, &data); err != nil {
			return false
		}
		a.sfu.ServerDeafenUser(channelID, data.User, data.Deafened)

	case int(mqmsg.EventTypeRTCServerKickUser):
		if !hasPerm(perms, permissions.PermVoiceMoveMembers) {
			a.log.Warn("kick denied: insufficient permissions", slog.Int64("user", uid))
			return false
		}
		var data kickUserData
		if err := json.Unmarshal(env.D, &data); err != nil {
			return false
		}
		a.sfu.KickUser(channelID, data.User)

	case int(mqmsg.EventTypeRTCServerBlockUser):
		if !hasPerm(perms, permissions.PermVoiceMoveMembers) {
			a.log.Warn("block denied: insufficient permissions", slog.Int64("user", uid))
			return false
		}
		var data blockEvent
		if err := json.Unmarshal(env.D, &data); err != nil {
			return false
		}
		a.sfu.BlockUser(channelID, data.UserId, data.Block)
	}
	return false
}

// parseLegacySDPType maps a string SDP type to webrtc.SDPType, defaulting to answer.
func parseLegacySDPType(t string) webrtc.SDPType {
	switch {
	case strings.EqualFold(t, webrtc.SDPTypeOffer.String()):
		return webrtc.SDPTypeOffer
	case strings.EqualFold(t, webrtc.SDPTypePranswer.String()):
		return webrtc.SDPTypePranswer
	case strings.EqualFold(t, webrtc.SDPTypeRollback.String()):
		return webrtc.SDPTypeRollback
	default:
		return webrtc.SDPTypeAnswer
	}
}

// ---------------------------------------------------------------------------
// Admin endpoint
// ---------------------------------------------------------------------------

// handleAdminCloseChannel closes all peer connections in a voice channel.
// Requires a valid admin JWT in the Authorization header.
func (a *App) handleAdminCloseChannel(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	channelID, err := a.validateAdminToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req CloseChannelRequest
	if err := c.BodyParser(&req); err != nil || req.ChannelID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	// Honour channel_id from both the token and the body; they must match.
	if channelID != 0 && channelID != req.ChannelID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "channel mismatch"})
	}
	a.sfu.KickAll(req.ChannelID)
	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Discovery heartbeat
// ---------------------------------------------------------------------------

func (a *App) discoveryHeartbeat() {
	if a.cfg.WebhookURL == "" {
		a.log.Warn("discovery heartbeat disabled", slog.String("reason", "webhook url missing"))
		return
	}

	url := buildSignalURL(a.cfg.PublicBaseURL)

	client := a.sfu.httpClient
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	type heartbeatPayload struct {
		ID     string `json:"id"`
		Region string `json:"region"`
		URL    string `json:"url"`
		Load   int64  `json:"load"`
	}

	for {
		select {
		case <-ticker.C:
			payload := heartbeatPayload{
				ID:     a.instID,
				Region: a.cfg.Region,
				URL:    url,
				Load:   a.totalPeers.Load(),
			}
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-Webhook-Token", a.cfg.WebhookToken).
				SetBody(payload).
				Post(a.cfg.WebhookURL + "/api/v1/webhook/sfu/heartbeat")
			if err != nil {
				a.log.Error("heartbeat request failed", slog.String("error", err.Error()))
				continue
			}
			if resp.StatusCode() != 204 {
				a.log.Warn("heartbeat unexpected status", slog.Int("status", resp.StatusCode()), slog.String("body", resp.String()))
				continue
			}
			a.discoverLog.Do(func() {
				a.log.Info("Service registered and discoverable")
			})

		case <-a.sfu.done:
			return
		}
	}
}

// buildSignalURL converts a public base URL to a WebSocket signal endpoint URL.
func buildSignalURL(publicBaseURL string) string {
	if publicBaseURL == "" {
		return "ws://localhost:3300/signal"
	}
	url := publicBaseURL
	lower := strings.ToLower(url)
	switch {
	case strings.HasPrefix(lower, "https://"):
		url = "wss://" + url[len("https://"):]
	case strings.HasPrefix(lower, "http://"):
		url = "ws://" + url[len("http://"):]
	}
	if !strings.HasSuffix(url, "/signal") {
		url = strings.TrimRight(url, "/") + "/signal"
	}
	return url
}

// ---------------------------------------------------------------------------
// Handshake helpers
// ---------------------------------------------------------------------------

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
	if tokChannel != 0 && tokChannel != env.D.Channel {
		return 0, 0, nil, 0, false, fmt.Errorf("unauthorized")
	}
	return uid, env.D.Channel, tokGuild, perms, moved, nil
}
