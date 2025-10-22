package main

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	cfgpkg "github.com/FlameInTheDark/gochat/cmd/sfu/config"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type roomManager struct {
	mu    sync.RWMutex
	rooms map[int64]*room
	log   *slog.Logger
	cfg   *cfgpkg.Config
}

func newRoomManager(log *slog.Logger, cfg *cfgpkg.Config) *roomManager {
	return &roomManager{rooms: make(map[int64]*room), log: log, cfg: cfg}
}

func (m *roomManager) getOrCreate(roomID int64) *room {
	if r := m.get(roomID); r != nil {
		return r
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if r := m.rooms[roomID]; r != nil {
		return r
	}
	r := newRoom(roomID, m.log, m.cfg)
	m.rooms[roomID] = r
	return r
}

func (m *roomManager) get(roomID int64) *room {
	m.mu.RLock()
	r := m.rooms[roomID]
	m.mu.RUnlock()
	return r
}

func (m *roomManager) remove(roomID int64) {
	m.mu.Lock()
	delete(m.rooms, roomID)
	m.mu.Unlock()
}

type room struct {
	id           int64
	mu           sync.RWMutex
	peers        map[int64]*peer
	log          *slog.Logger
	cfg          *cfgpkg.Config
	pubs         []*publication
	smuted       map[int64]struct{}
	sdeaf        map[int64]struct{}
	cleanupTimer *time.Timer
	blocked      map[int64]struct{}
}

func newRoom(id int64, log *slog.Logger, cfg *cfgpkg.Config) *room {
	return &room{
		id:      id,
		peers:   make(map[int64]*peer),
		log:     log,
		cfg:     cfg,
		smuted:  make(map[int64]struct{}),
		sdeaf:   make(map[int64]struct{}),
		blocked: make(map[int64]struct{}),
	}
}
func (r *room) addPeer(p *peer) {
	r.mu.Lock()
	r.peers[p.userID] = p
	if r.cleanupTimer != nil {
		r.cleanupTimer.Stop()
		r.cleanupTimer = nil
	}
	r.mu.Unlock()
}
func (r *room) removePeer(uid int64) {
	r.mu.Lock()
	delete(r.peers, uid)
	// Clean any sender references for this peer
	for i := range r.pubs {
		r.pubs[i].mu.Lock()
		delete(r.pubs[i].sends, uid)
		r.pubs[i].mu.Unlock()
	}
	// Clear server-deafen entry for this user
	delete(r.sdeaf, uid)
	r.mu.Unlock()
}
func (r *room) listPeers(except int64) []*peer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*peer, 0, len(r.peers))
	for uid, p := range r.peers {
		if uid != except {
			out = append(out, p)
		}
	}
	return out
}

func (r *room) peerCount() int {
	r.mu.RLock()
	n := len(r.peers)
	r.mu.RUnlock()
	return n
}

func (r *room) maybeCleanup(m *roomManager, delay time.Duration) {
	r.mu.Lock()
	if len(r.peers) == 0 && r.cleanupTimer == nil {
		r.cleanupTimer = time.AfterFunc(delay, func() {
			if r.peerCount() == 0 {
				m.remove(r.id)
			}
			r.mu.Lock()
			r.cleanupTimer = nil
			r.mu.Unlock()
		})
	}
	r.mu.Unlock()
}

type publication struct {
	from  int64
	local *webrtc.TrackLocalStaticRTP
	sends map[int64]*webrtc.RTPSender
	mu    sync.Mutex
}

func (r *room) publishTrack(log *slog.Logger, publisher *peer, tr *webrtc.TrackRemote) error {
	codec := tr.Codec().RTPCodecCapability
	trackID := fmt.Sprintf("pub-%d-%s", publisher.userID, tr.ID())
	streamID := fmt.Sprintf("user-%d", publisher.userID)

	local, err := webrtc.NewTrackLocalStaticRTP(codec, trackID, streamID)
	if err != nil {
		return err
	}

	pub := &publication{from: publisher.userID, local: local, sends: make(map[int64]*webrtc.RTPSender)}
	log.Info("publication created", slog.Int64("from", publisher.userID), slog.String("track", trackID))

	r.mu.Lock()
	r.pubs = append(r.pubs, pub)
	recipients := make([]*peer, 0, len(r.peers))
	for uid, peer := range r.peers {
		if uid == publisher.userID {
			continue
		}
		recipients = append(recipients, peer)
	}
	r.mu.Unlock()

	for _, p := range recipients {
		r.attachPublicationToPeer(pub, p)
	}

	go func() {
		defer r.unpublish(pub)

		buf := make([]byte, 1500)
		pkt := &rtp.Packet{}

		for {
			n, _, readErr := tr.Read(buf)
			if readErr != nil {
				return
			}

			if err := pkt.Unmarshal(buf[:n]); err != nil {
				log.Error("rtp unmarshal failed", slog.String("error", err.Error()))
				continue
			}

			pkt.Extension = false
			pkt.Extensions = nil

			if publisher.IsSelfMuted() || r.isServerMuted(publisher.userID) {
				continue
			}

			if werr := local.WriteRTP(pkt); werr != nil {
				// Do not exit on write errors (e.g., no active senders yet or renegotiation).
				// Keep reading so forwarding resumes once a sender attaches or negotiation completes.
				log.Debug("forward write error (will retry)", slog.Int64("from", publisher.userID), slog.String("err", werr.Error()))
				continue
			}
		}
	}()

	r.dispatchKeyFrame()

	return nil
}

func (r *room) unpublish(pub *publication) {
	r.mu.Lock()
	pub.mu.Lock()
	uids := make([]int64, 0, len(pub.sends))
	for uid := range pub.sends {
		uids = append(uids, uid)
	}
	pub.mu.Unlock()

	for i, candidate := range r.pubs {
		if candidate == pub {
			r.pubs = append(r.pubs[:i], r.pubs[i+1:]...)
			break
		}
	}

	peers := make([]*peer, 0, len(uids))
	for _, uid := range uids {
		if peer := r.peers[uid]; peer != nil {
			peers = append(peers, peer)
		}
	}
	r.mu.Unlock()

	for _, peer := range peers {
		r.detachPublicationFromPeer(pub, peer)
	}
}

func (r *room) dispatchKeyFrame() {
	r.mu.RLock()
	peers := make([]*peer, 0, len(r.peers))
	for _, p := range r.peers {
		peers = append(peers, p)
	}
	r.mu.RUnlock()

	for _, p := range peers {
		for _, receiver := range p.pc.GetReceivers() {
			track := receiver.Track()
			if track == nil || track.Kind() != webrtc.RTPCodecTypeVideo {
				continue
			}
			_ = p.pc.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())},
			})
		}
	}
}

// attachPublicationToPeer adds the publication to the given peer if not muted by the peer
func (r *room) attachPublicationToPeer(pub *publication, p *peer) {
	if p.IsUserMuted(pub.from) || r.isServerDeafened(p.userID) {
		r.log.Debug("skip attach (muted or deafened)", slog.Int64("from", pub.from), slog.Int64("to", p.userID))
		return
	}

	pub.mu.Lock()
	_, already := pub.sends[p.userID]
	pub.mu.Unlock()
	if already {
		return
	}
	sender, err := p.pc.AddTrack(pub.local)
	if err != nil {
		r.log.Error("addtrack to peer failed", slog.String("error", err.Error()))
		return
	}

	pub.mu.Lock()
	pub.sends[p.userID] = sender
	pub.mu.Unlock()

	r.log.Debug("attached existing pub", slog.Int64("from", pub.from), slog.Int64("to", p.userID))
	p.requestNegotiation()
}

// attachPublicationToPeerNoNeg attaches a publication to a peer without triggering negotiation.
// This is useful during initial join to avoid negotiation glare when the client is about to send an offer.
func (r *room) attachPublicationToPeerNoNeg(pub *publication, p *peer) {
	if p.IsUserMuted(pub.from) || r.isServerDeafened(p.userID) {
		return
	}
	pub.mu.Lock()
	_, already := pub.sends[p.userID]
	pub.mu.Unlock()
	if already {
		return
	}
	sender, err := p.pc.AddTrack(pub.local)
	if err != nil {
		r.log.Error("addtrack to peer failed", slog.String("error", err.Error()))
		return
	}
	pub.mu.Lock()
	pub.sends[p.userID] = sender
	pub.mu.Unlock()
}

// hasSendersForPeer reports whether the room has any publications attached to the given peer.
func (r *room) hasSendersForPeer(p *peer) bool {
	r.mu.RLock()
	pubs := make([]*publication, len(r.pubs))
	copy(pubs, r.pubs)
	r.mu.RUnlock()

	for _, pub := range pubs {
		pub.mu.Lock()
		_, ok := pub.sends[p.userID]
		pub.mu.Unlock()
		if ok {
			return true
		}
	}
	return false
}

// detachPublicationFromPeer removes the publication from the given peer if attached
func (r *room) detachPublicationFromPeer(pub *publication, p *peer) {
	pub.mu.Lock()
	sender, ok := pub.sends[p.userID]
	if ok {
		delete(pub.sends, p.userID)
	}
	pub.mu.Unlock()
	if !ok {
		return
	}

	if err := p.pc.RemoveTrack(sender); err != nil {
		r.log.Error("remove track failed", slog.String("error", err.Error()))
	}
	r.log.Debug("detached pub", slog.Int64("from", pub.from), slog.Int64("to", p.userID))
	p.requestNegotiation()
}

func (r *room) isServerMuted(uid int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.smuted[uid]
	return ok
}
func (r *room) isServerDeafened(uid int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.sdeaf[uid]
	return ok
}
func (r *room) setServerMuted(uid int64, v bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v {
		r.smuted[uid] = struct{}{}
	} else {
		delete(r.smuted, uid)
	}
}
func (r *room) setServerDeafened(uid int64, v bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v {
		r.sdeaf[uid] = struct{}{}
	} else {
		delete(r.sdeaf, uid)
	}
}
func (r *room) getPeer(uid int64) *peer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.peers[uid]
}

func (r *room) isBlocked(uid int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.blocked[uid]
	return ok
}

func (r *room) setBlocked(uid int64, v bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if v {
		r.blocked[uid] = struct{}{}
	} else {
		delete(r.blocked, uid)
	}
}
