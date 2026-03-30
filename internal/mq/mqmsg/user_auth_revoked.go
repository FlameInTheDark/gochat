package mqmsg

import "encoding/json"

type UserAuthRevoked struct {
	SessionVersion int64 `json:"session_version"`
}

func (m *UserAuthRevoked) EventType() *EventType {
	e := EventTypeUserAuthRevoked
	return &e
}

func (m *UserAuthRevoked) Operation() OPCodeType {
	return OpCodeDispatch
}

func (m *UserAuthRevoked) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
