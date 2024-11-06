package message

import (
	"fmt"
	"github.com/FlameInTheDark/gochat/internal/mq"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/attachment"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channel"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channelroleperm"
	"github.com/FlameInTheDark/gochat/internal/database/entities/channeluserperm"
	"github.com/FlameInTheDark/gochat/internal/database/entities/discriminator"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/entities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/entities/message"
	"github.com/FlameInTheDark/gochat/internal/database/entities/role"
	"github.com/FlameInTheDark/gochat/internal/database/entities/user"
	"github.com/FlameInTheDark/gochat/internal/database/entities/userrole"
	"github.com/FlameInTheDark/gochat/internal/s3"
	"github.com/FlameInTheDark/gochat/internal/server"
)

const entityName = "message"

func (e *entity) Init(router fiber.Router) {
	router.Post("/channel/:channel_id<int>", e.Send)
	router.Post("/channel/:channel_id<int>/attachment", e.Attachment)
}

type entity struct {
	name        string
	uploadLimit int64
	log         *slog.Logger
	storage     *s3.Client
	msgmq       *mq.Channel
	user        *user.Entity
	disc        *discriminator.Entity
	ch          *channel.Entity
	g           *guild.Entity
	gc          *guildchannels.Entity
	msg         *message.Entity
	at          *attachment.Entity
	uperm       *channeluserperm.Entity
	rperm       *channelroleperm.Entity
	role        *role.Entity
	ur          *userrole.Entity
}

func (e *entity) Name() string {
	return e.name
}

func New(dbcon *db.CQLCon, storage *s3.Client, queue *mq.Queue, uploadLimit int64, log *slog.Logger) server.Entity {
	msgmq, err := queue.InitChannel(mq.MessagesExchange)
	if err != nil {
		panic(fmt.Errorf("unable to create mq channel: %w", err))
	}
	return &entity{
		name:        entityName,
		log:         log,
		storage:     storage,
		msgmq:       msgmq,
		uploadLimit: uploadLimit,
		user:        user.New(dbcon),
		disc:        discriminator.New(dbcon),
		ch:          channel.New(dbcon),
		g:           guild.New(dbcon),
		gc:          guildchannels.New(dbcon),
		msg:         message.New(dbcon),
		at:          attachment.New(dbcon),
		uperm:       channeluserperm.New(dbcon),
		rperm:       channelroleperm.New(dbcon),
		role:        role.New(dbcon),
		ur:          userrole.New(dbcon),
	}
}
