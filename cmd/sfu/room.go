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

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
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

type publication struct {
	id    string
	from  int64
	kind  webrtc.RTPCodecType
	local *webrtc.TrackLocalStaticRTP
}

type room struct {
	id           int64
	mu           sync.RWMutex
	peers        map[int64]*peer
	log          *slog.Logger
	cfg          *cfgpkg.Config
	publications map[string]*publication
	smuted       map[int64]struct{}
	sdeaf        map[int64]struct{}
	blocked      map[int64]struct{}
	cleanupTimer *time.Timer
}

func newRoom(id int64, log *slog.Logger, cfg *cfgpkg.Config) *room {
	return &room{
		id:           id,
		peers:        make(map[int64]*peer),
		log:          log,
		cfg:          cfg,
		publications: make(map[string]*publication),
		smuted:       make(map[int64]struct{}),
		sdeaf:        make(map[int64]struct{}),
		blocked:      make(map[int64]struct{}),
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
	r.signalPeers()
}

func (r *room) removePeer(uid int64) {
	r.mu.Lock()
	delete(r.peers, uid)
	delete(r.sdeaf, uid)
	r.mu.Unlock()
	r.signalPeers()
}

func (r *room) listPeers(except int64) []*peer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*peer, 0, len(r.peers))
	for uid, p := range r.peers {
		if uid == except {
			continue
		}
		out = append(out, p)
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

func (r *room) publishTrack(log *slog.Logger, publisher *peer, tr *webrtc.TrackRemote) error {
	codec := tr.Codec().RTPCodecCapability
	trackID := fmt.Sprintf("pub-%d-%s", publisher.userID, tr.ID())
	streamID := fmt.Sprintf("user-%d", publisher.userID)

	local, err := webrtc.NewTrackLocalStaticRTP(codec, trackID, streamID)
	if err != nil {
		return err
	}

	pub := &publication{id: trackID, from: publisher.userID, kind: tr.Kind(), local: local}

	r.mu.Lock()
	r.publications[trackID] = pub
	r.mu.Unlock()

	log.Info("publication created", slog.Int64("from", publisher.userID), slog.String("track", trackID))

	go r.forwardPublication(pub, publisher, tr)

	r.signalPeers()

	return nil
}

func (r *room) forwardPublication(pub *publication, publisher *peer, tr *webrtc.TrackRemote) {
	defer r.unpublish(pub.id)

	buf := make([]byte, 1500)
	pkt := &rtp.Packet{}

	for {
		n, _, err := tr.Read(buf)
		if err != nil {
			return
		}
		if err := pkt.Unmarshal(buf[:n]); err != nil {
			r.log.Error("rtp unmarshal failed", slog.String("error", err.Error()))
			continue
		}
		pkt.Extension = false
		pkt.Extensions = nil

		if publisher.IsSelfMuted() || r.isServerMuted(publisher.userID) {
			continue
		}

		if err := pub.local.WriteRTP(pkt); err != nil {
			r.log.Debug("forward write error", slog.Int64("from", publisher.userID), slog.String("error", err.Error()))
			continue
		}
	}
}

func (r *room) unpublish(id string) {
	r.mu.Lock()
	if _, ok := r.publications[id]; !ok {
		r.mu.Unlock()
		return
	}
	delete(r.publications, id)
	r.mu.Unlock()
	r.signalPeers()
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
	pub.mu.Lock()
	pub.sends[p.userID] = sender
	pub.mu.Unlock()
}

func (r *room) signalPeers() {
	r.mu.Lock()
	defer func() {
		r.mu.Unlock()
		r.dispatchKeyFrame()
	}()

	attemptSync := func() bool {
		for uid, p := range r.peers {
			if p.pc.ConnectionState() == webrtc.PeerConnectionStateClosed {
				delete(r.peers, uid)
				return true
			}

			desired := make(map[string]*publication)
			if _, deaf := r.sdeaf[p.userID]; !deaf {
				for id, pub := range r.publications {
					if pub.from == p.userID {
						continue
					}
					if p.IsUserMuted(pub.from) {
						continue
					}
					if _, muted := r.smuted[pub.from]; muted {
						continue
					}
					desired[id] = pub
				}
			}

			existing := map[string]*webrtc.RTPSender{}
			for _, sender := range p.pc.GetSenders() {
				track := sender.Track()
				if track == nil {
					continue
				}
				existing[track.ID()] = sender
				if _, keep := desired[track.ID()]; !keep {
					if err := p.pc.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			if _, deaf := r.sdeaf[p.userID]; deaf {
				if err := r.pushOffer(p); err != nil {
					r.log.Warn("offer failed", slog.Int64("user", p.userID), slog.String("error", err.Error()))
				}
				continue
			}

			for id, pub := range desired {
				if _, ok := existing[id]; ok {
					continue
				}
				if _, err := p.pc.AddTrack(pub.local); err != nil {
					return true
				}
			}

			if err := r.pushOffer(p); err != nil {
				r.log.Warn("offer failed", slog.Int64("user", p.userID), slog.String("error", err.Error()))
			}
		}
		return false
	}

	for attempt := 0; attempt < 25; attempt++ {
		if !attemptSync() {
			return
		}
	}

	go func() {
		time.Sleep(3 * time.Second)
		r.signalPeers()
	}()
}

func (r *room) pushOffer(p *peer) error {
	offer, err := p.pc.CreateOffer(nil)
	if err != nil {
		return err
	}
	if err := p.pc.SetLocalDescription(offer); err != nil {
		return err
	}
	if err := p.send(int(mqmsg.OPCodeRTC), int(mqmsg.EventTypeRTCOffer), rtcOffer{SDP: offer.SDP}); err != nil {
		return err
	}
	return nil
}

func (r *room) isServerMuted(uid int64) bool {
	r.mu.RLock()
	_, ok := r.smuted[uid]
	r.mu.RUnlock()
	return ok
}

func (r *room) isServerDeafened(uid int64) bool {
	r.mu.RLock()
	_, ok := r.sdeaf[uid]
	r.mu.RUnlock()
	return ok
}

func (r *room) setServerMuted(uid int64, v bool) {
	r.mu.Lock()
	if v {
		r.smuted[uid] = struct{}{}
	} else {
		delete(r.smuted, uid)
	}
	r.mu.Unlock()
	r.signalPeers()
}

func (r *room) setServerDeafened(uid int64, v bool) {
	r.mu.Lock()
	if v {
		r.sdeaf[uid] = struct{}{}
	} else {
		delete(r.sdeaf, uid)
	}
	r.mu.Unlock()
	r.signalPeers()
}

func (r *room) getPeer(uid int64) *peer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.peers[uid]
}

func (r *room) isBlocked(uid int64) bool {
	r.mu.RLock()
	_, ok := r.blocked[uid]
	r.mu.RUnlock()
	return ok
}

func (r *room) setBlocked(uid int64, v bool) {
	r.mu.Lock()
	if v {
		r.blocked[uid] = struct{}{}
	} else {
		delete(r.blocked, uid)
	}
	r.mu.Unlock()
}
