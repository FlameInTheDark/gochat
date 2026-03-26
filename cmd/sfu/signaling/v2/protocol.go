package v2

import (
	"encoding/json"

	"github.com/FlameInTheDark/gochat/internal/helper"
)

const (
	ProtocolVersion = 2
	MaxDAVEProtocol = 1
)

const (
	OpIdentify                 = 0
	OpSelectProtocol           = 1
	OpReady                    = 2
	OpHeartbeat                = 3
	OpSessionDescription       = 4
	OpSpeaking                 = 5
	OpHeartbeatACK             = 6
	OpResume                   = 7
	OpHello                    = 8
	OpResumed                  = 9
	OpClientsConnect           = 11
	OpClientDisconnect         = 13
	OpDAVEPrepareTransition    = 21
	OpDAVEExecuteTransition    = 22
	OpDAVETransitionReady      = 23
	OpDAVEPrepareEpoch         = 24
	OpDAVEInvalidCommitWelcome = 31
)

const (
	CloseCodeUnknownOpcode     = 4001
	CloseCodeInvalidPayload    = 4002
	CloseCodeUnauthorized      = 4003
	CloseCodeHeartbeatTimeout  = 4009
	CloseCodeSessionExpired    = 4016
	CloseCodeDAVERequired      = 4017
	CloseCodeWrongPhase        = 4020
	CloseCodeUnsupportedMedium = 4021
)

type Packet struct {
	Op  int `json:"op"`
	Seq int `json:"seq,omitempty"`
	D   any `json:"d,omitempty"`
}

type IncomingPacket struct {
	Op  int             `json:"op"`
	Seq int             `json:"seq,omitempty"`
	D   json.RawMessage `json:"d"`
}

type Hello struct {
	V                 int    `json:"v"`
	HeartbeatInterval int64  `json:"heartbeat_interval"`
	SessionID         string `json:"session_id"`
}

type Identify struct {
	ServerID                  string             `json:"server_id,omitempty"`
	ChannelID                 helper.StringInt64 `json:"channel_id"`
	UserID                    string             `json:"user_id,omitempty"`
	SessionID                 string             `json:"session_id,omitempty"`
	Token                     string             `json:"token"`
	MaxDAVEProtocolVersion    int                `json:"max_dave_protocol_version"`
	SupportsEncodedTransforms bool               `json:"supports_encoded_transforms"`
	DAVESupported             bool               `json:"dave_supported,omitempty"`
	Video                     bool               `json:"video,omitempty"`
	Streams                   []Stream           `json:"streams,omitempty"`
	IdentityKey               *IdentityKey       `json:"identity_key,omitempty"`
}

type Resume struct {
	SessionID string             `json:"session_id"`
	ChannelID helper.StringInt64 `json:"channel_id"`
	Token     string             `json:"token"`
}

type Resumed struct {
	SessionID           string `json:"session_id"`
	DAVEProtocolVersion int    `json:"dave_protocol_version"`
	DAVEEpoch           uint64 `json:"dave_epoch"`
}

type Stream struct {
	Type    string `json:"type,omitempty"`
	RID     string `json:"rid,omitempty"`
	Quality int    `json:"quality,omitempty"`
}

type Codec struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	PayloadType    uint8  `json:"payload_type,omitempty"`
	RTXPayloadType uint8  `json:"rtx_payload_type,omitempty"`
	Priority       int    `json:"priority,omitempty"`
}

type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type IdentityKey struct {
	Type      string `json:"type,omitempty"`
	PublicKey []byte `json:"public_key,omitempty"`
	Version   uint32 `json:"version,omitempty"`
}

type Ready struct {
	ICEServers          []ICEServer `json:"ice_servers"`
	SupportedCodecs     []Codec     `json:"supported_codecs"`
	CanPublishAudio     bool        `json:"can_publish_audio"`
	CanPublishVideo     bool        `json:"can_publish_video"`
	MaxAudioBitrateKbps int         `json:"max_audio_bitrate_kbps"`
	Experiments         []string    `json:"experiments"`
	DAVEEnabled         bool        `json:"dave_enabled"`
	DAVERequired        bool        `json:"dave_required"`
	AllowAV1UnderDAVE   bool        `json:"allow_av1_under_dave"`
}

type SelectProtocol struct {
	Protocol        string   `json:"protocol"`
	Data            string   `json:"data,omitempty"`
	SDP             string   `json:"sdp"`
	Type            string   `json:"type,omitempty"`
	RTCConnectionID string   `json:"rtc_connection_id"`
	Codecs          []Codec  `json:"codecs,omitempty"`
	Streams         []Stream `json:"streams,omitempty"`
}

type SessionDescription struct {
	Type                string `json:"type"`
	SDP                 string `json:"sdp"`
	RTCConnectionID     string `json:"rtc_connection_id"`
	MediaSessionID      string `json:"media_session_id"`
	AudioCodec          string `json:"audio_codec,omitempty"`
	VideoCodec          string `json:"video_codec,omitempty"`
	DAVEProtocolVersion int    `json:"dave_protocol_version"`
	DAVEEpoch           uint64 `json:"dave_epoch,omitempty"`
}

type Heartbeat struct {
	T      int64 `json:"t"`
	SeqAck int   `json:"seq_ack,omitempty"`
}

type HeartbeatACK struct {
	T int64 `json:"t"`
}

type Speaking struct {
	UserID   string `json:"user_id,omitempty"`
	Speaking int    `json:"speaking"`
	Delay    int    `json:"delay,omitempty"`
	SSRC     uint32 `json:"ssrc,omitempty"`
}

type ClientsConnect struct {
	UserIDs []string `json:"user_ids"`
}

type ClientDisconnect struct {
	UserID string `json:"user_id"`
}

type PrepareTransition struct {
	ProtocolVersion int    `json:"protocol_version"`
	TransitionID    uint16 `json:"transition_id"`
}

type ExecuteTransition struct {
	TransitionID uint16 `json:"transition_id"`
}

type TransitionReady struct {
	TransitionID uint16 `json:"transition_id"`
}

type PrepareEpoch struct {
	ProtocolVersion int    `json:"protocol_version"`
	Epoch           uint64 `json:"epoch"`
}

type InvalidCommitWelcome struct {
	TransitionID uint16 `json:"transition_id"`
}
