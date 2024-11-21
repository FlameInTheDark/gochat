package rabbit

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
	GuildsExchange   ExchangeChannel = "gochat.guilds"
	UsersExchange    ExchangeChannel = "gochat.users"
)

type Queue struct {
	c *amqp.Connection
}

type Channel struct {
	c       *amqp.Channel
	msgExch ExchangeChannel
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

func (q *Queue) InitChannel() (*Channel, error) {
	ch, err := q.c.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

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
		err = errors.Join(err, ch.Close())
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	err = ch.ExchangeDeclare(
		string(GuildsExchange),
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

	err = ch.ExchangeDeclare(
		string(UsersExchange),
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
		c: ch,
	}, nil
}

func (c *Channel) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}

	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal message body: %w", err)
	}

	return c.c.Publish(
		string(MessagesExchange),
		fmt.Sprintf("channel.%d", channelId),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
}

func (c *Channel) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}

	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal message body: %w", err)
	}

	return c.c.Publish(
		string(GuildsExchange),
		fmt.Sprintf("guild.%d", guildId),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
}

func (c *Channel) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}

	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal message body: %w", err)
	}

	return c.c.Publish(
		string(GuildsExchange),
		fmt.Sprintf("user.%d", userId),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		},
	)
}
