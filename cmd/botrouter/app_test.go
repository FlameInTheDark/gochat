package main

import (
	"encoding/json"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
)

func TestRequiredPermissionsForMessageEvents(t *testing.T) {
	tp := mqmsg.EventTypeMessageCreate
	raw, err := json.Marshal(mqmsg.Message{Operation: mqmsg.OpCodeDispatch, EventType: &tp})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := requiredPermissionsForEvent(raw, true)
	if len(got) != 2 || got[0] != permissions.PermServerViewChannels || got[1] != permissions.PermTextReadMessageHistory {
		t.Fatalf("unexpected permissions: %#v", got)
	}
}

func TestAllowedDMEvent(t *testing.T) {
	tp := mqmsg.EventTypeUserDMMessage
	raw, err := json.Marshal(mqmsg.Message{Operation: mqmsg.OpCodeDispatch, EventType: &tp})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !allowedDMEvent(raw) {
		t.Fatalf("expected dm event to be allowed")
	}
}
