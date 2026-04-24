package main

import (
	"context"
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
	recm "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	daveserver "github.com/FlameInTheDark/go-dave/server"
	"github.com/FlameInTheDark/gochat/cmd/stream/config"
	"github.com/FlameInTheDark/gochat/internal/observability"
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
	telemetry   *observability.SFUTelemetry

	dave                 *daveserver.Coordinator
	signalHeartbeatGrace time.Duration
	signalV2Mu           sync.Mutex
	signalV2Sessions     map[string]*signalV2Session
}

// websocketMessage is the simple event-based message format used over WebSocket.
type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

// NewApp creates a fully configured SFU application.
func NewApp(shut *shutter.Shut, logger *slog.Logger, cfg *config.Config) *App {
	if cfg == nil {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			logger.Error("unable to load config", slog.String("error", err.Error()))
			panic(err)
		}
	}

	iceCfg := buildICEConfig(cfg.STUNServers)
	dtlsCertificates, err := loadDTLSCertificates(logger, cfg)
	if err != nil {
		logger.Error("unable to configure dtls certificate", slog.String("error", err.Error()))
		panic(err)
	}
	iceCfg.Certificates = dtlsCertificates
	api, err := buildWebRTCAPI(logger, cfg.DAVEAllowAV1, cfg.ICEPublicIP, cfg.UDPPortRangeStart, cfg.UDPPortRangeEnd)
	if err != nil {
		logger.Error("unable to configure webrtc api", slog.String("error", err.Error()))
		panic(err)
	}

	fiberApp := fiber.New(fiber.Config{DisableStartupMessage: true})
	fiberApp.Use(observability.RequestContextMiddleware())
	fiberApp.Use(observability.RequestLogger(logger))
	fiberApp.Use(observability.NewHTTPServerTelemetry("gochat-stream").Middleware())
	fiberApp.Use(recm.New())

	// Compute max audio bitrate in bps (0 means disabled)
	var maxAudioBps uint64
	if cfg.MaxAudioBitrateKbps > 0 {
		maxAudioBps = uint64(cfg.MaxAudioBitrateKbps) * 1000
	}
	var maxVideoBps uint64
	if cfg.MaxVideoBitrateKbps > 0 {
		maxVideoBps = uint64(cfg.MaxVideoBitrateKbps) * 1000
	}
	// Clamp margin to [0,100]
	marginPct := cfg.AudioBitrateMarginPercent
	if marginPct < 0 {
		marginPct = 0
	} else if marginPct > 100 {
		marginPct = 100
	}
	telemetry := observability.NewSFUTelemetry("gochat-stream",
		attribute.String("voice.region", cfg.Region),
		attribute.String("service.instance.id", cfg.ServiceID),
	)
	signalURL := buildSignalURL(cfg.PublicBaseURL)
	sfu := NewSFU(cfg.WebhookURL, cfg.WebhookToken, cfg.ServiceID, signalURL, cfg.Region, logger, maxAudioBps, maxVideoBps, cfg.EnforceAudioBitrate, marginPct, telemetry)

	a := &App{
		app:                  fiberApp,
		cfg:                  cfg,
		log:                  logger,
		shut:                 shut,
		sfu:                  sfu,
		instID:               cfg.ServiceID,
		iceConfig:            iceCfg,
		webrtcAPI:            api,
		telemetry:            telemetry,
		signalHeartbeatGrace: signalHeartbeatGrace,
		signalV2Sessions:     make(map[string]*signalV2Session),
	}
	a.dave = daveserver.NewCoordinator(daveserver.Config{
		Enabled:             cfg.DAVEEnabled,
		RequiredByDefault:   cfg.DAVERequiredDefault,
		TransitionTimeout:   time.Duration(cfg.DAVETransitionTimeoutMS) * time.Millisecond,
		OldRatchetRetention: time.Duration(cfg.DAVEOldRatchetWindowMS) * time.Millisecond,
		AllowAV1:            cfg.DAVEAllowAV1,
	}, a)

	fiberApp.Use("/signal", func(c *fiber.Ctx) error {
		if _, ok := parseSignalProtocolVersion(c.Query("v")); !ok {
			return fiber.NewError(fiber.StatusBadRequest, "unsupported signal protocol version")
		}
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	fiberApp.Get("/signal", websocket.New(a.handleSignalWS, websocket.Config{
		ReadBufferSize:  32768, // 32KB — large enough for video SDP renegotiations
		WriteBufferSize: 32768,
	}))
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

// ---------------------------------------------------------------------------
// Webhook notification helpers
// ---------------------------------------------------------------------------

// notifyUserJoin sends an async webhook notification for a user joining voice.
func (a *App) notifyUserJoin(ctx context.Context, uid, streamID, voiceChannelID int64, guildID *int64, sourceType, audioMode, publisherSessionID string) {
	go func() {
		reqCtx := observability.BackgroundFromContext(ctx)
		resp, err := a.sfu.httpClient.R().
			SetContext(reqCtx).
			SetHeader("Content-Type", "application/json").
			SetHeader("X-Webhook-Token", a.cfg.WebhookToken).
			SetBody(UserJoinNotify{
				StreamId:           streamID,
				UserId:             uid,
				ChannelId:          voiceChannelID,
				GuildId:            guildID,
				SourceType:         sourceType,
				AudioMode:          audioMode,
				PublisherSessionID: publisherSessionID,
				RouteID:            a.cfg.ServiceID,
				RouteURL:           buildSignalURL(a.cfg.PublicBaseURL),
				Region:             a.cfg.Region,
			}).
			Post(a.cfg.WebhookURL + "/api/v1/webhook/stream/start")
		if err != nil {
			a.log.Error("user join notify failed", slog.String("error", err.Error()))
		} else if resp.StatusCode() != 200 {
			a.log.Warn("user join notify unexpected status", slog.Int("status", resp.StatusCode()))
		}
	}()
}

// notifyUserLeave sends a synchronous webhook notification for a user leaving voice.
// Called in a defer, so it runs before the WebSocket is torn down.
func (a *App) notifyUserLeave(ctx context.Context, uid, streamID, voiceChannelID int64, guildID *int64, publisherSessionID, reason string) {
	resp, err := a.sfu.httpClient.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Webhook-Token", a.cfg.WebhookToken).
		SetBody(UserLeaveNotify{
			StreamId:           streamID,
			UserId:             uid,
			ChannelId:          voiceChannelID,
			GuildId:            guildID,
			PublisherSessionID: publisherSessionID,
			Reason:             reason,
		}).
		Post(a.cfg.WebhookURL + "/api/v1/webhook/stream/stop")
	if err != nil {
		a.log.Error("user leave notify failed", slog.String("error", err.Error()))
	} else if resp.StatusCode() != 200 {
		a.log.Warn("user leave notify unexpected status", slog.Int("status", resp.StatusCode()))
	}
}

// handleInboundTrack processes a single inbound track, forwarding RTP packets to
// a local track while enforcing permissions and bitrate limits.
func (a *App) handleInboundTrack(
	ctx context.Context,
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

	trackCtx, trackSpan := observability.Tracer("gochat/sfu").Start(ctx, "sfu.track")
	trackSpan.SetAttributes(
		attribute.String("track.kind", t.Kind().String()),
		attribute.String("track.id", t.ID()),
		attribute.Int64("voice.channel_id", channelID),
		attribute.Int64("user.id", uid),
	)
	defer trackSpan.End()

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

	trackLocal := a.sfu.AddTrack(trackCtx, channelID, uid, t)
	if trackLocal == nil {
		a.log.Warn("failed to create forwarding track", slog.Int64("user", uid), slog.Int64("channel", channelID), slog.String("track", t.ID()), slog.String("kind", t.Kind().String()))
		return
	}
	defer a.sfu.RemoveTrack(trackCtx, channelID, trackLocal)

	a.forwardRTP(trackCtx, pc, t, trackLocal, uid, channelID)
}

// forwardRTP reads RTP packets from the remote track and writes them to the local track.
// Handles audio bitrate enforcement when configured.
func (a *App) forwardRTP(
	ctx context.Context,
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
					a.telemetry.BitrateDisconnect(ctx,
						attribute.Int64("voice.channel_id", channelID),
						attribute.Int64("user.id", uid),
						attribute.String("track.kind", remote.Kind().String()),
					)
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
// Admin endpoint
// ---------------------------------------------------------------------------

// handleAdminCloseChannel closes all peer connections in a voice channel.
// Requires a valid admin JWT in the Authorization header.
func (a *App) handleAdminCloseChannel(c *fiber.Ctx) error {
	started := time.Now()
	token := c.Get("Authorization")
	channelID, err := a.validateAdminToken(token)
	if err != nil {
		a.telemetry.AdminClose(c.UserContext(), started, "unauthorized", attribute.Int64("voice.channel_id", channelID))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req CloseChannelRequest
	if err := c.BodyParser(&req); err != nil || req.ChannelID == 0 {
		a.telemetry.AdminClose(c.UserContext(), started, "bad_request", attribute.Int64("voice.channel_id", req.ChannelID))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	// Honour channel_id from both the token and the body; they must match.
	if channelID != 0 && channelID != req.ChannelID {
		a.telemetry.AdminClose(c.UserContext(), started, "forbidden", attribute.Int64("voice.channel_id", req.ChannelID))
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "channel mismatch"})
	}
	a.sfu.KickAll(c.UserContext(), req.ChannelID)
	a.telemetry.AdminClose(c.UserContext(), started, "ok", attribute.Int64("voice.channel_id", req.ChannelID))
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
			heartbeatStart := time.Now()
			hbCtx, hbSpan := observability.Tracer("gochat/sfu").Start(context.Background(), "sfu.discovery_heartbeat")
			hbSpan.SetAttributes(
				attribute.String("voice.region", a.cfg.Region),
				attribute.String("service.instance.id", a.cfg.ServiceID),
			)
			payload := heartbeatPayload{
				ID:     a.instID,
				Region: a.cfg.Region,
				URL:    url,
				Load:   a.totalPeers.Load(),
			}
			resp, err := client.R().
				SetContext(hbCtx).
				SetHeader("Content-Type", "application/json").
				SetHeader("X-Webhook-Token", a.cfg.WebhookToken).
				SetBody(payload).
				Post(a.cfg.WebhookURL + "/api/v1/webhook/stream/heartbeat")
			if err != nil {
				hbSpan.RecordError(err)
				hbSpan.SetStatus(codes.Error, err.Error())
				a.telemetry.Heartbeat(hbCtx, heartbeatStart, "error")
				hbSpan.End()
				a.log.Error("heartbeat request failed", slog.String("error", err.Error()))
				continue
			}
			if resp.StatusCode() != 204 {
				hbSpan.SetStatus(codes.Error, fmt.Sprintf("status_%d", resp.StatusCode()))
				a.telemetry.Heartbeat(hbCtx, heartbeatStart, fmt.Sprintf("status_%d", resp.StatusCode()))
				hbSpan.End()
				a.log.Warn("heartbeat unexpected status", slog.Int("status", resp.StatusCode()), slog.String("body", resp.String()))
				continue
			}
			hbSpan.SetStatus(codes.Ok, "registered")
			a.telemetry.Heartbeat(hbCtx, heartbeatStart, "ok")
			hbSpan.End()
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
		return "ws://localhost:3310/signal"
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
