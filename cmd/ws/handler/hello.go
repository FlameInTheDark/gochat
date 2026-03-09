package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	crand "crypto/rand"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"

	pgmodel "github.com/FlameInTheDark/gochat/internal/database/model"
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

	// --- Parallel DB fetch: user + guilds ---
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(h.hbTimeout))
	defer cancel()

	type userResult struct {
		user pgmodel.User
		err  error
	}
	type guildsResult struct {
		guilds []pgmodel.UserGuild
		err    error
	}
	userCh := make(chan userResult, 1)
	guildsCh := make(chan guildsResult, 1)

	go func() {
		u, e := h.u.GetUserById(ctx, token.UserID)
		userCh <- userResult{u, e}
	}()
	go func() {
		g, e := h.m.GetUserGuilds(ctx, token.UserID)
		guildsCh <- guildsResult{g, e}
	}()

	ur := <-userCh
	gr := <-guildsCh

	if ur.err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error getting user", "error", ur.err)
		return
	}
	if gr.err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error getting user's guilds", "error", gr.err)
		return
	}

	h.initTimer.Stop()

	h.user = &dto.User{
		Id:   ur.user.Id,
		Name: ur.user.Name,
	}

	// Establish or reuse session ID (UUID v4 style). Presence will be set only after client PresenceUpdate.
	if m.HeartbeatSessionID != "" {
		h.sessionID = m.HeartbeatSessionID
	} else {
		h.sessionID = newSessionID()
	}

	// Do not auto-set presence here. Presence is set only after client sends PresenceUpdate.
	hellomsg, err := mqmsg.BuildEventMessage(&mqmsg.HeartbeatInterval{HeartbeatInterval: h.hbTimeout, SessionID: h.sessionID})
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		return
	}
	err = h.sendJSON(hellomsg)
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error sending hello message", "error", err)
		return
	}
	h.hTimer = time.AfterFunc(time.Millisecond*time.Duration(h.hbTimeout+10000), func() {
		h.log.Warn("Heartbeat timeout; closing WS", "user_id", func() any {
			if h.user != nil {
				return h.user.Id
			}
			return int64(0)
		}())
		err := h.Close()
		if err != nil {
			h.log.Error("Error closing WS connection after timeout", "error", err)
		}
	})

	// Subscribe to personal user topic
	err = h.sub.Subscribe("user", fmt.Sprintf("user.%d", token.UserID))
	if err != nil {
		h.initTimer.Stop()
		h.closer()
		h.log.Error("Error subscribing to user", "error", err)
		return
	}

	// Subscribe to all guilds (hub registrations are fast in-memory ops)
	for _, g := range gr.guilds {
		if err := h.sub.Subscribe(fmt.Sprintf("guild.%d", g.GuildId), fmt.Sprintf("guild.%d", g.UserId)); err != nil {
			h.log.Warn("Error subscribing to guild", "error", err, "guild_id", g.GuildId)
		}
	}
}

// newSessionID generates a random UUIDv4-like string without external deps.
func newSessionID() string {
	var b [16]byte
	if n, err := randRead(b[:]); err == nil && n == len(b) {
		b[6] = (b[6] & 0x0f) | 0x40
		b[8] = (b[8] & 0x3f) | 0x80
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]),
			uint16(b[4])<<8|uint16(b[5]),
			uint16(b[6])<<8|uint16(b[7]),
			uint16(b[8])<<8|uint16(b[9]),
			uint64(b[10])<<40|uint64(b[11])<<32|uint64(b[12])<<24|uint64(b[13])<<16|uint64(b[14])<<8|uint64(b[15]),
		)
	}

	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// indirection to avoid importing crypto/rand in multiple places
var randRead = func(p []byte) (int, error) {
	return crand.Read(p)
}
