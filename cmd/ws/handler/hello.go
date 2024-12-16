package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

func (h *Handler) hello(msg *mqmsg.Message) {
	var m helloMessage
	err := json.Unmarshal(msg.Data, &m)
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error unmarshalling hello message", "error", err)
		return
	}
	token, err := h.jwt.Parse(m.Token)
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error parsing token", "error", err)
		return
	}
	usr, err := helper.GetUserFromToken(token)
	if err != nil || usr == nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error getting user from token", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.hbTimeout))
	defer cancel()
	dbuser, err := h.u.GetUserById(ctx, usr.Id)
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error getting user", "error", err)
		return
	}

	h.initTimer.Stop()

	h.user = &dto.User{
		Id:   dbuser.Id,
		Name: dbuser.Name,
	}
	err = h.ws.WriteJSON(helloResponse{
		HeartbeatInterval: 15000,
	})
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error sending hello message", "error", err)
		return
	}
	h.hTimer = time.AfterFunc(time.Second*time.Duration(h.hbTimeout), func() {
		err := h.Close()
		if err != nil {
			h.log.Error("Error closing WS connection after timeout", "error", err)
		}
	})
	err = h.sub.Subscribe("user", fmt.Sprintf("user.%d", usr.Id))
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error subscribing to user", "error", err)
		return
	}

	ctx, mcancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.hbTimeout))
	defer mcancel()
	guilds, err := h.m.GetUserGuilds(ctx, usr.Id)
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error getting user's guilds", "error", err)
		return
	}
	for _, g := range guilds {
		err := h.sub.Subscribe(fmt.Sprintf("guild.%d", g.GuildId), fmt.Sprintf("guild.%d", g.UserId))
		if err != nil {
			h.initTimer.Stop()
			h.closer()
			h.log.Error("Error subscribing to guild", "error", err)
			return
		}
	}
}
