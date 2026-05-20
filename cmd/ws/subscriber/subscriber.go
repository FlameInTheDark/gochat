package subscriber

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/FlameInTheDark/gochat/cmd/ws/hub"
	"github.com/FlameInTheDark/gochat/internal/observability"
)

// Subscriber manages topic subscriptions for a single WebSocket connection
// by delegating to a shared Hub. The Hub ensures that each unique NATS topic
// has at most one NATS subscription per server instance, fanning messages out
// to all local connections in-memory.
type Subscriber struct {
	hub       *hub.Hub
	conn      hub.Conn
	topics    map[string]string // key → NATS topic (for unsubscribe tracking)
	telemetry *observability.WSTelemetry
	ctx       func() context.Context
	mx        sync.Mutex
}

// New creates a subscriber backed by the shared hub for the given connection.
func New(h *hub.Hub, conn hub.Conn, telemetry *observability.WSTelemetry, ctx func() context.Context) *Subscriber {
	return &Subscriber{
		hub:       h,
		conn:      conn,
		topics:    make(map[string]string),
		telemetry: telemetry,
		ctx:       ctx,
	}
}

// Subscribe registers this connection for the given NATS topic under a logical
// key. If a previous subscription existed for the same key, it is replaced.
func (s *Subscriber) Subscribe(ctx context.Context, key, topic string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	// Unsubscribe old topic for this key if it differs.
	if old, ok := s.topics[key]; ok {
		if old == topic {
			return nil // already subscribed to the exact same topic
		}
		s.hub.Unregister(s.conn, old)
		s.recordSubscriptionDelta(ctx, old, -1)
	}

	if err := s.hub.Register(s.conn, topic); err != nil {
		return fmt.Errorf("subscribe to '%s' error: %w", topic, err)
	}
	s.topics[key] = topic
	s.recordSubscriptionDelta(ctx, topic, 1)
	return nil
}

// Unsubscribe removes the subscription for the given key.
func (s *Subscriber) Unsubscribe(ctx context.Context, key string) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	if topic, ok := s.topics[key]; ok {
		s.hub.Unregister(s.conn, topic)
		delete(s.topics, key)
		s.recordSubscriptionDelta(ctx, topic, -1)
	}
	return nil
}

// Close removes this connection from all topics.
func (s *Subscriber) Close() error {
	s.mx.Lock()
	defer s.mx.Unlock()
	ctx := s.currentContext()
	for _, topic := range s.topics {
		s.recordSubscriptionDelta(ctx, topic, -1)
	}
	s.hub.UnregisterAll(s.conn)
	s.topics = make(map[string]string)
	return nil
}

func (s *Subscriber) currentContext() context.Context {
	ctx := context.Background()
	if s.ctx != nil {
		if current := s.ctx(); current != nil {
			ctx = observability.BackgroundFromContext(current)
		}
	}
	return ctx
}

func (s *Subscriber) recordSubscriptionDelta(ctx context.Context, topic string, delta int64) {
	if s.telemetry == nil || delta == 0 {
		return
	}
	ctx = context.WithoutCancel(ctx)
	s.telemetry.SubscriptionDelta(ctx, subscriptionKind(topic), delta)
}

func subscriptionKind(topic string) string {
	switch {
	case strings.HasPrefix(topic, "guild."):
		return "guild"
	case strings.HasPrefix(topic, "channel."):
		return "channel"
	case strings.HasPrefix(topic, "presence."):
		return "presence"
	case strings.HasPrefix(topic, "user."):
		return "user"
	default:
		return "other"
	}
}
