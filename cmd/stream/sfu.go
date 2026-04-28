package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"go.opentelemetry.io/otel/attribute"
	"resty.dev/v3"

	voicev2 "github.com/FlameInTheDark/gochat/cmd/stream/signaling/v2"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/observability"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

const periodicKeyFrameInterval = 10 * time.Second

// ---------------------------------------------------------------------------
// threadSafeWriter wraps a websocket.Conn with a mutex for concurrent writes.
// ---------------------------------------------------------------------------

type threadSafeWriter struct {
	conn   *websocket.Conn
	mu     sync.Mutex
	closed atomic.Bool
}

func (t *threadSafeWriter) WriteJSON(v any) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed.Load() {
		return fmt.Errorf("websocket closed")
	}
	if t.conn == nil {
		return fmt.Errorf("websocket closed")
	}
	return t.conn.WriteJSON(v)
}

func (t *threadSafeWriter) WriteMessage(messageType int, payload []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed.Load() {
		return fmt.Errorf("websocket closed")
	}
	if t.conn == nil {
		return fmt.Errorf("websocket closed")
	}
	return t.conn.WriteMessage(messageType, payload)
}

func (t *threadSafeWriter) SendEnvelope(env OutEnvelope) error {
	return t.WriteJSON(env)
}

func (t *threadSafeWriter) SendVoiceGatewayPacket(op int, payload any) error {
	return t.WriteJSON(voicev2.Packet{Op: op, D: payload})
}

func (t *threadSafeWriter) SendBinaryPacket(payload []byte) error {
	return t.WriteMessage(websocket.BinaryMessage, payload)
}

func (t *threadSafeWriter) SendRTCOffer(desc webrtc.SessionDescription) error {
	payload := rtcOffer{SDP: desc.SDP, Type: desc.Type.String()}
	env := OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCOffer), D: payload}
	return t.SendEnvelope(env)
}

func (t *threadSafeWriter) SendRTCCandidate(c *webrtc.ICECandidate) error {
	if c == nil {
		return nil
	}
	cand := c.ToJSON()
	payload := rtcCandidate{Candidate: cand.Candidate, SDPMid: cand.SDPMid, SDPMLineIndex: cand.SDPMLineIndex}
	env := OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCCandidate), D: payload}
	return t.SendEnvelope(env)
}

func (t *threadSafeWriter) SendClose(code int, text string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed.Load() {
		return fmt.Errorf("websocket closed")
	}
	if t.conn == nil {
		return fmt.Errorf("websocket closed")
	}
	return t.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, text), time.Now().Add(time.Second))
}

func (t *threadSafeWriter) ReplaceConn(conn *websocket.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed.Load() {
		return
	}
	t.conn = conn
}

func (t *threadSafeWriter) Detach() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.conn = nil
}

// Close marks the writer as closed. Subsequent writes return immediately.
func (t *threadSafeWriter) Close() {
	t.Detach()
	t.closed.Store(true)
}

// ---------------------------------------------------------------------------
// peerConnectionState holds per-peer state within a channel.
// ---------------------------------------------------------------------------

type peerConnectionState struct {
	peerConnection  *webrtc.PeerConnection
	websocket       *threadSafeWriter
	rtcConnectionID string
	mediaSessionID  string
	metaMu          sync.RWMutex
	userID          int64
	perms           int64 // voice permission bitmask from JWT
	offeredRevision uint64
	appliedRevision uint64
	signalVersion   int
	daveProtocol    int
	daveEpoch       uint64
	serverMuted     bool // server-wide mute (admin action)
	serverDeafened  bool // server-wide deafen (admin action)
	negotiated      bool // true after the peer has answered at least one offer
	forceOffer      bool // explicit renegotiation request for this peer
}

func (p *peerConnectionState) sendDescription(desc webrtc.SessionDescription) error {
	if p.signalVersion == signalProtocolVersion2 {
		rtcConnectionID, mediaSessionID, daveProtocol, daveEpoch := p.sessionDescriptionMetadata()
		audioCodec, videoCodec := detectNegotiatedCodecs(desc.SDP)
		return p.websocket.SendVoiceGatewayPacket(voicev2.OpSessionDescription, voicev2.SessionDescription{
			Type:                desc.Type.String(),
			SDP:                 desc.SDP,
			RTCConnectionID:     rtcConnectionID,
			MediaSessionID:      mediaSessionID,
			AudioCodec:          audioCodec,
			VideoCodec:          videoCodec,
			DAVEProtocolVersion: daveProtocol,
			DAVEEpoch:           daveEpoch,
		})
	}
	return p.websocket.SendRTCOffer(desc)
}

func (p *peerConnectionState) sessionDescriptionMetadata() (rtcConnectionID, mediaSessionID string, daveProtocol int, daveEpoch uint64) {
	p.metaMu.RLock()
	defer p.metaMu.RUnlock()

	return p.rtcConnectionID, p.mediaSessionID, p.daveProtocol, p.daveEpoch
}

func (p *peerConnectionState) setSessionDescriptionMetadata(rtcConnectionID, mediaSessionID string, daveProtocol int, daveEpoch uint64) {
	p.metaMu.Lock()
	defer p.metaMu.Unlock()

	p.rtcConnectionID = rtcConnectionID
	p.mediaSessionID = mediaSessionID
	p.daveProtocol = daveProtocol
	p.daveEpoch = daveEpoch
}

func (p *peerConnectionState) setDAVEState(daveProtocol int, daveEpoch uint64) {
	p.metaMu.Lock()
	defer p.metaMu.Unlock()

	p.daveProtocol = daveProtocol
	p.daveEpoch = daveEpoch
}

func (p *peerConnectionState) daveState() (int, uint64) {
	p.metaMu.RLock()
	defer p.metaMu.RUnlock()

	return p.daveProtocol, p.daveEpoch
}

func (p *peerConnectionState) sendSpeaking(fromUser int64, speaking int) error {
	if p.signalVersion == signalProtocolVersion2 {
		return p.websocket.SendVoiceGatewayPacket(voicev2.OpSpeaking, voicev2.Speaking{
			UserID:   strconv.FormatInt(fromUser, 10),
			Speaking: speaking,
		})
	}
	return p.websocket.SendEnvelope(OutEnvelope{
		OP: int(mqmsg.OPCodeRTC),
		T:  int(mqmsg.EventTypeRTCSpeaking),
		D:  speakingEvent{UserId: fromUser, Speaking: speaking},
	})
}

func (p *peerConnectionState) sendMuteState(userID int64, muted bool) error {
	return p.websocket.SendEnvelope(OutEnvelope{
		OP: int(mqmsg.OPCodeRTC),
		T:  int(mqmsg.EventTypeRTCServerMuteUser),
		D:  muteEvent{UserId: userID, Muted: muted},
	})
}

func (p *peerConnectionState) sendDeafenState(userID int64, deafened bool) error {
	return p.websocket.SendEnvelope(OutEnvelope{
		OP: int(mqmsg.OPCodeRTC),
		T:  int(mqmsg.EventTypeRTCServerDeafenUser),
		D:  deafenEvent{UserId: userID, Deafened: deafened},
	})
}

func (p *peerConnectionState) sendKick(targetUserID int64) error {
	return p.websocket.SendEnvelope(OutEnvelope{
		OP: int(mqmsg.OPCodeRTC),
		T:  int(mqmsg.EventTypeRTCServerKickUser),
		D:  kickEvent{UserId: targetUserID},
	})
}

// ---------------------------------------------------------------------------
// trackLocalEntry associates a local track with its owner.
// ---------------------------------------------------------------------------

type trackLocalEntry struct {
	track *webrtc.TrackLocalStaticRTP
	kind  string
	owner int64
}

// ---------------------------------------------------------------------------
// channelState manages all peers and tracks within a single voice channel.
// Uses RWMutex for read-heavy workloads (speaking broadcasts, blocked checks).
// ---------------------------------------------------------------------------

type channelState struct {
	log         *slog.Logger
	ttlTicker   *time.Ticker
	ttlStopChan chan struct{}

	// Debounced signaling: write to signalCh to request a sync.
	// A dedicated goroutine reads from it with debounce.
	signalCh   chan struct{}
	signalStop chan struct{}

	// Per-channel blocked users set
	blockedUsers map[int64]bool

	telemetry   *observability.SFUTelemetry
	trackLocals map[string]trackLocalEntry
	peers       []*peerConnectionState

	id               int64
	topologyRevision uint64
	// Configured limits
	maxAudioBitrateBps uint64
	maxVideoBitrateBps uint64
	mu                 sync.RWMutex
	stopped            atomic.Bool
}

func newChannelState(id int64, httpClient *resty.Client, webhookUrl, webhookToken, routeID, routeURL, routeRegion string, log *slog.Logger, maxAudioBitrateBps, maxVideoBitrateBps uint64, telemetry *observability.SFUTelemetry) *channelState {
	t := time.NewTicker(time.Minute)
	stop := make(chan struct{})
	go func(channelId int64, ch chan struct{}) {
		for {
			select {
			case <-t.C:
				resp, err := httpClient.R().
					SetHeader("Content-Type", "application/json").
					SetHeader("X-Webhook-Token", webhookToken).
					SetBody(ChannelAliveNotify{
						StreamId: channelId,
						RouteID:  routeID,
						RouteURL: routeURL,
						Region:   routeRegion,
					}).
					Post(webhookUrl + "/api/v1/webhook/stream/alive")
				if err != nil {
					log.Error("channel alive request failed", slog.String("error", err.Error()))
				} else if resp.StatusCode() != 200 {
					log.Warn("channel alive unexpected status", slog.Int("status", resp.StatusCode()))
				}
			case <-ch:
				t.Stop()
				log.Info("channel liveness update stopped", slog.Int64("channel_id", channelId))
				return
			}
		}
	}(id, stop)

	sigCh := make(chan struct{}, 1)
	sigStop := make(chan struct{})
	cs := &channelState{
		id:                 id,
		log:                log,
		trackLocals:        make(map[string]trackLocalEntry),
		blockedUsers:       make(map[int64]bool),
		ttlTicker:          t,
		ttlStopChan:        stop,
		signalCh:           sigCh,
		signalStop:         sigStop,
		maxAudioBitrateBps: maxAudioBitrateBps,
		maxVideoBitrateBps: maxVideoBitrateBps,
		telemetry:          telemetry,
	}

	// Dedicated goroutine for debounced signaling.
	// Coalesces rapid signal requests into one sync pass with 50ms debounce.
	go func() {
		for {
			select {
			case <-sigCh:
				// Debounce: wait briefly to coalesce rapid signals
				time.Sleep(50 * time.Millisecond)
				// Drain any queued signals
				select {
				case <-sigCh:
				default:
				}
				cs.doSignalPeerConnections()
			case <-sigStop:
				return
			}
		}
	}()

	return cs
}

// stop terminates the channel's background goroutines. Safe to call once.
func (c *channelState) stop() {
	if c.stopped.CompareAndSwap(false, true) {
		close(c.ttlStopChan)
		close(c.signalStop)
	}
}

func (c *channelState) addPeer(state *peerConnectionState) {
	c.mu.Lock()
	c.peers = append(c.peers, state)
	rev := c.bumpTopologyRevisionLocked()
	n := len(c.peers)
	c.mu.Unlock()
	c.log.Debug("peer added", slog.Int64("channel", c.id), slog.Int64("user", state.userID), slog.Int("total_peers", n), slog.Uint64("revision", rev))
}

func (c *channelState) removePeer(pc *webrtc.PeerConnection) (removedUser int64, removedTracks []trackLocalEntry, removed bool, empty bool) {
	var rev uint64
	c.mu.Lock()
	for i := range c.peers {
		if c.peers[i].peerConnection == pc {
			removedUser = c.peers[i].userID
			// Swap with last element and truncate (order doesn't matter)
			last := len(c.peers) - 1
			c.peers[i] = c.peers[last]
			c.peers[last] = nil // help GC
			c.peers = c.peers[:last]
			removed = true
			rev = c.bumpTopologyRevisionLocked()
			break
		}
	}
	if removed && len(c.peers) == 0 && len(c.trackLocals) > 0 {
		removedTracks = make([]trackLocalEntry, 0, len(c.trackLocals))
		for _, entry := range c.trackLocals {
			removedTracks = append(removedTracks, entry)
		}
		c.trackLocals = make(map[string]trackLocalEntry)
	}
	empty = len(c.peers) == 0 && len(c.trackLocals) == 0
	n := len(c.peers)
	c.mu.Unlock()
	if removed {
		c.log.Debug("peer removed", slog.Int64("channel", c.id), slog.Int64("user", removedUser), slog.Int("total_peers", n), slog.Uint64("revision", rev))
	}
	return removedUser, removedTracks, removed, empty
}

func (c *channelState) addTrack(userID int64, t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	streamID, trackID := forwardedTrackIdentifiers(userID, t.ID())
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, trackID, streamID)
	if err != nil {
		c.log.Warn("failed to create local track", slog.Int64("channel", c.id), slog.Int64("user", userID), slog.String("track", trackID), slog.String("error", err.Error()))
		return nil
	}

	c.mu.Lock()
	c.trackLocals[trackID] = trackLocalEntry{track: trackLocal, owner: userID, kind: t.Kind().String()}
	rev := c.bumpTopologyRevisionLocked()
	c.mu.Unlock()
	c.log.Debug("track added", slog.Int64("channel", c.id), slog.Int64("user", userID), slog.String("track", trackID), slog.String("kind", t.Kind().String()), slog.Uint64("revision", rev))
	return trackLocal
}

func (c *channelState) removeTrack(track *webrtc.TrackLocalStaticRTP) (kind string, owner int64, removed bool, empty bool) {
	if track == nil {
		return "", 0, false, false
	}
	var rev uint64
	c.mu.Lock()
	if entry, ok := c.trackLocals[track.ID()]; ok {
		delete(c.trackLocals, track.ID())
		kind = entry.kind
		owner = entry.owner
		removed = true
		rev = c.bumpTopologyRevisionLocked()
	}
	empty = len(c.peers) == 0 && len(c.trackLocals) == 0
	c.mu.Unlock()
	if removed {
		c.log.Debug("track removed", slog.Int64("channel", c.id), slog.String("track", track.ID()), slog.Uint64("revision", rev))
	}
	return kind, owner, removed, empty
}

func forwardedTrackIdentifiers(userID int64, remoteTrackID string) (streamID string, trackID string) {
	// Use streamID to carry the sender's raw user ID so browser clients can map
	// remote tracks directly to application users via event.streams[0].id.
	streamID = fmt.Sprintf("%d", userID)
	// Ensure unique Track ID per user to avoid collisions across peers (e.g. "video").
	trackID = fmt.Sprintf("%d-%s", userID, remoteTrackID)
	return streamID, trackID
}

func (c *channelState) bumpTopologyRevisionLocked() uint64 {
	c.topologyRevision++
	return c.topologyRevision
}

func (c *channelState) requestPeerNegotiation(pc *webrtc.PeerConnection) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, state := range c.peers {
		if state.peerConnection == pc {
			state.forceOffer = true
			return true
		}
	}
	return false
}

func (c *channelState) applyPeerAnswer(pc *webrtc.PeerConnection) (needsSignal bool, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, state := range c.peers {
		if state.peerConnection != pc {
			continue
		}
		state.negotiated = true
		if state.offeredRevision > state.appliedRevision {
			state.appliedRevision = state.offeredRevision
		}
		needsSignal = state.forceOffer || state.appliedRevision < c.topologyRevision
		return needsSignal, true
	}
	return false, false
}

func (c *channelState) preparePeerInitialSync(state *peerConnectionState) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.syncPeerSenders(state)
	return c.topologyRevision
}

// signalPeerConnections enqueues a signal request to the dedicated goroutine.
// Non-blocking: if a signal is already pending it is coalesced.
func (c *channelState) signalPeerConnections() {
	if c.stopped.Load() {
		return
	}
	select {
	case c.signalCh <- struct{}{}:
	default:
		// Already pending, will be handled
	}
}

// doSignalPeerConnections performs the actual sync pass.
// Called by the dedicated signal goroutine only.
//
// The logic:
//  1. Under lock, remove ALL closed/failed peers in a single sweep (no retry loop).
//  2. For each peer that still needs negotiation, reconcile senders with the
//     current track set and create an offer if the effective sender layout changed
//     or the peer explicitly needs an offer.
//  3. Release the lock and send offers over WebSocket.
func (c *channelState) doSignalPeerConnections() {
	c.mu.Lock()

	// Step 1: Remove all closed/failed peers in one pass.
	n := 0
	for _, p := range c.peers {
		st := p.peerConnection.ConnectionState()
		if st != webrtc.PeerConnectionStateClosed && st != webrtc.PeerConnectionStateFailed {
			c.peers[n] = p
			n++
		}
	}
	// Nil out removed tail entries to help GC.
	for i := n; i < len(c.peers); i++ {
		c.peers[i] = nil
	}
	c.peers = c.peers[:n]

	// Step 2: Build offers for each signaling-stable peer.
	type peerWork struct {
		state *peerConnectionState
		offer webrtc.SessionDescription
	}
	work := make([]peerWork, 0, len(c.peers))
	currentRevision := c.topologyRevision
	c.log.Debug("signaling peers", slog.Int64("channel", c.id), slog.Int("peers", len(c.peers)), slog.Int("tracks", len(c.trackLocals)), slog.Uint64("revision", currentRevision))

	for _, state := range c.peers {
		needsOffer := !state.negotiated || state.forceOffer || state.appliedRevision < currentRevision
		if !needsOffer {
			continue
		}
		if state.peerConnection.SignalingState() != webrtc.SignalingStateStable {
			continue
		}

		sendersChanged := c.syncPeerSenders(state)
		if !state.forceOffer && state.negotiated && !sendersChanged {
			state.appliedRevision = currentRevision
			continue
		}

		offer, err := state.peerConnection.CreateOffer(nil)
		if err != nil {
			c.log.Warn("failed to create offer", slog.Int64("channel", c.id), slog.String("error", err.Error()))
			continue
		}
		if err = state.peerConnection.SetLocalDescription(offer); err != nil {
			c.log.Warn("failed to set local description", slog.Int64("channel", c.id), slog.String("error", err.Error()))
			continue
		}

		offerToSend := offer
		// Keep Pion's local description pristine and only munge the copy sent over
		// the wire. Re-marshaled SDP can be rejected by SetLocalDescription.
		if c.maxAudioBitrateBps > 0 {
			offerToSend.SDP = limitAudioBitrateInSDP(offerToSend.SDP, c.maxAudioBitrateBps)
		}
		if c.maxVideoBitrateBps > 0 {
			offerToSend.SDP = limitVideoBitrateInSDP(offerToSend.SDP, c.maxVideoBitrateBps)
		}
		// Strip any m-lines carrying the receiver's own forwarded track — sending
		// a user's own audio/video back wastes bandwidth and confuses DAVE E2EE.
		offerToSend.SDP = stripSelfTracksFromSDP(offerToSend.SDP, state.userID)

		state.offeredRevision = currentRevision
		state.forceOffer = false
		work = append(work, peerWork{state: state, offer: offerToSend})
	}
	c.mu.Unlock()

	// Step 3: Send offers outside the lock
	for _, w := range work {
		if err := w.state.sendDescription(w.offer); err != nil {
			c.log.Warn("failed to send offer", slog.Int64("channel", c.id), slog.String("error", err.Error()))
			continue
		}
		if c.telemetry != nil {
			c.telemetry.Offer(context.Background(), "outbound",
				attribute.Int64("voice.channel_id", c.id),
				attribute.Int64("user.id", w.state.userID),
			)
		}
	}
}

func (c *channelState) syncPeerSenders(state *peerConnectionState) (changed bool) {
	existingSenders := make(map[string]bool)
	// Remove senders that should no longer be sent (track gone, or belongs to same user)
	for _, sender := range state.peerConnection.GetSenders() {
		if sender.Track() == nil {
			continue
		}
		trackID := sender.Track().ID()
		entry, exists := c.trackLocals[trackID]
		// Remove if: track no longer exists, belongs to the same user,
		// or the receiver is server-deafened (should receive nothing).
		if !exists || entry.owner == state.userID || state.serverDeafened {
			if err := state.peerConnection.RemoveTrack(sender); err != nil {
				c.log.Warn("failed to remove sender", slog.Int64("channel", c.id), slog.String("error", err.Error()))
			} else {
				changed = true
			}
			continue
		}
		existingSenders[trackID] = true
	}

	// Add missing tracks for other users (skip if receiver is deafened)
	if !state.serverDeafened {
		for id, entry := range c.trackLocals {
			if entry.owner == state.userID {
				continue
			}
			if existingSenders[id] {
				continue
			}
			if _, err := state.peerConnection.AddTrack(entry.track); err != nil {
				c.log.Warn("failed to add track to peer", slog.Int64("channel", c.id), slog.String("error", err.Error()))
			} else {
				changed = true
			}
		}
	}

	return changed
}

func (c *channelState) dispatchKeyFrame() {
	c.mu.RLock()
	peers := make([]*peerConnectionState, len(c.peers))
	copy(peers, c.peers)
	c.mu.RUnlock()

	for _, p := range peers {
		for _, receiver := range p.peerConnection.GetReceivers() {
			track := receiver.Track()
			if track == nil || track.Kind() != webrtc.RTPCodecTypeVideo {
				continue
			}
			_ = p.peerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())},
			})
		}
	}
}

func (c *channelState) isEmpty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.peers) == 0 && len(c.trackLocals) == 0
}

// isBlocked checks if a user is in this channel's block list.
func (c *channelState) isBlocked(userID int64) bool {
	c.mu.RLock()
	blocked := c.blockedUsers[userID]
	c.mu.RUnlock()
	return blocked
}

// snapshotPeers returns a shallow copy of the peer slice for iteration outside the lock.
func (c *channelState) snapshotPeers() []*peerConnectionState {
	c.mu.RLock()
	peers := make([]*peerConnectionState, len(c.peers))
	copy(peers, c.peers)
	c.mu.RUnlock()
	return peers
}

// broadcastSpeaking relays speaking state to all peers in the channel except the origin.
func (c *channelState) broadcastSpeaking(fromUser int64, speaking int) {
	peers := c.snapshotPeers()
	for _, p := range peers {
		if p.userID == fromUser {
			continue
		}
		_ = p.sendSpeaking(fromUser, speaking)
	}
}

func (c *channelState) broadcastMuteState(userID int64, muted bool) {
	peers := c.snapshotPeers()
	for _, p := range peers {
		_ = p.sendMuteState(userID, muted)
	}
}

func (c *channelState) broadcastDeafenState(userID int64, deafened bool) {
	peers := c.snapshotPeers()
	for _, p := range peers {
		_ = p.sendDeafenState(userID, deafened)
	}
}

// serverMuteUser sets/unsets server-wide mute on a target user.
// When muted, the user's audio tracks are removed so no one receives them.
func (c *channelState) serverMuteUser(targetUserID int64, muted bool) {
	var changedTopology bool
	c.mu.Lock()
	for _, p := range c.peers {
		if p.userID == targetUserID {
			p.serverMuted = muted
			break
		}
	}
	// If muting, remove the user's tracks so they stop being forwarded
	if muted {
		for id, entry := range c.trackLocals {
			if entry.owner == targetUserID {
				delete(c.trackLocals, id)
				changedTopology = true
			}
		}
	}
	if changedTopology {
		c.bumpTopologyRevisionLocked()
	}
	c.mu.Unlock()
	c.log.Info("server mute user", slog.Int64("channel", c.id), slog.Int64("user", targetUserID), slog.Bool("muted", muted))
	// Notify all peers about the mute state and renegotiate
	c.broadcastMuteState(targetUserID, muted)
	if changedTopology {
		c.signalPeerConnections()
	}
}

// serverDeafenUser sets/unsets server-wide deafen on a target user.
// When deafened, the user receives no audio/video from anyone.
func (c *channelState) serverDeafenUser(targetUserID int64, deafened bool) {
	var changedTopology bool
	c.mu.Lock()
	for _, p := range c.peers {
		if p.userID == targetUserID {
			if p.serverDeafened != deafened {
				changedTopology = true
			}
			p.serverDeafened = deafened
			break
		}
	}
	if changedTopology {
		c.bumpTopologyRevisionLocked()
	}
	c.mu.Unlock()
	c.log.Info("server deafen user", slog.Int64("channel", c.id), slog.Int64("user", targetUserID), slog.Bool("deafened", deafened))
	// Notify all peers and renegotiate (deafened user gets no senders)
	c.broadcastDeafenState(targetUserID, deafened)
	if changedTopology {
		c.signalPeerConnections()
	}
}

// kickUser closes the peer connection of the target user.
func (c *channelState) kickUser(targetUserID int64) {
	c.mu.RLock()
	var target *peerConnectionState
	for _, p := range c.peers {
		if p.userID == targetUserID {
			target = p
			break
		}
	}
	c.mu.RUnlock()

	if target == nil {
		c.log.Warn("kick target not found", slog.Int64("channel", c.id), slog.Int64("user", targetUserID))
		return
	}
	c.log.Info("kicking user", slog.Int64("channel", c.id), slog.Int64("user", targetUserID))
	// Notify the target they are being kicked
	_ = target.sendKick(targetUserID)
	// Close their peer connection (triggers cleanup via OnConnectionStateChange)
	_ = target.peerConnection.Close()
}

// blockUser adds or removes a user from the channel's block list.
func (c *channelState) blockUser(targetUserID int64, block bool) {
	c.mu.Lock()
	if block {
		c.blockedUsers[targetUserID] = true
	} else {
		delete(c.blockedUsers, targetUserID)
	}
	c.mu.Unlock()
	// If blocking, also kick them out
	if block {
		c.kickUser(targetUserID)
	}
}

// ---------------------------------------------------------------------------
// SFU is the top-level manager of voice channels.
// ---------------------------------------------------------------------------

type SFU struct {
	log        *slog.Logger
	httpClient *resty.Client // Shared HTTP client for all webhook calls
	channels   map[int64]*channelState
	telemetry  *observability.SFUTelemetry
	// Graceful shutdown
	done         chan struct{}
	webhookUrl   string
	webhookToken string
	routeID      string
	routeURL     string
	routeRegion  string

	maxAudioBitrateBps    uint64
	maxVideoBitrateBps    uint64
	mu                    sync.RWMutex
	audioBitrateMarginPct int
	enforceAudioBitrate   bool
}

func NewSFU(webhookUrl, webhookToken, routeID, routeURL, routeRegion string, log *slog.Logger, maxAudioBitrateBps, maxVideoBitrateBps uint64, enforceAudioBitrate bool, audioBitrateMarginPct int, telemetry *observability.SFUTelemetry) *SFU {
	return &SFU{
		log:                   log,
		channels:              make(map[int64]*channelState),
		webhookUrl:            webhookUrl,
		webhookToken:          webhookToken,
		routeID:               routeID,
		routeURL:              routeURL,
		routeRegion:           routeRegion,
		httpClient:            resty.New().SetTransport(observability.NewHTTPTransport("gochat-sfu-webhook", http.DefaultTransport)).SetTimeout(5 * time.Second),
		maxAudioBitrateBps:    maxAudioBitrateBps,
		maxVideoBitrateBps:    maxVideoBitrateBps,
		enforceAudioBitrate:   enforceAudioBitrate,
		audioBitrateMarginPct: audioBitrateMarginPct,
		telemetry:             telemetry,
		done:                  make(chan struct{}),
	}
}

// Close stops all background goroutines (key-frame ticker) and cleans up channels.
func (s *SFU) Close() {
	select {
	case <-s.done:
		return // already closed
	default:
		close(s.done)
	}

	s.mu.Lock()
	for id, ch := range s.channels {
		ch.stop()
		delete(s.channels, id)
	}
	s.mu.Unlock()
}

func (s *SFU) getOrCreateChannel(channelID int64) (*channelState, bool) {
	// Fast path: read lock
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if ok {
		return ch, false
	}

	// Slow path: write lock + double-check
	s.mu.Lock()
	ch, ok = s.channels[channelID]
	created := false
	if !ok {
		ch = newChannelState(channelID, s.httpClient, s.webhookUrl, s.webhookToken, s.routeID, s.routeURL, s.routeRegion, s.log, s.maxAudioBitrateBps, s.maxVideoBitrateBps, s.telemetry)
		s.channels[channelID] = ch
		created = true
	}
	s.mu.Unlock()
	return ch, created
}

func (s *SFU) AddPeer(ctx context.Context, channelID int64, state *peerConnectionState) *channelState {
	ch, created := s.getOrCreateChannel(channelID)
	if created && s.telemetry != nil {
		s.telemetry.ChannelDelta(ctx, 1)
	}
	ch.addPeer(state)
	if s.telemetry != nil {
		s.telemetry.PeerDelta(ctx, 1,
			attribute.Int64("voice.channel_id", channelID),
			attribute.Int64("user.id", state.userID),
		)
	}
	return ch
}

func (s *SFU) RemovePeer(ctx context.Context, channelID int64, pc *webrtc.PeerConnection) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		s.log.Debug("remove peer: channel not found", slog.Int64("channel", channelID))
		return
	}

	removedUser, removedTracks, removed, empty := ch.removePeer(pc)
	if removed {
		if s.telemetry != nil {
			s.telemetry.PeerDelta(ctx, -1,
				attribute.Int64("voice.channel_id", channelID),
				attribute.Int64("user.id", removedUser),
			)
			for _, entry := range removedTracks {
				s.telemetry.TrackDelta(ctx, entry.kind, -1,
					attribute.Int64("voice.channel_id", channelID),
					attribute.Int64("user.id", entry.owner),
				)
			}
		}
		ch.signalPeerConnections()
	}
	if empty {
		s.cleanupChannel(ctx, channelID, ch)
	}
}

func (s *SFU) GetChannel(channelID int64) *channelState {
	s.mu.RLock()
	ch := s.channels[channelID]
	s.mu.RUnlock()
	return ch
}

func (s *SFU) ChannelRevision(channelID int64) uint64 {
	s.mu.RLock()
	ch := s.channels[channelID]
	s.mu.RUnlock()
	if ch == nil {
		return 0
	}
	ch.mu.RLock()
	rev := ch.topologyRevision
	ch.mu.RUnlock()
	return rev
}

func (s *SFU) AddTrack(ctx context.Context, channelID int64, userID int64, t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	ch, _ := s.getOrCreateChannel(channelID)
	track := ch.addTrack(userID, t)
	if track != nil {
		if s.telemetry != nil {
			s.telemetry.TrackDelta(ctx, t.Kind().String(), 1,
				attribute.Int64("voice.channel_id", channelID),
				attribute.Int64("user.id", userID),
			)
		}
		ch.signalPeerConnections()
	}
	return track
}

func (s *SFU) RemoveTrack(ctx context.Context, channelID int64, track *webrtc.TrackLocalStaticRTP) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	kind, owner, removed, empty := ch.removeTrack(track)
	if removed {
		if s.telemetry != nil {
			s.telemetry.TrackDelta(ctx, kind, -1,
				attribute.Int64("voice.channel_id", channelID),
				attribute.Int64("user.id", owner),
			)
		}
		ch.signalPeerConnections()
	}
	if empty {
		s.cleanupChannel(ctx, channelID, ch)
	}
}

// cleanupChannel stops goroutines and removes a channel from the map.
// Uses double-check under write lock to prevent races.
func (s *SFU) cleanupChannel(ctx context.Context, channelID int64, ch *channelState) {
	s.mu.Lock()
	// Double-check: another goroutine may have added a new peer between
	// the empty check and acquiring this write lock.
	if current, ok := s.channels[channelID]; ok && current == ch && ch.isEmpty() {
		ch.stop()
		delete(s.channels, channelID)
		if s.telemetry != nil {
			s.telemetry.ChannelDelta(ctx, -1)
		}
	}
	s.mu.Unlock()
}

func (s *SFU) SignalChannel(ctx context.Context, channelID int64) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	if s.telemetry != nil {
		s.telemetry.Renegotiation(ctx, attribute.Int64("voice.channel_id", channelID))
	}
	ch.signalPeerConnections()
}

func (s *SFU) SignalPeer(ctx context.Context, channelID int64, pc *webrtc.PeerConnection) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	if !ch.requestPeerNegotiation(pc) {
		return
	}
	if s.telemetry != nil {
		s.telemetry.Renegotiation(ctx, attribute.Int64("voice.channel_id", channelID))
	}
	ch.signalPeerConnections()
}

func (s *SFU) ApplyAnswer(ctx context.Context, channelID int64, pc *webrtc.PeerConnection) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	needsSignal, found := ch.applyPeerAnswer(pc)
	if !found {
		return
	}
	ch.dispatchKeyFrame()
	if !needsSignal {
		return
	}
	if s.telemetry != nil {
		s.telemetry.Renegotiation(ctx, attribute.Int64("voice.channel_id", channelID))
	}
	ch.signalPeerConnections()
}

func (s *SFU) RequestKeyFrame(channelID int64) {
	s.mu.RLock()
	ch := s.channels[channelID]
	s.mu.RUnlock()
	if ch == nil {
		return
	}
	ch.dispatchKeyFrame()
}

func (s *SFU) dispatchKeyFrameAll() {
	s.mu.RLock()
	channels := make([]*channelState, 0, len(s.channels))
	for _, ch := range s.channels {
		channels = append(channels, ch)
	}
	s.mu.RUnlock()

	for _, ch := range channels {
		ch.dispatchKeyFrame()
	}
}

// BroadcastSpeaking relays speaking state to all peers in the channel except the origin.
func (s *SFU) BroadcastSpeaking(_ context.Context, channelID int64, fromUser int64, speaking int) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.broadcastSpeaking(fromUser, speaking)
}

// RunKeyFrameTicker periodically requests key frames from all peers.
// Stops when the SFU's done channel is closed.
func (s *SFU) RunKeyFrameTicker() {
	ticker := time.NewTicker(periodicKeyFrameInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.dispatchKeyFrameAll()
		case <-s.done:
			return
		}
	}
}

// hasPerm checks if a permission bitmask includes a specific voice permission.
// PermAdministrator overrides all checks.
func hasPerm(perms int64, required permissions.RolePermission) bool {
	if perms&int64(permissions.PermAdministrator) != 0 {
		return true
	}
	return perms&int64(required) != 0
}

// ServerMuteUser sets/unsets server-wide mute on a target user in a channel.
func (s *SFU) ServerMuteUser(ctx context.Context, channelID int64, targetUserID int64, muted bool) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.serverMuteUser(targetUserID, muted)
	if s.telemetry != nil {
		s.telemetry.Renegotiation(ctx, attribute.Int64("voice.channel_id", channelID))
	}
}

// ServerDeafenUser sets/unsets server-wide deafen on a target user.
func (s *SFU) ServerDeafenUser(ctx context.Context, channelID int64, targetUserID int64, deafened bool) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.serverDeafenUser(targetUserID, deafened)
	if s.telemetry != nil {
		s.telemetry.Renegotiation(ctx, attribute.Int64("voice.channel_id", channelID))
	}
}

// KickUser closes the peer connection of the target user, removing them from the channel.
func (s *SFU) KickUser(_ context.Context, channelID int64, targetUserID int64) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.kickUser(targetUserID)
}

// BlockUser adds or removes a user from the channel's block list.
func (s *SFU) BlockUser(_ context.Context, channelID int64, targetUserID int64, block bool) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.blockUser(targetUserID, block)
}

// KickAll sends a kick envelope to every peer in the channel and closes their peer connections.
// Used when the channel's SFU region changes and this instance is the old SFU.
func (s *SFU) KickAll(_ context.Context, channelID int64) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	peers := ch.snapshotPeers()
	for _, p := range peers {
		_ = p.sendKick(p.userID)
		_ = p.peerConnection.Close()
	}
}

// IsBlocked checks if a user is blocked from a channel.
func (s *SFU) IsBlocked(channelID int64, userID int64) bool {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return false
	}
	return ch.isBlocked(userID)
}
