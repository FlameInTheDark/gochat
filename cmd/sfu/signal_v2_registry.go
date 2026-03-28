package main

import (
	"time"

	voicev2 "github.com/FlameInTheDark/gochat/cmd/sfu/signaling/v2"
)

const signalV2ResumeWindow = 30 * time.Second

func (a *App) SendJSON(sessionID string, op int, payload any) error {
	session := a.getSignalV2Session(sessionID)
	if session == nil {
		return nil
	}
	if err := session.writer.SendVoiceGatewayPacket(op, payload); err != nil {
		return err
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	switch op {
	case voicev2.OpDAVEPrepareEpoch:
		if msg, ok := payload.(voicev2.PrepareEpoch); ok {
			session.davePendingProtocol = msg.ProtocolVersion
			session.davePendingEpoch = msg.Epoch
		}
	case voicev2.OpDAVEPrepareTransition:
		if msg, ok := payload.(voicev2.PrepareTransition); ok {
			session.davePendingProtocol = msg.ProtocolVersion
			if msg.ProtocolVersion == 0 {
				session.davePendingEpoch = 0
			}
		}
	case voicev2.OpDAVEExecuteTransition:
		if _, ok := payload.(voicev2.ExecuteTransition); ok {
			session.daveProtocolVersion = session.davePendingProtocol
			session.daveEpoch = session.davePendingEpoch
			if session.state != nil {
				session.state.daveProtocol = session.daveProtocolVersion
				session.state.daveEpoch = session.daveEpoch
			}
		}
	}

	return nil
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
