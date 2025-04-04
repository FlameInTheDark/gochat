package main

import (
	"encoding/json"
	"time"

	"github.com/gofiber/contrib/websocket"

	"github.com/FlameInTheDark/gochat/cmd/ws/handler"
	"github.com/FlameInTheDark/gochat/cmd/ws/subscriber"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

func (a *App) wsHandler(c *websocket.Conn) {
	defer func() {
		if c.Conn != nil {
			err := c.Close()
			if err != nil && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				a.log.Error("Error closing WebSocket", "error", err)
			}
		}
	}()

	subs := subscriber.New(c, a.natsConn)
	defer func() {
		cerr := subs.Close()
		if cerr != nil {
			a.log.Error("Error closing subscriber", "error", cerr)
		}
	}()

	h := handler.New(a.cdb, subs, c, a.jwt, a.cfg.HearthBeatTimeout, func() {
		if c.Conn != nil {
			err := c.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Closed"),
				time.Now().Add(1*time.Second),
			)
			if err != nil {
				a.log.Error("Error writing close WebSocket message", "error", err)
			}
			cerr := c.Close()
			if cerr != nil {
				a.log.Error("Error closing WebSocket during auth timeout", "error", err)
			}
		}
	}, a.log)

	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseProtocolError,
				websocket.CloseNoStatusReceived,
				websocket.CloseGoingAway,
			) {
				return
			}
			a.log.Error("Read WS message error", "error", err)
			continue
		}
		switch mt {
		case websocket.TextMessage:
			var message mqmsg.Message
			if err := json.Unmarshal(msg, &message); err != nil {
				a.log.Error("Error unmarshalling message", "error", err)
				continue
			}
			h.HandleMessage(message)

		case websocket.BinaryMessage:
			a.log.Info("Received binary message", "length", len(msg))

		case -1: // in case of dropped connection
			fallthrough
		case websocket.CloseMessage:
			return
		}
	}
}
