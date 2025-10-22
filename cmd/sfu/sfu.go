package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

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

func (t *threadSafeWriter) sendDual(envelope any, eventName, eventData string) error {
	if t.conn == nil {
		return fmt.Errorf("websocket closed")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := t.conn.WriteJSON(envelope); err != nil {
		return err
	}
	if eventName != "" && eventData != "" {
		msg := websocketMessage{Event: eventName, Data: eventData}
		if err := t.conn.WriteJSON(&msg); err != nil {
			return err
		}
	}
	return nil
}

func (t *threadSafeWriter) SendEnvelope(env OutEnvelope) error {
	return t.sendDual(env, "", "")
}

func (t *threadSafeWriter) SendRTCOffer(desc webrtc.SessionDescription) error {
	payload := rtcOffer{SDP: desc.SDP}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	env := OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCOffer), D: payload}
	return t.sendDual(env, "offer", string(data))
}

func (t *threadSafeWriter) SendRTCCandidate(c *webrtc.ICECandidate) error {
	if c == nil {
		return nil
	}
	cand := c.ToJSON()
	payload := rtcCandidate{Candidate: cand.Candidate, SDPMid: cand.SDPMid, SDPMLineIndex: cand.SDPMLineIndex}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	env := OutEnvelope{OP: int(mqmsg.OPCodeRTC), T: int(mqmsg.EventTypeRTCCandidate), D: payload}
	return t.sendDual(env, "candidate", string(data))
}

type peerConnectionState struct {
	peerConnection *webrtc.PeerConnection
	websocket      *threadSafeWriter
}

type channelState struct {
	id          int64
	log         *slog.Logger
	mu          sync.Mutex
	peers       []*peerConnectionState
	trackLocals map[string]*webrtc.TrackLocalStaticRTP
}

func newChannelState(id int64, log *slog.Logger) *channelState {
	return &channelState{
		id:          id,
		log:         log,
		trackLocals: make(map[string]*webrtc.TrackLocalStaticRTP),
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
	empty = len(c.peers) == 0 && len(c.trackLocals) == 0
	c.mu.Unlock()
	return removed, empty
}

func (c *channelState) addTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		c.log.Warn("failed to create local track", slog.Int64("channel", c.id), slog.String("error", err.Error()))
		return nil
	}

	c.mu.Lock()
	c.trackLocals[t.ID()] = trackLocal
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

			existingSenders := map[string]bool{}
			for _, sender := range state.peerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				trackID := sender.Track().ID()
				existingSenders[trackID] = true

				if _, ok := c.trackLocals[trackID]; !ok {
					if err := state.peerConnection.RemoveTrack(sender); err != nil {
						c.log.Warn("failed to remove sender", slog.Int64("channel", c.id), slog.String("error", err.Error()))
						return true
					}
				}
			}

			for _, receiver := range state.peerConnection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}
				existingSenders[receiver.Track().ID()] = true
			}

			for id, track := range c.trackLocals {
				if _, ok := existingSenders[id]; !ok {
					if _, err := state.peerConnection.AddTrack(track); err != nil {
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
}

func NewSFU(log *slog.Logger) *SFU {
	return &SFU{
		log:      log,
		channels: make(map[int64]*channelState),
	}
}

func (s *SFU) getOrCreateChannel(channelID int64) *channelState {
	s.mu.Lock()
	ch, ok := s.channels[channelID]
	if !ok {
		ch = newChannelState(channelID, s.log)
		s.channels[channelID] = ch
	}
	s.mu.Unlock()
	return ch
}

func (s *SFU) AddPeer(channelID int64, state *peerConnectionState) *channelState {
	ch := s.getOrCreateChannel(channelID)
	ch.addPeer(state)
	ch.signalPeerConnections()
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
			delete(s.channels, channelID)
		}
		s.mu.Unlock()
	}
}

func (s *SFU) AddTrack(channelID int64, t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	ch := s.getOrCreateChannel(channelID)
	track := ch.addTrack(t)
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

func (s *SFU) RunKeyFrameTicker() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.dispatchKeyFrameAll()
	}
}

func (s *SFU) PeerCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	total := 0
	for _, ch := range s.channels {
		ch.mu.Lock()
		total += len(ch.peers)
		ch.mu.Unlock()
	}
	return total
}
