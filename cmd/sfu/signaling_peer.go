package main

import (
	"fmt"
	"log/slog"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/pion/webrtc/v3"
)

// ensureRecvonly makes sure there is at least one transceiver of the given kind on pc.
// It avoids adding duplicate recvonly transceivers on subsequent offers.
func ensureRecvonly(pc *webrtc.PeerConnection, kind webrtc.RTPCodecType) {
	for _, tr := range pc.GetTransceivers() {
		if tr.Kind() == kind {
			return
		}
	}
	_, _ = pc.AddTransceiverFromKind(kind, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly})
}

// setupPeer builds the PeerConnection, peer wrapper, and signal handlers.
func (a *App) setupPeer(room *room, uid int64, perms int64, send func(any) error) (*webrtc.PeerConnection, *peer, error) {
	// Enforce connect permission (Administrator overrides). Bypass if token is marked as moved.
	// Note: moved bypass is enforced before calling this; here we keep the check for clarity.
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
	p := &peer{userID: uid, pc: pc, log: a.log, send: func(op int, t int, data any) error { return send(OutEnvelope{OP: op, T: t, D: data}) }}
	// Helpful: if the remote client can't handle server-initiated offers reliably,
	// this will still be driven by explicit requestNegotiation() calls.
	pc.OnNegotiationNeeded(func() {
		// Coalesce via requestNegotiation state machine
		p.requestNegotiation()
	})

	pc.OnTrack(func(tr *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		a.log.Info("inbound track",
			slog.Int64("user", uid),
			slog.String("kind", tr.Kind().String()),
			slog.String("id", tr.ID()),
		)
		// Enforce media permissions per kind
		if tr.Kind() == webrtc.RTPCodecTypeAudio {
			if !permissions.CheckPermissions(perms, permissions.PermVoiceSpeak) {
				a.log.Warn("reject audio (no PermVoiceSpeak)", slog.Int64("user", uid))
				return
			}
		}
		if tr.Kind() == webrtc.RTPCodecTypeVideo {
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

// attachExistingPublications attaches current room publications to the new peer.
func (a *App) attachExistingPublications(room *room, p *peer) {
	room.mu.RLock()
	pubs := make([]*publication, len(room.pubs))
	copy(pubs, room.pubs)
	room.mu.RUnlock()
	for _, pub := range pubs {
		// Attach without immediate negotiation to avoid glare with the client's initial offer
		room.attachPublicationToPeerNoNeg(pub, p)
	}
}
