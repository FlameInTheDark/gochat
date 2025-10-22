package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	recm "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	slogfiber "github.com/samber/slog-fiber"

	"github.com/FlameInTheDark/gochat/cmd/sfu/config"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

type App struct {
	app  *fiber.App
	cfg  *config.Config
	log  *slog.Logger
	shut *shutter.Shut
	sfu  *SFU
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) *App {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("unable to load config", slog.String("error", err.Error()))
		panic(err)
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
		app:  fiberApp,
		cfg:  cfg,
		log:  logger,
		shut: shut,
		sfu:  sfu,
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

	go a.heartbeatLoop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func (a *App) Close() error { return a.app.Shutdown() }

func (a *App) handleSignalWS(c *websocket.Conn) {
	defer c.Close()

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
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
		if i == nil {
			return
		}
		candidateJSON, err := json.Marshal(i.ToJSON())
		if err != nil {
			a.log.Warn("failed to marshal candidate", slog.String("error", err.Error()))
			return
		}
		if err := writer.WriteJSON(&websocketMessage{Event: "candidate", Data: string(candidateJSON)}); err != nil {
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
			a.sfu.SignalPeerConnections()
		}
	})

	pc.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		a.log.Info("inbound track", slog.String("kind", t.Kind().String()), slog.String("id", t.ID()))
		trackLocal := a.sfu.AddTrack(t)
		if trackLocal == nil {
			return
		}
		defer a.sfu.RemoveTrack(trackLocal)

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

	a.sfu.AddPeer(state)
	defer a.sfu.RemovePeer(pc)

	a.sfu.SignalPeerConnections()

	for {
		_, raw, err := c.ReadMessage()
		if err != nil {
			a.log.Warn("failed to read message", slog.String("error", err.Error()))
			return
		}

		var msg websocketMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			a.log.Warn("failed to unmarshal message", slog.String("error", err.Error()))
			return
		}

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
			if err := pc.SetRemoteDescription(answer); err != nil {
				a.log.Warn("failed to set remote description", slog.String("error", err.Error()))
				return
			}
		default:
			a.log.Warn("unknown message", slog.String("event", msg.Event))
		}
	}
}

func (a *App) heartbeatLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		load := a.sfu.PeerCount()
		a.log.Info("heartbeat", slog.Int("peers", load))
	}
}
