package embedmq

import (
	"context"
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/nats-io/nats.go"

	"github.com/FlameInTheDark/gochat/internal/observability"
)

const MakeEmbedSubject = "embed.make"

type MakeEmbedRequest struct {
	GuildId *int64      `json:"guild_id,omitempty"`
	Message dto.Message `json:"message"`
}

type Queue struct {
	conn *nats.Conn
}

func New(conn string) (*Queue, error) {
	nc, err := nats.Connect(conn, nats.Compression(true))
	if err != nil {
		return nil, err
	}
	return &Queue{conn: nc}, nil
}

func (q *Queue) MakeEmbed(msg MakeEmbedRequest) error {
	return q.MakeEmbedContext(context.Background(), msg)
}

func (q *Queue) MakeEmbedContext(ctx context.Context, msg MakeEmbedRequest) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, finish := observability.StartNATSPublishSpan(ctx, MakeEmbedSubject)
	defer func() {
		finish(err)
	}()

	headers := observability.InjectNATSHeaders(ctx, nil)
	return q.conn.PublishMsg(&nats.Msg{
		Subject: MakeEmbedSubject,
		Header:  headers,
		Data:    data,
	})
}

func (q *Queue) Close() error {
	q.conn.Close()
	return nil
}
