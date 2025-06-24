package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	nq "github.com/nats-io/nats.go"
	"github.com/opensearch-project/opensearch-go"

	"github.com/FlameInTheDark/gochat/cmd/indexer/config"
	"github.com/FlameInTheDark/gochat/internal/dto"
)

const indexQueue = "indexer.message"

type App struct {
	logger *slog.Logger

	osc  *opensearch.Client
	conn *nq.Conn

	sub *nq.Subscription
}

func NewApp(logger *slog.Logger) (*App, error) {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		return nil, err
	}

	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.OSInsecureSkipVerify},
		},
		Addresses: cfg.OSAddresses,
		Username:  cfg.OSUsername,
		Password:  cfg.OSPassword,
	})
	if err != nil {
		return nil, err
	}

	res, err := client.Indices.Exists([]string{"messages"})
	if err != nil {
		return nil, fmt.Errorf("failed to check if index exists: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		ctx := context.Background()
		mapping := strings.NewReader(
			`{
  "settings": {
    "index": {
      "number_of_shards": 5,
      "number_of_replicas": 1
    }
  },
  "mappings": {
    "_routing": {
      "required": true
    },
    "properties": {
      "message_id": { "type": "long" },
      "user_id": { "type": "long" },
      "channel_id": { "type": "long" },
      "guild_id": { "type": "long" },
      "mentions": { "type": "long" },
      "has": { "type": "keyword" },
      "content": { "type": "text" }
    }
  }
}`)
		_, err = client.Indices.Create(
			"messages",
			client.Indices.Create.WithBody(mapping),
			client.Indices.Create.WithContext(ctx),
		)
		if err != nil {
			return nil, err
		}
	}

	c, err := nq.Connect(cfg.NatsConnString, nq.Compression(true))
	if err != nil {
		return nil, err
	}

	return &App{
		logger: logger,
		osc:    client,
		conn:   c,
	}, nil
}

func (a *App) Start() error {
	a.logger.Info("Starting service")
	sub, err := a.conn.Subscribe(indexQueue, func(msg *nq.Msg) {
		a.logger.Info("Received message", slog.String("body", string(msg.Data)))
		var indexMsg dto.IndexMessage
		err := json.Unmarshal(msg.Data, &indexMsg)
		if err != nil {
			a.logger.Error(err.Error())
			a.logger.Debug("Error unmarshalling message", slog.String("data", string(msg.Data)))
			return
		}

		index, err := a.osc.Index(
			"messages",
			strings.NewReader(string(msg.Data)),
			a.osc.Index.WithDocumentID(fmt.Sprintf("%d", indexMsg.MessageId)),
			a.osc.Index.WithRouting(fmt.Sprintf("%d", indexMsg.ChannelId)),
		)
		if index.IsError() || err != nil {
			a.logger.Error(err.Error())
			a.logger.Debug("Error indexing message", slog.String("data", string(msg.Data)))
			return
		}
		a.logger.Info("Message sent to index", slog.String("resp", index.String()))
		defer index.Body.Close()
	})
	if err != nil {
		a.logger.Error(err.Error())
		return err
	}

	a.sub = sub
	return nil
}

func (a *App) Close() error {
	a.logger.Debug("Closing app")
	err := a.sub.Unsubscribe()
	if err != nil {
		a.logger.Error("Unable to unsubscribe", slog.String("error", err.Error()))
	}
	a.conn.Close()
	return nil
}
