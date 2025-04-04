package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/contrib/websocket"

	"github.com/FlameInTheDark/gochat/cmd/ws/auth"
	"github.com/FlameInTheDark/gochat/cmd/ws/subscriber"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/entities/member"
	"github.com/FlameInTheDark/gochat/internal/database/entities/user"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

type helloMessage struct {
	Token string `json:"token"`
}

type helloResponse struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type heartbeatMessage struct {
	// Seconds since connection opened
	LastEventId int64 `json:"e"`
}

type Handler struct {
	user *dto.User
	sub  *subscriber.Subscriber
	g    *guild.Entity
	m    *member.Entity
	u    *user.Entity
	jwt  *auth.Auth
	ws   *websocket.Conn

	lastEventId int64
	hbTimeout   int64
	hTimer      *time.Timer
	initTimer   *time.Timer
	closer      func()
	log         *slog.Logger
}

func New(c *db.CQLCon, sub *subscriber.Subscriber, ws *websocket.Conn, jwt *auth.Auth, hbTimeout int64, closer func(), logger *slog.Logger) *Handler {
	initTimer := time.AfterFunc(time.Second*5, closer)
	return &Handler{
		sub: sub,
		g:   guild.New(c),
		m:   member.New(c),
		u:   user.New(c),
		jwt: jwt,
		ws:  ws,

		hbTimeout: hbTimeout,
		initTimer: initTimer,
		closer:    closer,
		log:       logger,
	}
}

func (h *Handler) HandleMessage(e mqmsg.Message) {
	if e.Operation != mqmsg.OPCodeHello && h.user == nil {
		return
	}
	switch e.Operation {
	case mqmsg.OPCodeHello:
		h.hello(&e)
	case mqmsg.OPCodeHeartBeat:
		var m heartbeatMessage
		err := json.Unmarshal(e.Data, &m)
		if err != nil {
			h.log.Warn("Error unmarshalling heart beat msg", "error", err)
			return
		}
		if m.LastEventId >= h.lastEventId {
			h.hTimer.Reset(time.Second * time.Duration(h.hbTimeout))
		}
	case mqmsg.OPCodeChannelSubscription:
		var m mqmsg.Subscribe
		err := json.Unmarshal(e.Data, &m)
		if err != nil {
			h.log.Warn("Error unmarshalling channel subscription msg", "error", err)
			return
		}
		// TODO: check if user has access
		if m.Channel != nil {
			err := h.sub.Subscribe("channel", fmt.Sprintf("channel.%d", *m.Channel))
			if err != nil {
				h.log.Warn("Error subscribing to channel", "error", err)
			}
		}
		for _, g := range m.Guilds {
			err := h.sub.Subscribe(fmt.Sprintf("guild.%d", g), fmt.Sprintf("guild.%d", g))
			if err != nil {
				h.log.Warn("Error subscribing to channel", "error", err)
			}
		}

	default:
		h.log.Warn("Unknown operation", "operation", e.Operation)
	}
}

func (h *Handler) Close() error {
	h.closer()
	return nil
}
