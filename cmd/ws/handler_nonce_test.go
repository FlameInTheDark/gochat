package main

import (
	"encoding/json"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/dto"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
)

func mustNonce(t *testing.T, raw string) *helper.MessageNonce {
	t.Helper()
	var nonce helper.MessageNonce
	if err := json.Unmarshal([]byte(raw), &nonce); err != nil {
		t.Fatalf("failed to unmarshal nonce: %v", err)
	}
	return &nonce
}

func decodeCreateMessage(t *testing.T, payload []byte) mqmsg.CreateMessage {
	t.Helper()
	var envelope mqmsg.Message
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v", err)
	}

	var message mqmsg.CreateMessage
	if err := json.Unmarshal(envelope.Data, &message); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	return message
}

func TestPersonalizeMessageForRecipientKeepsNonceForAuthor(t *testing.T) {
	event, err := mqmsg.BuildEventMessage(&mqmsg.CreateMessage{
		Message: dto.Message{
			Id:      1,
			Author:  dto.User{Id: 42},
			Content: "hello",
			Nonce:   mustNonce(t, `"draft-1"`),
		},
	})
	if err != nil {
		t.Fatalf("BuildEventMessage returned error: %v", err)
	}

	wire, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	personalized := personalizeMessageForRecipient("channel.99", 42, wire)
	got := decodeCreateMessage(t, personalized)
	if got.Message.Nonce == nil || string(*got.Message.Nonce) != `"draft-1"` {
		t.Fatalf("expected author to receive nonce, got %#v", got.Message.Nonce)
	}
}

func TestPersonalizeMessageForRecipientStripsNonceForOtherUsers(t *testing.T) {
	event, err := mqmsg.BuildEventMessage(&mqmsg.CreateMessage{
		Message: dto.Message{
			Id:      1,
			Author:  dto.User{Id: 42},
			Content: "hello",
			Nonce:   mustNonce(t, `"draft-1"`),
		},
	})
	if err != nil {
		t.Fatalf("BuildEventMessage returned error: %v", err)
	}

	wire, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	personalized := personalizeMessageForRecipient("channel.99", 7, wire)
	got := decodeCreateMessage(t, personalized)
	if got.Message.Nonce != nil {
		t.Fatalf("expected nonce to be stripped, got %#v", got.Message.Nonce)
	}
}

func TestPersonalizeMessageForRecipientLeavesUserTopicsUntouched(t *testing.T) {
	event, err := mqmsg.BuildEventMessage(&mqmsg.CreateMessage{
		Message: dto.Message{
			Id:      1,
			Author:  dto.User{Id: 42},
			Content: "hello",
			Nonce:   mustNonce(t, `"draft-1"`),
		},
	})
	if err != nil {
		t.Fatalf("BuildEventMessage returned error: %v", err)
	}

	wire, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	personalized := personalizeMessageForRecipient("user.42", 7, wire)
	got := decodeCreateMessage(t, personalized)
	if got.Message.Nonce == nil || string(*got.Message.Nonce) != `"draft-1"` {
		t.Fatalf("expected user-topic payload to stay untouched, got %#v", got.Message.Nonce)
	}
}
