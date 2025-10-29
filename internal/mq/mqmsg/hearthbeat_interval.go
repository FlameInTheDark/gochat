package mqmsg

import (
	"encoding/json"
)

type HeartbeatInterval struct {
	HeartbeatInterval int64  `json:"heartbeat_interval"`
	SessionID         string `json:"session_id,omitempty"`
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
