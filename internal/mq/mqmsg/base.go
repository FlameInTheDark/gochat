package mqmsg

import "encoding/json"

type EventDataMessage interface {
	EventType() *EventType
	Operation() OPCodeType
	Marshal() ([]byte, error)
}

type OPCodeType int

const (
	OpCodeDispatch OPCodeType = iota
	OPCodeHello
	OPCodeHeartBeat
	OPCodePresenceUpdate
	OPCodeGuildUpdateSubscription
	OPCodeChannelSubscription
	OPCodePresenceSubscription
	// WebRTC signaling over existing WS (single opcode)
	OPCodeRTC
)

type EventType int

const (
	EventTypeMessageCreate EventType = 100 + iota
	EventTypeMessageUpdate
	EventTypeMessageDelete
	EventTypeGuildCreate
	EventTypeGuildUpdate
	EventTypeGuildDelete
	EventTypeChannelCreate
	EventTypeChannelUpdate
	EventTypeChannelOrderUpdate
	EventTypeChannelDelete
	EventTypeGuildRoleCreate
	EventTypeGuildRoleUpdate
	EventTypeGuildRoleDelete
	EventTypeThreadCreate
	EventTypeThreadUpdate
	EventTypeThreadDelete
)

const (
	EventTypeGuildMemberAdd EventType = 200 + iota
	EventTypeGuildMemberUpdate
	EventTypeGuildMemberRemove
	EventTypeGuildMemberAddRole
	EventTypeGuildMemberRemoveRole
)

const (
	EventTypeGuildChannelMessage EventType = 300 + iota
)

const (
	EventTypeUserUpdateReadState EventType = 400 + iota
	EventTypeUserUpdateSettings
	EventTypeUserFriendRequest
	EventTypeUserFriendAdded
	EventTypeUserFriendRemoved
	EventTypeUserDMMessage
	EventTypeUserUpdate
)

// RTC signaling event types (client <-> SFU via WS)
const (
	EventTypeRTCJoin EventType = 500 + iota
	EventTypeRTCOffer
	EventTypeRTCAnswer
	EventTypeRTCCandidate
	EventTypeRTCLeave
)

// Extended RTC signaling/control events (client <-> SFU)
const (
	// Client mutes/unmutes self (local microphone)
	EventTypeRTCMuteSelf EventType = 505 + iota
	// Client mutes/unmutes specific user locally (does not affect others)
	EventTypeRTCMuteUser
	// Privileged: server-side mute a user for everyone
	EventTypeRTCServerMuteUser
	// Privileged: server-side deafen a user (they receive no one)
	EventTypeRTCServerDeafenUser
	// Keep binding alive while users are in the call; clients send periodically
	EventTypeRTCBindingAlive
	// Privileged: kick a user from the room (server instructs client to leave)
	EventTypeRTCServerKickUser
	// Privileged: block/unblock a user from joining this room
	EventTypeRTCServerBlockUser
	// Server -> client notification: you were moved to another channel (close WS and reconnect)
	EventTypeRTCMoved
	// Server -> clients in a voice channel: SFU route changed; reconnect by rejoining
	EventTypeRTCServerRebind
)

type Message struct {
	Operation OPCodeType      `json:"op"`
	Data      json.RawMessage `json:"d"`
	EventType *EventType      `json:"t,omitempty"`
}

func BuildEventMessage(data EventDataMessage) (msg Message, err error) {
	msg.EventType = data.EventType()
	msg.Operation = data.Operation()
	msg.Data, err = data.Marshal()
	return
}

// UnmarshalJSON supports clients that send "data" instead of "d" for the payload field.
func (m *Message) UnmarshalJSON(b []byte) error {
	var aux struct {
		Operation OPCodeType      `json:"op"`
		D         json.RawMessage `json:"d"`
		Data      json.RawMessage `json:"data"`
		T         *EventType      `json:"t"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	m.Operation = aux.Operation
	m.EventType = aux.T
	if len(aux.D) > 0 {
		m.Data = aux.D
	} else {
		m.Data = aux.Data
	}
	return nil
}
