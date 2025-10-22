package main

import (
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/pion/webrtc/v4"
)

type peer struct {
	userID int64
	pc     *webrtc.PeerConnection
	send   func(op int, t int, d any) error
	close  func()
	log    *slog.Logger

	audioMuted       atomic.Int32
	initialOfferSent atomic.Int32

	mu    sync.Mutex
	muted map[int64]struct{}
}

func (p *peer) SetSelfMuted(v bool) {
	if v {
		p.audioMuted.Store(1)
	} else {
		p.audioMuted.Store(0)
	}
}

func (p *peer) IsSelfMuted() bool { return p.audioMuted.Load() == 1 }

func (p *peer) NeedsInitialOffer() bool { return p.initialOfferSent.Load() == 0 }

func (p *peer) MarkInitialOfferSent() { p.initialOfferSent.Store(1) }

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
