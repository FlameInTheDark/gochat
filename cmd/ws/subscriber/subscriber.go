package subscriber

import (
	"fmt"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	emit func([]byte) error
	nc   *nats.Conn
	subs map[string]*nats.Subscription
	mx   sync.Mutex
}

// New creates a subscriber that forwards incoming NATS messages to the provided
// emitter function (typically a websocket writer pump).
func New(emit func([]byte) error, natsCon *nats.Conn) *Subscriber {
	return &Subscriber{
		emit: emit,
		nc:   natsCon,
		subs: make(map[string]*nats.Subscription),
	}
}

func (s *Subscriber) Subscribe(key, topic string) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	if _, ok := s.subs[key]; ok {
		err := s.subs[key].Unsubscribe()
		if err != nil {
			log.Println("Unsubscribe from old error:", err)
		}
	}
	delete(s.subs, key)
	sub, err := s.nc.Subscribe(topic, func(msg *nats.Msg) {
		if err := s.emit(msg.Data); err != nil {
			log.Println("Emit message error:", err)
		}
	})
	if err != nil {
		return err
	}
	s.subs[key] = sub
	return nil
}

// WriteLock returns a pointer to the write mutex used to serialize
// all writes to the websocket connection. This allows other components
// (like the ws handler) to coordinate writes with subscriber callbacks.
// no WriteLock: writes are serialized by the emitter/writer pump

func (s *Subscriber) Unsubscribe(key string) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	if _, ok := s.subs[key]; ok {
		err := s.subs[key].Unsubscribe()
		if err != nil {
			return fmt.Errorf("Unsubscribe from '%s' error: %s\n", key, err)
		}
		delete(s.subs, key)
	}
	return nil
}

func (s *Subscriber) Close() error {
	s.mx.Lock()
	defer s.mx.Unlock()
	var cerr error
	for i, _ := range s.subs {
		err := s.subs[i].Unsubscribe()
		if err != nil {
			cerr = err
		}
		delete(s.subs, i)
	}
	return cerr
}
