package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/contrib/websocket"

	"github.com/FlameInTheDark/gochat/cmd/ws/auth"
	"github.com/FlameInTheDark/gochat/cmd/ws/subscriber"
	"github.com/FlameInTheDark/gochat/internal/database/db"
	"github.com/FlameInTheDark/gochat/internal/database/entities/rolecheck"
	"github.com/FlameInTheDark/gochat/internal/database/pgdb"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guild"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/guildchannels"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/member"
	"github.com/FlameInTheDark/gochat/internal/database/pgentities/user"
	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

type helloMessage struct {
	Token string `json:"token"`
}

type helloResponse struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type heartbeatMessage struct {
	// Seconds since connection opened
	LastEventId int64 `json:"e"`
}

type Handler struct {
	user *dto.User
	sub  *subscriber.Subscriber
	g    *guild.Entity
	m    *member.Entity
	u    *user.Entity
	gc   *guildchannels.Entity
	perm *rolecheck.Entity
	jwt  *auth.Auth
	ws   *websocket.Conn

	lastEventId int64
	// Timeout to close connection
	hbTimeout int64
	// Heartbeat timer, close connection if no heartbeat or messages are received in the heartbeat window
	hTimer *time.Timer
	// Timer to receive hello message, close connection if no hello message was received
	initTimer *time.Timer
	// Close connection handler function
	closer func()
	log    *slog.Logger
}

func New(c *db.CQLCon, pg *pgdb.DB, sub *subscriber.Subscriber, ws *websocket.Conn, jwt *auth.Auth, hbTimeout int64, closer func(), logger *slog.Logger) *Handler {
	initTimer := time.AfterFunc(time.Second*5, closer)
	return &Handler{
		sub:  sub,
		g:    guild.New(pg.Conn()),
		m:    member.New(pg.Conn()),
		u:    user.New(pg.Conn()),
		gc:   guildchannels.New(pg.Conn()),
		perm: rolecheck.New(c, pg),
		jwt:  jwt,
		ws:   ws,

		hbTimeout: hbTimeout,
		initTimer: initTimer,
		closer:    closer,
		log:       logger,
	}
}

func (h *Handler) HandleMessage(e mqmsg.Message) {
	if e.Operation != mqmsg.OPCodeHello && h.user == nil {
		return
	}
	switch e.Operation {
	case mqmsg.OPCodeHello:
		h.hello(&e)
	case mqmsg.OPCodeHeartBeat:
		var m heartbeatMessage
		err := json.Unmarshal(e.Data, &m)
		if err != nil {
			h.log.Warn("Error unmarshalling heart beat msg", "error", err)
			return
		}
		if m.LastEventId >= h.lastEventId {
			h.hTimer.Reset(time.Second * time.Duration(h.hbTimeout))
		}
	case mqmsg.OPCodeChannelSubscription:
		var m mqmsg.Subscribe
		err := json.Unmarshal(e.Data, &m)
		if err != nil {
			h.log.Warn("Error unmarshalling channel subscription msg", "error", err)
			return
		}

		// Check if user has access to the channel
		if m.Channel != nil {
			// Get the guild channel to find the guild ID
			gc, err := h.gc.GetGuildByChannel(context.Background(), *m.Channel)
			if err != nil {
				h.log.Warn("Error getting guild channel", "error", err, "channel_id", *m.Channel)
			} else {
				// Check if user has permission to view the channel
				_, _, _, ok, err := h.perm.ChannelPerm(context.Background(), gc.GuildId, gc.ChannelId, h.user.Id, permissions.PermServerViewChannels)
				if err != nil {
					h.log.Warn("Error checking channel permissions", "error", err)
				} else if ok {
					// User has permission, subscribe to the channel
					err := h.sub.Subscribe("channel", fmt.Sprintf("channel.%d", *m.Channel))
					if err != nil {
						h.log.Warn("Error subscribing to channel", "error", err)
					}
				} else {
					h.log.Warn("User does not have permission to view channel", "user_id", h.user.Id, "channel_id", *m.Channel)
				}
			}
		}

		// Check if user has access to the guilds
		for _, guildID := range m.Guilds {
			// Check if user is a member of the guild
			_, ok, err := h.perm.GuildPerm(context.Background(), guildID, h.user.Id, permissions.PermServerViewChannels)
			if err != nil {
				h.log.Warn("Error checking guild permissions", "error", err)
			} else if ok {
				// User has permission, subscribe to the guild
				err := h.sub.Subscribe(fmt.Sprintf("guild.%d", guildID), fmt.Sprintf("guild.%d", guildID))
				if err != nil {
					h.log.Warn("Error subscribing to guild", "error", err)
				}
			} else {
				h.log.Warn("User does not have permission to view guild", "user_id", h.user.Id, "guild_id", guildID)
			}
		}

	default:
		h.log.Warn("Unknown operation", "operation", e.Operation)
	}
}

func (h *Handler) Close() error {
	h.closer()
	return nil
}
