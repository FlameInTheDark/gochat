package main

import (
	"encoding/json"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/pion/webrtc/v3"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

// messageLoop reads envelopes and handles heartbeat and RTC messages.
func (a *App) messageLoop(c *websocket.Conn, room *room, p *peer, pc *webrtc.PeerConnection, perms int64, send func(any) error) {
	for {
		var env envelope
		if err := c.ReadJSON(&env); err != nil {
			return
		}
		if env.OP == int(mqmsg.OPCodeHeartBeat) {
			var hb heartbeatData
			_ = json.Unmarshal(env.D, &hb)
			_ = send(OutEnvelope{OP: int(mqmsg.OPCodeHeartBeat), D: HeartbeatReply{Pong: true, ServerTS: time.Now().UnixMilli(), Nonce: hb.Nonce, TS: hb.TS}})
			continue
		}
		if env.OP != int(mqmsg.OPCodeRTC) {
			continue
		}
		switch env.T {
		case int(mqmsg.EventTypeRTCOffer):
			var payload rtcOffer
			if err := json.Unmarshal(env.D, &payload); err != nil {
				continue
			}
			if payload.SDP == "" {
				continue
			}
			a.log.Debug("client offer", slog.Int64("user", p.userID))
			offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: payload.SDP}
			if err := pc.SetRemoteDescription(offer); err != nil {
				_ = send(ErrorResponse{Error: err.Error()})
				continue
			}
			// Do not add transceivers here; answer must mirror the offer's m-line order/count.
			// Pion will create the necessary receivers for offered m-lines during CreateAnswer.
			answer, aerr := pc.CreateAnswer(nil)
			if aerr != nil {
				_ = send(ErrorResponse{Error: aerr.Error()})
				continue
			}
			if err := pc.SetLocalDescription(answer); err != nil {
				_ = send(ErrorResponse{Error: err.Error()})
				continue
			}
			_ = p.send(int(mqmsg.OPCodeRTC), int(mqmsg.EventTypeRTCAnswer), rtcAnswer{SDP: answer.SDP})
			a.log.Debug("server answer sent", slog.Int64("user", p.userID))
			// After answering the client's offer, if we already attached existing publications
			// to this peer during join, negotiate once to deliver them, avoiding glare.
			if room.hasSendersForPeer(p) {
				p.requestNegotiation()
			}
		case int(mqmsg.EventTypeRTCAnswer):
			var payload rtcAnswer
			if err := json.Unmarshal(env.D, &payload); err != nil {
				continue
			}
			if payload.SDP == "" {
				continue
			}
			ans := webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: payload.SDP}
			_ = pc.SetRemoteDescription(ans)
			a.log.Debug("client answer applied", slog.Int64("user", p.userID))
			// Mark negotiation round complete and trigger a follow-up if pending
			p.onAnswerProcessed()
		case int(mqmsg.EventTypeRTCCandidate):
			var payload rtcCandidate
			if err := json.Unmarshal(env.D, &payload); err != nil {
				continue
			}
			if payload.Candidate == "" {
				continue
			}
			_ = pc.AddICECandidate(webrtc.ICECandidateInit{Candidate: payload.Candidate, SDPMid: payload.SDPMid, SDPMLineIndex: payload.SDPMLineIndex})
			a.log.Debug("client candidate", slog.Int64("user", p.userID))
		case int(mqmsg.EventTypeRTCLeave):
			return
		case int(mqmsg.EventTypeRTCMuteSelf):
			var payload rtcMuteSelf
			if err := json.Unmarshal(env.D, &payload); err == nil {
				p.SetSelfMuted(payload.Muted)
			}
		case int(mqmsg.EventTypeRTCMuteUser):
			var payload rtcMuteUser
			if err := json.Unmarshal(env.D, &payload); err == nil {
				uid := payload.User
				muted := payload.Muted
				p.SetUserMuted(uid, muted)
				// Apply to existing publications in this room
				room.mu.RLock()
				pubs := make([]*publication, len(room.pubs))
				copy(pubs, room.pubs)
				room.mu.RUnlock()
				for _, pub := range pubs {
					if pub.from != uid {
						continue
					}
					if muted {
						room.detachPublicationFromPeer(pub, p)
					} else {
						room.attachPublicationToPeer(pub, p)
					}
				}
			}
		case int(mqmsg.EventTypeRTCServerMuteUser):
			// Requires PermVoiceMuteMembers (Administrator overrides)
			if !permissions.CheckPermissions(perms, permissions.PermVoiceMuteMembers) {
				break
			}
			var payload rtcMuteUser
			if err := json.Unmarshal(env.D, &payload); err == nil {
				room.setServerMuted(payload.User, payload.Muted)
			}
		case int(mqmsg.EventTypeRTCServerDeafenUser):
			// Requires PermVoiceDeafenMembers (Administrator overrides)
			if !permissions.CheckPermissions(perms, permissions.PermVoiceDeafenMembers) {
				break
			}
			var payload rtcServerDeafenUser
			if err := json.Unmarshal(env.D, &payload); err == nil {
				// Toggle deafen by detaching/attaching all publications to this peer
				if target := room.getPeer(payload.User); target != nil {
					room.mu.RLock()
					pubs := make([]*publication, len(room.pubs))
					copy(pubs, room.pubs)
					room.mu.RUnlock()
					for _, pub := range pubs {
						if payload.Deafened {
							room.detachPublicationFromPeer(pub, target)
						} else {
							room.attachPublicationToPeer(pub, target)
						}
					}
					room.setServerDeafened(payload.User, payload.Deafened)
				}
			}
		case int(mqmsg.EventTypeRTCServerKickUser):
			// Requires PermVoiceMoveMembers (Administrator overrides)
			if !permissions.CheckPermissions(perms, permissions.PermVoiceMoveMembers) {
				break
			}
			var payload rtcKickUser
			if err := json.Unmarshal(env.D, &payload); err == nil {
				if target := room.getPeer(payload.User); target != nil {
					_ = target.send(int(mqmsg.OPCodeRTC), int(mqmsg.EventTypeRTCServerKickUser), ErrorResponse{Error: "kicked"})
					if target.close != nil {
						target.close()
					}
				}
			}
		case int(mqmsg.EventTypeRTCServerBlockUser):
			// Requires PermVoiceMoveMembers (Administrator overrides)
			if !permissions.CheckPermissions(perms, permissions.PermVoiceMoveMembers) {
				break
			}
			var payload rtcBlockUser
			if err := json.Unmarshal(env.D, &payload); err == nil {
				room.setBlocked(payload.User, payload.Block)
				if payload.Block {
					if target := room.getPeer(payload.User); target != nil {
						_ = target.send(int(mqmsg.OPCodeRTC), int(mqmsg.EventTypeRTCServerBlockUser), ErrorResponse{Error: "blocked"})
						if target.close != nil {
							target.close()
						}
					}
				}
			}
		case int(mqmsg.EventTypeRTCMoved):
			// Requires PermVoiceMoveMembers (Administrator overrides) — admin instructs server to move target
			if !permissions.CheckPermissions(perms, permissions.PermVoiceMoveMembers) {
				break
			}
			var payload struct {
				User    int64 `json:"user"`
				Channel int64 `json:"channel"`
			}
			if err := json.Unmarshal(env.D, &payload); err == nil {
				if target := room.getPeer(payload.User); target != nil {
					_ = target.send(int(mqmsg.OPCodeRTC), int(mqmsg.EventTypeRTCMoved), rtcMoved{Channel: payload.Channel})
					if target.close != nil {
						target.close()
					}
				}
			}
		}
	}
}
