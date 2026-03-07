package main

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"resty.dev/v3"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

// ---------------------------------------------------------------------------
// threadSafeWriter wraps a websocket.Conn with a mutex for concurrent writes.
// ---------------------------------------------------------------------------

type threadSafeWriter struct {
	conn   *websocket.Conn
	mu     sync.Mutex
	closed atomic.Bool
}

func (t *threadSafeWriter) WriteJSON(v any) error {
	if t.closed.Load() {
		return fmt.Errorf("websocket closed")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		return fmt.Errorf("websocket closed")
	}
	return t.conn.WriteJSON(v)
}

func (t *threadSafeWriter) SendEnvelope(env OutEnvelope) error {
	return t.WriteJSON(env)
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

// Close marks the writer as closed. Subsequent writes return immediately.
func (t *threadSafeWriter) Close() {
	t.closed.Store(true)
}

// ---------------------------------------------------------------------------
// peerConnectionState holds per-peer state within a channel.
// ---------------------------------------------------------------------------

type peerConnectionState struct {
	peerConnection *webrtc.PeerConnection
	websocket      *threadSafeWriter
	userID         int64
	perms          int64 // voice permission bitmask from JWT
	serverMuted    bool  // server-wide mute (admin action)
	serverDeafened bool  // server-wide deafen (admin action)
}

// ---------------------------------------------------------------------------
// trackLocalEntry associates a local track with its owner.
// ---------------------------------------------------------------------------

type trackLocalEntry struct {
	track *webrtc.TrackLocalStaticRTP
	owner int64
}

// ---------------------------------------------------------------------------
// channelState manages all peers and tracks within a single voice channel.
// Uses RWMutex for read-heavy workloads (speaking broadcasts, blocked checks).
// ---------------------------------------------------------------------------

type channelState struct {
	id  int64
	log *slog.Logger

	mu          sync.RWMutex
	peers       []*peerConnectionState
	trackLocals map[string]trackLocalEntry

	ttlTicker   *time.Ticker
	ttlStopChan chan struct{}

	// Debounced signaling: write to signalCh to request a sync.
	// A dedicated goroutine reads from it with debounce.
	signalCh   chan struct{}
	signalStop chan struct{}
	stopped    atomic.Bool

	// Per-channel blocked users set
	blockedUsers map[int64]bool

	// Configured limits
	maxAudioBitrateBps uint64
}

func newChannelState(id int64, httpClient *resty.Client, webhookUrl, webhookToken string, log *slog.Logger, maxAudioBitrateBps uint64) *channelState {
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
						GuildId:   nil,
						ChannelId: channelId,
					}).
					Post(webhookUrl + "/api/v1/webhook/sfu/channel/alive")
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
	n := len(c.peers)
	c.mu.Unlock()
	c.log.Debug("peer added", slog.Int64("channel", c.id), slog.Int64("user", state.userID), slog.Int("total_peers", n))
}

func (c *channelState) removePeer(pc *webrtc.PeerConnection) (removed bool, empty bool) {
	var removedUser int64
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
			break
		}
	}
	if removed && len(c.peers) == 0 && len(c.trackLocals) > 0 {
		c.trackLocals = make(map[string]trackLocalEntry)
	}
	empty = len(c.peers) == 0 && len(c.trackLocals) == 0
	n := len(c.peers)
	c.mu.Unlock()
	if removed {
		c.log.Debug("peer removed", slog.Int64("channel", c.id), slog.Int64("user", removedUser), slog.Int("total_peers", n))
	}
	return removed, empty
}

func (c *channelState) addTrack(userID int64, t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	// Use streamID to carry the sender's user ID so receivers can map tracks to users.
	// Keep the original track ID for uniqueness.
	streamID := fmt.Sprintf("u:%d", userID)
	// Ensure unique Track ID per user to avoid collisions across peers (e.g. "video")
	trackID := fmt.Sprintf("%d-%s", userID, t.ID())
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, trackID, streamID)
	if err != nil {
		c.log.Warn("failed to create local track", slog.Int64("channel", c.id), slog.Int64("user", userID), slog.String("track", trackID), slog.String("error", err.Error()))
		return nil
	}

	c.mu.Lock()
	c.trackLocals[trackID] = trackLocalEntry{track: trackLocal, owner: userID}
	c.mu.Unlock()
	c.log.Debug("track added", slog.Int64("channel", c.id), slog.Int64("user", userID), slog.String("track", trackID), slog.String("kind", t.Kind().String()))
	return trackLocal
}

func (c *channelState) removeTrack(track *webrtc.TrackLocalStaticRTP) (removed bool, empty bool) {
	if track == nil {
		return false, false
	}
	c.mu.Lock()
	if _, ok := c.trackLocals[track.ID()]; ok {
		delete(c.trackLocals, track.ID())
		removed = true
	}
	empty = len(c.peers) == 0 && len(c.trackLocals) == 0
	c.mu.Unlock()
	if removed {
		c.log.Debug("track removed", slog.Int64("channel", c.id), slog.String("track", track.ID()))
	}
	return removed, empty
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
//  2. For each remaining peer in stable signaling state, reconcile senders
//     with the current track set, create an offer, and set local description.
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
	c.log.Debug("signaling peers", slog.Int64("channel", c.id), slog.Int("peers", len(c.peers)), slog.Int("tracks", len(c.trackLocals)))

	for _, state := range c.peers {
		if state.peerConnection.SignalingState() != webrtc.SignalingStateStable {
			continue
		}

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
				}
			}
		}

		offer, err := state.peerConnection.CreateOffer(nil)
		if err != nil {
			c.log.Warn("failed to create offer", slog.Int64("channel", c.id), slog.String("error", err.Error()))
			continue
		}
		if c.maxAudioBitrateBps > 0 {
			offer.SDP = limitAudioBitrateInSDP(offer.SDP, c.maxAudioBitrateBps)
		}
		if err = state.peerConnection.SetLocalDescription(offer); err != nil {
			c.log.Warn("failed to set local description", slog.Int64("channel", c.id), slog.String("error", err.Error()))
			continue
		}

		work = append(work, peerWork{state: state, offer: offer})
	}
	c.mu.Unlock()

	// Step 3: Send offers outside the lock
	for _, w := range work {
		if err := w.state.websocket.SendRTCOffer(w.offer); err != nil {
			c.log.Warn("failed to send offer", slog.Int64("channel", c.id), slog.String("error", err.Error()))
		}
	}
}

func (c *channelState) dispatchKeyFrame() {
	c.mu.RLock()
	peers := make([]*peerConnectionState, len(c.peers))
	copy(peers, c.peers)
	c.mu.RUnlock()

	for _, p := range peers {
		for _, receiver := range p.peerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}
			_ = p.peerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{MediaSSRC: uint32(receiver.Track().SSRC())},
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
	payload := speakingEvent{UserId: fromUser, Speaking: speaking}
	env := OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCSpeaking), D: payload}

	for _, p := range peers {
		if p.userID == fromUser {
			continue
		}
		_ = p.websocket.SendEnvelope(env)
	}
}

func (c *channelState) broadcastMuteState(userID int64, muted bool) {
	peers := c.snapshotPeers()
	env := OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCServerMuteUser), D: muteEvent{UserId: userID, Muted: muted}}
	for _, p := range peers {
		_ = p.websocket.SendEnvelope(env)
	}
}

func (c *channelState) broadcastDeafenState(userID int64, deafened bool) {
	peers := c.snapshotPeers()
	env := OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCServerDeafenUser), D: deafenEvent{UserId: userID, Deafened: deafened}}
	for _, p := range peers {
		_ = p.websocket.SendEnvelope(env)
	}
}

// serverMuteUser sets/unsets server-wide mute on a target user.
// When muted, the user's audio tracks are removed so no one receives them.
func (c *channelState) serverMuteUser(targetUserID int64, muted bool) {
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
			}
		}
	}
	c.mu.Unlock()
	c.log.Info("server mute user", slog.Int64("channel", c.id), slog.Int64("user", targetUserID), slog.Bool("muted", muted))
	// Notify all peers about the mute state and renegotiate
	c.broadcastMuteState(targetUserID, muted)
	c.signalPeerConnections()
}

// serverDeafenUser sets/unsets server-wide deafen on a target user.
// When deafened, the user receives no audio/video from anyone.
func (c *channelState) serverDeafenUser(targetUserID int64, deafened bool) {
	c.mu.Lock()
	for _, p := range c.peers {
		if p.userID == targetUserID {
			p.serverDeafened = deafened
			break
		}
	}
	c.mu.Unlock()
	c.log.Info("server deafen user", slog.Int64("channel", c.id), slog.Int64("user", targetUserID), slog.Bool("deafened", deafened))
	// Notify all peers and renegotiate (deafened user gets no senders)
	c.broadcastDeafenState(targetUserID, deafened)
	c.signalPeerConnections()
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
	_ = target.websocket.SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCServerKickUser), D: kickEvent{UserId: targetUserID}})
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
	log      *slog.Logger
	mu       sync.RWMutex
	channels map[int64]*channelState

	webhookUrl   string
	webhookToken string
	httpClient   *resty.Client // Shared HTTP client for all webhook calls

	maxAudioBitrateBps    uint64
	enforceAudioBitrate   bool
	audioBitrateMarginPct int

	// Graceful shutdown
	done chan struct{}
}

func NewSFU(webhookUrl, webhookToken string, log *slog.Logger, maxAudioBitrateBps uint64, enforceAudioBitrate bool, audioBitrateMarginPct int) *SFU {
	return &SFU{
		log:                   log,
		channels:              make(map[int64]*channelState),
		webhookUrl:            webhookUrl,
		webhookToken:          webhookToken,
		httpClient:            resty.New().SetTimeout(5 * time.Second),
		maxAudioBitrateBps:    maxAudioBitrateBps,
		enforceAudioBitrate:   enforceAudioBitrate,
		audioBitrateMarginPct: audioBitrateMarginPct,
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

func (s *SFU) getOrCreateChannel(channelID int64) *channelState {
	// Fast path: read lock
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if ok {
		return ch
	}

	// Slow path: write lock + double-check
	s.mu.Lock()
	ch, ok = s.channels[channelID]
	if !ok {
		ch = newChannelState(channelID, s.httpClient, s.webhookUrl, s.webhookToken, s.log, s.maxAudioBitrateBps)
		s.channels[channelID] = ch
	}
	s.mu.Unlock()
	return ch
}

func (s *SFU) AddPeer(channelID int64, state *peerConnectionState) *channelState {
	ch := s.getOrCreateChannel(channelID)
	ch.addPeer(state)
	return ch
}

func (s *SFU) RemovePeer(channelID int64, pc *webrtc.PeerConnection) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		s.log.Debug("remove peer: channel not found", slog.Int64("channel", channelID))
		return
	}

	removed, empty := ch.removePeer(pc)
	if removed {
		ch.signalPeerConnections()
	}
	if empty {
		s.cleanupChannel(channelID, ch)
	}
}

func (s *SFU) AddTrack(channelID int64, userID int64, t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	ch := s.getOrCreateChannel(channelID)
	track := ch.addTrack(userID, t)
	if track != nil {
		ch.signalPeerConnections()
	}
	return track
}

func (s *SFU) RemoveTrack(channelID int64, track *webrtc.TrackLocalStaticRTP) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	removed, empty := ch.removeTrack(track)
	if removed {
		ch.signalPeerConnections()
	}
	if empty {
		s.cleanupChannel(channelID, ch)
	}
}

// cleanupChannel stops goroutines and removes a channel from the map.
// Uses double-check under write lock to prevent races.
func (s *SFU) cleanupChannel(channelID int64, ch *channelState) {
	s.mu.Lock()
	// Double-check: another goroutine may have added a new peer between
	// the empty check and acquiring this write lock.
	if current, ok := s.channels[channelID]; ok && current == ch && ch.isEmpty() {
		ch.stop()
		delete(s.channels, channelID)
	}
	s.mu.Unlock()
}

func (s *SFU) SignalChannel(channelID int64) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.signalPeerConnections()
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
func (s *SFU) BroadcastSpeaking(channelID int64, fromUser int64, speaking int) {
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
	ticker := time.NewTicker(3 * time.Second)
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
func (s *SFU) ServerMuteUser(channelID int64, targetUserID int64, muted bool) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.serverMuteUser(targetUserID, muted)
}

// ServerDeafenUser sets/unsets server-wide deafen on a target user.
func (s *SFU) ServerDeafenUser(channelID int64, targetUserID int64, deafened bool) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.serverDeafenUser(targetUserID, deafened)
}

// KickUser closes the peer connection of the target user, removing them from the channel.
func (s *SFU) KickUser(channelID int64, targetUserID int64) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.kickUser(targetUserID)
}

// BlockUser adds or removes a user from the channel's block list.
func (s *SFU) BlockUser(channelID int64, targetUserID int64, block bool) {
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
func (s *SFU) KickAll(channelID int64) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	peers := ch.snapshotPeers()
	for _, p := range peers {
		_ = p.websocket.SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCServerKickUser), D: kickEvent{UserId: p.userID}})
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
