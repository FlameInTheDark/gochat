package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/pion/webrtc/v4"
	"go.opentelemetry.io/otel/attribute"
)

// setupTransceivers adds audio and video sendrecv transceivers to the peer connection.
func (a *App) setupTransceivers(pc *webrtc.PeerConnection) error {
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := pc.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionSendrecv,
		}); err != nil {
			return fmt.Errorf("add transceiver %s: %w", typ, err)
		}
	}
	return nil
}

// registerPeerCallbacks sets up OnICECandidate, OnConnectionStateChange, and OnTrack.
func (a *App) registerPeerCallbacks(
	ctx context.Context,
	pc *webrtc.PeerConnection,
	writer *threadSafeWriter,
	state *peerConnectionState,
	uid, channelID, perms int64,
) {
	peerAttrs := []attribute.KeyValue{
		attribute.Int64("voice.channel_id", channelID),
		attribute.Int64("user.id", uid),
	}
	pc.OnICECandidate(func(i *webrtc.ICECandidate) {
		if state.signalVersion == signalProtocolVersion2 {
			return
		}
		if err := writer.SendRTCCandidate(i); err != nil {
			a.log.Warn("failed to send candidate", slog.String("error", err.Error()))
			return
		}
		if i != nil {
			a.telemetry.Candidate(ctx, "outbound", peerAttrs...)
		}
	})

	pc.OnConnectionStateChange(func(connState webrtc.PeerConnectionState) {
		a.telemetry.ConnectionState(ctx, connState.String(), peerAttrs...)
		a.log.Info("connection state change", slog.String("state", connState.String()), slog.Int64("user", uid))
		switch connState {
		case webrtc.PeerConnectionStateFailed:
			if err := pc.Close(); err != nil {
				a.log.Warn("failed to close peer connection", slog.String("error", err.Error()))
			}
		case webrtc.PeerConnectionStateClosed:
			a.sfu.SignalChannel(ctx, channelID)
		}
	})

	pc.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		a.handleInboundTrack(ctx, pc, state, t, uid, channelID, perms)
	})
}

// parseSpeakingData extracts a speaking indicator (0 or 1) from various payload formats.
func parseSpeakingData(data string) int {
	switch data {
	case "1", "\"1\"":
		return 1
	case "0", "\"0\"", "":
		return 0
	}
	var aux struct {
		Speaking int `json:"speaking"`
	}
	_ = json.Unmarshal([]byte(data), &aux)
	if aux.Speaking != 0 {
		return 1
	}
	return 0
}

// parseLegacySDPType maps a string SDP type to webrtc.SDPType, defaulting to answer.
func parseLegacySDPType(t string) webrtc.SDPType {
	switch {
	case strings.EqualFold(t, webrtc.SDPTypeOffer.String()):
		return webrtc.SDPTypeOffer
	case strings.EqualFold(t, webrtc.SDPTypePranswer.String()):
		return webrtc.SDPTypePranswer
	case strings.EqualFold(t, webrtc.SDPTypeRollback.String()):
		return webrtc.SDPTypeRollback
	default:
		return webrtc.SDPTypeAnswer
	}
}
