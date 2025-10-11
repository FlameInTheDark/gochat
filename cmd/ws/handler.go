package main

import (
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/contrib/websocket"

	"github.com/FlameInTheDark/gochat/cmd/ws/handler"
	"github.com/FlameInTheDark/gochat/cmd/ws/subscriber"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

func (a *App) wsHandler(c *websocket.Conn) {
	type outMsg struct {
		kind int
		data []byte
		v    any
		done chan error
	}
	out := make(chan outMsg, 256)
	writerClosed := make(chan struct{})
	go func() {
		for m := range out {
			var err error
			switch m.kind {
			case 1:
				if c.Conn != nil {
					_ = c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				}
				err = c.WriteMessage(websocket.TextMessage, m.data)
			case 2:
				if c.Conn != nil {
					_ = c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				}
				err = c.WriteJSON(m.v)
			case 3:
				err = c.WriteControl(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, string(m.data)),
					time.Now().Add(1*time.Second),
				)
				_ = c.Close()
			}
			if m.done != nil {
				m.done <- err
			}
			if m.kind == 3 {
				return
			}
		}
	}()

	var closed int32
	var closeOnce sync.Once
	errWriterClosed := errors.New("ws writer closed")
	emitText := func(b []byte) error {
		if atomic.LoadInt32(&closed) == 1 {
			return errWriterClosed
		}
		done := make(chan error, 1)
		select {
		case out <- outMsg{kind: 1, data: b, done: done}:
			return <-done
		case <-writerClosed:
			atomic.StoreInt32(&closed, 1)
			return errWriterClosed
		}
	}
	sendJSON := func(v any) error {
		if atomic.LoadInt32(&closed) == 1 {
			return errWriterClosed
		}
		done := make(chan error, 1)
		select {
		case out <- outMsg{kind: 2, v: v, done: done}:
			return <-done
		case <-writerClosed:
			atomic.StoreInt32(&closed, 1)
			return errWriterClosed
		}
	}
	sendClose := func(reason string) {
		closeOnce.Do(func() {
			atomic.StoreInt32(&closed, 1)
			close(writerClosed)
			done := make(chan error, 1)
			out <- outMsg{kind: 3, data: []byte(reason), done: done}
			<-done
		})
	}

	defer func() {
		if c.Conn != nil {
			err := c.Close()
			if err != nil && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				a.log.Error("Error closing WebSocket", "error", err)
			}
		}
	}()

	subs := subscriber.New(emitText, a.natsConn)
	defer func() {
		sendClose("Closed")
		cerr := subs.Close()
		if cerr != nil {
			a.log.Error("Error closing subscriber", "error", cerr)
		}
	}()

	h := handler.New(a.cdb, a.pg, subs, sendJSON, a.jwt, a.cfg.HearthBeatTimeout, func() {
		sendClose("Closed")
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

		case -1:
			fallthrough
		case websocket.CloseMessage:
			return
		}
	}
}
