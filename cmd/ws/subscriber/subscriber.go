package subscriber

import (
	"fmt"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/nats-io/nats.go"
)

type Subscriber struct {
	c    *websocket.Conn
	nc   *nats.Conn
	subs map[string]*nats.Subscription
	mx   sync.Mutex
}

func New(c *websocket.Conn, natsCon *nats.Conn) *Subscriber {
	return &Subscriber{
		c:    c,
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
		err := s.c.WriteMessage(websocket.TextMessage, msg.Data)
		if err != nil {
			log.Println("Write message error:", err)
		}
	})
	if err != nil {
		return err
	}
	s.subs[key] = sub
	return nil
}

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
