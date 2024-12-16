package handler

import (
	"encoding/json"
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
	Since int64 `json:"since"`
}

type Handler struct {
	user *dto.User
	sub  *subscriber.Subscriber
	g    *guild.Entity
	m    *member.Entity
	u    *user.Entity
	jwt  *auth.Auth
	ws   *websocket.Conn

	hbTimeout int
	hTimer    *time.Timer
	initTimer *time.Timer
	closer    func()
	log       *slog.Logger
}

func New(c *db.CQLCon, sub *subscriber.Subscriber, ws *websocket.Conn, jwt *auth.Auth, hbTimeout int, closer func(), logger *slog.Logger) *Handler {
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
		h.hTimer.Reset(time.Second * time.Duration(h.hbTimeout))
	default:
		h.log.Warn("Unknown operation", "operation", e.Operation)
	}
}

func (h *Handler) Close() error {
	var err error
	err = h.sub.Close()
	h.closer()
	return err
}
