package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"reflect"
	"time"

	"github.com/gocql/gocql"
	nq "github.com/nats-io/nats.go"

	"github.com/FlameInTheDark/gochat/cmd/embedder/config"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	messageentity "github.com/FlameInTheDark/gochat/internal/database/entities/message"
	"github.com/FlameInTheDark/gochat/internal/database/model"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/embed"
	"github.com/FlameInTheDark/gochat/internal/embedgen"
	"github.com/FlameInTheDark/gochat/internal/embedmq"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	mqnats "github.com/FlameInTheDark/gochat/internal/mq/nats"
)

const embedQueueGroup = "embedder"

type App struct {
	log  *slog.Logger
	db   *db.CQLCon
	msg  messageentity.Message
	gen  *embedgen.Generator
	conn *nq.Conn
	mqt  *mqnats.NatsQueue

	subs []*nq.Subscription
}

func NewApp(logger *slog.Logger) (*App, error) {
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		return nil, err
	}

	logger.Info("Connecting to ScyllaDB")
	database, err := db.NewCQLCon(cfg.ClusterKeyspace, db.NewDBLogger(logger), cfg.Cluster...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connecting to NATS subscriber")
	conn, err := nq.Connect(cfg.NatsConnString, nq.Compression(true))
	if err != nil {
		_ = database.Close()
		return nil, err
	}

	logger.Info("Connecting to NATS publisher")
	transport, err := mqnats.New(cfg.NatsConnString)
	if err != nil {
		conn.Close()
		_ = database.Close()
		return nil, err
	}

	generator := embedgen.New(embedgen.Config{
		AllowPrivateHosts:     cfg.AllowPrivateHosts,
		FetchTimeout:          cfg.FetchTimeout,
		MaxBodyBytes:          cfg.MaxBodyBytes,
		YouTubeOEmbedEndpoint: cfg.YouTubeOEmbedEndpoint,
		YouTubeEmbedBaseURL:   cfg.YouTubeEmbedBaseURL,
	})

	return &App{
		log:  logger,
		db:   database,
		msg:  messageentity.New(database),
		gen:  generator,
		conn: conn,
		mqt:  transport,
	}, nil
}

func (a *App) Start() error {
	a.log.Info("Starting service")
	sub, err := a.conn.QueueSubscribe(embedmq.MakeEmbedSubject, embedQueueGroup, func(msg *nq.Msg) {
		var request embedmq.MakeEmbedRequest
		if err := json.Unmarshal(msg.Data, &request); err != nil {
			a.log.Error("failed to decode embed request", slog.String("error", err.Error()))
			return
		}
		if err := a.processRequest(request); err != nil {
			a.log.Error("failed to process embed request",
				slog.Int64("message_id", request.Message.Id),
				slog.Int64("channel_id", request.Message.ChannelId),
				slog.String("error", err.Error()))
		}
	})
	if err != nil {
		return err
	}
	a.subs = append(a.subs, sub)
	return nil
}

func (a *App) processRequest(request embedmq.MakeEmbedRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	currentMessage, err := a.msg.GetMessage(ctx, request.Message.Id, request.Message.ChannelId)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil
		}
		return err
	}

	manualEmbeds, err := embed.ParseEmbeds(currentMessage.EmbedsJSON)
	if err != nil {
		return err
	}
	existingAutoEmbeds, err := embed.ParseEmbeds(currentMessage.AutoEmbedsJSON)
	if err != nil {
		return err
	}

	flags := model.NormalizeMessageFlags(currentMessage.Flags)
	if model.HasMessageFlag(flags, model.MessageFlagSuppressEmbeds) {
		if len(existingAutoEmbeds) == 0 {
			return nil
		}
		return a.persistAndPublish(ctx, request, currentMessage, manualEmbeds, nil, flags)
	}

	generatedEmbeds, err := a.gen.Generate(ctx, currentMessage.Content, manualEmbeds)
	if err != nil {
		a.log.Warn("embed generation failed",
			slog.Int64("message_id", currentMessage.Id),
			slog.Int64("channel_id", currentMessage.ChannelId),
			slog.String("error", err.Error()))
		return nil
	}
	if reflect.DeepEqual(existingAutoEmbeds, generatedEmbeds) {
		return nil
	}

	return a.persistAndPublish(ctx, request, currentMessage, manualEmbeds, generatedEmbeds, flags)
}

func (a *App) persistAndPublish(ctx context.Context, request embedmq.MakeEmbedRequest, currentMessage model.Message, manualEmbeds, generatedEmbeds []embed.Embed, flags int) error {
	autoEmbedsJSON, err := embed.MarshalEmbeds(generatedEmbeds)
	if err != nil {
		return err
	}
	if err := a.msg.UpdateGeneratedEmbeds(ctx, currentMessage.Id, currentMessage.ChannelId, autoEmbedsJSON); err != nil {
		return err
	}

	author := request.Message.Author
	if author.Id == 0 {
		author = dto.User{Id: currentMessage.UserId}
	}
	updatedMessage := dto.Message{
		Id:          currentMessage.Id,
		ChannelId:   currentMessage.ChannelId,
		Author:      author,
		Content:     currentMessage.Content,
		Attachments: request.Message.Attachments,
		Embeds:      embed.MergeEmbeds(manualEmbeds, generatedEmbeds),
		Flags:       flags,
		Type:        currentMessage.Type,
		UpdatedAt:   currentMessage.EditedAt,
	}
	return a.mqt.SendChannelMessage(currentMessage.ChannelId, &mqmsg.UpdateMessage{
		GuildId: request.GuildId,
		Message: updatedMessage,
	})
}

func (a *App) Close() error {
	for _, sub := range a.subs {
		if err := sub.Unsubscribe(); err != nil {
			a.log.Error("unable to unsubscribe", slog.String("error", err.Error()))
		}
	}
	if a.conn != nil {
		a.conn.Close()
	}
	if a.mqt != nil {
		_ = a.mqt.Close()
	}
	if a.db != nil {
		_ = a.db.Close()
	}
	return nil
}
