package goclient

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// VERSION is the semantic version of this client package.
const VERSION = "0.1.0"

const (
	// DefaultEndpoint is the default GoChat service URL.
	DefaultEndpoint = "https://gochat.anticode.dev"
	// DefaultAPIEndpoint is the default bot REST API base URL.
	DefaultAPIEndpoint = DefaultEndpoint
	// DefaultGatewayEndpoint is the default bot WebSocket gateway URL.
	DefaultGatewayEndpoint = "wss://gochat.anticode.dev/bot/ws"
)

// Option configures a Session during construction.
type Option func(*Session)

// WithEndpoint sets the base GoChat service URL for both REST and gateway use.
func WithEndpoint(endpoint string) Option {
	return func(s *Session) {
		s.APIEndpoint = normalizeEndpoint(endpoint)
		s.GatewayEndpoint = gatewayEndpointFromBase(endpoint)
	}
}

// WithAPIEndpoint sets the bot REST API base URL.
func WithAPIEndpoint(endpoint string) Option {
	return func(s *Session) {
		s.APIEndpoint = normalizeEndpoint(endpoint)
	}
}

// WithGatewayEndpoint sets the bot gateway WebSocket URL.
func WithGatewayEndpoint(endpoint string) Option {
	return func(s *Session) {
		s.GatewayEndpoint = endpoint
	}
}

// WithHTTPClient sets the HTTP client used for REST requests.
func WithHTTPClient(client *http.Client) Option {
	return func(s *Session) {
		if client != nil {
			s.Client = client
		}
	}
}

// WithWebsocketDialer sets the dialer used for gateway connections.
func WithWebsocketDialer(dialer *websocket.Dialer) Option {
	return func(s *Session) {
		if dialer != nil {
			s.Dialer = dialer
		}
	}
}

// WithShard configures the shard identity sent during gateway identify.
func WithShard(id, count int) Option {
	return func(s *Session) {
		s.ShardID = id
		s.ShardCount = count
	}
}

// New creates a new GoChat bot session with the provided token.
//
// The token may be either the raw gcb_ token or an Authorization header value
// already prefixed with "Bot ".
func New(token string, options ...Option) (*Session, error) {
	s := &Session{
		Token:                  normalizeBotToken(token),
		APIEndpoint:            DefaultAPIEndpoint,
		GatewayEndpoint:        DefaultGatewayEndpoint,
		UserAgent:              "GoChatBot (github.com/FlameInTheDark/gochat/clients/bot/goclient, v" + VERSION + ")",
		ShardID:                0,
		ShardCount:             1,
		ShouldReconnectOnError: true,
		MaxRestRetries:         3,
		Client:                 &http.Client{Timeout: 20 * time.Second},
		Dialer:                 websocket.DefaultDialer,
		LastHeartbeatAck:       time.Now().UTC(),
	}
	for _, option := range options {
		option(s)
	}
	return s, nil
}

func normalizeBotToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" || strings.HasPrefix(strings.ToLower(token), "bot ") {
		return token
	}
	return "Bot " + token
}
