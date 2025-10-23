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
	recm "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	slogfiber "github.com/samber/slog-fiber"
	"resty.dev/v3"

	"github.com/FlameInTheDark/gochat/cmd/sfu/config"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

type App struct {
	app  *fiber.App
	cfg  *config.Config
	log  *slog.Logger
	shut *shutter.Shut
	sfu  *SFU

	iceConfig webrtc.Configuration

	instID      string
	totalPeers  atomic.Int64
	discoverLog sync.Once
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) *App {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("unable to load config", slog.String("error", err.Error()))
		panic(err)
	}

	iceCfg := webrtc.Configuration{}
	for _, raw := range cfg.STUNServers {
		url := strings.TrimSpace(raw)
		if url == "" {
			continue
		}
		iceCfg.ICEServers = append(iceCfg.ICEServers, webrtc.ICEServer{URLs: []string{url}})
	}
	if len(iceCfg.ICEServers) == 0 {
		iceCfg.ICEServers = []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}}
	}

	fiberApp := fiber.New(fiber.Config{DisableStartupMessage: true})
	lm := slogfiber.NewWithFilters(
		logger,
		slogfiber.IgnorePath("/metrics"),
	)
	fiberApp.Use(lm)
	fiberApp.Use(recm.New())

	sfu := NewSFU(logger)

	a := &App{
		app:       fiberApp,
		cfg:       cfg,
		log:       logger,
		shut:      shut,
		sfu:       sfu,
		instID:    cfg.ServiceID,
		iceConfig: iceCfg,
	}

	fiberApp.Get("/signal", websocket.New(a.handleSignalWS, websocket.Config{}))
	go sfu.RunKeyFrameTicker()

	return a
}

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

func (a *App) Close() error { return a.app.Shutdown() }

func (a *App) handleSignalWS(c *websocket.Conn) {
	defer c.Close()

	joinEnv, err := a.readJoinEnvelope(c)
	if err != nil {
		a.log.Warn("invalid join envelope", slog.String("error", err.Error()))
		_ = (&threadSafeWriter{conn: c.Conn}).SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: ErrorResponse{Error: "invalid message"}})
		return
	}
	uid, channelID, _, _, err := a.authorizeJoin(joinEnv)
	if err != nil {
		a.log.Warn("join unauthorized", slog.String("error", err.Error()))
		_ = (&threadSafeWriter{conn: c.Conn}).SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: ErrorResponse{Error: err.Error()}})
		return
	}

	pc, err := webrtc.NewPeerConnection(a.iceConfig)
	if err != nil {
		a.log.Error("failed to create peer connection", slog.String("error", err.Error()))
		return
	}
	defer pc.Close()

	writer := &threadSafeWriter{conn: c.Conn}
	state := &peerConnectionState{peerConnection: pc, websocket: writer}

	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := pc.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			a.log.Error("failed to add transceiver", slog.String("error", err.Error()))
			return
		}
	}

	pc.OnICECandidate(func(i *webrtc.ICECandidate) {
		if err := writer.SendRTCCandidate(i); err != nil {
			a.log.Warn("failed to send candidate", slog.String("error", err.Error()))
		}
	})

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		a.log.Info("connection state change", slog.String("state", state.String()))
		if state == webrtc.PeerConnectionStateFailed {
			if err := pc.Close(); err != nil {
				a.log.Warn("failed to close peer connection", slog.String("error", err.Error()))
			}
		}
		if state == webrtc.PeerConnectionStateClosed {
			a.sfu.SignalChannel(channelID)
		}
	})

	pc.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		a.log.Info("inbound track", slog.String("kind", t.Kind().String()), slog.String("id", t.ID()))
		trackLocal := a.sfu.AddTrack(channelID, t)
		if trackLocal == nil {
			return
		}
		defer a.sfu.RemoveTrack(channelID, trackLocal)

		buf := make([]byte, 1500)
		rtpPacket := &rtp.Packet{}

		for {
			n, _, err := t.Read(buf)
			if err != nil {
				return
			}
			if err = rtpPacket.Unmarshal(buf[:n]); err != nil {
				a.log.Warn("failed to unmarshal rtp packet", slog.String("error", err.Error()))
				return
			}
			rtpPacket.Extension = false
			rtpPacket.Extensions = nil
			if err = trackLocal.WriteRTP(rtpPacket); err != nil {
				return
			}
		}
	})

	a.sfu.AddPeer(channelID, state)
	a.totalPeers.Add(1)
	defer func() {
		a.sfu.RemovePeer(channelID, pc)
		a.totalPeers.Add(-1)
	}()

	if err := writer.SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: JoinAck{Ok: true}}); err != nil {
		a.log.Warn("failed to send join ack", slog.String("error", err.Error()))
		return
	}

	a.log.Info("client joined", slog.Int64("user", uid), slog.Int64("channel", channelID))

	a.sfu.SignalChannel(channelID)

	for {
		_, raw, err := c.ReadMessage()
		if err != nil {
			a.log.Warn("failed to read message", slog.String("error", err.Error()))
			return
		}

		var msg websocketMessage
		if err := json.Unmarshal(raw, &msg); err == nil && msg.Event != "" {
			switch msg.Event {
			case "candidate":
				var cand webrtc.ICECandidateInit
				if err := json.Unmarshal([]byte(msg.Data), &cand); err != nil {
					a.log.Warn("failed to parse candidate", slog.String("error", err.Error()))
					return
				}
				if err := pc.AddICECandidate(cand); err != nil {
					a.log.Warn("failed to add candidate", slog.String("error", err.Error()))
					return
				}
			case "answer":
				var answer webrtc.SessionDescription
				if err := json.Unmarshal([]byte(msg.Data), &answer); err != nil {
					a.log.Warn("failed to parse answer", slog.String("error", err.Error()))
					return
				}
				if answer.Type == webrtc.SDPType(0) {
					answer.Type = webrtc.SDPTypeAnswer
				}
				if err := pc.SetRemoteDescription(answer); err != nil {
					a.log.Warn("failed to set remote description", slog.String("error", err.Error()))
					return
				}
			default:
				a.log.Warn("unknown message", slog.String("event", msg.Event))
			}
			continue
		}

		var env envelope
		if err := json.Unmarshal(raw, &env); err == nil && env.OP != 0 {
			if a.handleLegacyEnvelope(env, pc, writer) {
				return
			}
			continue
		}

		a.log.Warn("unrecognized message format")
	}
}

func (a *App) discoveryHeartbeat() {
	if a.cfg.WebhookURL == "" {
		a.log.Warn("discovery heartbeat disabled", slog.String("reason", "webhook url missing"))
		return
	}

	url := a.cfg.PublicBaseURL
	if url == "" {
		url = "ws://localhost:3300/signal"
	} else {
		lower := strings.ToLower(url)
		switch {
		case strings.HasPrefix(lower, "https://"):
			url = "wss://" + strings.TrimPrefix(url, "https://")
		case strings.HasPrefix(lower, "http://"):
			url = "ws://" + strings.TrimPrefix(url, "http://")
		}
		if !strings.HasSuffix(url, "/signal") {
			url = strings.TrimRight(url, "/") + "/signal"
		}
	}

	client := resty.New().SetTimeout(5 * time.Second)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	type heartbeatPayload struct {
		ID     string `json:"id"`
		Region string `json:"region"`
		URL    string `json:"url"`
		Load   int64  `json:"load"`
	}

	for range ticker.C {
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
			Post(a.cfg.WebhookURL)
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
	}
}

func (a *App) readJoinEnvelope(c *websocket.Conn) (rtcJoinEnvelope, error) {
	var env rtcJoinEnvelope
	if c.Conn != nil {
		_ = c.Conn.SetReadDeadline(time.Now().Add(joinHandshakeTimeout))
	}
	if err := c.ReadJSON(&env); err != nil {
		return rtcJoinEnvelope{}, err
	}
	if c.Conn != nil {
		_ = c.Conn.SetReadDeadline(time.Time{})
	}
	return env, nil
}

func (a *App) authorizeJoin(env rtcJoinEnvelope) (int64, int64, int64, bool, error) {
	if env.OP != int(mqmsg.OPCodeRTC) || env.T != int(mqmsg.EventTypeRTCJoin) || env.D.Token == "" || env.D.Channel == 0 {
		return 0, 0, 0, false, fmt.Errorf("expected join")
	}
	uid, tokChannel, perms, moved, err := a.validateJoinToken(env.D.Token)
	if err != nil {
		return 0, 0, 0, false, fmt.Errorf("unauthorized")
	}
	if tokChannel != 0 && tokChannel != env.D.Channel {
		return 0, 0, 0, false, fmt.Errorf("unauthorized")
	}
	return uid, env.D.Channel, perms, moved, nil
}

func (a *App) handleLegacyEnvelope(env envelope, pc *webrtc.PeerConnection, writer *threadSafeWriter) bool {
	switch env.OP {
	case int(mqmsg.OPCodeHeartBeat):
		var hb heartbeatData
		_ = json.Unmarshal(env.D, &hb)
		_ = writer.SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeHeartBeat), D: HeartbeatReply{Pong: true, ServerTS: time.Now().UnixMilli(), Nonce: hb.Nonce, TS: hb.TS}})
		return false
	case int(mqmsg.OPCodeRTC):
		switch env.T {
		case int(mqmsg.EventTypeRTCAnswer):
			var ans rtcAnswer
			if err := json.Unmarshal(env.D, &ans); err != nil || ans.SDP == "" {
				return false
			}
			descType := webrtc.SDPTypeAnswer
			switch {
			case strings.EqualFold(ans.Type, webrtc.SDPTypeOffer.String()):
				descType = webrtc.SDPTypeOffer
			case strings.EqualFold(ans.Type, webrtc.SDPTypePranswer.String()):
				descType = webrtc.SDPTypePranswer
			case strings.EqualFold(ans.Type, webrtc.SDPTypeRollback.String()):
				descType = webrtc.SDPTypeRollback
			case strings.EqualFold(ans.Type, webrtc.SDPTypeAnswer.String()):
				descType = webrtc.SDPTypeAnswer
			}
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
		}
	}
	return false
}
