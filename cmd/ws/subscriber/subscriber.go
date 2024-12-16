package subscriber

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/nats-io/nats.go"
	"log"
)

type Subscriber struct {
	c    *websocket.Conn
	nc   *nats.Conn
	subs map[string]*nats.Subscription
}

func New(c *websocket.Conn, natsCon *nats.Conn) *Subscriber {
	return &Subscriber{
		c:    c,
		nc:   natsCon,
		subs: make(map[string]*nats.Subscription),
	}
}

func (s *Subscriber) Subscribe(key, topic string) error {
	if oldSub, ok := s.subs[key]; ok {
		err := oldSub.Unsubscribe()
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
	if sub, ok := s.subs[key]; ok {
		err := sub.Unsubscribe()
		if err != nil {
			return err
		}
		delete(s.subs, key)
	}
	return nil
}

func (s *Subscriber) Close() error {
	var cerr error
	for _, sub := range s.subs {
		err := sub.Unsubscribe()
		if err != nil {
			cerr = err
			log.Println("Unsubscribe error:", err)
		}
	}
	return cerr
}
