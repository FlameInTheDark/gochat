package goclient

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Session represents a GoChat bot client session.
type Session struct {
	sync.RWMutex

	// Token is the Authorization header value used for REST and gateway auth.
	Token string
	// APIEndpoint is the base URL for bot REST requests.
	APIEndpoint string
	// GatewayEndpoint is the bot gateway WebSocket URL.
	GatewayEndpoint string
	// UserAgent is sent on REST and gateway connections.
	UserAgent string
	// ShardID is the shard number sent during identify.
	ShardID int
	// ShardCount is the total shard count sent during identify.
	ShardCount int
	// Presence is the optional initial presence sent during identify.
	Presence *Presence
	// ShouldReconnectOnError controls automatic gateway reconnects after read errors.
	ShouldReconnectOnError bool
	// MaxRestRetries controls transient REST retries for 429 and 5xx responses.
	MaxRestRetries int
	// Client is the HTTP client used for REST requests.
	Client httpClient
	// Dialer is the WebSocket dialer used for gateway connections.
	Dialer *websocket.Dialer
	// LastHeartbeatSent is the most recent heartbeat send time.
	LastHeartbeatSent time.Time
	// LastHeartbeatAck is the most recent heartbeat ack time.
	LastHeartbeatAck time.Time

	wsConn    *websocket.Conn
	wsMu      sync.Mutex
	closeCh   chan struct{}
	manual    bool
	sessionID string

	handlersMu    sync.RWMutex
	handlers      []eventHandlerInstance
	nextHandlerID uint64
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// SessionID returns the current gateway session ID, if any.
func (s *Session) SessionID() string {
	s.RLock()
	defer s.RUnlock()
	return s.sessionID
}

// HeartbeatLatency returns the latency between the last heartbeat send and ack.
func (s *Session) HeartbeatLatency() time.Duration {
	s.RLock()
	defer s.RUnlock()
	return s.LastHeartbeatAck.Sub(s.LastHeartbeatSent)
}

// IsOpen reports whether a gateway WebSocket is currently attached.
func (s *Session) IsOpen() bool {
	s.RLock()
	defer s.RUnlock()
	return s.wsConn != nil
}

// Open creates a gateway connection with context.Background.
func (s *Session) Open() error {
	return s.OpenWithContext(context.Background())
}
