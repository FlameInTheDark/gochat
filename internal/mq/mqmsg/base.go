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
