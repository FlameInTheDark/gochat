package subscriber

import (
	"fmt"
	"sync"

	"github.com/FlameInTheDark/gochat/cmd/ws/hub"
)

// Subscriber manages topic subscriptions for a single WebSocket connection
// by delegating to a shared Hub. The Hub ensures that each unique NATS topic
// has at most one NATS subscription per server instance, fanning messages out
// to all local connections in-memory.
type Subscriber struct {
	hub    *hub.Hub
	conn   hub.Conn
	topics map[string]string // key → NATS topic (for unsubscribe tracking)
	mx     sync.Mutex
}

// New creates a subscriber backed by the shared hub for the given connection.
func New(h *hub.Hub, conn hub.Conn) *Subscriber {
	return &Subscriber{
		hub:    h,
		conn:   conn,
		topics: make(map[string]string),
	}
}

// Subscribe registers this connection for the given NATS topic under a logical
// key. If a previous subscription existed for the same key, it is replaced.
func (s *Subscriber) Subscribe(key, topic string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	// Unsubscribe old topic for this key if it differs.
	if old, ok := s.topics[key]; ok {
		if old == topic {
			return nil // already subscribed to the exact same topic
		}
		s.hub.Unregister(s.conn, old)
	}

	if err := s.hub.Register(s.conn, topic); err != nil {
		return fmt.Errorf("subscribe to '%s' error: %w", topic, err)
	}
	s.topics[key] = topic
	return nil
}

// Unsubscribe removes the subscription for the given key.
func (s *Subscriber) Unsubscribe(key string) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	if topic, ok := s.topics[key]; ok {
		s.hub.Unregister(s.conn, topic)
		delete(s.topics, key)
	}
	return nil
}

// Close removes this connection from all topics.
func (s *Subscriber) Close() error {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.hub.UnregisterAll(s.conn)
	s.topics = make(map[string]string)
	return nil
}
