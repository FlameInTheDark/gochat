package mqmsg

import (
	"encoding/json"

	streammeta "github.com/FlameInTheDark/gochat/internal/stream"
)

type DMCallSummary struct {
	CallID       int64           `json:"call_id"`
	ChannelID    int64           `json:"channel_id"`
	CallerID     int64           `json:"caller_id"`
	RecipientID  int64           `json:"recipient_id"`
	Region       string          `json:"region,omitempty"`
	Participants map[int64]int64 `json:"participants,omitempty"`
	StartedAt    int64           `json:"started_at"`
	SoloSince    int64           `json:"solo_since,omitempty"`
	Dismissed    bool            `json:"dismissed,omitempty"`
}

type dmCallEvent struct {
	Type   EventType
	Call   DMCallSummary `json:"call"`
	UserID int64         `json:"user_id,omitempty"`
	Reason string        `json:"reason,omitempty"`
}

func (m *dmCallEvent) EventType() *EventType    { return &m.Type }
func (m *dmCallEvent) Operation() OPCodeType    { return OpCodeDispatch }
func (m *dmCallEvent) Marshal() ([]byte, error) { return json.Marshal(m) }

func NewDMCallStarted(call DMCallSummary) EventDataMessage {
	return &dmCallEvent{Type: EventTypeUserDMCallStarted, Call: call}
}

func NewDMCallJoined(call DMCallSummary, userID int64) EventDataMessage {
	return &dmCallEvent{Type: EventTypeUserDMCallJoined, Call: call, UserID: userID}
}

func NewDMCallDeclined(call DMCallSummary, userID int64) EventDataMessage {
	return &dmCallEvent{Type: EventTypeUserDMCallDeclined, Call: call, UserID: userID}
}

func NewDMCallLeft(call DMCallSummary, userID int64) EventDataMessage {
	return &dmCallEvent{Type: EventTypeUserDMCallLeft, Call: call, UserID: userID}
}

func NewDMCallEnded(call DMCallSummary, reason string) EventDataMessage {
	return &dmCallEvent{Type: EventTypeUserDMCallEnded, Call: call, Reason: reason}
}

type DMCallStreamEvent struct {
	Call     DMCallSummary           `json:"call"`
	UserID   int64                   `json:"user_id"`
	Stream   streammeta.ActiveStream `json:"stream,omitempty"`
	StreamID int64                   `json:"stream_id,omitempty"`
	Reason   string                  `json:"reason,omitempty"`
	started  bool
}

func (m *DMCallStreamEvent) EventType() *EventType {
	if m.started {
		e := EventTypeUserDMCallStreamStarted
		return &e
	}
	e := EventTypeUserDMCallStreamStopped
	return &e
}

func (m *DMCallStreamEvent) Operation() OPCodeType    { return OpCodeDispatch }
func (m *DMCallStreamEvent) Marshal() ([]byte, error) { return json.Marshal(m) }

func NewDMCallStreamStarted(call DMCallSummary, userID int64, stream streammeta.ActiveStream) EventDataMessage {
	return &DMCallStreamEvent{Call: call, UserID: userID, Stream: stream, started: true}
}

func NewDMCallStreamStopped(call DMCallSummary, userID, streamID int64, reason string) EventDataMessage {
	return &DMCallStreamEvent{Call: call, UserID: userID, StreamID: streamID, Reason: reason}
}
