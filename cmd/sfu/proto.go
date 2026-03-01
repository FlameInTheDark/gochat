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

type UserJoinNotify struct {
	UserId    int64  `json:"user_id"`
	ChannelId int64  `json:"channel_id"`
	GuildId   *int64 `json:"guild_id"`
}

type UserLeaveNotify struct {
	UserId    int64  `json:"user_id"`
	ChannelId int64  `json:"channel_id"`
	GuildId   *int64 `json:"guild_id"`
}

type ChannelAliveNotify struct {
	GuildId   *int64 `json:"guild_id"`
	ChannelId int64  `json:"channel_id"`
}

// speakingEvent is sent to clients in the same channel to indicate
// that a user started or stopped speaking. Speaking is 1 (active) or 0 (inactive).
type speakingEvent struct {
	UserId   int64 `json:"user_id"`
	Speaking int   `json:"speaking"`
}

// muteEvent is broadcast when a user is server-muted or unmuted.
type muteEvent struct {
	UserId int64 `json:"user_id"`
	Muted  bool  `json:"muted"`
}

// deafenEvent is broadcast when a user is server-deafened or undeafened.
type deafenEvent struct {
	UserId   int64 `json:"user_id"`
	Deafened bool  `json:"deafened"`
}

// kickEvent is sent to a user being kicked from the channel.
type kickEvent struct {
	UserId int64 `json:"user_id"`
}

// blockEvent payload for block/unblock requests.
type blockEvent struct {
	UserId int64 `json:"user_id"`
	Block  bool  `json:"block"`
}

// CloseChannelRequest is the body for the admin /admin/channel/close endpoint.
type CloseChannelRequest struct {
	ChannelID int64 `json:"channel_id"`
}

// muteSelfData payload for self-mute toggle.
type muteSelfData struct {
	Muted bool `json:"muted"`
}

// muteUserData payload for local/server mute of another user.
type muteUserData struct {
	User  int64 `json:"user"`
	Muted bool  `json:"muted"`
}

// deafenUserData payload for server deafen of a user.
type deafenUserData struct {
	User     int64 `json:"user"`
	Deafened bool  `json:"deafened"`
}

// kickUserData payload for kicking a user.
type kickUserData struct {
	User int64 `json:"user"`
}
