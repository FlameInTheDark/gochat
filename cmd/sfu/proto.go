package main

import (
	"encoding/json"
)

// Incoming envelopes and payloads

// rtcJoinEnvelope represents the expected first message from client.
// It must be an RTC join envelope: {op: OPCodeRTC, t: EventTypeRTCJoin, d: {channel, token}}
type rtcJoinEnvelope struct {
	OP int `json:"op"`
	T  int `json:"t"`
	D  struct {
		Channel int64  `json:"channel"`
		Token   string `json:"token"`
	} `json:"d"`
}

// envelope is a generic wrapper for subsequent messages
type envelope struct {
	OP int             `json:"op"`
	T  int             `json:"t"`
	D  json.RawMessage `json:"d"`
}

type rtcOffer struct {
	SDP string `json:"sdp"`
}

type rtcAnswer struct {
	SDP string `json:"sdp"`
}
type rtcCandidate struct {
	Candidate     string  `json:"candidate"`
	SDPMid        *string `json:"sdpMid,omitempty"`
	SDPMLineIndex *uint16 `json:"sdpMLineIndex,omitempty"`
}
type rtcMuteSelf struct {
	Muted bool `json:"muted"`
}

type rtcMuteUser struct {
	User  int64 `json:"user"`
	Muted bool  `json:"muted"`
}

type rtcServerDeafenUser struct {
	User     int64 `json:"user"`
	Deafened bool  `json:"deafened"`
}

type heartbeatData struct {
	Nonce any `json:"nonce,omitempty"`
	TS    any `json:"ts,omitempty"`
}

// Admin control payloads
type rtcKickUser struct {
	User int64 `json:"user"`
}

type rtcBlockUser struct {
	User  int64 `json:"user"`
	Block bool  `json:"block"`
}

// Server notification for move
type rtcMoved struct {
	Channel int64 `json:"channel"`
}

// Outgoing envelopes and payloads

// OutEnvelope is a generic envelope for outgoing messages.
// "t" is omitted when zero (e.g., heartbeat messages).
type OutEnvelope struct {
	OP int `json:"op"`
	T  int `json:"t,omitempty"`
	D  any `json:"d"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
type HeartbeatReply struct {
	Pong     bool  `json:"pong"`
	ServerTS int64 `json:"server_ts"`
	Nonce    any   `json:"nonce,omitempty"`
	TS       any   `json:"ts,omitempty"`
}
type JoinAck struct {
	Ok bool `json:"ok"`
}
