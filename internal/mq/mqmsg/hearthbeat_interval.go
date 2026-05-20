package mqmsg

import (
	"encoding/json"
)

type HeartbeatInterval struct {
	HeartbeatInterval int64  `json:"heartbeat_interval"`
	SessionID         string `json:"session_id,omitempty"`
	ConnectionID      string `json:"connection_id,omitempty"`
	Generation        int64  `json:"generation,omitempty"`
	ProtocolVersion   int    `json:"protocol_version,omitempty"`
}

func (m *HeartbeatInterval) EventType() *EventType {
	return nil
}

func (m *HeartbeatInterval) Operation() OPCodeType {
	return OPCodeHello
}

func (m *HeartbeatInterval) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

type HeartbeatAck struct {
	LastEventID  int64  `json:"e"`
	ServerTime   int64  `json:"server_time"`
	ConnectionID string `json:"connection_id,omitempty"`
}

func (m *HeartbeatAck) EventType() *EventType {
	return nil
}

func (m *HeartbeatAck) Operation() OPCodeType {
	return OPCodeHeartbeatAck
}

func (m *HeartbeatAck) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
