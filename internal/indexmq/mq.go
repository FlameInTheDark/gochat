package indexmq

import (
	"encoding/json"

	"github.com/nats-io/nats.go"

	"github.com/FlameInTheDark/gochat/internal/dto"
)

const indexerQueue = "indexer.message"

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
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	err = i.conn.Publish(indexerQueue, data)
	if err != nil {
		return err
	}
	return nil
}

func (i *IndexMQ) Close() error {
	i.conn.Close()
	return nil
}
