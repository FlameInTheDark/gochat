package main

import (
	crand "crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
	"go.opentelemetry.io/otel/attribute"

	voicev2 "github.com/FlameInTheDark/gochat/cmd/sfu/signaling/v2"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/voice/dave"
)

func (a *App) authorizeJoinFields(channel helper.StringInt64, token string) (int64, int64, *int64, int64, bool, error) {
	if token == "" || channel == 0 {
		return 0, 0, nil, 0, false, fmt.Errorf("expected join")
	}
	uid, tokChannel, tokGuild, perms, moved, err := a.validateJoinToken(token)
	if err != nil {
		return 0, 0, nil, 0, false, fmt.Errorf("unauthorized")
	}
	if tokChannel != 0 && tokChannel != int64(channel) {
		return 0, 0, nil, 0, false, fmt.Errorf("unauthorized")
	}
	return uid, int64(channel), tokGuild, perms, moved, nil
}

func encodeVoiceGatewayICEServers(servers []webrtc.ICEServer) []voicev2.ICEServer {
	out := make([]voicev2.ICEServer, 0, len(servers))
	for _, server := range servers {
		item := voicev2.ICEServer{
			URLs:     append([]string(nil), server.URLs...),
			Username: server.Username,
		}
		if credential, ok := server.Credential.(string); ok {
			item.Credential = credential
		}
		out = append(out, item)
	}
	return out
}

func supportedVoiceGatewayCodecs(allowAV1 bool) []voicev2.Codec {
	codecs := []voicev2.Codec{
		{Name: "opus", Type: "audio", PayloadType: 111, Priority: 1000},
		{Name: "H264", Type: "video", PayloadType: 102, RTXPayloadType: 103, Priority: 1500},
		{Name: "VP8", Type: "video", PayloadType: 96, RTXPayloadType: 97, Priority: 2000},
		{Name: "VP9", Type: "video", PayloadType: 98, RTXPayloadType: 99, Priority: 3000},
	}
	if allowAV1 {
		codecs = append(codecs, voicev2.Codec{Name: "AV1", Type: "video", PayloadType: 45, RTXPayloadType: 46, Priority: 4000})
	}
	return codecs
}

func (a *App) signalV2ReadyPayload(perms int64) voicev2.Ready {
	return voicev2.Ready{
		ICEServers:          encodeVoiceGatewayICEServers(a.iceConfig.ICEServers),
		SupportedCodecs:     supportedVoiceGatewayCodecs(a.cfg.DAVEAllowAV1),
		CanPublishAudio:     hasPerm(perms, permissions.PermVoiceSpeak),
		CanPublishVideo:     hasPerm(perms, permissions.PermVoiceVideo),
		MaxAudioBitrateKbps: a.cfg.MaxAudioBitrateKbps,
		Experiments:         []string{},
		DAVEEnabled:         a.cfg.DAVEEnabled,
		DAVERequired:        a.cfg.DAVERequiredDefault,
		AllowAV1UnderDAVE:   a.cfg.DAVEAllowAV1,
	}
}

func detectNegotiatedCodecs(sessionSDP string) (string, string) {
	var desc sdp.SessionDescription
	if err := desc.UnmarshalString(sessionSDP); err != nil {
		return "", ""
	}

	var audioCodec string
	var videoCodec string

	for _, md := range desc.MediaDescriptions {
		if md == nil {
			continue
		}

		ptToCodec := make(map[string]string)
		for _, attr := range md.Attributes {
			if !strings.EqualFold(attr.Key, "rtpmap") {
				continue
			}
			parts := strings.Fields(attr.Value)
			if len(parts) < 2 {
				continue
			}
			codecName := strings.SplitN(parts[1], "/", 2)[0]
			ptToCodec[parts[0]] = codecName
		}

		for _, format := range md.MediaName.Formats {
			codecName := ptToCodec[format]
			if codecName == "" {
				continue
			}
			if strings.EqualFold(codecName, "rtx") || strings.EqualFold(codecName, "red") || strings.EqualFold(codecName, "ulpfec") {
				continue
			}

			switch {
			case strings.EqualFold(md.MediaName.Media, "audio") && audioCodec == "":
				audioCodec = codecName
			case strings.EqualFold(md.MediaName.Media, "video") && videoCodec == "":
				videoCodec = codecName
			}
			break
		}
	}

	return audioCodec, videoCodec
}

func waitForGatheredLocalDescription(pc *webrtc.PeerConnection, fallback webrtc.SessionDescription) webrtc.SessionDescription {
	gatheringComplete := webrtc.GatheringCompletePromise(pc)
	select {
	case <-gatheringComplete:
	case <-time.After(2 * time.Second):
	}
	if local := pc.LocalDescription(); local != nil {
		return *local
	}
	return fallback
}

func parseVoiceGatewaySDPType(raw string, fallback webrtc.SDPType) webrtc.SDPType {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case webrtc.SDPTypeOffer.String():
		return webrtc.SDPTypeOffer
	case webrtc.SDPTypeAnswer.String():
		return webrtc.SDPTypeAnswer
	case webrtc.SDPTypePranswer.String():
		return webrtc.SDPTypePranswer
	case webrtc.SDPTypeRollback.String():
		return webrtc.SDPTypeRollback
	default:
		return fallback
	}
}

func (a *App) buildDAVEParticipant(session *signalV2Session) dave.Participant {
	var identityKey *dave.IdentityKey
	if session.identityKey != nil {
		identityKey = &dave.IdentityKey{
			Type:      session.identityKey.Type,
			PublicKey: append([]byte(nil), session.identityKey.PublicKey...),
			Version:   session.identityKey.Version,
		}
	}
	return dave.Participant{
		SessionID:                 session.sessionID,
		UserID:                    session.userID,
		ChannelID:                 session.channelID,
		SignalVersion:             signalProtocolVersion2,
		DAVESupported:             session.supportsDAVE,
		SupportsEncodedTransforms: session.supportsEncodedTransforms,
		MaxDAVEProtocolVersion:    session.maxDAVEProtocolVersion,
		IdentityKey:               identityKey,
	}
}

func (a *App) finalizeSignalV2Session(session *signalV2Session) {
	session.mu.Lock()
	if session.finalized {
		session.mu.Unlock()
		return
	}
	session.finalized = true
	if session.resumeTimer != nil {
		session.resumeTimer.Stop()
		session.resumeTimer = nil
	}
	writer := session.writer
	pc := session.pc
	peerAdded := session.peerAdded
	channelID := session.channelID
	userID := session.userID
	guildID := session.guildID
	startedAt := session.startedAt
	joinNotified := session.joinNotified
	sessionID := session.sessionID
	session.mu.Unlock()

	a.deleteSignalV2Session(sessionID)
	_ = a.dave.Disconnect(sessionID)

	if peerAdded {
		a.sfu.RemovePeer(session.ctx, channelID, pc)
		a.totalPeers.Add(-1)
		a.telemetry.Leave(session.ctx, startedAt, signalPeerAttrs(channelID, userID, guildID)...)
	}
	if joinNotified {
		a.notifyUserLeave(session.ctx, userID, channelID, guildID)
	}
	if pc != nil {
		_ = pc.Close()
	}
	writer.Close()
}

func (a *App) detachSignalV2Session(session *signalV2Session) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.finalized || session.explicitClose || session.phase != signalV2PhaseEstablished {
		return
	}
	session.writer.Detach()
	if session.resumeTimer != nil {
		session.resumeTimer.Stop()
	}
	session.resumeTimer = time.AfterFunc(signalV2ResumeWindow, func() {
		a.finalizeSignalV2Session(session)
	})
}

func (a *App) bindSignalV2Session(session *signalV2Session, conn *websocket.Conn) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.resumeTimer != nil {
		session.resumeTimer.Stop()
		session.resumeTimer = nil
	}
	session.writer.ReplaceConn(conn.Conn)
	session.explicitClose = false
}

func (a *App) closeSignalV2Session(session *signalV2Session, code int, reason string) error {
	session.mu.Lock()
	session.explicitClose = true
	session.mu.Unlock()
	return session.writer.SendClose(code, reason)
}

func signalPeerAttrs(channelID, userID int64, guildID *int64) []attribute.KeyValue {
	attrs := []attribute.KeyValue{}
	if channelID != 0 {
		attrs = append(attrs, attribute.Int64("voice.channel_id", channelID))
	}
	if userID != 0 {
		attrs = append(attrs, attribute.Int64("user.id", userID))
	}
	if guildID != nil {
		attrs = append(attrs, attribute.Int64("guild.id", *guildID))
	}
	return attrs
}

func newSignalSessionID() string {
	var b [16]byte
	if _, err := crand.Read(b[:]); err == nil {
		b[6] = (b[6] & 0x0f) | 0x40
		b[8] = (b[8] & 0x3f) | 0x80
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]),
			uint16(b[4])<<8|uint16(b[5]),
			uint16(b[6])<<8|uint16(b[7]),
			uint16(b[8])<<8|uint16(b[9]),
			uint64(b[10])<<40|uint64(b[11])<<32|uint64(b[12])<<24|uint64(b[13])<<16|uint64(b[14])<<8|uint64(b[15]),
		)
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func isExpectedWSReadError(err error) bool {
	if err == nil {
		return false
	}

	if websocket.IsCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseProtocolError,
		websocket.CloseNoStatusReceived,
		websocket.CloseGoingAway,
		websocket.CloseAbnormalClosure,
	) {
		return true
	}

	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, net.ErrClosed)
}
