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
)

type Message struct {
	Operation OPCodeType      `json:"op"`
	Data      json.RawMessage `json:"d"`
	EventType *EventType      `json:"t"`
}

func BuildEventMessage(data EventDataMessage) (msg Message, err error) {
	msg.EventType = data.EventType()
	msg.Operation = data.Operation()
	msg.Data, err = data.Marshal()
	return
}
