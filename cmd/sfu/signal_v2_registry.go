package main

import (
	"time"
)

const signalV2ResumeWindow = 30 * time.Second

func (a *App) SendJSON(sessionID string, op int, payload any) error {
	session := a.getSignalV2Session(sessionID)
	if session == nil {
		return nil
	}
	return session.writer.SendVoiceGatewayPacket(op, payload)
}

func (a *App) SendBinary(sessionID string, payload []byte) error {
	session := a.getSignalV2Session(sessionID)
	if session == nil {
		return nil
	}
	return session.writer.SendBinaryPacket(payload)
}

func (a *App) registerSignalV2Session(session *signalV2Session) {
	a.signalV2Mu.Lock()
	defer a.signalV2Mu.Unlock()
	a.signalV2Sessions[session.sessionID] = session
}

func (a *App) getSignalV2Session(sessionID string) *signalV2Session {
	a.signalV2Mu.Lock()
	defer a.signalV2Mu.Unlock()
	return a.signalV2Sessions[sessionID]
}

func (a *App) deleteSignalV2Session(sessionID string) {
	a.signalV2Mu.Lock()
	defer a.signalV2Mu.Unlock()
	delete(a.signalV2Sessions, sessionID)
}
