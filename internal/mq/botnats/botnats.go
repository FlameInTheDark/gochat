package botnats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FlameInTheDark/gochat/internal/botgateway"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	nq "github.com/nats-io/nats.go"
)

type Queue struct {
	conn       *nq.Conn
	partitions int
}

func New(conn string, partitions int) (*Queue, error) {
	c, err := nq.Connect(conn, nq.Compression(true))
	if err != nil {
		return nil, err
	}
	return &Queue{conn: c, partitions: botgateway.NormalizePartitions(partitions)}, nil
}

func (q *Queue) Close() error {
	q.conn.Close()
	return nil
}

func (q *Queue) Conn() *nq.Conn {
	if q == nil {
		return nil
	}
	return q.conn
}

func (q *Queue) Ping(ctx context.Context) error {
	if ctx == nil {
		return context.Canceled
	}
	return q.conn.FlushWithContext(ctx)
}

func (q *Queue) SendChannelMessage(channelId int64, message mqmsg.EventDataMessage) error {
	return q.SendChannelMessageContext(context.Background(), channelId, message)
}

func (q *Queue) SendChannelMessageContext(ctx context.Context, channelId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}
	guildID, ok := guildIDFromChannelMessage(message)
	if !ok {
		return nil
	}
	return q.publish(ctx, botgateway.GuildChannelSubject(guildID, channelId, q.partitions), msg)
}

func (q *Queue) SendGuildUpdate(guildId int64, message mqmsg.EventDataMessage) error {
	return q.SendGuildUpdateContext(context.Background(), guildId, message)
}

func (q *Queue) SendGuildUpdateContext(ctx context.Context, guildId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}
	return q.publish(ctx, botgateway.GuildSubject(guildId, q.partitions), msg)
}

func (q *Queue) SendUserUpdate(userId int64, message mqmsg.EventDataMessage) error {
	return q.SendUserUpdateContext(context.Background(), userId, message)
}

func (q *Queue) SendUserUpdateContext(ctx context.Context, userId int64, message mqmsg.EventDataMessage) error {
	msg, err := mqmsg.BuildEventMessage(message)
	if err != nil {
		return err
	}
	return q.publish(ctx, botgateway.UserDMSubject(userId, q.partitions), msg)
}

func (q *Queue) publish(ctx context.Context, subject string, msg mqmsg.Message) error {
	messageBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal message body: %w", err)
	}

	ctx, finish := observability.StartNATSPublishSpan(observability.BackgroundFromContext(ctx), subject)
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

func guildIDFromChannelMessage(message mqmsg.EventDataMessage) (int64, bool) {
	switch m := message.(type) {
	case *mqmsg.CreateMessage:
		return optionalGuildID(m.GuildId)
	case *mqmsg.UpdateMessage:
		return optionalGuildID(m.GuildId)
	case *mqmsg.DeleteMessage:
		return optionalGuildID(m.GuildId)
	case *mqmsg.MessageReactionAdd:
		return optionalGuildID(m.GuildId)
	case *mqmsg.MessageReactionRemove:
		return optionalGuildID(m.GuildId)
	case *mqmsg.GuildChannelMessage:
		return optionalGuildID(m.GuildId)
	case *mqmsg.ChannelUserTyping:
		return optionalGuildID(m.GuildId)
	default:
		return 0, false
	}
}

func optionalGuildID(id *int64) (int64, bool) {
	if id == nil || *id == 0 {
		return 0, false
	}
	return *id, true
}
