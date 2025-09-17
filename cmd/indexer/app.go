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

const (
	indexQueue       = "indexer.message"
	indexDeleteQueue = "indexer.delete"
	indexUpdateQueue = "indexer.update"
)

type App struct {
	logger *slog.Logger

	search *msgsearch.Search
	conn   *nq.Conn

	subs []*nq.Subscription
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
		err = a.search.IndexMessage(ctx, msgsearch.Message{
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
	a.subs = append(a.subs, sub)

	delsub, err := a.conn.Subscribe(indexDeleteQueue, func(msg *nq.Msg) {
		a.logger.Info("Received delete message", slog.String("body", string(msg.Data)))
		var indexMsg dto.IndexDeleteMessage
		err := json.Unmarshal(msg.Data, &indexMsg)
		if err != nil {
			a.logger.Error(err.Error())
			a.logger.Debug("Error unmarshalling message", slog.String("data", string(msg.Data)))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err = a.search.DeleteMessage(ctx, msgsearch.DeleteMessage{
			ChannelId: indexMsg.ChannelId,
			MessageId: indexMsg.MessageId,
		})
		if err != nil {
			a.logger.Error("Error deleting message", slog.String("error", err.Error()))
			return
		}
	})
	if err != nil {
		a.logger.Error(err.Error())
		return err
	}
	a.subs = append(a.subs, delsub)

	updsub, err := a.conn.Subscribe(indexUpdateQueue, func(msg *nq.Msg) {
		a.logger.Info("Received update message", slog.String("body", string(msg.Data)))
		var indexMsg dto.IndexMessage
		err := json.Unmarshal(msg.Data, &indexMsg)
		if err != nil {
			a.logger.Error(err.Error())
			a.logger.Debug("Error unmarshalling message", slog.String("data", string(msg.Data)))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err = a.search.UpdateMessage(ctx, msgsearch.Message{
			GuildId:   indexMsg.GuildId,
			ChannelId: indexMsg.ChannelId,
			UserId:    indexMsg.UserId,
			MessageId: indexMsg.MessageId,
			Has:       indexMsg.Has,
			Mentions:  indexMsg.Mentions,
			Content:   indexMsg.Content,
		})
		if err != nil {
			a.logger.Error("Error updating message", slog.String("error", err.Error()))
			return
		}
	})
	if err != nil {
		a.logger.Error(err.Error())
		return err
	}
	a.subs = append(a.subs, updsub)

	return nil
}

func (a *App) Close() error {
	a.logger.Debug("Closing app")
	for _, sub := range a.subs {
		err := sub.Unsubscribe()
		if err != nil {
			a.logger.Error("Unable to unsubscribe", slog.String("error", err.Error()))
		}
	}
	a.conn.Close()
	return nil
}
