package indexmq

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/observability"
)

const (
	indexerQueue       = "indexer.message"
	indexerDeleteQueue = "indexer.delete"
	indexerUpdateQueue = "indexer.update"
)

type IndexMQ struct {
	conn *nats.Conn
}

func NewIndexMQ(conn string) (*IndexMQ, error) {
	c, err := nats.Connect(conn, nats.Compression(true))
	if err != nil {
		return nil, err
	}
	return &IndexMQ{
		conn: c,
	}, nil
}

func (i *IndexMQ) IndexMessage(msg dto.IndexMessage) error {
	return i.IndexMessageContext(context.Background(), msg)
}

func (i *IndexMQ) IndexMessageContext(ctx context.Context, msg dto.IndexMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return i.publish(ctx, indexerQueue, data)
}

func (i *IndexMQ) IndexDeleteMessage(msg dto.IndexDeleteMessage) error {
	return i.IndexDeleteMessageContext(context.Background(), msg)
}

func (i *IndexMQ) IndexDeleteMessageContext(ctx context.Context, msg dto.IndexDeleteMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return i.publish(ctx, indexerDeleteQueue, data)
}

func (i *IndexMQ) UpdateMessage(msg dto.IndexMessage) error {
	return i.UpdateMessageContext(context.Background(), msg)
}

func (i *IndexMQ) UpdateMessageContext(ctx context.Context, msg dto.IndexMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return i.publish(ctx, indexerUpdateQueue, data)
}

func (i *IndexMQ) publish(ctx context.Context, subject string, data []byte) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, finish := observability.StartNATSPublishSpan(ctx, subject)
	defer func() {
		finish(err)
	}()

	headers := observability.InjectNATSHeaders(ctx, nil)
	return i.conn.PublishMsg(&nats.Msg{
		Subject: subject,
		Header:  headers,
		Data:    data,
	})
}

func (i *IndexMQ) Close() error {
	i.conn.Close()
	return nil
}
