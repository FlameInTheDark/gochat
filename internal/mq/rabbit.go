package mq

import (
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

type ExchangeChannel string

const (
	MessagesExchange ExchangeChannel = "gochat.messages"
)

type Queue struct {
	c *amqp.Connection
}

type Channel struct {
	exch ExchangeChannel
	c    *amqp.Channel
}

func (c *Channel) Close() error {
	return c.c.Close()
}

func New(host string, port int, username string, password string) (*Queue, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Queue{c: conn}, nil
}

func (q *Queue) Close() error {
	return q.c.Close()
}

func (q *Queue) InitChannel(channel ExchangeChannel) (*Channel, error) {
	ch, err := q.c.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		string(channel),
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		err = errors.Join(err, ch.Close())
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &Channel{
		exch: channel,
		c:    ch,
	}, nil
}

func (c *Channel) PublishMessage(channel int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}

	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal message body: %w", err)
	}

	return c.c.Publish(
		string(c.exch),
		fmt.Sprintf("channel.%d", channel),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
}

func (q *Queue) PublishMessage(channel int64, message mqmsg.EventDataMessage) error {
	ch, err := q.c.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		string(MessagesExchange),
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	messageBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("unable to marshal message body: %w", err)
	}

	return ch.Publish(
		string(MessagesExchange),
		fmt.Sprintf("channel.%d", channel),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
}
