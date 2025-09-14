package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/dto"
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
	token, err := h.jwt.ParseAccess(m.Token)
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error parsing token", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.hbTimeout))
	defer cancel()
	dbuser, err := h.u.GetUserById(ctx, token.UserID)
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
	hellomsg, err := mqmsg.BuildEventMessage(&mqmsg.HeartbeatInterval{HeartbeatInterval: h.hbTimeout})
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		return
	}
	err = h.ws.WriteJSON(hellomsg)
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error sending hello message", "error", err)
		return
	}
	h.hTimer = time.AfterFunc(time.Second*time.Duration(h.hbTimeout+2000), func() {
		err := h.Close()
		if err != nil {
			h.log.Error("Error closing WS connection after timeout", "error", err)
		}
	})
	err = h.sub.Subscribe("user", fmt.Sprintf("user.%d", token.UserID))
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error subscribing to user", "error", err)
		return
	}

	ctx, mcancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.hbTimeout))
	defer mcancel()
	guilds, err := h.m.GetUserGuilds(ctx, token.UserID)
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
