package main

import (
	"fmt"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/pion/webrtc/v4"
)

func ensureRecvonly(pc *webrtc.PeerConnection, kind webrtc.RTPCodecType) {
	for _, tr := range pc.GetTransceivers() {
		if tr.Kind() == kind {
			return
		}
	}
	_, _ = pc.AddTransceiverFromKind(kind, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly})
}

func (a *App) setupPeer(room *room, uid int64, perms int64, send func(any) error) (*webrtc.PeerConnection, *peer, error) {
	if !permissions.CheckPermissions(perms, permissions.PermVoiceConnect) {
		return nil, nil, fmt.Errorf("forbidden")
	}

	conf := webrtc.Configuration{}
	if len(a.cfg.STUNServers) > 0 {
		conf.ICEServers = []webrtc.ICEServer{{URLs: a.cfg.STUNServers}}
	}

	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, nil, fmt.Errorf("media init error")
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))
	pc, err := api.NewPeerConnection(conf)
	if err != nil {
		return nil, nil, fmt.Errorf("pc create failed")
	}

	for _, kind := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeAudio, webrtc.RTPCodecTypeVideo} {
		ensureRecvonly(pc, kind)
	}

	p := &peer{
		userID: uid,
		pc:     pc,
		log:    a.log,
		send: func(op int, t int, data any) error {
			return send(OutEnvelope{OP: op, T: t, D: data})
		},
	}

	pc.OnSignalingStateChange(func(state webrtc.SignalingState) {
		if state != webrtc.SignalingStateStable {
			return
		}
		p.resumePendingNegotiation("signaling stable")
	})

	pc.OnTrack(func(tr *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		a.log.Info("inbound track",
			slog.Int64("user", uid),
			slog.String("kind", tr.Kind().String()),
			slog.String("id", tr.ID()),
		)

		switch tr.Kind() {
		case webrtc.RTPCodecTypeAudio:
			if !permissions.CheckPermissions(perms, permissions.PermVoiceSpeak) {
				a.log.Warn("reject audio (no PermVoiceSpeak)", slog.Int64("user", uid))
				return
			}
		case webrtc.RTPCodecTypeVideo:
			if !permissions.CheckPermissions(perms, permissions.PermVoiceVideo) {
				a.log.Warn("reject video (no PermVoiceVideo)", slog.Int64("user", uid))
				return
			}
		}

		if err := room.publishTrack(a.log, p, tr); err != nil {
			a.log.Error("publish failed", slog.String("error", err.Error()))
		}
	})

	pc.OnICECandidate(func(cand *webrtc.ICECandidate) {
		if cand == nil {
			return
		}
		cj := cand.ToJSON()
		_ = p.send(int(mqmsg.OPCodeRTC), int(mqmsg.EventTypeRTCCandidate), rtcCandidate{Candidate: cj.Candidate, SDPMid: cj.SDPMid, SDPMLineIndex: cj.SDPMLineIndex})
	})

	pc.OnICEConnectionStateChange(func(s webrtc.ICEConnectionState) {
		a.log.Info("ice state", slog.Int64("user", uid), slog.String("state", s.String()))
	})

	return pc, p, nil
}
