package main

import (
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
	slogfiber "github.com/samber/slog-fiber"
	"resty.dev/v3"

	"github.com/FlameInTheDark/gochat/cmd/sfu/config"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/shutter"
)

type App struct {
	app  *fiber.App
	cfg  *config.Config
	log  *slog.Logger
	shut *shutter.Shut

	rooms  *roomManager
	instID string
	// totalPeers tracks total connected peers across all rooms.
	totalPeers atomic.Int64
}

func NewApp(shut *shutter.Shut, logger *slog.Logger) *App {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("unable to load config", slog.String("error", err.Error()))
		panic(err)
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	lm := slogfiber.NewWithFilters(
		logger,
		slogfiber.IgnorePath("/metrics"),
	)
	app.Use(lm)
	app.Use(recm.New())

	a := &App{app: app, cfg: cfg, log: logger, shut: shut, rooms: newRoomManager(logger, cfg), instID: cfg.ServiceID}

	app.Get("/signal", websocket.New(a.handleSignalWS, websocket.Config{}))
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
	// Block until termination signal so the process doesn't exit immediately
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}

func (a *App) Close() error { return a.app.Shutdown() }

// ----- signaling handler -----
func (a *App) handleSignalWS(c *websocket.Conn) {
	// Single writer goroutine to avoid concurrent writes
	writeDone := make(chan struct{})
	var closeOnce sync.Once
	out := make(chan any, 64)
	go func() {
		for {
			select {
			case v := <-out:
				if c.Conn != nil {
					_ = c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				}
				if err := c.WriteJSON(v); err != nil {
					a.log.Warn("ws write failed", slog.String("error", err.Error()))
					closeOnce.Do(func() { close(writeDone); _ = c.Close() })
					return
				}
			case <-writeDone:
				return
			}
		}
	}()
	send := func(v any) error {
		select {
		case <-writeDone:
			return nil
		default:
		}
		select {
		case out <- v:
		case <-writeDone:
		}
		return nil
	}

	// 1) Read and validate join
	first, err := a.readJoinEnvelope(c)
	if err != nil {
		_ = send(ErrorResponse{Error: "invalid message"})
		closeOnce.Do(func() { close(writeDone); _ = c.Close() })
		return
	}
	uid, chID, perms, moved, err := a.authorizeJoin(first)
	if err != nil {
		_ = send(ErrorResponse{Error: err.Error()})
		closeOnce.Do(func() { close(writeDone); _ = c.Close() })
		return
	}
	room := a.rooms.getOrCreate(chID)
	// Enforce channel-level block, unless join token is marked as moved
	if room.isBlocked(uid) && !moved {
		_ = send(ErrorResponse{Error: "blocked"})
		closeOnce.Do(func() { close(writeDone); _ = c.Close() })
		return
	}

	// Allow moved users to bypass connect and grant media publish perms (audio/video)
	if moved {
		perms = permissions.AddPermissions(
			perms,
			permissions.PermVoiceConnect,
			permissions.PermVoiceSpeak,
			permissions.PermVoiceVideo,
		)
	}
	// 2) Prepare peer connection and handlers
	pc, p, err := a.setupPeer(room, uid, perms, send)
	if err != nil {
		_ = send(ErrorResponse{Error: err.Error()})
		closeOnce.Do(func() { close(writeDone); _ = c.Close() })
		return
	}
	p.close = func() { closeOnce.Do(func() { close(writeDone); _ = c.Close() }) }

	// 3) Add to room and ack
	room.addPeer(p)
	a.totalPeers.Add(1)
	_ = send(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCJoin), D: JoinAck{Ok: true}})
	defer func() {
		room.removePeer(uid)
		a.totalPeers.Add(-1)
		// Schedule cleanup with a small grace period to handle brief reconnects
		room.maybeCleanup(a.rooms, 10*time.Second)
		_ = pc.Close()
		closeOnce.Do(func() { close(writeDone); _ = c.Close() })
	}()

	// 4) Attach existing publications
	a.attachExistingPublications(room, p)

	// 5) Enter message loop
	a.messageLoop(c, room, p, pc, perms, send)
}

// ----- discovery -----

func (a *App) heartbeatLoop() {
	url := a.cfg.PublicBaseURL
	if url == "" {
		url = "ws://localhost:3300/signal"
	} else {
		if strings.HasPrefix(strings.ToLower(url), "https://") {
			url = "wss://" + strings.TrimPrefix(url, "https://")
		} else if strings.HasPrefix(strings.ToLower(url), "http://") {
			url = "ws://" + strings.TrimPrefix(url, "http://")
		}
		if !strings.HasSuffix(url, "/signal") {
			url = strings.TrimRight(url, "/") + "/signal"
		}
	}

	client := resty.New().
		SetTimeout(5 * time.Second)
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	type hb struct {
		ID     string `json:"id"`
		Region string `json:"region"`
		URL    string `json:"url"`
		Load   int64  `json:"load"`
	}
	ok := sync.OnceFunc(func() {
		a.log.Info("Service registered and discoverable")
	})
	for {
		payload := hb{ID: a.instID, Region: a.cfg.Region, URL: url, Load: a.totalPeers.Load()}
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("X-Webhook-Token", a.cfg.WebhookToken).
			SetBody(payload).
			SetTimeout(time.Second * 5).
			Post(a.cfg.WebhookURL)
		if err != nil {
			a.log.Error("heartbeat request failed", slog.String("error", err.Error()))
		} else if resp.StatusCode() != 204 {
			a.log.Warn("heartbeat unexpected status", slog.Int("status", resp.StatusCode()), slog.String("body", resp.String()))
		} else if resp.StatusCode() == 204 {
			ok()
		}

		<-tick.C
	}
}
