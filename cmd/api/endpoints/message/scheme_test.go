package message

import (
	"testing"

	"github.com/FlameInTheDark/gochat/internal/embed"
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
