package nats

import (
	"context"
	"encoding/json"
	"fmt"

	nq "github.com/nats-io/nats.go"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
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

func (q *NatsQueue) Ping(ctx context.Context) error {
	if ctx == nil {
		return context.Canceled
	}
	return q.conn.FlushWithContext(ctx)
}

func (q *NatsQueue) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	return q.SendChannelMessageContext(context.Background(), channelId, message)
}

func (q *NatsQueue) SendChannelMessageContext(ctx context.Context, channelId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}
	return q.publish(ctx, fmt.Sprintf("channel.%d", channelId), msg)
}

func (q *NatsQueue) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	return q.SendGuildUpdateContext(context.Background(), guildId, message)
}

func (q *NatsQueue) SendGuildUpdateContext(ctx context.Context, guildId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}
	return q.publish(ctx, fmt.Sprintf("guild.%d", guildId), msg)
}

func (q *NatsQueue) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	return q.SendUserUpdateContext(context.Background(), userId, message)
}

func (q *NatsQueue) SendUserUpdateContext(ctx context.Context, userId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}
	return q.publish(ctx, fmt.Sprintf("user.%d", userId), msg)
}

func (q *NatsQueue) publish(ctx context.Context, subject string, msg mqmsg.Message) error {
	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal message body: %w", err)
	}

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, finish := observability.StartNATSPublishSpan(ctx, subject)
	defer func() {
		finish(err)
	}()

	headers := observability.InjectNATSHeaders(ctx, nil)
	return q.conn.PublishMsg(&nq.Msg{
		Subject: subject,
		Header:  headers,
		Data:    messageBody,
	})
}
