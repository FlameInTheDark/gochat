package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/FlameInTheDark/gochat/cmd/indexer/config"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/msgsearch"
	nq "github.com/nats-io/nats.go"
)

const indexQueue = "indexer.message"

type App struct {
	logger *slog.Logger

	search *msgsearch.Search
	conn   *nq.Conn

	sub *nq.Subscription
}

func NewApp(logger *slog.Logger) (*App, error) {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		return nil, err
	}

	search, err := msgsearch.NewSearch(cfg.OSAddresses, cfg.OSInsecureSkipVerify, cfg.OSUsername, cfg.OSPassword)
	if err != nil {
		return nil, err
	}

	c, err := nq.Connect(cfg.NatsConnString, nq.Compression(true))
	if err != nil {
		return nil, err
	}

	return &App{
		logger: logger,
		search: search,
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

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err = a.search.IndexMessage(ctx, msgsearch.AddMessage{
			GuildId:   indexMsg.GuildId,
			ChannelId: indexMsg.ChannelId,
			UserId:    indexMsg.UserId,
			MessageId: indexMsg.MessageId,
			Has:       indexMsg.Has,
			Mentions:  indexMsg.Mentions,
			Content:   indexMsg.Content,
		})
		if err != nil {
			a.logger.Error("Error indexing message", slog.String("error", err.Error()))
			return
		}
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
