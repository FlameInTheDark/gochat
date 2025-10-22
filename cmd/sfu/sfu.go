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

type SFU struct {
	log         *slog.Logger
	mu          sync.Mutex
	peers       []*peerConnectionState
	trackLocals map[string]*webrtc.TrackLocalStaticRTP
}

func NewSFU(log *slog.Logger) *SFU {
	return &SFU{
		log:         log,
		trackLocals: make(map[string]*webrtc.TrackLocalStaticRTP),
	}
}

func (s *SFU) AddPeer(state *peerConnectionState) {
	s.mu.Lock()
	s.peers = append(s.peers, state)
	s.mu.Unlock()
}

func (s *SFU) RemovePeer(pc *webrtc.PeerConnection) {
	s.mu.Lock()
	removed := false
	for i := range s.peers {
		if s.peers[i].peerConnection == pc {
			s.peers = append(s.peers[:i], s.peers[i+1:]...)
			removed = true
			break
		}
	}
	s.mu.Unlock()
	if removed {
		s.SignalPeerConnections()
	}
}

func (s *SFU) AddTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		s.SignalPeerConnections()
	}()

	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		s.log.Warn("failed to create local track", slog.String("error", err.Error()))
		return nil
	}

	s.trackLocals[t.ID()] = trackLocal
	return trackLocal
}

func (s *SFU) RemoveTrack(track *webrtc.TrackLocalStaticRTP) {
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		s.SignalPeerConnections()
	}()

	delete(s.trackLocals, track.ID())
}

func (s *SFU) SignalPeerConnections() {
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		s.dispatchKeyFrame()
	}()

	attemptSync := func() bool {
		for i := 0; i < len(s.peers); i++ {
			state := s.peers[i]
			if state.peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				s.peers = append(s.peers[:i], s.peers[i+1:]...)
				return true
			}

			existingSenders := map[string]bool{}
			for _, sender := range state.peerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				trackID := sender.Track().ID()
				existingSenders[trackID] = true

				if _, ok := s.trackLocals[trackID]; !ok {
					if err := state.peerConnection.RemoveTrack(sender); err != nil {
						s.log.Warn("failed to remove sender", slog.String("error", err.Error()))
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

			for id, track := range s.trackLocals {
				if _, ok := existingSenders[id]; !ok {
					if _, err := state.peerConnection.AddTrack(track); err != nil {
						s.log.Warn("failed to add track to peer", slog.String("error", err.Error()))
						return true
					}
				}
			}

			offer, err := state.peerConnection.CreateOffer(nil)
			if err != nil {
				s.log.Warn("failed to create offer", slog.String("error", err.Error()))
				return true
			}
			if err = state.peerConnection.SetLocalDescription(offer); err != nil {
				s.log.Warn("failed to set local description", slog.String("error", err.Error()))
				return true
			}

			if err = state.websocket.SendRTCOffer(offer); err != nil {
				s.log.Warn("failed to send offer", slog.String("error", err.Error()))
				return true
			}
		}
		return false
	}

	for attempts := 0; ; attempts++ {
		if attempts == 25 {
			go func() {
				time.Sleep(3 * time.Second)
				s.SignalPeerConnections()
			}()
			return
		}
		if !attemptSync() {
			break
		}
	}
}

func (s *SFU) dispatchKeyFrame() {
	s.mu.Lock()
	peers := append([]*peerConnectionState(nil), s.peers...)
	s.mu.Unlock()

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

func (s *SFU) RunKeyFrameTicker() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.dispatchKeyFrame()
	}
}

func (s *SFU) PeerCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.peers)
}
