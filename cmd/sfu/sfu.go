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

type threadSafeWriter struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (t *threadSafeWriter) WriteJSON(v any) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		return fmt.Errorf("websocket closed")
	}
	return t.conn.WriteJSON(v)
}

func (t *threadSafeWriter) SendEnvelope(env OutEnvelope) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conn == nil {
		return fmt.Errorf("websocket closed")
	}
	return t.conn.WriteJSON(env)
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

type peerConnectionState struct {
	peerConnection *webrtc.PeerConnection
	websocket      *threadSafeWriter
	userID         int64
	perms          int64 // voice permission bitmask from JWT
	serverMuted    bool  // server-wide mute (admin action)
	serverDeafened bool  // server-wide deafen (admin action)
}

type channelState struct {
	id          int64
	log         *slog.Logger
	mu          sync.Mutex
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

type trackLocalEntry struct {
	track *webrtc.TrackLocalStaticRTP
	owner int64
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

func (c *channelState) addPeer(state *peerConnectionState) {
	c.mu.Lock()
	c.peers = append(c.peers, state)
	c.mu.Unlock()
}

func (c *channelState) removePeer(pc *webrtc.PeerConnection) (removed bool, empty bool) {
	c.mu.Lock()
	for i := range c.peers {
		if c.peers[i].peerConnection == pc {
			c.peers = append(c.peers[:i], c.peers[i+1:]...)
			removed = true
			break
		}
	}
	if removed && len(c.peers) == 0 && len(c.trackLocals) > 0 {
		c.trackLocals = make(map[string]trackLocalEntry)
	}
	empty = len(c.peers) == 0 && len(c.trackLocals) == 0
	c.mu.Unlock()
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
		c.log.Warn("failed to create local track", slog.Int64("channel", c.id), slog.String("error", err.Error()))
		return nil
	}

	c.mu.Lock()
	c.trackLocals[trackID] = trackLocalEntry{track: trackLocal, owner: userID}
	c.mu.Unlock()
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
func (c *channelState) doSignalPeerConnections() {
	for attempts := 0; attempts < 25; attempts++ {
		// Step 1: snapshot state under lock, perform mutations, release lock.
		c.mu.Lock()

		// Remove closed peers
		restart := false
		for i := 0; i < len(c.peers); i++ {
			if c.peers[i].peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				c.peers = append(c.peers[:i], c.peers[i+1:]...)
				restart = true
				break
			}
		}
		if restart {
			c.mu.Unlock()
			continue
		}

		// Snapshot peers and tracks for signaling
		type peerWork struct {
			state *peerConnectionState
			offer webrtc.SessionDescription
			ok    bool
		}
		work := make([]peerWork, 0, len(c.peers))

		for _, state := range c.peers {
			if state.peerConnection.SignalingState() != webrtc.SignalingStateStable {
				continue
			}

			existingSenders := map[string]bool{}
			// Remove senders that no longer should be sent (missing or belongs to same user)
			for _, sender := range state.peerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}
				trackID := sender.Track().ID()
				entry, ok := c.trackLocals[trackID]
				if !ok || entry.owner == state.userID {
					if err := state.peerConnection.RemoveTrack(sender); err != nil {
						c.log.Warn("failed to remove sender", slog.Int64("channel", c.id), slog.String("error", err.Error()))
					}
					continue
				}
				existingSenders[trackID] = true
			}

			// Add missing tracks for other users
			for id, entry := range c.trackLocals {
				if entry.owner == state.userID {
					continue
				}
				if _, ok := existingSenders[id]; !ok {
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

			work = append(work, peerWork{state: state, offer: offer, ok: true})
		}
		c.mu.Unlock()

		// Step 2: Send offers outside the lock
		for _, w := range work {
			if !w.ok {
				continue
			}
			if err := w.state.websocket.SendRTCOffer(w.offer); err != nil {
				c.log.Warn("failed to send offer", slog.Int64("channel", c.id), slog.String("error", err.Error()))
			}
		}
		return
	}
	c.log.Warn("signalPeerConnections exceeded max attempts", slog.Int64("channel", c.id))
}

func (c *channelState) dispatchKeyFrame() {
	c.mu.Lock()
	peers := append([]*peerConnectionState(nil), c.peers...)
	c.mu.Unlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.peers) == 0 && len(c.trackLocals) == 0
}

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
	}
}

func (s *SFU) getOrCreateChannel(channelID int64) *channelState {
	s.mu.Lock()
	ch, ok := s.channels[channelID]
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
func (s *SFU) cleanupChannel(channelID int64, ch *channelState) {
	s.mu.Lock()
	if ch.isEmpty() {
		ch.stopped.Store(true)
		close(ch.ttlStopChan)
		close(ch.signalStop)
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

func (c *channelState) broadcastSpeaking(fromUser int64, speaking int) {
	// snapshot peers to avoid holding lock while writing
	c.mu.Lock()
	peers := append([]*peerConnectionState(nil), c.peers...)
	c.mu.Unlock()

	payload := speakingEvent{UserId: fromUser, Speaking: speaking}
	env := OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCSpeaking), D: payload}

	for _, p := range peers {
		if p.userID == fromUser {
			continue
		}
		_ = p.websocket.SendEnvelope(env)
	}
}

func (s *SFU) RunKeyFrameTicker() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.dispatchKeyFrameAll()
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
// When muted, the user's audio tracks are removed from the channel so no one receives them.
func (s *SFU) ServerMuteUser(channelID int64, targetUserID int64, muted bool) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.serverMuteUser(targetUserID, muted)
}

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
	// Notify all peers about the mute state and renegotiate
	c.broadcastMuteState(targetUserID, muted)
	c.signalPeerConnections()
}

func (c *channelState) broadcastMuteState(userID int64, muted bool) {
	c.mu.Lock()
	peers := append([]*peerConnectionState(nil), c.peers...)
	c.mu.Unlock()

	env := OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCServerMuteUser), D: muteEvent{UserId: userID, Muted: muted}}
	for _, p := range peers {
		_ = p.websocket.SendEnvelope(env)
	}
}

// ServerDeafenUser sets/unsets server-wide deafen on a target user.
// When deafened, the user receives no audio/video from anyone.
func (s *SFU) ServerDeafenUser(channelID int64, targetUserID int64, deafened bool) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.serverDeafenUser(targetUserID, deafened)
}

func (c *channelState) serverDeafenUser(targetUserID int64, deafened bool) {
	c.mu.Lock()
	for _, p := range c.peers {
		if p.userID == targetUserID {
			p.serverDeafened = deafened
			break
		}
	}
	c.mu.Unlock()
	// Notify all peers and renegotiate (deafened user gets no senders)
	c.broadcastDeafenState(targetUserID, deafened)
	c.signalPeerConnections()
}

func (c *channelState) broadcastDeafenState(userID int64, deafened bool) {
	c.mu.Lock()
	peers := append([]*peerConnectionState(nil), c.peers...)
	c.mu.Unlock()

	env := OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCServerDeafenUser), D: deafenEvent{UserId: userID, Deafened: deafened}}
	for _, p := range peers {
		_ = p.websocket.SendEnvelope(env)
	}
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

func (c *channelState) kickUser(targetUserID int64) {
	c.mu.Lock()
	var target *peerConnectionState
	for _, p := range c.peers {
		if p.userID == targetUserID {
			target = p
			break
		}
	}
	c.mu.Unlock()
	if target == nil {
		return
	}
	// Notify the target they are being kicked
	_ = target.websocket.SendEnvelope(OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCServerKickUser), D: kickEvent{UserId: targetUserID}})
	// Close their peer connection (triggers cleanup via OnConnectionStateChange)
	_ = target.peerConnection.Close()
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

// KickAll sends a kick envelope to every peer in the channel and closes their peer connections.
// Used when the channel's SFU region changes and this instance is the old SFU.
// Cleanup happens naturally via OnConnectionStateChange → RemovePeer flow.
func (s *SFU) KickAll(channelID int64) {
	s.mu.RLock()
	ch, ok := s.channels[channelID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	ch.mu.Lock()
	peers := append([]*peerConnectionState(nil), ch.peers...)
	ch.mu.Unlock()
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
	ch.mu.Lock()
	blocked := ch.blockedUsers[userID]
	ch.mu.Unlock()
	return blocked
}
