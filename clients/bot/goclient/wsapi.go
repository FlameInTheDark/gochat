package goclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Gateway errors.
var (
	ErrWSAlreadyOpen  = errors.New("web socket already opened")
	ErrWSNotFound     = errors.New("no websocket connection exists")
	ErrWSShardBounds  = errors.New("ShardID must be less than ShardCount")
	ErrWSShardInvalid = errors.New("ShardCount must be greater than zero")
)

type identifyPayload struct {
	ShardID    int       `json:"shard_id"`
	ShardCount int       `json:"shard_count"`
	SessionID  string    `json:"session_id,omitempty"`
	Presence   *Presence `json:"presence,omitempty"`
}

// OpenWithContext creates a bot gateway connection.
func (s *Session) OpenWithContext(ctx context.Context) error {
	if ctx == nil {
		return ErrNilContext
	}
	if s.ShardCount <= 0 {
		return ErrWSShardInvalid
	}
	if s.ShardID < 0 || s.ShardID >= s.ShardCount {
		return ErrWSShardBounds
	}

	s.Lock()
	if s.wsConn != nil {
		s.Unlock()
		return ErrWSAlreadyOpen
	}
	s.manual = false
	s.Unlock()

	header := http.Header{}
	header.Set("Authorization", s.Token)
	header.Set("User-Agent", s.UserAgent)
	conn, _, err := s.Dialer.DialContext(ctx, s.GatewayEndpoint, header)
	if err != nil {
		return fmt.Errorf("connect bot gateway: %w", err)
	}

	closeCh := make(chan struct{})
	s.Lock()
	s.wsConn = conn
	s.closeCh = closeCh
	sessionID := s.sessionID
	s.Unlock()

	if err := s.gatewayWrite(conn, GatewayMessage{
		Operation: OPCodeHello,
		Data: mustJSON(identifyPayload{
			ShardID:    s.ShardID,
			ShardCount: s.ShardCount,
			SessionID:  sessionID,
			Presence:   s.Presence,
		}),
	}); err != nil {
		_ = conn.Close()
		s.Lock()
		if s.wsConn == conn {
			s.wsConn = nil
			s.closeCh = nil
		}
		s.Unlock()
		return fmt.Errorf("identify bot gateway: %w", err)
	}

	go s.listen(context.WithoutCancel(ctx), conn, closeCh)
	return nil
}

// Close closes the gateway connection normally.
func (s *Session) Close() error {
	return s.CloseWithCode(websocket.CloseNormalClosure)
}

// CloseWithCode closes the gateway connection with the provided WebSocket code.
func (s *Session) CloseWithCode(closeCode int) error {
	s.Lock()
	conn := s.wsConn
	closeCh := s.closeCh
	s.manual = true
	s.wsConn = nil
	s.closeCh = nil
	s.Unlock()
	if conn == nil {
		return nil
	}
	safeClose(closeCh)
	s.wsMu.Lock()
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, ""))
	s.wsMu.Unlock()
	return conn.Close()
}

// GatewayWriteStruct writes a raw struct directly to the gateway.
func (s *Session) GatewayWriteStruct(data any) error {
	s.RLock()
	conn := s.wsConn
	s.RUnlock()
	if conn == nil {
		return ErrWSNotFound
	}
	return s.gatewayWrite(conn, data)
}

// GatewayWrite writes an operation with a JSON payload to the gateway.
func (s *Session) GatewayWrite(op OPCodeType, payload any) error {
	return s.GatewayWriteStruct(GatewayMessage{Operation: op, Data: mustJSON(payload)})
}

// UpdatePresence updates the bot user's gateway presence.
func (s *Session) UpdatePresence(status, customStatusText string) error {
	return s.GatewayWrite(OPCodePresenceUpdate, PresenceUpdateRequest{
		Status:           status,
		CustomStatusText: customStatusText,
	})
}

func (s *Session) listen(ctx context.Context, conn *websocket.Conn, closeCh chan struct{}) {
	shouldReconnect := atomic.Bool{}
	shouldReconnect.Store(true)
	defer func() {
		s.Lock()
		same := s.wsConn == conn
		manual := s.manual
		if same {
			s.wsConn = nil
			s.closeCh = nil
		}
		s.Unlock()
		if same {
			safeClose(closeCh)
		}
		_ = conn.Close()
		if same && !manual && s.ShouldReconnectOnError && shouldReconnect.Load() {
			go s.reconnect(ctx)
		}
	}()

	for {
		messageType, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}
		if err := s.onGatewayMessage(conn, closeCh, raw); err != nil {
			s.handle(&Event{RawData: raw})
		}
	}
}

func (s *Session) onGatewayMessage(conn *websocket.Conn, closeCh chan struct{}, raw []byte) error {
	var msg GatewayMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return err
	}
	switch msg.Operation {
	case OPCodeHello:
		var hello HeartbeatInterval
		if err := json.Unmarshal(msg.Data, &hello); err != nil {
			return err
		}
		if hello.SessionID != "" {
			s.Lock()
			s.sessionID = hello.SessionID
			s.Unlock()
		}
		go s.heartbeat(conn, closeCh, time.Duration(hello.HeartbeatInterval)*time.Millisecond)
	case OPCodeHeartbeatAck:
		var ack HeartbeatAck
		_ = json.Unmarshal(msg.Data, &ack)
		s.Lock()
		s.LastHeartbeatAck = time.Now().UTC()
		s.Unlock()
	case OPCodePresenceUpdate:
		var update PresenceUpdate
		if err := json.Unmarshal(msg.Data, &update); err != nil {
			return err
		}
		s.handle(&update)
	case OpCodeDispatch, OPCodeRTC:
		return s.dispatch(msg)
	}
	return nil
}

func (s *Session) dispatch(msg GatewayMessage) error {
	event := &Event{Operation: msg.Operation, Type: msg.EventType, RawData: msg.Data}
	if msg.EventType == nil {
		s.handle(event)
		return nil
	}
	constructor := eventConstructors[*msg.EventType]
	if constructor == nil {
		s.handle(event)
		return nil
	}
	payload := constructor()
	if err := json.Unmarshal(msg.Data, payload); err != nil {
		return err
	}
	if ready, ok := payload.(*Ready); ok {
		s.Lock()
		s.sessionID = ready.SessionID
		s.Unlock()
	}
	event.Struct = payload
	s.handle(event)
	s.handle(payload)
	return nil
}

func (s *Session) heartbeat(conn *websocket.Conn, closeCh chan struct{}, interval time.Duration) {
	if interval <= 0 {
		interval = 45 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-closeCh:
			return
		case <-ticker.C:
			s.Lock()
			lastAck := s.LastHeartbeatAck
			s.LastHeartbeatSent = time.Now().UTC()
			s.Unlock()
			if err := s.gatewayWrite(conn, GatewayMessage{Operation: OPCodeHeartBeat, Data: mustJSON(map[string]int64{"client_time": time.Now().UnixMilli()})}); err != nil {
				_ = conn.Close()
				return
			}
			if !lastAck.IsZero() && time.Since(lastAck) > interval*3 {
				_ = conn.Close()
				return
			}
		}
	}
}

func (s *Session) reconnect(ctx context.Context) {
	wait := time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		s.RLock()
		manual := s.manual
		s.RUnlock()
		if manual {
			return
		}
		if err := s.OpenWithContext(ctx); err == nil || errors.Is(err, ErrWSAlreadyOpen) {
			return
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		wait *= 2
		if wait > time.Minute {
			wait = time.Minute
		}
	}
}

func (s *Session) gatewayWrite(conn *websocket.Conn, data any) error {
	s.wsMu.Lock()
	defer s.wsMu.Unlock()
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteJSON(data)
}

func mustJSON(v any) json.RawMessage {
	raw, _ := json.Marshal(v)
	return raw
}

func safeClose(ch chan struct{}) {
	if ch == nil {
		return
	}
	select {
	case <-ch:
	default:
		close(ch)
	}
}
