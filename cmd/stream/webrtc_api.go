package main

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

var videoRTCPFeedback = []webrtc.RTCPFeedback{
	{Type: "goog-remb"},
	{Type: "ccm", Parameter: "fir"},
	{Type: "nack"},
	{Type: "nack", Parameter: "pli"},
	{Type: "transport-cc"},
}

// buildWebRTCAPI creates a WebRTC API tuned for live voice/video publishing.
// We keep the codec list explicit so signaling can advertise the same payload
// types, while still enabling RTX, NACK/PLI, TWCC and the rest of Pion's
// default interceptor stack that browsers rely on for stable 720p30 delivery.
func buildWebRTCAPI(logger *slog.Logger, allowAV1 bool, icePublicIP string, udpPortRangeStart, udpPortRangeEnd int) (*webrtc.API, error) {
	me := &webrtc.MediaEngine{}
	registerSFUCodecs(logger, me, allowAV1)

	ir := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(me, ir); err != nil {
		logger.Error("failed to register default webrtc interceptors", slog.String("error", err.Error()))
	}

	settingEngine := webrtc.SettingEngine{}
	if udpPortRangeStart != 0 || udpPortRangeEnd != 0 {
		if udpPortRangeStart == 0 || udpPortRangeEnd == 0 {
			return nil, fmt.Errorf("udp port range must set both start and end")
		}
		if err := settingEngine.SetEphemeralUDPPortRange(uint16(udpPortRangeStart), uint16(udpPortRangeEnd)); err != nil {
			return nil, fmt.Errorf("set udp port range: %w", err)
		}
		logger.Info("configured webrtc udp port range",
			slog.Int("start", udpPortRangeStart),
			slog.Int("end", udpPortRangeEnd),
		)
	}
	if publicIP := strings.TrimSpace(icePublicIP); publicIP != "" {
		settingEngine.SetNAT1To1IPs([]string{publicIP}, webrtc.ICECandidateTypeHost)
		logger.Info("configured webrtc public ice ip",
			slog.String("public_ip", publicIP),
			slog.String("candidate_type", webrtc.ICECandidateTypeHost.String()),
		)
	}

	return webrtc.NewAPI(
		webrtc.WithMediaEngine(me),
		webrtc.WithInterceptorRegistry(ir),
		webrtc.WithSettingEngine(settingEngine),
	), nil
}

func registerSFUCodecs(logger *slog.Logger, me *webrtc.MediaEngine, allowAV1 bool) {
	registerCodec(logger, me, "Opus", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeOpus,
			ClockRate:   48000,
			Channels:    2,
			SDPFmtpLine: "minptime=10;useinbandfec=1",
		},
		PayloadType: 111,
	}, webrtc.RTPCodecTypeAudio)

	if allowAV1 {
		registerCodec(logger, me, "AV1", webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeAV1,
				ClockRate:    90000,
				RTCPFeedback: videoRTCPFeedback,
			},
			PayloadType: 45,
		}, webrtc.RTPCodecTypeVideo)
		registerCodec(logger, me, "RTX for AV1", webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:    webrtc.MimeTypeRTX,
				ClockRate:   90000,
				SDPFmtpLine: "apt=45",
			},
			PayloadType: 46,
		}, webrtc.RTPCodecTypeVideo)
	}

	registerCodec(logger, me, "H264", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
			RTCPFeedback: videoRTCPFeedback,
		},
		PayloadType: 102,
	}, webrtc.RTPCodecTypeVideo)
	registerCodec(logger, me, "RTX for H264", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeRTX,
			ClockRate:   90000,
			SDPFmtpLine: "apt=102",
		},
		PayloadType: 103,
	}, webrtc.RTPCodecTypeVideo)
	registerCodec(logger, me, "H264 packetization-mode 0", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42001f",
			RTCPFeedback: videoRTCPFeedback,
		},
		PayloadType: 104,
	}, webrtc.RTPCodecTypeVideo)
	registerCodec(logger, me, "RTX for H264 packetization-mode 0", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeRTX,
			ClockRate:   90000,
			SDPFmtpLine: "apt=104",
		},
		PayloadType: 105,
	}, webrtc.RTPCodecTypeVideo)
	registerCodec(logger, me, "H264 constrained baseline", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
			RTCPFeedback: videoRTCPFeedback,
		},
		PayloadType: 106,
	}, webrtc.RTPCodecTypeVideo)
	registerCodec(logger, me, "RTX for H264 constrained baseline", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeRTX,
			ClockRate:   90000,
			SDPFmtpLine: "apt=106",
		},
		PayloadType: 107,
	}, webrtc.RTPCodecTypeVideo)
	registerCodec(logger, me, "H264 constrained baseline packetization-mode 0", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42e01f",
			RTCPFeedback: videoRTCPFeedback,
		},
		PayloadType: 108,
	}, webrtc.RTPCodecTypeVideo)
	registerCodec(logger, me, "RTX for H264 constrained baseline packetization-mode 0", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeRTX,
			ClockRate:   90000,
			SDPFmtpLine: "apt=108",
		},
		PayloadType: 109,
	}, webrtc.RTPCodecTypeVideo)

	registerCodec(logger, me, "VP8", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeVP8,
			ClockRate:    90000,
			RTCPFeedback: videoRTCPFeedback,
		},
		PayloadType: 96,
	}, webrtc.RTPCodecTypeVideo)
	registerCodec(logger, me, "RTX for VP8", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeRTX,
			ClockRate:   90000,
			SDPFmtpLine: "apt=96",
		},
		PayloadType: 97,
	}, webrtc.RTPCodecTypeVideo)

	registerCodec(logger, me, "VP9", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeVP9,
			ClockRate:    90000,
			SDPFmtpLine:  "profile-id=0",
			RTCPFeedback: videoRTCPFeedback,
		},
		PayloadType: 98,
	}, webrtc.RTPCodecTypeVideo)
	registerCodec(logger, me, "RTX for VP9", webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeRTX,
			ClockRate:   90000,
			SDPFmtpLine: "apt=98",
		},
		PayloadType: 99,
	}, webrtc.RTPCodecTypeVideo)
}

func registerCodec(logger *slog.Logger, me *webrtc.MediaEngine, label string, codec webrtc.RTPCodecParameters, typ webrtc.RTPCodecType) {
	if err := me.RegisterCodec(codec, typ); err != nil {
		logger.Error("failed to register codec", slog.String("codec", label), slog.String("error", err.Error()))
	}
}
