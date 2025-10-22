package main

import (
	"fmt"
	"time"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/gofiber/contrib/websocket"
)

const joinHandshakeTimeout = 5 * time.Second

// readJoinEnvelope reads the first join message from the websocket.
func (a *App) readJoinEnvelope(c *websocket.Conn) (rtcJoinEnvelope, error) {
	var first rtcJoinEnvelope
	// Enforce a short deadline for the initial join envelope
	if c.Conn != nil {
		_ = c.Conn.SetReadDeadline(time.Now().Add(joinHandshakeTimeout))
	}
	if err := c.ReadJSON(&first); err != nil {
		return rtcJoinEnvelope{}, err
	}
	// Clear the read deadline once handshake is complete
	if c.Conn != nil {
		_ = c.Conn.SetReadDeadline(time.Time{})
	}
	return first, nil
}

// authorizeJoin validates the join envelope and token; returns uid, channel, perms.
func (a *App) authorizeJoin(first rtcJoinEnvelope) (int64, int64, int64, bool, error) {
	if first.OP != int(mqmsg.OPCodeRTC) || first.T != int(mqmsg.EventTypeRTCJoin) || first.D.Token == "" || first.D.Channel == 0 {
		return 0, 0, 0, false, fmt.Errorf("expected join")
	}
	uid, tokCh, perms, moved, err := a.validateJoinToken(first.D.Token)
	if err != nil || (tokCh != 0 && tokCh != first.D.Channel) {
		return 0, 0, 0, false, fmt.Errorf("unauthorized")
	}
	return uid, first.D.Channel, perms, moved, nil
}
