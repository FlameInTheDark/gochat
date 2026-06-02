package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

func TestHeartbeatDeadlineToleratesBrowserTimerThrottling(t *testing.T) {
	t.Parallel()

	h := &Handler{hbTimeout: 35_000}
	if got, want := h.heartbeatDeadline(), 105*time.Second; got != want {
		t.Fatalf("heartbeatDeadline() = %s, want %s", got, want)
	}
}

func TestStaleHeartbeatStillGetsAcked(t *testing.T) {
	t.Parallel()

	var sent mqmsg.Message
	h := &Handler{
		user:        &dto.User{Id: 1},
		lastEventId: 10,
		hbTimeout:   35_000,
		hTimer:      time.NewTimer(time.Hour),
		sendJSON: func(v any) error {
			var ok bool
			sent, ok = v.(mqmsg.Message)
			if !ok {
				t.Fatalf("sent payload type = %T, want mqmsg.Message", v)
			}
			return nil
		},
	}
	defer h.hTimer.Stop()

	data, err := json.Marshal(heartbeatMessage{LastEventId: 5})
	if err != nil {
		t.Fatal(err)
	}
	h.HandleMessage(mqmsg.Message{Operation: mqmsg.OPCodeHeartBeat, Data: data})

	if h.lastEventId != 10 {
		t.Fatalf("lastEventId regressed to %d", h.lastEventId)
	}
	if sent.Operation != mqmsg.OPCodeHeartbeatAck {
		t.Fatalf("sent op = %d, want %d", sent.Operation, mqmsg.OPCodeHeartbeatAck)
	}
	var ack mqmsg.HeartbeatAck
	if err := json.Unmarshal(sent.Data, &ack); err != nil {
		t.Fatal(err)
	}
	if ack.LastEventID != 5 {
		t.Fatalf("ack e = %d, want 5", ack.LastEventID)
	}
}
