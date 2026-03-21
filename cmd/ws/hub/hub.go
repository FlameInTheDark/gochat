package hub

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/FlameInTheDark/gochat/internal/observability"
)

// Conn is the interface each WebSocket connection must implement to receive
// messages from the hub. Send must be non-blocking (drop if the connection's
// buffer is full).
type Conn interface {
	// Send delivers a message envelope to the connection.
	// Implementations must be non-blocking: if the outbound buffer is full
	// the message should be dropped (and the connection optionally evicted).
	Send(delivery Delivery)
}

type Delivery struct {
	Topic   string
	Data    []byte
	Headers nats.Header
	Context context.Context
}

// topicEntry tracks a shared NATS subscription and all local connections
// interested in this topic.
type topicEntry struct {
	mu    sync.RWMutex
	sub   *nats.Subscription
	conns map[Conn]struct{}
}

// Hub manages shared NATS subscriptions and fans messages out to local
// WebSocket connections. Instead of each connection creating its own NATS
// subscription, the hub creates **one** NATS subscription per unique topic
// and delivers received messages to every registered local connection in-memory.
type Hub struct {
	nc     *nats.Conn
	mu     sync.RWMutex
	topics map[string]*topicEntry
	log    *slog.Logger
}

// New creates a new Hub backed by the given NATS connection.
func New(nc *nats.Conn, logger *slog.Logger) *Hub {
	if logger == nil {
		logger = observability.Logger()
	}
	return &Hub{
		nc:     nc,
		topics: make(map[string]*topicEntry),
		log:    logger,
	}
}

// Register associates a connection with a topic. The first registration for a
// topic creates a shared NATS subscription; subsequent registrations only add
// the connection to the local fan-out set.
func (h *Hub) Register(conn Conn, topic string) error {
	h.mu.Lock()
	te, ok := h.topics[topic]
	if ok {
		h.mu.Unlock()
		// Topic already subscribed — just add the connection.
		te.mu.Lock()
		te.conns[conn] = struct{}{}
		te.mu.Unlock()
		return nil
	}

	// First subscriber for this topic — create the shared NATS subscription.
	te = &topicEntry{
		conns: map[Conn]struct{}{conn: {}},
	}
	h.topics[topic] = te
	h.mu.Unlock()

	sub, err := h.nc.Subscribe(topic, func(msg *nats.Msg) {
		ctx := observability.ExtractNATSContext(context.Background(), msg)
		ctx, finish := observability.StartNATSConsumeSpan(ctx, topic)
		defer finish(nil)

		te.mu.RLock()
		defer te.mu.RUnlock()
		for c := range te.conns {
			c.Send(Delivery{
				Topic:   topic,
				Data:    msg.Data,
				Headers: msg.Header,
				Context: ctx,
			})
		}
	})
	if err != nil {
		// Roll back.
		h.mu.Lock()
		delete(h.topics, topic)
		h.mu.Unlock()
		return err
	}

	te.mu.Lock()
	te.sub = sub
	te.mu.Unlock()
	return nil
}

// Unregister removes a connection from a topic. When the last connection for a
// topic is removed, the shared NATS subscription is unsubscribed.
func (h *Hub) Unregister(conn Conn, topic string) {
	h.mu.RLock()
	te, ok := h.topics[topic]
	h.mu.RUnlock()
	if !ok {
		return
	}

	te.mu.Lock()
	delete(te.conns, conn)
	empty := len(te.conns) == 0
	te.mu.Unlock()

	if empty {
		h.mu.Lock()
		// Double-check under write lock.
		te.mu.RLock()
		stillEmpty := len(te.conns) == 0
		te.mu.RUnlock()
		if stillEmpty {
			delete(h.topics, topic)
			h.mu.Unlock()
			if te.sub != nil {
				if err := te.sub.Unsubscribe(); err != nil {
					if !strings.Contains(err.Error(), "invalid subscription") {
						h.log.Warn("hub unsubscribe error", slog.String("error", err.Error()), slog.String("topic", topic))
					}
				}
			}
		} else {
			h.mu.Unlock()
		}
	}
}

// UnregisterAll removes a connection from every topic it is subscribed to.
// This should be called when a WebSocket connection closes.
func (h *Hub) UnregisterAll(conn Conn) {
	h.mu.RLock()
	// Snapshot the topic keys so we can iterate without holding the lock
	// during potentially slow unsubscribe operations.
	topics := make([]string, 0, len(h.topics))
	for t, te := range h.topics {
		te.mu.RLock()
		if _, ok := te.conns[conn]; ok {
			topics = append(topics, t)
		}
		te.mu.RUnlock()
	}
	h.mu.RUnlock()

	for _, t := range topics {
		h.Unregister(conn, t)
	}
}

// Stats returns the number of active topics and total connection-topic pairs.
func (h *Hub) Stats() (topics int, pairs int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	topics = len(h.topics)
	for _, te := range h.topics {
		te.mu.RLock()
		pairs += len(te.conns)
		te.mu.RUnlock()
	}
	return
}
