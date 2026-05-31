package searchmq

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/observability"
)

const (
	GuildUpsertSubject = "search.guild.upsert"
	GuildDeleteSubject = "search.guild.delete"
	BotUpsertSubject   = "search.bot.upsert"
	BotDeleteSubject   = "search.bot.delete"
)

type Queue struct {
	conn *nats.Conn
}

func New(conn string) (*Queue, error) {
	c, err := nats.Connect(conn, nats.Compression(true))
	if err != nil {
		return nil, err
	}
	return &Queue{conn: c}, nil
}

func (q *Queue) UpsertGuild(ctx context.Context, msg dto.GuildIndexMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return q.publish(ctx, GuildUpsertSubject, data)
}

func (q *Queue) DeleteGuild(ctx context.Context, msg dto.GuildIndexDeleteMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return q.publish(ctx, GuildDeleteSubject, data)
}

func (q *Queue) UpsertBot(ctx context.Context, msg dto.BotIndexMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return q.publish(ctx, BotUpsertSubject, data)
}

func (q *Queue) DeleteBot(ctx context.Context, msg dto.BotIndexDeleteMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return q.publish(ctx, BotDeleteSubject, data)
}

func (q *Queue) publish(ctx context.Context, subject string, data []byte) (err error) {
	ctx, finish := observability.StartNATSPublishSpan(observability.BackgroundFromContext(ctx), subject)
	defer func() {
		finish(err)
	}()

	headers := observability.InjectNATSHeaders(ctx, nil)
	return q.conn.PublishMsg(&nats.Msg{
		Subject: subject,
		Header:  headers,
		Data:    data,
	})
}

func (q *Queue) Ping(ctx context.Context) error {
	if ctx == nil {
		return context.Canceled
	}
	return q.conn.FlushWithContext(ctx)
}

func (q *Queue) Close() error {
	q.conn.Close()
	return nil
}
