package main

import (
	"encoding/json"
	"time"
)

const joinHandshakeTimeout = 5 * time.Second

type rtcJoinEnvelope struct {
	OP int `json:"op"`
	T  int `json:"t"`
	D  struct {
		Channel int64  `json:"channel"`
		Token   string `json:"token"`
	} `json:"d"`
}

type envelope struct {
	OP int             `json:"op"`
	T  int             `json:"t"`
	D  json.RawMessage `json:"d"`
}

type rtcAnswer struct {
	SDP  string `json:"sdp"`
	Type string `json:"type,omitempty"`
}

type rtcOffer struct {
	SDP  string `json:"sdp"`
	Type string `json:"type,omitempty"`
}

type rtcCandidate struct {
	Candidate     string  `json:"candidate"`
	SDPMid        *string `json:"sdpMid,omitempty"`
	SDPMLineIndex *uint16 `json:"sdpMLineIndex,omitempty"`
}

type heartbeatData struct {
	Nonce any `json:"nonce,omitempty"`
	TS    any `json:"ts,omitempty"`
}

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
