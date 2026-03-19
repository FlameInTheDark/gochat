package message

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/FlameInTheDark/gochat/internal/embed"
	"github.com/FlameInTheDark/gochat/internal/helper"
)

func TestSendMessageRequestValidateAllowsEmbedsOnly(t *testing.T) {
	req := SendMessageRequest{
		Embeds: []embed.Embed{{
			Description: "hello from embed",
		}},
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestCreateThreadRequestValidateRequiresStarterPayload(t *testing.T) {
	req := CreateThreadRequest{}

	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreateThreadRequestValidateAllowsExplicitName(t *testing.T) {
	req := CreateThreadRequest{
		Name:    "thread title",
		Content: "first message",
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestSendMessageRequestValidateRequiresNonceWhenEnforced(t *testing.T) {
	req := SendMessageRequest{
		Content:      "hello",
		EnforceNonce: true,
	}

	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSendMessageRequestValidateAllowsReplyReference(t *testing.T) {
	reference := helper.StringInt64(42)
	req := SendMessageRequest{
		Content:   "reply",
		Reference: &reference,
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestSendMessageRequestValidateRejectsInvalidReplyReference(t *testing.T) {
	reference := helper.StringInt64(0)
	req := SendMessageRequest{
		Content:   "reply",
		Reference: &reference,
	}

	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCreateThreadRequestMessageRequestPreservesNonce(t *testing.T) {
	var nonce helper.MessageNonce
	if err := json.Unmarshal([]byte(`"draft-1"`), &nonce); err != nil {
		t.Fatalf("failed to unmarshal nonce: %v", err)
	}

	req := CreateThreadRequest{
		Content: "first message",
		Nonce:   &nonce,
	}

	messageReq := req.MessageRequest([]int64{1, 2})
	if messageReq.Nonce == nil || string(*messageReq.Nonce) != `"draft-1"` {
		t.Fatalf("expected nonce to be copied, got %#v", messageReq.Nonce)
	}

	(*messageReq.Nonce)[0] = 'x'
	if string(nonce) != `"draft-1"` {
		t.Fatalf("expected source nonce to stay unchanged, got %q", string(nonce))
	}
}

func TestCreateThreadRequestValidateRejectsLongName(t *testing.T) {
	req := CreateThreadRequest{
		Name:    strings.Repeat("a", maxThreadNameLength+1),
		Content: "first message",
	}

	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestUpdateMessageRequestValidateRequiresPayload(t *testing.T) {
	req := UpdateMessageRequest{}

	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSendMessageRequestValidateRejectsEmbedLimit(t *testing.T) {
	req := SendMessageRequest{
		Embeds: make([]embed.Embed, embed.MaxEmbedsPerMessage+1),
	}
	for i := range req.Embeds {
		req.Embeds[i] = embed.Embed{Description: "hello"}
	}

	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestUpdateMessageRequestValidateAllowsFlagsOnly(t *testing.T) {
	flags := 0
	req := UpdateMessageRequest{Flags: &flags}

	if err := req.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestUpdateMessageRequestValidateRejectsNegativeFlags(t *testing.T) {
	flags := -1
	req := UpdateMessageRequest{Flags: &flags}

	if err := req.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}
