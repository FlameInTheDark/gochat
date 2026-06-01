package goclient

import (
	"encoding/json"
	"testing"
)

func TestDispatchTypedEventsAndRawEvent(t *testing.T) {
	s, err := New("Bot gcb_test")
	if err != nil {
		t.Fatal(err)
	}
	var rawCount int
	var messageID int64
	var readySession string
	s.AddHandler(func(_ *Session, event *Event) {
		rawCount++
		if event.Struct == nil {
			t.Fatal("expected decoded payload on raw event")
		}
	})
	s.AddHandler(func(_ *Session, event *Ready) {
		readySession = event.SessionID
	})
	s.AddHandler(func(_ *Session, event *MessageCreate) {
		messageID = event.Message.ID
	})

	readyType := EventTypeGatewayReady
	if err := s.dispatch(GatewayMessage{
		Operation: OpCodeDispatch,
		EventType: &readyType,
		Data:      mustJSON(Ready{SessionID: "session-1", ShardID: 0, ShardCount: 1}),
	}); err != nil {
		t.Fatal(err)
	}
	messageType := EventTypeMessageCreate
	if err := s.dispatch(GatewayMessage{
		Operation: OpCodeDispatch,
		EventType: &messageType,
		Data:      mustJSON(MessageCreate{Message: Message{ID: 42, ChannelID: 7, Content: "hello"}}),
	}); err != nil {
		t.Fatal(err)
	}

	if rawCount != 2 {
		t.Fatalf("raw events = %d, want 2", rawCount)
	}
	if readySession != "session-1" || s.SessionID() != "session-1" {
		t.Fatalf("ready session = %q, stored = %q", readySession, s.SessionID())
	}
	if messageID != 42 {
		t.Fatalf("message id = %d, want 42", messageID)
	}
}

func TestAddHandlerOnceRemovesItself(t *testing.T) {
	s, err := New("gcb_test")
	if err != nil {
		t.Fatal(err)
	}
	var calls int
	s.AddHandlerOnce(func(_ *Session, _ *MessageDelete) {
		calls++
	})
	eventType := EventTypeMessageDelete
	msg := GatewayMessage{
		Operation: OpCodeDispatch,
		EventType: &eventType,
		Data:      mustJSON(MessageDelete{ChannelID: 1, MessageID: 2}),
	}
	if err := s.dispatch(msg); err != nil {
		t.Fatal(err)
	}
	if err := s.dispatch(msg); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestDispatchCurrentProtocolEdgeEvents(t *testing.T) {
	s, err := New("gcb_test")
	if err != nil {
		t.Fatal(err)
	}
	var settings string
	var callUserID int64
	var rtcChannel int64
	var rtcRawOP OPCodeType
	s.AddHandler(func(_ *Session, event *UpdateUserSettings) {
		settings = string(event.Settings)
	})
	s.AddHandler(func(_ *Session, event *DMCallJoined) {
		callUserID = event.UserID
	})
	s.AddHandler(func(_ *Session, event *Event) {
		if event.Operation == OPCodeRTC {
			rtcRawOP = event.Operation
		}
	})
	s.AddHandler(func(_ *Session, event *VoiceRebind) {
		rtcChannel = event.Channel
	})

	settingsType := EventTypeUserUpdateSettings
	if err := s.dispatch(GatewayMessage{
		Operation: OpCodeDispatch,
		EventType: &settingsType,
		Data:      []byte(`{"settings":{"language":"en"}}`),
	}); err != nil {
		t.Fatal(err)
	}
	callType := EventTypeUserDMCallJoined
	if err := s.dispatch(GatewayMessage{
		Operation: OpCodeDispatch,
		EventType: &callType,
		Data:      mustJSON(DMCallJoined{UserID: 88}),
	}); err != nil {
		t.Fatal(err)
	}
	rtcType := EventTypeRTCServerRebind
	rawRTC := mustJSON(GatewayMessage{
		Operation: OPCodeRTC,
		EventType: &rtcType,
		Data:      mustJSON(VoiceRebind{Channel: 99}),
	})
	if err := s.onGatewayMessage(nil, nil, rawRTC); err != nil {
		t.Fatal(err)
	}

	if settings != `{"language":"en"}` {
		t.Fatalf("settings = %s", settings)
	}
	if callUserID != 88 {
		t.Fatalf("call user id = %d, want 88", callUserID)
	}
	if rtcRawOP != OPCodeRTC || rtcChannel != 99 {
		t.Fatalf("rtc op = %d, channel = %d", rtcRawOP, rtcChannel)
	}
}

func TestGatewayMessageAcceptsDataAlias(t *testing.T) {
	var msg GatewayMessage
	if err := json.Unmarshal([]byte(`{"op":0,"t":100,"data":{"message":{"id":55}}}`), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.EventType == nil || *msg.EventType != EventTypeMessageCreate {
		t.Fatalf("event type = %#v", msg.EventType)
	}
	if string(msg.Data) != `{"message":{"id":55}}` {
		t.Fatalf("data = %s", msg.Data)
	}
}
