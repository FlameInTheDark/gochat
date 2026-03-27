package main

import (
	"log/slog"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

var videoRTCPFeedback = []webrtc.RTCPFeedback{
	{Type: "goog-remb"},
	{Type: "ccm", Parameter: "fir"},
	{Type: "nack"},
	{Type: "nack", Parameter: "pli"},
}

// buildWebRTCAPI creates a WebRTC API tuned for live voice/video publishing.
// We keep the codec list explicit so signaling can advertise the same payload
// types, while still enabling RTX, NACK/PLI, TWCC and the rest of Pion's
// default interceptor stack that browsers rely on for stable 720p30 delivery.
func buildWebRTCAPI(logger *slog.Logger, allowAV1 bool) *webrtc.API {
	me := &webrtc.MediaEngine{}
	registerSFUCodecs(logger, me, allowAV1)

	ir := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(me, ir); err != nil {
		logger.Error("failed to register default webrtc interceptors", slog.String("error", err.Error()))
	}

	return webrtc.NewAPI(
		webrtc.WithMediaEngine(me),
		webrtc.WithInterceptorRegistry(ir),
	)
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
}

func registerCodec(logger *slog.Logger, me *webrtc.MediaEngine, label string, codec webrtc.RTPCodecParameters, typ webrtc.RTPCodecType) {
	if err := me.RegisterCodec(codec, typ); err != nil {
		logger.Error("failed to register codec", slog.String("codec", label), slog.String("error", err.Error()))
	}
}
