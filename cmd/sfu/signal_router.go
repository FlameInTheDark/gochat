package main

import "github.com/gofiber/contrib/websocket"

func (a *App) handleSignalWS(c *websocket.Conn) {
	version, _ := parseSignalProtocolVersion(c.Query("v"))
	if version == signalProtocolVersion2 {
		a.handleSignalWSV2(c)
		return
	}
	a.handleSignalWSV1(c)
}
