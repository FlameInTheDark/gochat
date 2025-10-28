package main

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"resty.dev/v3"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
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
	if t.conn == nil {
		return fmt.Errorf("websocket closed")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
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
}

type channelState struct {
	id          int64
	log         *slog.Logger
	mu          sync.Mutex
	peers       []*peerConnectionState
	trackLocals map[string]trackLocalEntry
	ttlTicker   *time.Ticker
	ttlStopChan chan struct{}

	// Configured limits
	maxAudioBitrateBps uint64
}

type trackLocalEntry struct {
	track *webrtc.TrackLocalStaticRTP
	owner int64
}

func newChannelState(id int64, webhookUrl, webhookToken string, log *slog.Logger, maxAudioBitrateBps uint64) *channelState {
	t := time.NewTicker(time.Minute)
	stop := make(chan struct{})
	go func(channelId int64, ch chan struct{}) {
		for {
			select {
			case <-t.C:
				resp, err := resty.New().R().
					SetTimeout(5*time.Second).
					SetHeader("Content-Type", "application/json").
					SetHeader("X-Webhook-Token", webhookToken).
					SetBody(ChannelAliveNotify{
						GuildId:   nil,
						ChannelId: channelId,
					}).
					Post(webhookUrl + "/api/v1/webhook/sfu/channel/alive")
				if err != nil && resp.StatusCode() != 200 {
					log.Error("user join notify request failed", slog.String("error", err.Error()))
				}
			case <-ch:
				t.Stop()
				log.Info("channel liveness update stopped", slog.Int64("channel_id", channelId))
				return
			}
		}
	}(id, stop)
	return &channelState{
		id:                 id,
		log:                log,
		trackLocals:        make(map[string]trackLocalEntry),
		ttlTicker:          t,
		ttlStopChan:        stop,
		maxAudioBitrateBps: maxAudioBitrateBps,
	}
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

func (c *channelState) signalPeerConnections() {
	attemptSync := func() bool {
		c.mu.Lock()
		defer c.mu.Unlock()

		for i := 0; i < len(c.peers); i++ {
			state := c.peers[i]
			if state.peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				c.peers = append(c.peers[:i], c.peers[i+1:]...)
				return true
			}

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
						return true
					}
					continue
				}
				existingSenders[trackID] = true
			}

			// Add missing tracks for other users
			for id, entry := range c.trackLocals {
				if entry.owner == state.userID {
					continue // don't send user's own tracks back to them
				}
				if _, ok := existingSenders[id]; !ok {
					if _, err := state.peerConnection.AddTrack(entry.track); err != nil {
						c.log.Warn("failed to add track to peer", slog.Int64("channel", c.id), slog.String("error", err.Error()))
						return true
					}
				}
			}

			offer, err := state.peerConnection.CreateOffer(nil)
			if err != nil {
				c.log.Warn("failed to create offer", slog.Int64("channel", c.id), slog.String("error", err.Error()))
				return true
			}
			// Inject audio bitrate limits if configured (>0)
			if c.maxAudioBitrateBps > 0 {
				offer.SDP = limitAudioBitrateInSDP(offer.SDP, c.maxAudioBitrateBps)
			}
			if err = state.peerConnection.SetLocalDescription(offer); err != nil {
				c.log.Warn("failed to set local description", slog.Int64("channel", c.id), slog.String("error", err.Error()))
				return true
			}

			if err = state.websocket.SendRTCOffer(offer); err != nil {
				c.log.Warn("failed to send offer", slog.Int64("channel", c.id), slog.String("error", err.Error()))
				return true
			}
		}
		return false
	}

	for attempts := 0; ; attempts++ {
		if attempts == 25 {
			go func() {
				time.Sleep(3 * time.Second)
				c.signalPeerConnections()
			}()
			return
		}
		if !attemptSync() {
			break
		}
	}
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
	mu       sync.Mutex
	channels map[int64]*channelState

	webhookUrl   string
	webhookToken string

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
		maxAudioBitrateBps:    maxAudioBitrateBps,
		enforceAudioBitrate:   enforceAudioBitrate,
		audioBitrateMarginPct: audioBitrateMarginPct,
	}
}

func (s *SFU) getOrCreateChannel(channelID int64) *channelState {
	s.mu.Lock()
	ch, ok := s.channels[channelID]
	if !ok {
		ch = newChannelState(channelID, s.webhookUrl, s.webhookToken, s.log, s.maxAudioBitrateBps)
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
	s.mu.Lock()
	ch, ok := s.channels[channelID]
	s.mu.Unlock()
	if !ok {
		return
	}

	removed, empty := ch.removePeer(pc)
	if removed {
		ch.signalPeerConnections()
	}
	if empty {
		s.mu.Lock()
		if ch.isEmpty() {
			close(s.channels[channelID].ttlStopChan)
			delete(s.channels, channelID)
		}
		s.mu.Unlock()
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
	s.mu.Lock()
	ch, ok := s.channels[channelID]
	s.mu.Unlock()
	if !ok {
		return
	}
	removed, empty := ch.removeTrack(track)
	if removed {
		ch.signalPeerConnections()
	}
	if empty {
		s.mu.Lock()
		if ch.isEmpty() {
			delete(s.channels, channelID)
		}
		s.mu.Unlock()
	}
}

func (s *SFU) SignalChannel(channelID int64) {
	s.mu.Lock()
	ch, ok := s.channels[channelID]
	s.mu.Unlock()
	if !ok {
		return
	}
	ch.signalPeerConnections()
}

func (s *SFU) dispatchKeyFrameAll() {
	s.mu.Lock()
	channels := make([]*channelState, 0, len(s.channels))
	for _, ch := range s.channels {
		channels = append(channels, ch)
	}
	s.mu.Unlock()

	for _, ch := range channels {
		ch.dispatchKeyFrame()
	}
}

// BroadcastSpeaking relays speaking state to all peers in the channel except the origin.
func (s *SFU) BroadcastSpeaking(channelID int64, fromUser int64, speaking int) {
	s.mu.Lock()
	ch, ok := s.channels[channelID]
	s.mu.Unlock()
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
