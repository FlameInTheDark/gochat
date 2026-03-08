package embedmq

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/nats-io/nats.go"
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
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return q.conn.Publish(MakeEmbedSubject, data)
}

func (q *Queue) Close() error {
	q.conn.Close()
	return nil
}
