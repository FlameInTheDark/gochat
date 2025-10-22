package main

import (
	"encoding/json"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/pion/webrtc/v4"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

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
			if err := json.Unmarshal(env.D, &payload); err != nil || payload.SDP == "" {
				continue
			}
			a.log.Debug("client offer", slog.Int64("user", p.userID))
			offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: payload.SDP}
			if pc.SignalingState() == webrtc.SignalingStateHaveLocalOffer {
				if err := pc.SetLocalDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeRollback}); err != nil {
					a.log.Warn("rollback failed", slog.Int64("user", p.userID), slog.String("error", err.Error()))
					continue
				}
			}
			if err := pc.SetRemoteDescription(offer); err != nil {
				a.log.Warn("apply offer failed", slog.Int64("user", p.userID), slog.String("error", err.Error()))
				continue
			}
			answer, aerr := pc.CreateAnswer(nil)
			if aerr != nil {
				a.log.Warn("create answer failed", slog.Int64("user", p.userID), slog.String("error", aerr.Error()))
				continue
			}
			if err := pc.SetLocalDescription(answer); err != nil {
				a.log.Warn("set local failed", slog.Int64("user", p.userID), slog.String("error", err.Error()))
				continue
			}
			_ = p.send(int(mqmsg.OPCodeRTC), int(mqmsg.EventTypeRTCAnswer), rtcAnswer{SDP: answer.SDP})
			room.signalPeers()
		case int(mqmsg.EventTypeRTCAnswer):
			var payload rtcAnswer
			if err := json.Unmarshal(env.D, &payload); err != nil || payload.SDP == "" {
				continue
			}
			ans := webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: payload.SDP}
			if err := pc.SetRemoteDescription(ans); err != nil {
				a.log.Warn("apply answer failed", slog.Int64("user", p.userID), slog.String("error", err.Error()))
			} else {
				a.log.Debug("client answer applied", slog.Int64("user", p.userID))
				room.signalPeers()
			}
		case int(mqmsg.EventTypeRTCCandidate):
			var payload rtcCandidate
			if err := json.Unmarshal(env.D, &payload); err != nil || payload.Candidate == "" {
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
				p.SetUserMuted(payload.User, payload.Muted)
				room.signalPeers()
			}
		case int(mqmsg.EventTypeRTCServerMuteUser):
			if !permissions.CheckPermissions(perms, permissions.PermVoiceMuteMembers) {
				break
			}
			var payload rtcMuteUser
			if err := json.Unmarshal(env.D, &payload); err == nil {
				room.setServerMuted(payload.User, payload.Muted)
			}
		case int(mqmsg.EventTypeRTCServerDeafenUser):
			if !permissions.CheckPermissions(perms, permissions.PermVoiceDeafenMembers) {
				break
			}
			var payload rtcServerDeafenUser
			if err := json.Unmarshal(env.D, &payload); err == nil {
				room.setServerDeafened(payload.User, payload.Deafened)
			}
		case int(mqmsg.EventTypeRTCServerKickUser):
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
