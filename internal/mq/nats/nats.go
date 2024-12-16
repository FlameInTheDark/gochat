package nats

import (
	"encoding/json"
	"fmt"
	nq "github.com/nats-io/nats.go"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

type NatsQueue struct {
	conn *nq.Conn
}

func New(conn string) (*NatsQueue, error) {
	c, err := nq.Connect(conn, nq.Compression(true))
	if err != nil {
		return nil, err
	}
	return &NatsQueue{conn: c}, nil
}

func (q *NatsQueue) Close() error {
	q.conn.Close()
	return nil
}

func (q *NatsQueue) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}

	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal message body: %w", err)
	}

	err = q.conn.Publish(fmt.Sprintf("channel.%d", channelId), messageBody)
	if err != nil {
		return err
	}
	return nil
}

func (q *NatsQueue) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}
	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal message body: %w", err)
	}
	err = q.conn.Publish(fmt.Sprintf("guild.%d", guildId), messageBody)
	if err != nil {
		return err
	}
	return nil
}

func (q *NatsQueue) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}

	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal message body: %w", err)
	}

	err = q.conn.Publish(fmt.Sprintf("user.%d", userId), messageBody)
	if err != nil {
		return err
	}
	return nil
}
