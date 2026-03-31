package main

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/contrib/websocket"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/FlameInTheDark/gochat/cmd/ws/handler"
	"github.com/FlameInTheDark/gochat/cmd/ws/hub"
	"github.com/FlameInTheDark/gochat/cmd/ws/subscriber"
	cachei "github.com/FlameInTheDark/gochat/internal/cache"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/presence"
	"github.com/FlameInTheDark/gochat/internal/wsmsg"
)

// wsConn implements hub.Conn for a single WebSocket connection.
// It wraps the writer pump's outbound channel so the hub can deliver
// messages without blocking.
type wsConn struct {
	id        string
	out       chan<- outMsg
	userID    int64
	telemetry *observability.WSTelemetry
	cache     cachei.Cache
	close     func(reason string)
}

func (w *wsConn) Send(delivery hub.Delivery) {
	if isAuthRevokedEvent(delivery.Data) {
		if w.close != nil {
			w.close("Auth revoked")
		}
		return
	}
	payload := wsmsg.PersonalizeMessageForRecipientWithCache(w.cache, delivery.Topic, atomic.LoadInt64(&w.userID), delivery.Data)
	// Non-blocking: drop the message if the connection's buffer is full.
	select {
	case w.out <- outMsg{kind: 1, data: payload, topic: delivery.Topic, ctx: delivery.Context}:
	default:
		if w.telemetry != nil {
			w.telemetry.OutboundDrop(delivery.Context, delivery.Topic)
		}
		// Slow consumer - message dropped. The ping/heartbeat timeout
		// will eventually evict this connection.
	}
}

func (w *wsConn) SetUserID(userID int64) {
	atomic.StoreInt64(&w.userID, userID)
}

func isAuthRevokedEvent(data []byte) bool {
	var envelope mqmsg.Message
	if err := json.Unmarshal(data, &envelope); err != nil || envelope.EventType == nil {
		return false
	}
	return *envelope.EventType == mqmsg.EventTypeUserAuthRevoked
}

// outMsg is an internal message sent through the writer pump channel.
type outMsg struct {
	kind  int
	data  []byte
	v     any
	done  chan error
	topic string
	ctx   context.Context
}

func (a *App) wsHandler(c *websocket.Conn) {
	requestCtx := context.Background()
	if raw := c.Locals("request_context"); raw != nil {
		if current, ok := raw.(context.Context); ok && current != nil {
			requestCtx = observability.BackgroundFromContext(current)
		}
	}
	compressMode := strings.EqualFold(c.Query("compress"), "zlib-stream")
	connCtx, connSpan := observability.Tracer("gochat/ws").Start(requestCtx, "ws.connection")
	connSpan.SetAttributes(attribute.Bool("ws.compression", compressMode))
	connLog := helper.WithContext(a.log, connCtx)
	connectionStarted := time.Now()
	a.wsm.ConnectionOpened(connCtx, attribute.Bool("ws.compression", compressMode))
	defer func() {
		a.wsm.ConnectionClosed(connCtx, connectionStarted)
		connSpan.End()
	}()

	out := make(chan outMsg, 256)
	writerClosed := make(chan struct{})

	var zbuf bytes.Buffer
	var zw *zlib.Writer
	if compressMode {
		zw, _ = zlib.NewWriterLevel(&zbuf, zlib.BestSpeed)
	}

	go func() {
		for m := range out {
			writeCtx := m.ctx
			if writeCtx == nil {
				writeCtx = connCtx
			}
			writeCtx = observability.BackgroundFromContext(writeCtx)
			var err error
			var span trace.Span
			if m.kind == 1 || m.kind == 2 {
				writeCtx, span = observability.Tracer("gochat/ws").Start(writeCtx, "ws.write")
				span.SetAttributes(
					attribute.String("topic", m.topic),
					attribute.Int("ws.message.kind", m.kind),
				)
			}
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
			if span != nil {
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					a.wsm.MessageOut(writeCtx, m.topic, "error")
				} else {
					span.SetStatus(codes.Ok, "sent")
					a.wsm.MessageOut(writeCtx, m.topic, "sent")
				}
				span.End()
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
		case out <- outMsg{kind: 2, v: v, done: done, topic: "connection", ctx: connCtx}:
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
			out <- outMsg{kind: 3, data: []byte(reason), done: done, topic: "connection.close", ctx: connCtx}
			<-done
		})
	}

	defer func() {
		if c.Conn != nil {
			err := c.Close()
			if err != nil && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				connSpan.RecordError(err)
				connLog.Error("Error closing WebSocket", "error", err)
			}
		}
	}()

	conn := &wsConn{id: c.RemoteAddr().String(), out: out, telemetry: a.wsm, cache: a.cache, close: sendClose}
	subs := subscriber.New(a.hub, conn, a.wsm, func() context.Context { return connCtx })
	defer func() {
		cerr := subs.Close()
		if cerr != nil {
			connLog.Error("Error closing subscriber", "error", cerr)
		}
	}()
	pstore := presence.NewStore(a.cache)

	h := handler.New(a.cdb, a.pg, subs, sendJSON, a.jwt, a.cfg.HearthBeatTimeout, func() {
		sendClose("Closed")
	}, a.log, a.natsConn, pstore, a.cache, conn.SetUserID, connCtx, a.wsm)

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
				case out <- outMsg{kind: 4, done: done, topic: "connection.ping", ctx: connCtx}:
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
			if isExpectedWSReadError(err) {
				return
			}
			connSpan.RecordError(err)
			connLog.Error("Read WS message error", "error", err)
			return
		}

		switch mt {
		case websocket.TextMessage:
			var message mqmsg.Message
			if err := json.Unmarshal(msg, &message); err != nil {
				connLog.Error("Error unmarshalling message", "error", err)
				continue
			}
			h.HandleMessage(message)

		case websocket.BinaryMessage:
			connLog.Info("Received binary message", "length", len(msg))

		case -1:
			fallthrough
		case websocket.CloseMessage:
			return
		}
	}
}

func isExpectedWSReadError(err error) bool {
	if err == nil {
		return false
	}

	if websocket.IsCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseProtocolError,
		websocket.CloseNoStatusReceived,
		websocket.CloseGoingAway,
		websocket.CloseAbnormalClosure,
	) {
		return true
	}

	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, net.ErrClosed)
}
