package main

import (
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/pion/webrtc/v3"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

type peer struct {
	userID     int64
	pc         *webrtc.PeerConnection
	send       func(op int, t int, d any) error
	close      func()
	log        *slog.Logger
	audioMuted atomic.Int32
	mu         sync.Mutex
	muted      map[int64]struct{}

	// negotiation state
	negoMu      sync.Mutex
	negotiating bool
	negoPending bool
}

// requestNegotiation coalesces renegotiation requests so only one offer is in-flight per peer.
// If a negotiation is already running, marks it as pending to run another round after the answer.
func (p *peer) requestNegotiation() {
	p.negoMu.Lock()
	if p.negotiating {
		p.negoPending = true
		p.negoMu.Unlock()
		if p.log != nil {
			p.log.Debug("nego: request queued", slog.Int64("user", p.userID))
		}
		return
	}
	p.negotiating = true
	p.negoPending = false
	p.negoMu.Unlock()
	if p.log != nil {
		p.log.Debug("nego: start", slog.Int64("user", p.userID))
	}
	go p.doNegotiation()
}

// resumePendingNegotiation starts a pending negotiation if the signaling state is stable and no round is running.
func (p *peer) resumePendingNegotiation(reason string) bool {
	if p.pc.SignalingState() != webrtc.SignalingStateStable {
		return false
	}
	p.negoMu.Lock()
	if !p.negoPending || p.negotiating {
		p.negoMu.Unlock()
		return false
	}
	p.negoPending = false
	p.negotiating = true
	p.negoMu.Unlock()
	if p.log != nil {
		p.log.Debug("nego: resume", slog.Int64("user", p.userID), slog.String("reason", reason))
	}
	go p.doNegotiation()
	return true
}

func (p *peer) doNegotiation() {
	if p.pc.SignalingState() != webrtc.SignalingStateStable {
		p.negoMu.Lock()
		// ensure another attempt runs once we return to stable state
		p.negotiating = false
		p.negoPending = true
		p.negoMu.Unlock()
		if p.log != nil {
			p.log.Debug("nego: skip (not stable)", slog.Int64("user", p.userID))
		}
		return
	}
	offer, err := p.pc.CreateOffer(nil)
	if err != nil {
		p.negoMu.Lock()
		p.negotiating = false
		p.negoMu.Unlock()
		if p.log != nil {
			p.log.Warn("nego: create offer failed", slog.Int64("user", p.userID), slog.String("error", err.Error()))
		}
		return
	}
	if err := p.pc.SetLocalDescription(offer); err != nil {
		p.negoMu.Lock()
		p.negotiating = false
		p.negoPending = true
		p.negoMu.Unlock()
		if p.log != nil {
			p.log.Warn("nego: set local failed", slog.Int64("user", p.userID), slog.String("error", err.Error()))
		}
		return
	}
	if p.log != nil {
		p.log.Debug("nego: offer sent", slog.Int64("user", p.userID))
	}
	_ = p.send(int(mqmsg.OPCodeRTC), int(mqmsg.EventTypeRTCOffer), rtcOffer{SDP: offer.SDP})
}

// onAnswerProcessed should be called after remote answer is applied.
// It starts a new negotiation if there were pending requests.
func (p *peer) onAnswerProcessed() {
	p.negoMu.Lock()
	p.negotiating = false
	pending := p.negoPending
	p.negoMu.Unlock()
	if pending {
		if p.resumePendingNegotiation("follow-up") {
			return
		}
		if p.log != nil {
			p.log.Debug("nego: pending (waiting for stable)", slog.Int64("user", p.userID))
		}
	}
}

func (p *peer) SetSelfMuted(v bool) {
	if v {
		p.audioMuted.Store(1)
	} else {
		p.audioMuted.Store(0)
	}
}

func (p *peer) IsSelfMuted() bool { return p.audioMuted.Load() == 1 }

func (p *peer) SetUserMuted(uid int64, muted bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.muted == nil {
		p.muted = make(map[int64]struct{})
	}
	if muted {
		p.muted[uid] = struct{}{}
	} else {
		delete(p.muted, uid)
	}
}

func (p *peer) IsUserMuted(uid int64) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.muted == nil {
		return false
	}
	_, ok := p.muted[uid]
	return ok
}
