package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/FlameInTheDark/gochat/cmd/indexer/config"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/msgsearch"
	"github.com/FlameInTheDark/gochat/internal/observability"
	nq "github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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

	consumeDuration metric.Float64Histogram
	processSuccess  metric.Int64Counter
	processFailure  metric.Int64Counter
	decodeFailure   metric.Int64Counter
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
		logger:          logger,
		search:          search,
		conn:            c,
		consumeDuration: mustFloat64Histogram("gochat.indexer.consume.duration"),
		processSuccess:  mustInt64Counter("gochat.indexer.consume.success"),
		processFailure:  mustInt64Counter("gochat.indexer.consume.failure"),
		decodeFailure:   mustInt64Counter("gochat.indexer.decode.failure"),
	}, nil
}

func (a *App) Start() error {
	a.logger.Info("Starting service")
	sub, err := a.conn.Subscribe(indexQueue, func(msg *nq.Msg) {
		start := time.Now()
		ctx := observability.ExtractNATSContext(context.Background(), msg)
		ctx, finish := observability.StartNATSConsumeSpan(ctx, msg.Subject)
		var processErr error
		defer func() {
			a.consumeDuration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "index")))
			finish(processErr)
		}()

		var indexMsg dto.IndexMessage
		err := json.Unmarshal(msg.Data, &indexMsg)
		if err != nil {
			processErr = err
			a.decodeFailure.Add(ctx, 1, metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "index")))
			a.logger.ErrorContext(ctx, "failed to decode index message", slog.String("error", err.Error()))
			return
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
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
			processErr = err
			a.processFailure.Add(ctx, 1, metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "index")))
			a.logger.ErrorContext(ctx, "failed to index message", slog.Int64("message_id", indexMsg.MessageId), slog.Int64("channel_id", indexMsg.ChannelId), slog.String("error", err.Error()))
			return
		}
		a.processSuccess.Add(ctx, 1, metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "index")))
	})
	if err != nil {
		a.logger.Error(err.Error())
		return err
	}
	a.subs = append(a.subs, sub)

	delsub, err := a.conn.Subscribe(indexDeleteQueue, func(msg *nq.Msg) {
		start := time.Now()
		ctx := observability.ExtractNATSContext(context.Background(), msg)
		ctx, finish := observability.StartNATSConsumeSpan(ctx, msg.Subject)
		var processErr error
		defer func() {
			a.consumeDuration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "delete")))
			finish(processErr)
		}()

		var indexMsg dto.IndexDeleteMessage
		err := json.Unmarshal(msg.Data, &indexMsg)
		if err != nil {
			processErr = err
			a.decodeFailure.Add(ctx, 1, metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "delete")))
			a.logger.ErrorContext(ctx, "failed to decode index delete message", slog.String("error", err.Error()))
			return
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		err = a.search.DeleteMessage(ctx, msgsearch.DeleteMessage{
			ChannelId: indexMsg.ChannelId,
			MessageId: indexMsg.MessageId,
		})
		if err != nil {
			processErr = err
			a.processFailure.Add(ctx, 1, metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "delete")))
			a.logger.ErrorContext(ctx, "failed to delete indexed message", slog.Int64("message_id", indexMsg.MessageId), slog.Int64("channel_id", indexMsg.ChannelId), slog.String("error", err.Error()))
			return
		}
		a.processSuccess.Add(ctx, 1, metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "delete")))
	})
	if err != nil {
		a.logger.Error(err.Error())
		return err
	}
	a.subs = append(a.subs, delsub)

	updsub, err := a.conn.Subscribe(indexUpdateQueue, func(msg *nq.Msg) {
		start := time.Now()
		ctx := observability.ExtractNATSContext(context.Background(), msg)
		ctx, finish := observability.StartNATSConsumeSpan(ctx, msg.Subject)
		var processErr error
		defer func() {
			a.consumeDuration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "update")))
			finish(processErr)
		}()

		var indexMsg dto.IndexMessage
		err := json.Unmarshal(msg.Data, &indexMsg)
		if err != nil {
			processErr = err
			a.decodeFailure.Add(ctx, 1, metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "update")))
			a.logger.ErrorContext(ctx, "failed to decode index update message", slog.String("error", err.Error()))
			return
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
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
			processErr = err
			a.processFailure.Add(ctx, 1, metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "update")))
			a.logger.ErrorContext(ctx, "failed to update indexed message", slog.Int64("message_id", indexMsg.MessageId), slog.Int64("channel_id", indexMsg.ChannelId), slog.String("error", err.Error()))
			return
		}
		a.processSuccess.Add(ctx, 1, metric.WithAttributes(attribute.String("subject", msg.Subject), attribute.String("operation", "update")))
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

func mustFloat64Histogram(name string) metric.Float64Histogram {
	h, _ := observability.Meter("gochat-indexer").Float64Histogram(name)
	return h
}

func mustInt64Counter(name string) metric.Int64Counter {
	c, _ := observability.Meter("gochat-indexer").Int64Counter(name)
	return c
}
