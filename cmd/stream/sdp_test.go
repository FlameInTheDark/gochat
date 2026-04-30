package main

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/pion/webrtc/v4"
)

func newStreamTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestLimitVideoBitrateInSDPAddsOnlyVideoBandwidth(t *testing.T) {
	const input = "v=0\r\n" +
		"o=- 0 0 IN IP4 127.0.0.1\r\n" +
		"s=-\r\n" +
		"t=0 0\r\n" +
		"m=audio 9 UDP/TLS/RTP/SAVPF 111\r\n" +
		"c=IN IP4 0.0.0.0\r\n" +
		"a=rtpmap:111 opus/48000/2\r\n" +
		"m=video 9 UDP/TLS/RTP/SAVPF 96\r\n" +
		"c=IN IP4 0.0.0.0\r\n" +
		"a=rtpmap:96 VP8/90000\r\n"

	out := limitVideoBitrateInSDP(input, 100_000_000)

	audioSection := mediaSection(t, out, "audio")
	if strings.Contains(audioSection, "b=TIAS:") || strings.Contains(audioSection, "b=AS:") {
		t.Fatalf("audio section should not receive video bandwidth limits:\n%s", audioSection)
	}

	videoSection := mediaSection(t, out, "video")
	for _, want := range []string{"b=TIAS:100000000", "b=AS:100000"} {
		if !strings.Contains(videoSection, want) {
			t.Fatalf("expected video section to contain %q, got:\n%s", want, videoSection)
		}
	}
}

func TestLimitVideoBitrateInSDPUpdatesExistingBandwidth(t *testing.T) {
	const input = "v=0\r\n" +
		"o=- 0 0 IN IP4 127.0.0.1\r\n" +
		"s=-\r\n" +
		"t=0 0\r\n" +
		"m=video 9 UDP/TLS/RTP/SAVPF 96\r\n" +
		"c=IN IP4 0.0.0.0\r\n" +
		"b=AS:2500\r\n" +
		"b=TIAS:2500000\r\n" +
		"a=rtpmap:96 VP8/90000\r\n"

	videoSection := mediaSection(t, limitVideoBitrateInSDP(input, 60_000_000), "video")
	for _, want := range []string{"b=AS:60000", "b=TIAS:60000000"} {
		if !strings.Contains(videoSection, want) {
			t.Fatalf("expected video section to contain %q, got:\n%s", want, videoSection)
		}
	}
	for _, old := range []string{"b=AS:2500", "b=TIAS:2500000"} {
		if strings.Contains(videoSection, old) {
			t.Fatalf("expected old bandwidth %q to be replaced, got:\n%s", old, videoSection)
		}
	}
}

func TestBuildWebRTCAPIAdvertisesStreamVideoTransportCC(t *testing.T) {
	pc := newStreamTestPeerConnection(t)

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		t.Fatalf("create offer: %v", err)
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		t.Fatalf("set local description: %v", err)
	}

	sdp := pc.LocalDescription().SDP
	for _, want := range []string{
		"a=rtcp-fb:96 transport-cc",
		"a=rtcp-fb:102 transport-cc",
	} {
		if !strings.Contains(sdp, want) {
			t.Fatalf("expected offer SDP to contain %q, got:\n%s", want, sdp)
		}
	}
}

func TestSupportedStreamGatewayCodecsPreferAV1WhenAllowed(t *testing.T) {
	codecs := supportedVoiceGatewayCodecs(true)
	var videoCodecs []string
	for _, codec := range codecs {
		if codec.Type == "video" {
			videoCodecs = append(videoCodecs, codec.Name)
		}
	}

	want := []string{"AV1", "VP9", "VP8", "H264"}
	if strings.Join(videoCodecs, ",") != strings.Join(want, ",") {
		t.Fatalf("video codecs = %v, want %v", videoCodecs, want)
	}

	codecs = supportedVoiceGatewayCodecs(false)
	for _, codec := range codecs {
		if codec.Name == "AV1" {
			t.Fatalf("AV1 should not be advertised when disabled: %v", codecs)
		}
	}
}

func TestBuildWebRTCAPIAdvertisesAV1FirstWhenAllowed(t *testing.T) {
	api, err := buildWebRTCAPI(newStreamTestLogger(), true, "", 0, 0)
	if err != nil {
		t.Fatalf("build webrtc api: %v", err)
	}
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatalf("create peer connection: %v", err)
	}
	t.Cleanup(func() {
		_ = pc.Close()
	})

	if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendrecv,
	}); err != nil {
		t.Fatalf("add video transceiver: %v", err)
	}

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		t.Fatalf("create offer: %v", err)
	}
	videoSection := mediaSection(t, offer.SDP, "video")
	if !strings.Contains(videoSection, "a=rtpmap:45 AV1/90000") {
		t.Fatalf("expected AV1 payload in offer, got:\n%s", videoSection)
	}
	if !strings.HasPrefix(videoSection, "m=video 9 UDP/TLS/RTP/SAVPF 45 ") {
		t.Fatalf("expected AV1 payload to be first in video m-line, got:\n%s", videoSection)
	}
}

func TestViewerOfferCarriesConfiguredVideoBitrate(t *testing.T) {
	pc := newStreamTestPeerConnection(t)
	track, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeVP8,
			ClockRate:    90000,
			RTCPFeedback: videoRTCPFeedback,
		},
		"publisher-video",
		"111",
	)
	if err != nil {
		t.Fatalf("create local track: %v", err)
	}

	ch := &channelState{
		id:                 123,
		log:                newStreamTestLogger(),
		trackLocals:        map[string]trackLocalEntry{track.ID(): {track: track, owner: 111, kind: webrtc.RTPCodecTypeVideo.String()}},
		maxVideoBitrateBps: 100_000_000,
	}
	state := &peerConnectionState{peerConnection: pc, userID: 222}
	if !ch.syncPeerSenders(state) {
		t.Fatal("expected viewer sender layout to change")
	}

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		t.Fatalf("create viewer offer: %v", err)
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		t.Fatalf("set local description: %v", err)
	}

	offerToSend := offer
	offerToSend.SDP = limitVideoBitrateInSDP(offerToSend.SDP, ch.maxVideoBitrateBps)
	videoSection := mediaSection(t, offerToSend.SDP, "video")
	for _, want := range []string{"b=TIAS:100000000", "b=AS:100000"} {
		if !strings.Contains(videoSection, want) {
			t.Fatalf("expected viewer offer video section to contain %q, got:\n%s", want, videoSection)
		}
	}
}

func newStreamTestPeerConnection(t *testing.T) *webrtc.PeerConnection {
	t.Helper()

	api, err := buildWebRTCAPI(newStreamTestLogger(), false, "", 0, 0)
	if err != nil {
		t.Fatalf("build webrtc api: %v", err)
	}
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatalf("create peer connection: %v", err)
	}
	t.Cleanup(func() {
		_ = pc.Close()
	})

	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := pc.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendrecv,
		}); err != nil {
			t.Fatalf("add transceiver %s: %v", typ, err)
		}
	}
	return pc
}

func mediaSection(t *testing.T, sdpIn, media string) string {
	t.Helper()

	parts := strings.Split(sdpIn, "\r\nm=")
	for i, part := range parts {
		section := part
		if i > 0 {
			section = "m=" + section
		}
		if strings.HasPrefix(section, "m="+media+" ") {
			return section
		}
	}
	t.Fatalf("missing %s media section in:\n%s", media, sdpIn)
	return ""
}
