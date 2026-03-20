package main

import (
	"io"
	"log/slog"
	"testing"

	"github.com/pion/webrtc/v4"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestChannelState() *channelState {
	return &channelState{
		id:           1,
		log:          newTestLogger(),
		trackLocals:  make(map[string]trackLocalEntry),
		blockedUsers: make(map[int64]bool),
	}
}

func newTestPeerConnection(t *testing.T) *webrtc.PeerConnection {
	t.Helper()

	api := buildWebRTCAPI(newTestLogger())
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatalf("create peer connection: %v", err)
	}
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := pc.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendrecv,
		}); err != nil {
			t.Fatalf("add transceiver %s: %v", typ, err)
		}
	}
	t.Cleanup(func() {
		_ = pc.Close()
	})
	return pc
}

func newTestTrack(t *testing.T, id, stream string) *webrtc.TrackLocalStaticRTP {
	t.Helper()

	track, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000},
		id,
		stream,
	)
	if err != nil {
		t.Fatalf("create track: %v", err)
	}
	return track
}

func TestDoSignalPeerConnections_TargetedNegotiateOnlyOffersRequestingPeer(t *testing.T) {
	ch := newTestChannelState()
	pcA := newTestPeerConnection(t)
	pcB := newTestPeerConnection(t)

	stateA := &peerConnectionState{
		peerConnection:  pcA,
		websocket:       &threadSafeWriter{},
		userID:          1,
		negotiated:      true,
		offeredRevision: 1,
		appliedRevision: 1,
	}
	stateB := &peerConnectionState{
		peerConnection:  pcB,
		websocket:       &threadSafeWriter{},
		userID:          2,
		negotiated:      true,
		offeredRevision: 1,
		appliedRevision: 1,
	}
	ch.peers = []*peerConnectionState{stateA, stateB}
	ch.topologyRevision = 1

	if !ch.requestPeerNegotiation(pcA) {
		t.Fatal("expected targeted negotiation request to find peer")
	}

	ch.doSignalPeerConnections()

	if pcA.LocalDescription() == nil {
		t.Fatal("expected requesting peer to receive a fresh offer")
	}
	if stateA.forceOffer {
		t.Fatal("expected targeted request to be cleared after offer creation")
	}
	if pcB.LocalDescription() != nil {
		t.Fatal("did not expect non-requesting peer to receive an offer")
	}
}

func TestDoSignalPeerConnections_NoOpTopologyChangeAdvancesRevisionWithoutOffer(t *testing.T) {
	ch := newTestChannelState()
	pc := newTestPeerConnection(t)
	for _, sender := range pc.GetSenders() {
		if sender.Track() == nil {
			continue
		}
		if err := pc.RemoveTrack(sender); err != nil {
			t.Fatalf("remove reserved sender: %v", err)
		}
	}

	state := &peerConnectionState{
		peerConnection:  pc,
		websocket:       &threadSafeWriter{},
		userID:          1,
		negotiated:      true,
		offeredRevision: 1,
		appliedRevision: 1,
	}
	ch.peers = []*peerConnectionState{state}
	ch.topologyRevision = 2

	ch.doSignalPeerConnections()

	if pc.LocalDescription() != nil {
		t.Fatal("expected no-op topology change to avoid a new offer")
	}
	if state.appliedRevision != 2 {
		t.Fatalf("expected peer revision to advance without offer, got %d", state.appliedRevision)
	}
}

func TestApplyPeerAnswer_ResignalsWhenTopologyAdvancedWhilePeerWasBehind(t *testing.T) {
	ch := newTestChannelState()
	pc := newTestPeerConnection(t)

	state := &peerConnectionState{
		peerConnection:  pc,
		websocket:       &threadSafeWriter{},
		userID:          2,
		offeredRevision: 1,
	}
	ch.peers = []*peerConnectionState{state}
	ch.topologyRevision = 2
	track := newTestTrack(t, "1-video", "u:1")
	ch.trackLocals[track.ID()] = trackLocalEntry{track: track, owner: 1, kind: webrtc.RTPCodecTypeVideo.String()}

	needsSignal, ok := ch.applyPeerAnswer(pc)
	if !ok {
		t.Fatal("expected answer bookkeeping to find peer")
	}
	if !needsSignal {
		t.Fatal("expected answer bookkeeping to request another sync pass")
	}
	if !state.negotiated {
		t.Fatal("expected peer to be marked negotiated after answer")
	}
	if state.appliedRevision != 1 {
		t.Fatalf("expected applied revision to match last offer, got %d", state.appliedRevision)
	}

	ch.doSignalPeerConnections()

	if pc.LocalDescription() == nil {
		t.Fatal("expected follow-up offer after missed topology change")
	}
	if state.offeredRevision != 2 {
		t.Fatalf("expected follow-up offer to carry latest revision, got %d", state.offeredRevision)
	}
}

func TestDoSignalPeerConnections_NewPeerGetsInitialOfferForExistingTracks(t *testing.T) {
	ch := newTestChannelState()
	pc := newTestPeerConnection(t)

	state := &peerConnectionState{
		peerConnection: pc,
		websocket:      &threadSafeWriter{},
		userID:         2,
	}
	ch.peers = []*peerConnectionState{state}
	ch.topologyRevision = 1
	track := newTestTrack(t, "1-video", "u:1")
	ch.trackLocals[track.ID()] = trackLocalEntry{track: track, owner: 1, kind: webrtc.RTPCodecTypeVideo.String()}

	ch.doSignalPeerConnections()

	if pc.LocalDescription() == nil {
		t.Fatal("expected new peer to receive initial offer with current tracks")
	}
	if state.offeredRevision != 1 {
		t.Fatalf("expected initial offer revision 1, got %d", state.offeredRevision)
	}
}
