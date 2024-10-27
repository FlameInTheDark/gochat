package mq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Queue struct {
	c *amqp.Connection
}

type Channel struct {
	c *amqp.Channel
}

func New(host string, port int, username string, password string) (*Queue, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Queue{conn}, nil
}

func (q *Queue) Close() error {
	return q.c.Close()
}

func (c *Channel) SendToQueue(queue string, message interface{}) error {
	return nil
}
