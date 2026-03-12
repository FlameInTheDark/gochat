package main

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/contrib/websocket"

	"github.com/FlameInTheDark/gochat/cmd/ws/handler"
	"github.com/FlameInTheDark/gochat/cmd/ws/subscriber"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/presence"
)

// wsConn implements hub.Conn for a single WebSocket connection.
// It wraps the writer pump's outbound channel so the hub can deliver
// messages without blocking.
type wsConn struct {
	id     string
	out    chan<- outMsg
	userID int64
}

func (w *wsConn) Send(topic string, data []byte) {
	payload := personalizeMessageForRecipient(topic, atomic.LoadInt64(&w.userID), data)
	// Non-blocking: drop the message if the connection's buffer is full.
	select {
	case w.out <- outMsg{kind: 1, data: payload}:
	default:
		// Slow consumer вЂ” message dropped. The ping/heartbeat timeout
		// will eventually evict this connection.
	}
}

func (w *wsConn) SetUserID(userID int64) {
	atomic.StoreInt64(&w.userID, userID)
}

func personalizeMessageForRecipient(topic string, userID int64, data []byte) []byte {
	if !strings.HasPrefix(topic, "channel.") || userID == 0 {
		return cloneWSMessage(data)
	}

	var envelope mqmsg.Message
	if err := json.Unmarshal(data, &envelope); err != nil || envelope.EventType == nil {
		return cloneWSMessage(data)
	}

	switch *envelope.EventType {
	case mqmsg.EventTypeMessageCreate, mqmsg.EventTypeMessageUpdate:
	default:
		return cloneWSMessage(data)
	}

	var payload struct {
		GuildId *int64      `json:"guild_id"`
		Message dto.Message `json:"message"`
	}
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		return cloneWSMessage(data)
	}
	if payload.Message.Nonce == nil || payload.Message.Author.Id == userID {
		return cloneWSMessage(data)
	}

	payload.Message.Nonce = nil
	redactedData, err := json.Marshal(payload)
	if err != nil {
		return cloneWSMessage(data)
	}
	envelope.Data = redactedData

	out, err := json.Marshal(envelope)
	if err != nil {
		return cloneWSMessage(data)
	}
	return out
}

func cloneWSMessage(data []byte) []byte {
	cp := make([]byte, len(data))
	copy(cp, data)
	return cp
}

// outMsg is an internal message sent through the writer pump channel.
type outMsg struct {
	kind int
	data []byte
	v    any
	done chan error
}

func (a *App) wsHandler(c *websocket.Conn) {
	// Track active clients
	if a.wsActive != nil {
		a.wsActive.Inc()
		defer a.wsActive.Dec()
	}
	out := make(chan outMsg, 256)
	writerClosed := make(chan struct{})

	compressMode := strings.EqualFold(c.Query("compress"), "zlib-stream")
	var zbuf bytes.Buffer
	var zw *zlib.Writer
	if compressMode {
		zw, _ = zlib.NewWriterLevel(&zbuf, zlib.BestSpeed)
	}

	go func() {
		for m := range out {
			var err error
			switch m.kind {
			case 1:
				if compressMode {
					if c.Conn != nil {
						_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
					}
					if zw != nil {
						_, _ = zw.Write(m.data)
						_ = zw.Flush()
						chunk := zbuf.Bytes()
						err = c.WriteMessage(websocket.BinaryMessage, chunk)
						zbuf.Reset()
					}
				} else {
					if c.Conn != nil {
						_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
					}
					err = c.WriteMessage(websocket.TextMessage, m.data)
				}
			case 2:
				if compressMode {
					if c.Conn != nil {
						_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
					}
					b, jerr := json.Marshal(m.v)
					if jerr != nil {
						err = jerr
						break
					}
					if zw != nil {
						_, _ = zw.Write(b)
						_ = zw.Flush()
						chunk := zbuf.Bytes()
						err = c.WriteMessage(websocket.BinaryMessage, chunk)
						zbuf.Reset()
					}
				} else {
					if c.Conn != nil {
						_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
					}
					err = c.WriteJSON(m.v)
				}
			case 3:
				err = c.WriteControl(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, string(m.data)),
					time.Now().Add(1*time.Second),
				)
				_ = c.Close()
			case 4:
				if c.Conn != nil {
					_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
				}
				err = c.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
			}
			if m.done != nil {
				m.done <- err
			}
			if m.kind == 3 {
				if compressMode && zw != nil {
					_ = zw.Close()
				}
				return
			}
		}
	}()

	var closed int32
	var closeOnce sync.Once
	errWriterClosed := errors.New("ws writer closed")
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

	conn := &wsConn{id: c.RemoteAddr().String(), out: out}
	subs := subscriber.New(a.hub, conn)
	defer func() {
		cerr := subs.Close()
		if cerr != nil {
			a.log.Error("Error closing subscriber", "error", cerr)
		}
	}()
	pstore := presence.NewStore(a.cache)

	h := handler.New(a.cdb, a.pg, subs, sendJSON, a.jwt, a.cfg.HearthBeatTimeout, func() {
		sendClose("Closed")
	}, a.log, a.natsConn, pstore, a.cache, conn.SetUserID)

	defer func() { _ = h.Close() }()

	pingInterval := time.Second * 15
	if a.cfg.HearthBeatTimeout > 0 {
		half := time.Duration(a.cfg.HearthBeatTimeout/2) * time.Millisecond
		if half < pingInterval {
			pingInterval = half
		}
	}
	stopPing := make(chan struct{})
	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-writerClosed:
				return
			case <-stopPing:
				return
			case <-ticker.C:
				done := make(chan error, 1)
				select {
				case out <- outMsg{kind: 4, done: done}:
					<-done
				case <-writerClosed:
					return
				}
			}
		}
	}()
	defer close(stopPing)

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
