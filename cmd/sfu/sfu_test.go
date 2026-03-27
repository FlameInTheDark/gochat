package main

import (
	"io"
	"log/slog"
	"strings"
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

	api := buildWebRTCAPI(newTestLogger(), false)
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

func TestPreparePeerInitialSync_InitialAnswerContainsExistingTracks(t *testing.T) {
	ch := newTestChannelState()
	pc := newTestPeerConnection(t)
	state := &peerConnectionState{
		peerConnection: pc,
		websocket:      &threadSafeWriter{},
		userID:         2,
	}

	track := newTestTrack(t, "1-video", "u:1")
	ch.trackLocals[track.ID()] = trackLocalEntry{track: track, owner: 1, kind: webrtc.RTPCodecTypeVideo.String()}

	offerPC := newTestPeerConnection(t)
	offer, err := offerPC.CreateOffer(nil)
	if err != nil {
		t.Fatalf("create offer: %v", err)
	}
	if err := offerPC.SetLocalDescription(offer); err != nil {
		t.Fatalf("set offer local description: %v", err)
	}
	if err := pc.SetRemoteDescription(offer); err != nil {
		t.Fatalf("set remote description: %v", err)
	}

	_ = ch.preparePeerInitialSync(state)

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		t.Fatalf("create answer: %v", err)
	}
	if err := pc.SetLocalDescription(answer); err != nil {
		t.Fatalf("set local description: %v", err)
	}

	foundSender := false
	for _, sender := range pc.GetSenders() {
		if sender.Track() == nil {
			continue
		}
		if sender.Track().ID() == track.ID() {
			foundSender = true
			break
		}
	}
	if !foundSender {
		t.Fatal("expected initial sync to add existing track sender before answering")
	}
	if !strings.Contains(answer.SDP, "u:1") {
		t.Fatalf("expected answer SDP to reference existing stream id, got:\n%s", answer.SDP)
	}
}

func TestDoSignalPeerConnections_V2BootstrappedPeerRenegotiatesOnTopologyChange(t *testing.T) {
	ch := newTestChannelState()
	pc := newTestPeerConnection(t)

	state := &peerConnectionState{
		peerConnection:  pc,
		websocket:       &threadSafeWriter{},
		userID:          2,
		negotiated:      true,
		offeredRevision: 1,
		appliedRevision: 1,
	}
	ch.peers = []*peerConnectionState{state}
	ch.topologyRevision = 1

	ch.doSignalPeerConnections()

	if pc.LocalDescription() != nil {
		t.Fatal("did not expect immediate renegotiation without a topology change")
	}

	track := newTestTrack(t, "1-video", "u:1")
	ch.trackLocals[track.ID()] = trackLocalEntry{track: track, owner: 1, kind: webrtc.RTPCodecTypeVideo.String()}
	ch.topologyRevision = 2

	ch.doSignalPeerConnections()

	if pc.LocalDescription() == nil {
		t.Fatal("expected renegotiation offer after topology changed for a v2-bootstrapped peer")
	}
	if state.offeredRevision != 2 {
		t.Fatalf("expected offered revision 2, got %d", state.offeredRevision)
	}
}

func TestBuildWebRTCAPIAdvertisesVideoFeedbackAndRTX(t *testing.T) {
	pc := newTestPeerConnection(t)

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		t.Fatalf("create offer: %v", err)
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		t.Fatalf("set local description: %v", err)
	}

	sdp := pc.LocalDescription().SDP
	for _, want := range []string{
		"a=rtpmap:102 H264/90000",
		"a=rtpmap:97 rtx/90000",
		"a=rtpmap:103 rtx/90000",
		"a=rtcp-fb:96 nack pli",
		"a=rtcp-fb:96 transport-cc",
		"a=rtcp-fb:102 nack pli",
	} {
		if !strings.Contains(sdp, want) {
			t.Fatalf("expected offer SDP to contain %q, got:\n%s", want, sdp)
		}
	}
}

func TestSupportedVoiceGatewayCodecsMatchNegotiatedPayloadTypes(t *testing.T) {
	codecs := supportedVoiceGatewayCodecs(true)
	byName := make(map[string]struct {
		payload uint8
		rtx     uint8
	})
	for _, codec := range codecs {
		byName[codec.Name] = struct {
			payload uint8
			rtx     uint8
		}{
			payload: codec.PayloadType,
			rtx:     codec.RTXPayloadType,
		}
	}

	for name, want := range map[string]struct {
		payload uint8
		rtx     uint8
	}{
		"opus": {payload: 111},
		"H264": {payload: 102, rtx: 103},
		"VP8":  {payload: 96, rtx: 97},
		"VP9":  {payload: 98, rtx: 99},
		"AV1":  {payload: 45, rtx: 46},
	} {
		got, ok := byName[name]
		if !ok {
			t.Fatalf("expected codec %q to be advertised", name)
		}
		if got.payload != want.payload || got.rtx != want.rtx {
			t.Fatalf("codec %q payloads = (%d, %d), want (%d, %d)", name, got.payload, got.rtx, want.payload, want.rtx)
		}
	}
}
